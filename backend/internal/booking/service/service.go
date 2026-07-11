package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/backend/internal/booking"
	bookingrepo "github.com/meetoria/meetoria/backend/internal/booking/repository"
	customerrepo "github.com/meetoria/meetoria/backend/internal/customer/repository"
	employeerepo "github.com/meetoria/meetoria/backend/internal/employee/repository"
	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	redisclient "github.com/meetoria/meetoria/backend/internal/common/redis"
	"github.com/meetoria/meetoria/backend/internal/common/rabbitmq"
	notifservice "github.com/meetoria/meetoria/backend/internal/notification/service"
	servicerepo "github.com/meetoria/meetoria/backend/internal/service/repository"
	schedulerepo "github.com/meetoria/meetoria/backend/internal/schedule/repository"
	"github.com/meetoria/meetoria/backend/pkg/events"
)

type Service interface {
	Create(ctx context.Context, orgID uuid.UUID, req booking.CreateBookingRequest, correlationID uuid.UUID) (*booking.Booking, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*booking.Booking, error)
	Update(ctx context.Context, orgID, id uuid.UUID, req booking.UpdateBookingRequest, correlationID uuid.UUID) (*booking.Booking, error)
	Cancel(ctx context.Context, orgID, id uuid.UUID, req booking.CancelBookingRequest, correlationID uuid.UUID) (*booking.Booking, error)
	List(ctx context.Context, orgID uuid.UUID, filters bookingrepo.ListFilters, params commonmodel.PaginationParams) (commonmodel.PaginatedResponse[booking.Booking], error)
	GetAvailability(ctx context.Context, orgID uuid.UUID, req booking.AvailabilityRequest) ([]booking.TimeSlot, error)
	SendSMS(ctx context.Context, orgID, id uuid.UUID, correlationID uuid.UUID) error
	SendEmail(ctx context.Context, orgID, id uuid.UUID, correlationID uuid.UUID) error
}

type bookingService struct {
	repo         bookingrepo.Repository
	customerRepo customerrepo.Repository
	employeeRepo employeerepo.Repository
	serviceRepo  servicerepo.Repository
	scheduleRepo schedulerepo.Repository
	redis        *redisclient.Client
	publisher    *rabbitmq.Publisher
	notifService notifservice.Service
}

func NewService(
	repo bookingrepo.Repository,
	customerRepo customerrepo.Repository,
	employeeRepo employeerepo.Repository,
	serviceRepo servicerepo.Repository,
	scheduleRepo schedulerepo.Repository,
	redis *redisclient.Client,
	publisher *rabbitmq.Publisher,
	notifService notifservice.Service,
) Service {
	return &bookingService{
		repo:         repo,
		customerRepo: customerRepo,
		employeeRepo: employeeRepo,
		serviceRepo:  serviceRepo,
		scheduleRepo: scheduleRepo,
		redis:        redis,
		publisher:    publisher,
		notifService: notifService,
	}
}

func (s *bookingService) Create(ctx context.Context, orgID uuid.UUID, req booking.CreateBookingRequest, correlationID uuid.UUID) (*booking.Booking, error) {
	lockKey := fmt.Sprintf("booking:lock:%s:%s", req.EmployeeID, req.StartTime.Format(time.RFC3339))
	acquired, err := s.redis.AcquireLock(ctx, lockKey, 30*time.Second)
	if err != nil || !acquired {
		return nil, apperrors.Conflict("unable to acquire booking lock, please retry")
	}
	defer s.redis.ReleaseLock(ctx, lockKey)

	svc, err := s.serviceRepo.GetByID(ctx, orgID, req.ServiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("service not found")
		}
		return nil, apperrors.Internal("failed to get service", err)
	}

	if _, err := s.customerRepo.GetByID(ctx, orgID, req.CustomerID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("customer not found")
		}
		return nil, apperrors.Internal("failed to get customer", err)
	}

	if _, err := s.employeeRepo.GetByID(ctx, orgID, req.EmployeeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("employee not found")
		}
		return nil, apperrors.Internal("failed to get employee", err)
	}

	endTime := req.StartTime.Add(time.Duration(svc.DurationMinutes) * time.Minute)

	if err := s.validateSchedule(ctx, orgID, req.EmployeeID, req.StartTime, endTime); err != nil {
		return nil, err
	}

	var created *booking.Booking
	err = s.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
		overlap, err := s.repo.HasOverlap(ctx, tx, orgID, req.EmployeeID, req.StartTime, endTime, nil)
		if err != nil {
			return apperrors.Internal("failed to check availability", err)
		}
		if overlap {
			return apperrors.ErrDoubleBooking
		}

		b := &booking.Booking{
			OrganizationScoped: commonmodel.OrganizationScoped{OrganizationID: orgID},
			CustomerID:         req.CustomerID,
			EmployeeID:         req.EmployeeID,
			ServiceID:          req.ServiceID,
			StartTime:          req.StartTime,
			EndTime:            endTime,
			Status:             booking.StatusConfirmed,
			Notes:              req.Notes,
			Price:              svc.Price,
			Currency:           svc.Currency,
		}

		if err := s.repo.Create(ctx, tx, b); err != nil {
			return apperrors.Internal("failed to create booking", err)
		}
		created = b
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.publishBookingEvent(ctx, events.EventBookingCreated, created, correlationID)
	s.notifService.QueueBookingConfirmation(ctx, orgID, created, correlationID)

	return created, nil
}

func (s *bookingService) GetByID(ctx context.Context, orgID, id uuid.UUID) (*booking.Booking, error) {
	b, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("booking not found")
		}
		return nil, apperrors.Internal("failed to get booking", err)
	}
	return b, nil
}

func (s *bookingService) Update(ctx context.Context, orgID, id uuid.UUID, req booking.UpdateBookingRequest, correlationID uuid.UUID) (*booking.Booking, error) {
	b, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	if b.Status == booking.StatusCancelled || b.Status == booking.StatusCompleted {
		return nil, apperrors.Conflict("cannot update booking in current status")
	}

	if req.StartTime != nil {
		duration := b.EndTime.Sub(b.StartTime)
		newEnd := req.StartTime.Add(duration)

		if err := s.validateSchedule(ctx, orgID, b.EmployeeID, *req.StartTime, newEnd); err != nil {
			return nil, err
		}

		err = s.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
			overlap, err := s.repo.HasOverlap(ctx, tx, orgID, b.EmployeeID, *req.StartTime, newEnd, &b.ID)
			if err != nil {
				return apperrors.Internal("failed to check availability", err)
			}
			if overlap {
				return apperrors.ErrDoubleBooking
			}
			b.StartTime = *req.StartTime
			b.EndTime = newEnd
			return s.repo.Update(ctx, b)
		})
		if err != nil {
			return nil, err
		}
	}

	if req.ServiceID != nil && *req.ServiceID != b.ServiceID {
		svc, err := s.serviceRepo.GetByID(ctx, orgID, *req.ServiceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.NotFound("service not found")
			}
			return nil, apperrors.Internal("failed to get service", err)
		}

		newEnd := b.StartTime.Add(time.Duration(svc.DurationMinutes) * time.Minute)
		if err := s.validateSchedule(ctx, orgID, b.EmployeeID, b.StartTime, newEnd); err != nil {
			return nil, err
		}

		err = s.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
			overlap, err := s.repo.HasOverlap(ctx, tx, orgID, b.EmployeeID, b.StartTime, newEnd, &b.ID)
			if err != nil {
				return apperrors.Internal("failed to check availability", err)
			}
			if overlap {
				return apperrors.ErrDoubleBooking
			}
			b.ServiceID = svc.ID
			b.EndTime = newEnd
			b.Price = svc.Price
			b.Currency = svc.Currency
			return s.repo.Update(ctx, b)
		})
		if err != nil {
			return nil, err
		}
	}

	if req.EmployeeID != nil && *req.EmployeeID != b.EmployeeID {
		if _, err := s.employeeRepo.GetByID(ctx, orgID, *req.EmployeeID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.NotFound("employee not found")
			}
			return nil, apperrors.Internal("failed to get employee", err)
		}

		newEmployeeID := *req.EmployeeID
		if err := s.validateSchedule(ctx, orgID, newEmployeeID, b.StartTime, b.EndTime); err != nil {
			return nil, err
		}

		err := s.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
			overlap, err := s.repo.HasOverlap(ctx, tx, orgID, newEmployeeID, b.StartTime, b.EndTime, &b.ID)
			if err != nil {
				return apperrors.Internal("failed to check availability", err)
			}
			if overlap {
				return apperrors.ErrDoubleBooking
			}
			b.EmployeeID = newEmployeeID
			return s.repo.Update(ctx, b)
		})
		if err != nil {
			return nil, err
		}
	}

	if req.CustomerID != nil && *req.CustomerID != b.CustomerID {
		if _, err := s.customerRepo.GetByID(ctx, orgID, *req.CustomerID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.NotFound("customer not found")
			}
			return nil, apperrors.Internal("failed to get customer", err)
		}
		b.CustomerID = *req.CustomerID
	}

	if req.Notes != nil {
		b.Notes = *req.Notes
	}
	if req.Status != nil {
		if *req.Status == booking.StatusCancelled {
			return nil, apperrors.Validation("cancellation reason is required")
		}
		b.Status = *req.Status
	}

	if err := s.repo.Update(ctx, b); err != nil {
		return nil, apperrors.Internal("failed to update booking", err)
	}

	s.publishBookingEvent(ctx, events.EventBookingUpdated, b, correlationID)
	return b, nil
}

func (s *bookingService) Cancel(ctx context.Context, orgID, id uuid.UUID, req booking.CancelBookingRequest, correlationID uuid.UUID) (*booking.Booking, error) {
	b, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	if b.Status == booking.StatusCancelled {
		return nil, apperrors.Conflict("booking is already cancelled")
	}

	b.Status = booking.StatusCancelled
	b.CancellationReason = req.Reason

	if err := s.repo.Update(ctx, b); err != nil {
		return nil, apperrors.Internal("failed to cancel booking", err)
	}

	s.publishBookingEvent(ctx, events.EventBookingCancelled, b, correlationID)
	return b, nil
}

func (s *bookingService) List(ctx context.Context, orgID uuid.UUID, filters bookingrepo.ListFilters, params commonmodel.PaginationParams) (commonmodel.PaginatedResponse[booking.Booking], error) {
	bookings, total, err := s.repo.List(ctx, orgID, filters, params.Offset(), params.Limit)
	if err != nil {
		return commonmodel.PaginatedResponse[booking.Booking]{}, apperrors.Internal("failed to list bookings", err)
	}
	return commonmodel.NewPaginatedResponse(bookings, total, params.Page, params.Limit), nil
}

func (s *bookingService) GetAvailability(ctx context.Context, orgID uuid.UUID, req booking.AvailabilityRequest) ([]booking.TimeSlot, error) {
	svc, err := s.serviceRepo.GetByID(ctx, orgID, req.ServiceID)
	if err != nil {
		return nil, apperrors.NotFound("service not found")
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, apperrors.Validation("invalid date format, use YYYY-MM-DD")
	}

	dayOfWeek := int(date.Weekday())
	workingHours, err := s.scheduleRepo.GetWorkingHours(ctx, orgID, req.EmployeeID, dayOfWeek)
	if err != nil || len(workingHours) == 0 {
		return []booking.TimeSlot{}, nil
	}

	existingBookings, err := s.repo.GetByEmployeeAndDate(ctx, orgID, req.EmployeeID, date)
	if err != nil {
		return nil, apperrors.Internal("failed to get existing bookings", err)
	}

	breaks, _ := s.scheduleRepo.GetBreaks(ctx, orgID, req.EmployeeID, dayOfWeek)
	duration := time.Duration(svc.DurationMinutes) * time.Minute
	slotInterval := 15 * time.Minute

	var slots []booking.TimeSlot
	for _, wh := range workingHours {
		current := combineDateAndTime(date, wh.StartTime)
		dayEnd := combineDateAndTime(date, wh.EndTime)

		for current.Add(duration).Before(dayEnd) || current.Add(duration).Equal(dayEnd) {
			slotEnd := current.Add(duration)
			available := true

			for _, b := range existingBookings {
				if current.Before(b.EndTime) && slotEnd.After(b.StartTime) {
					available = false
					break
				}
			}

			for _, br := range breaks {
				brStart := combineDateAndTime(date, br.StartTime)
				brEnd := combineDateAndTime(date, br.EndTime)
				if current.Before(brEnd) && slotEnd.After(brStart) {
					available = false
					break
				}
			}

			slots = append(slots, booking.TimeSlot{
				StartTime: current,
				EndTime:   slotEnd,
				Available: available,
			})

			current = current.Add(slotInterval)
		}
	}

	return slots, nil
}

func (s *bookingService) validateSchedule(ctx context.Context, orgID, employeeID uuid.UUID, start, end time.Time) error {
	dayOfWeek := int(start.Weekday())
	workingHours, err := s.scheduleRepo.GetWorkingHours(ctx, orgID, employeeID, dayOfWeek)
	if err != nil {
		return apperrors.Internal("failed to check working hours", err)
	}
	if len(workingHours) == 0 {
		return nil
	}

	startTime := time.Date(0, 1, 1, start.Hour(), start.Minute(), 0, 0, time.UTC)
	endTime := time.Date(0, 1, 1, end.Hour(), end.Minute(), 0, 0, time.UTC)

	withinHours := false
	for _, wh := range workingHours {
		if !startTime.Before(wh.StartTime) && !endTime.After(wh.EndTime) {
			withinHours = true
			break
		}
	}
	if !withinHours {
		return apperrors.ErrOutsideWorkingHours
	}

	isHoliday, err := s.scheduleRepo.IsHoliday(ctx, orgID, employeeID, start)
	if err != nil {
		return apperrors.Internal("failed to check holidays", err)
	}
	if isHoliday {
		return apperrors.ErrOutsideWorkingHours
	}

	return nil
}

func (s *bookingService) SendSMS(ctx context.Context, orgID, id uuid.UUID, correlationID uuid.UUID) error {
	b, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}
	return s.notifService.QueueBookingConfirmationSMS(ctx, orgID, b, correlationID)
}

func (s *bookingService) SendEmail(ctx context.Context, orgID, id uuid.UUID, correlationID uuid.UUID) error {
	b, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}
	return s.notifService.QueueBookingConfirmationEmail(ctx, orgID, b, correlationID)
}

func (s *bookingService) publishBookingEvent(ctx context.Context, eventType string, b *booking.Booking, correlationID uuid.UUID) {
	event := events.BookingEvent{
		BaseEvent:      events.NewBaseEvent(eventType, "meetoria-api", correlationID),
		OrganizationID: b.OrganizationID,
		BookingID:      b.ID,
		CustomerID:     b.CustomerID,
		EmployeeID:     b.EmployeeID,
		ServiceID:      b.ServiceID,
		StartTime:      b.StartTime,
		EndTime:        b.EndTime,
		Status:         string(b.Status),
	}

	body, err := events.MarshalEvent(event)
	if err != nil {
		return
	}
	_ = s.publisher.Publish(ctx, eventType, body)
}

func combineDateAndTime(date time.Time, t time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location())
}
