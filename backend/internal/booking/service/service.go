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
	branchservice "github.com/meetoria/meetoria/backend/internal/branch/service"
	customerrepo "github.com/meetoria/meetoria/backend/internal/customer/repository"
	"github.com/meetoria/meetoria/backend/internal/customer"
	employeerepo "github.com/meetoria/meetoria/backend/internal/employee/repository"
	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	redisclient "github.com/meetoria/meetoria/backend/internal/common/redis"
	"github.com/meetoria/meetoria/backend/internal/common/rabbitmq"
	notifservice "github.com/meetoria/meetoria/backend/internal/notification/service"
	"github.com/meetoria/meetoria/backend/internal/organization"
	orgrepo "github.com/meetoria/meetoria/backend/internal/organization/repository"
	"github.com/meetoria/meetoria/backend/internal/schedule"
	"github.com/meetoria/meetoria/backend/pkg/phone"
	servicerepo "github.com/meetoria/meetoria/backend/internal/service/repository"
	schedulerepo "github.com/meetoria/meetoria/backend/internal/schedule/repository"
	"github.com/meetoria/meetoria/backend/pkg/events"
)

type Service interface {
	Create(ctx context.Context, orgID uuid.UUID, req booking.CreateBookingRequest, correlationID uuid.UUID) (*booking.Booking, error)
	CreatePublic(ctx context.Context, org *organization.Organization, req booking.PublicCreateBookingRequest, correlationID uuid.UUID) (*booking.Booking, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*booking.Booking, error)
	Update(ctx context.Context, orgID, id uuid.UUID, req booking.UpdateBookingRequest, correlationID uuid.UUID) (*booking.Booking, error)
	Cancel(ctx context.Context, orgID, id uuid.UUID, req booking.CancelBookingRequest, correlationID uuid.UUID) (*booking.Booking, error)
	List(ctx context.Context, orgID uuid.UUID, filters bookingrepo.ListFilters, params commonmodel.PaginationParams) (commonmodel.PaginatedResponse[booking.Booking], error)
	GetAvailability(ctx context.Context, orgID uuid.UUID, req booking.AvailabilityRequest) ([]booking.TimeSlot, error)
	GetPublicAvailability(ctx context.Context, orgID uuid.UUID, req booking.PublicAvailabilityRequest, minNoticeMinutes int) ([]booking.PublicTimeSlot, error)
	SendSMS(ctx context.Context, orgID, id uuid.UUID, correlationID uuid.UUID) error
	SendEmail(ctx context.Context, orgID, id uuid.UUID, correlationID uuid.UUID) error
}

type bookingService struct {
	repo          bookingrepo.Repository
	customerRepo  customerrepo.Repository
	employeeRepo  employeerepo.Repository
	serviceRepo   servicerepo.Repository
	scheduleRepo  schedulerepo.Repository
	orgRepo       orgrepo.Repository
	branchService branchservice.Service
	redis         *redisclient.Client
	publisher     *rabbitmq.Publisher
	notifService  notifservice.Service
}

func NewService(
	repo bookingrepo.Repository,
	customerRepo customerrepo.Repository,
	employeeRepo employeerepo.Repository,
	serviceRepo servicerepo.Repository,
	scheduleRepo schedulerepo.Repository,
	orgRepo orgrepo.Repository,
	branchService branchservice.Service,
	redis *redisclient.Client,
	publisher *rabbitmq.Publisher,
	notifService notifservice.Service,
) Service {
	return &bookingService{
		repo:          repo,
		customerRepo:  customerRepo,
		employeeRepo:  employeeRepo,
		serviceRepo:   serviceRepo,
		scheduleRepo:  scheduleRepo,
		orgRepo:       orgRepo,
		branchService: branchService,
		redis:         redis,
		publisher:     publisher,
		notifService:  notifService,
	}
}

func (s *bookingService) Create(ctx context.Context, orgID uuid.UUID, req booking.CreateBookingRequest, correlationID uuid.UUID) (*booking.Booking, error) {
	branchID, err := s.branchService.ResolveBranchID(ctx, orgID, req.BranchID)
	if err != nil {
		return nil, err
	}
	return s.createBooking(ctx, orgID, branchID, req.CustomerID, req.EmployeeID, req.ServiceID, req.StartTime, req.Notes, booking.StatusConfirmed, nil, correlationID)
}

func (s *bookingService) CreatePublic(ctx context.Context, org *organization.Organization, req booking.PublicCreateBookingRequest, correlationID uuid.UUID) (*booking.Booking, error) {
	settings := organization.ParseSettings(org.Settings)
	if !settings.Booking.Enabled {
		return nil, apperrors.Forbidden("public booking is not enabled for this organization")
	}
	if settings.Booking.EmailRequired && req.Customer.Email == "" {
		return nil, apperrors.Validation("email is required")
	}

	normalizedPhone, err := phone.NormalizeE164(req.Customer.Phone)
	if err != nil {
		return nil, err
	}
	req.Customer.Phone = normalizedPhone

	svc, err := s.serviceRepo.GetByID(ctx, org.ID, req.ServiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("service not found")
		}
		return nil, apperrors.Internal("failed to get service", err)
	}
	if !svc.IsActive {
		return nil, apperrors.NotFound("service not found")
	}

	if _, err := s.branchService.GetByID(ctx, org.ID, req.BranchID); err != nil {
		return nil, err
	}

	endTime := req.StartTime.Add(time.Duration(svc.DurationMinutes) * time.Minute)
	if err := s.validateBookingWindow(req.StartTime, endTime, settings.Booking); err != nil {
		return nil, err
	}

	employeeID, err := s.resolveEmployeeForPublicBooking(ctx, org.ID, req, svc.DurationMinutes)
	if err != nil {
		return nil, err
	}

	cust, err := s.findOrCreatePublicCustomer(ctx, org.ID, req.Customer)
	if err != nil {
		return nil, err
	}

	status := booking.BookingStatus(settings.Booking.InitialBookingStatus())
	return s.createBooking(ctx, org.ID, req.BranchID, cust.ID, employeeID, req.ServiceID, req.StartTime, req.Notes, status, &settings.Booking, correlationID)
}

func (s *bookingService) createBooking(
	ctx context.Context,
	orgID uuid.UUID,
	branchID uuid.UUID,
	customerID, employeeID, serviceID uuid.UUID,
	startTime time.Time,
	notes string,
	status booking.BookingStatus,
	bookingSettings *organization.BookingSettings,
	correlationID uuid.UUID,
) (*booking.Booking, error) {
	lockKey := fmt.Sprintf("booking:lock:%s:%s", employeeID, startTime.Format(time.RFC3339))
	acquired, err := s.redis.AcquireLock(ctx, lockKey, 30*time.Second)
	if err != nil || !acquired {
		return nil, apperrors.Conflict("unable to acquire booking lock, please retry")
	}
	defer s.redis.ReleaseLock(ctx, lockKey)

	svc, err := s.serviceRepo.GetByID(ctx, orgID, serviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("service not found")
		}
		return nil, apperrors.Internal("failed to get service", err)
	}

	if _, err := s.customerRepo.GetByID(ctx, orgID, customerID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("customer not found")
		}
		return nil, apperrors.Internal("failed to get customer", err)
	}

	emp, err := s.employeeRepo.GetByID(ctx, orgID, employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("employee not found")
		}
		return nil, apperrors.Internal("failed to get employee", err)
	}
	if emp.BranchID != branchID {
		return nil, apperrors.Validation("employee does not belong to the specified branch")
	}

	hasService, err := s.branchService.HasService(ctx, orgID, branchID, serviceID)
	if err != nil {
		return nil, err
	}
	if !hasService {
		return nil, apperrors.Validation("service is not available at this branch")
	}

	endTime := startTime.Add(time.Duration(svc.DurationMinutes) * time.Minute)

	if bookingSettings != nil {
		if err := s.validateBookingWindow(startTime, endTime, *bookingSettings); err != nil {
			return nil, err
		}
	}

	if err := s.validateSchedule(ctx, orgID, branchID, employeeID, startTime, endTime); err != nil {
		return nil, err
	}

	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("organization not found")
		}
		return nil, apperrors.Internal("failed to get organization", err)
	}
	currency := org.Currency
	if currency == "" {
		currency = "EUR"
	}

	var created *booking.Booking
	err = s.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
		overlap, err := s.repo.HasOverlap(ctx, tx, orgID, employeeID, startTime, endTime, nil)
		if err != nil {
			return apperrors.Internal("failed to check availability", err)
		}
		if overlap {
			return apperrors.ErrDoubleBooking
		}

		b := &booking.Booking{
			OrganizationScoped: commonmodel.OrganizationScoped{OrganizationID: orgID},
			BranchID:           branchID,
			CustomerID:         customerID,
			EmployeeID:         employeeID,
			ServiceID:          serviceID,
			StartTime:          startTime,
			EndTime:            endTime,
			Status:             status,
			Notes:              notes,
			Price:              svc.Price,
			Currency:           currency,
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
	if created.Status == booking.StatusConfirmed {
		s.notifService.QueueBookingConfirmation(ctx, orgID, created, correlationID)
	}

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

		if err := s.validateSchedule(ctx, orgID, b.BranchID, b.EmployeeID, *req.StartTime, newEnd); err != nil {
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
		if err := s.validateSchedule(ctx, orgID, b.BranchID, b.EmployeeID, b.StartTime, newEnd); err != nil {
			return nil, err
		}

		org, err := s.orgRepo.GetByID(ctx, orgID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.NotFound("organization not found")
			}
			return nil, apperrors.Internal("failed to get organization", err)
		}
		currency := org.Currency
		if currency == "" {
			currency = "EUR"
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
			b.Currency = currency
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
		if err := s.validateSchedule(ctx, orgID, b.BranchID, newEmployeeID, b.StartTime, b.EndTime); err != nil {
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
	branchID, err := s.branchService.ResolveBranchID(ctx, orgID, req.BranchID)
	if err != nil {
		return nil, err
	}

	svc, err := s.serviceRepo.GetByID(ctx, orgID, req.ServiceID)
	if err != nil {
		return nil, apperrors.NotFound("service not found")
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, apperrors.Validation("invalid date format, use YYYY-MM-DD")
	}

	dayOfWeek := int(date.Weekday())
	workingHours, err := s.scheduleRepo.GetWorkingHours(ctx, orgID, &branchID, req.EmployeeID, dayOfWeek)
	if err != nil {
		return nil, apperrors.Internal("failed to get working hours", err)
	}
	if len(workingHours) == 0 {
		workingHours = schedule.DefaultHoursForDay(dayOfWeek)
	}
	if len(workingHours) == 0 {
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
		current := combineDateAndTime(date, wh.StartTime.Time)
		dayEnd := combineDateAndTime(date, wh.EndTime.Time)

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
				brStart := combineDateAndTime(date, br.StartTime.Time)
				brEnd := combineDateAndTime(date, br.EndTime.Time)
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

func (s *bookingService) GetPublicAvailability(ctx context.Context, orgID uuid.UUID, req booking.PublicAvailabilityRequest, minNoticeMinutes int) ([]booking.PublicTimeSlot, error) {
	branchID, err := req.ParsedBranchID()
	if err != nil {
		return nil, apperrors.Validation("invalid branch_id")
	}

	if _, err := s.branchService.GetByID(ctx, orgID, branchID); err != nil {
		return nil, err
	}

	serviceID, err := req.ParsedServiceID()
	if err != nil {
		return nil, apperrors.Validation("invalid service_id")
	}

	employeeID, err := req.ParsedEmployeeID()
	if err != nil {
		return nil, apperrors.Validation("invalid employee_id")
	}

	if employeeID != nil {
		slots, err := s.GetAvailability(ctx, orgID, booking.AvailabilityRequest{
			BranchID:   &branchID,
			EmployeeID: *employeeID,
			ServiceID:  serviceID,
			Date:       req.Date,
		})
		if err != nil {
			return nil, err
		}
		result := make([]booking.PublicTimeSlot, len(slots))
		for i, slot := range slots {
			result[i] = booking.PublicTimeSlot{
				StartTime: slot.StartTime,
				EndTime:   slot.EndTime,
				Available: slot.Available,
			}
			if slot.Available {
				result[i].EmployeeIDs = []uuid.UUID{*employeeID}
			}
		}
		applyMinNoticeFilter(result, minNoticeMinutes)
		return result, nil
	}

	employees, err := s.employeeRepo.ListByBranchAndService(ctx, orgID, branchID, serviceID)
	if err != nil {
		return nil, apperrors.Internal("failed to list employees for service", err)
	}
	if len(employees) == 0 {
		employees, _, err = s.employeeRepo.List(ctx, orgID, &branchID, 0, 1000, true)
		if err != nil {
			return nil, apperrors.Internal("failed to list employees", err)
		}
	}

	slotMap := make(map[string]*booking.PublicTimeSlot)
	for _, emp := range employees {
		slots, err := s.GetAvailability(ctx, orgID, booking.AvailabilityRequest{
			BranchID:   &branchID,
			EmployeeID: emp.ID,
			ServiceID:  serviceID,
			Date:       req.Date,
		})
		if err != nil {
			return nil, err
		}
		for _, slot := range slots {
			key := slot.StartTime.Format(time.RFC3339)
			existing, ok := slotMap[key]
			if !ok {
				slotMap[key] = &booking.PublicTimeSlot{
					StartTime: slot.StartTime,
					EndTime:   slot.EndTime,
					Available: slot.Available,
				}
				existing = slotMap[key]
			}
			if slot.Available {
				existing.Available = true
				existing.EmployeeIDs = append(existing.EmployeeIDs, emp.ID)
			}
		}
	}

	result := make([]booking.PublicTimeSlot, 0, len(slotMap))
	for _, slot := range slotMap {
		result = append(result, *slot)
	}

	sortPublicTimeSlots(result)
	applyMinNoticeFilter(result, minNoticeMinutes)
	return result, nil
}

func applyMinNoticeFilter(slots []booking.PublicTimeSlot, minNoticeMinutes int) {
	if minNoticeMinutes <= 0 {
		return
	}
	minStart := time.Now().Add(time.Duration(minNoticeMinutes) * time.Minute)
	for i := range slots {
		if slots[i].Available && slots[i].StartTime.Before(minStart) {
			slots[i].Available = false
			slots[i].EmployeeIDs = nil
		}
	}
}

func (s *bookingService) findOrCreatePublicCustomer(ctx context.Context, orgID uuid.UUID, info booking.PublicCustomerInfo) (*customer.Customer, error) {
	existing, err := s.customerRepo.FindByPhoneOrEmail(ctx, orgID, info.Phone, info.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.Internal("failed to find customer", err)
	}
	if existing != nil {
		updated := false
		if info.FirstName != "" && existing.FirstName != info.FirstName {
			existing.FirstName = info.FirstName
			updated = true
		}
		if info.LastName != "" && existing.LastName != info.LastName {
			existing.LastName = info.LastName
			updated = true
		}
		if info.Email != "" && existing.Email != info.Email {
			existing.Email = info.Email
			updated = true
		}
		if info.Phone != "" && existing.Phone != info.Phone {
			existing.Phone = info.Phone
			updated = true
		}
		if updated {
			if err := s.customerRepo.Update(ctx, existing); err != nil {
				return nil, apperrors.Internal("failed to update customer", err)
			}
		}
		return existing, nil
	}

	c := &customer.Customer{
		OrganizationScoped: commonmodel.OrganizationScoped{OrganizationID: orgID},
		FirstName:          info.FirstName,
		LastName:           info.LastName,
		Email:              info.Email,
		Phone:              info.Phone,
	}
	if err := s.customerRepo.Create(ctx, c); err != nil {
		return nil, apperrors.Internal("failed to create customer", err)
	}
	return c, nil
}

func (s *bookingService) resolveEmployeeForPublicBooking(ctx context.Context, orgID uuid.UUID, req booking.PublicCreateBookingRequest, durationMinutes int) (uuid.UUID, error) {
	endTime := req.StartTime.Add(time.Duration(durationMinutes) * time.Minute)

	if req.EmployeeID != nil {
		emp, err := s.employeeRepo.GetByID(ctx, orgID, *req.EmployeeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return uuid.Nil, apperrors.NotFound("employee not found")
			}
			return uuid.Nil, apperrors.Internal("failed to get employee", err)
		}
		if emp.BranchID != req.BranchID {
			return uuid.Nil, apperrors.Validation("employee does not belong to the specified branch")
		}
		if err := s.validateSchedule(ctx, orgID, req.BranchID, *req.EmployeeID, req.StartTime, endTime); err != nil {
			return uuid.Nil, err
		}
		overlap, err := s.repo.HasOverlap(ctx, nil, orgID, *req.EmployeeID, req.StartTime, endTime, nil)
		if err != nil {
			return uuid.Nil, apperrors.Internal("failed to check availability", err)
		}
		if overlap {
			return uuid.Nil, apperrors.ErrDoubleBooking
		}
		return *req.EmployeeID, nil
	}

	employees, err := s.employeeRepo.ListByBranchAndService(ctx, orgID, req.BranchID, req.ServiceID)
	if err != nil {
		return uuid.Nil, apperrors.Internal("failed to list employees for service", err)
	}
	if len(employees) == 0 {
		employees, _, err = s.employeeRepo.List(ctx, orgID, &req.BranchID, 0, 1000, true)
		if err != nil {
			return uuid.Nil, apperrors.Internal("failed to list employees", err)
		}
	}

	for _, emp := range employees {
		if err := s.validateSchedule(ctx, orgID, req.BranchID, emp.ID, req.StartTime, endTime); err != nil {
			continue
		}
		overlap, err := s.repo.HasOverlap(ctx, nil, orgID, emp.ID, req.StartTime, endTime, nil)
		if err != nil {
			return uuid.Nil, apperrors.Internal("failed to check availability", err)
		}
		if !overlap {
			return emp.ID, nil
		}
	}

	return uuid.Nil, apperrors.ErrDoubleBooking
}

func (s *bookingService) validateBookingWindow(start, end time.Time, settings organization.BookingSettings) error {
	now := time.Now()
	minStart := now.Add(time.Duration(settings.MinNoticeMinutes) * time.Minute)
	if start.Before(minStart) {
		return apperrors.Validation("booking is too soon")
	}

	maxStart := now.AddDate(0, 0, settings.BookingWindowDays)
	if settings.MaxNoticeMinutes != nil && *settings.MaxNoticeMinutes > 0 {
		maxFromMinutes := now.Add(time.Duration(*settings.MaxNoticeMinutes) * time.Minute)
		if maxFromMinutes.Before(maxStart) {
			maxStart = maxFromMinutes
		}
	}
	if start.After(maxStart) {
		return apperrors.Validation("booking is too far in the future")
	}

	_ = end
	return nil
}

func sortPublicTimeSlots(slots []booking.PublicTimeSlot) {
	for i := 0; i < len(slots); i++ {
		for j := i + 1; j < len(slots); j++ {
			if slots[j].StartTime.Before(slots[i].StartTime) {
				slots[i], slots[j] = slots[j], slots[i]
			}
		}
	}
}

func (s *bookingService) validateSchedule(ctx context.Context, orgID uuid.UUID, branchID uuid.UUID, employeeID uuid.UUID, start, end time.Time) error {
	dayOfWeek := int(start.Weekday())
	workingHours, err := s.scheduleRepo.GetWorkingHours(ctx, orgID, &branchID, employeeID, dayOfWeek)
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
		if !startTime.Before(wh.StartTime.Time) && !endTime.After(wh.EndTime.Time) {
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
