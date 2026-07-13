package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	bookingrepo "github.com/meetoria/meetoria/backend/internal/booking/repository"
	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	"github.com/meetoria/meetoria/backend/internal/customer"
	"github.com/meetoria/meetoria/backend/internal/customer/repository"
	notifservice "github.com/meetoria/meetoria/backend/internal/notification/service"
)

type Service interface {
	Create(ctx context.Context, orgID uuid.UUID, req customer.CreateCustomerRequest) (*customer.Customer, error)
	FindOrCreate(ctx context.Context, orgID uuid.UUID, req customer.CreateCustomerRequest) (*customer.Customer, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*customer.Customer, error)
	Update(ctx context.Context, orgID, id uuid.UUID, req customer.UpdateCustomerRequest) (*customer.Customer, error)
	CheckDeletion(ctx context.Context, orgID, id uuid.UUID) (commonmodel.DeletionCheck, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID, params commonmodel.PaginationParams, search string) (commonmodel.PaginatedResponse[customer.ListItem], error)
	SendSMS(ctx context.Context, orgID, customerID, correlationID uuid.UUID) error
	SendEmail(ctx context.Context, orgID, customerID, correlationID uuid.UUID) error
}

type customerService struct {
	repo         repository.Repository
	bookingRepo  bookingrepo.Repository
	notifService notifservice.Service
}

func NewService(repo repository.Repository, bookingRepo bookingrepo.Repository, notifService notifservice.Service) Service {
	return &customerService{
		repo:         repo,
		bookingRepo:  bookingRepo,
		notifService: notifService,
	}
}

func (s *customerService) Create(ctx context.Context, orgID uuid.UUID, req customer.CreateCustomerRequest) (*customer.Customer, error) {
	c := &customer.Customer{
		OrganizationScoped: commonmodel.OrganizationScoped{OrganizationID: orgID},
		UserID:             req.UserID,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Email:              req.Email,
		Phone:              req.Phone,
		Notes:              req.Notes,
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return nil, apperrors.Internal("failed to create customer", err)
	}
	return c, nil
}

func (s *customerService) FindOrCreate(ctx context.Context, orgID uuid.UUID, req customer.CreateCustomerRequest) (*customer.Customer, error) {
	existing, err := s.repo.FindByPhoneOrEmail(ctx, orgID, req.Phone, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.Internal("failed to find customer", err)
	}
	if existing != nil {
		updated := false
		if req.FirstName != "" && existing.FirstName != req.FirstName {
			existing.FirstName = req.FirstName
			updated = true
		}
		if req.LastName != "" && existing.LastName != req.LastName {
			existing.LastName = req.LastName
			updated = true
		}
		if req.Email != "" && existing.Email != req.Email {
			existing.Email = req.Email
			updated = true
		}
		if req.Phone != "" && existing.Phone != req.Phone {
			existing.Phone = req.Phone
			updated = true
		}
		if updated {
			if err := s.repo.Update(ctx, existing); err != nil {
				return nil, apperrors.Internal("failed to update customer", err)
			}
		}
		return existing, nil
	}

	return s.Create(ctx, orgID, req)
}

func (s *customerService) GetByID(ctx context.Context, orgID, id uuid.UUID) (*customer.Customer, error) {
	c, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("customer not found")
		}
		return nil, apperrors.Internal("failed to get customer", err)
	}
	return c, nil
}

func (s *customerService) Update(ctx context.Context, orgID, id uuid.UUID, req customer.UpdateCustomerRequest) (*customer.Customer, error) {
	c, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	if req.FirstName != nil {
		c.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		c.LastName = *req.LastName
	}
	if req.Email != nil {
		c.Email = *req.Email
	}
	if req.Phone != nil {
		c.Phone = *req.Phone
	}
	if req.Notes != nil {
		c.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, c); err != nil {
		return nil, apperrors.Internal("failed to update customer", err)
	}
	return c, nil
}

func (s *customerService) CheckDeletion(ctx context.Context, orgID, id uuid.UUID) (commonmodel.DeletionCheck, error) {
	if _, err := s.GetByID(ctx, orgID, id); err != nil {
		return commonmodel.DeletionCheck{}, err
	}

	count, err := s.bookingRepo.CountByCustomerID(ctx, orgID, id)
	if err != nil {
		return commonmodel.DeletionCheck{}, apperrors.Internal("failed to check customer bookings", err)
	}

	check := commonmodel.DeletionCheck{
		CanDelete:     count == 0,
		BookingsCount: count,
	}
	if count > 0 {
		check.Message = fmt.Sprintf("Cannot delete customer with %d booking(s).", count)
	}
	return check, nil
}

func (s *customerService) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	check, err := s.CheckDeletion(ctx, orgID, id)
	if err != nil {
		return err
	}
	if !check.CanDelete {
		return apperrors.Conflict(check.Message)
	}

	if err := s.repo.Delete(ctx, orgID, id); err != nil {
		return apperrors.Internal("failed to delete customer", err)
	}
	return nil
}

func (s *customerService) List(ctx context.Context, orgID uuid.UUID, params commonmodel.PaginationParams, search string) (commonmodel.PaginatedResponse[customer.ListItem], error) {
	customers, total, err := s.repo.List(ctx, orgID, params.Offset(), params.Limit, search)
	if err != nil {
		return commonmodel.PaginatedResponse[customer.ListItem]{}, apperrors.Internal("failed to list customers", err)
	}

	customerIDs := make([]uuid.UUID, len(customers))
	for i, c := range customers {
		customerIDs[i] = c.ID
	}

	stats, err := s.repo.GetBookingStats(ctx, orgID, customerIDs)
	if err != nil {
		return commonmodel.PaginatedResponse[customer.ListItem]{}, apperrors.Internal("failed to get customer booking stats", err)
	}

	items := make([]customer.ListItem, len(customers))
	for i, c := range customers {
		itemStats := stats[c.ID]
		items[i] = customer.ListItem{
			Customer:           c,
			BookingsCount:      itemStats.BookingsCount,
			CancellationsCount: itemStats.CancellationsCount,
		}
	}

	return commonmodel.NewPaginatedResponse(items, total, params.Page, params.Limit), nil
}

func (s *customerService) SendSMS(ctx context.Context, orgID, customerID, correlationID uuid.UUID) error {
	c, err := s.GetByID(ctx, orgID, customerID)
	if err != nil {
		return err
	}
	if c.Phone == "" {
		return apperrors.Validation("customer has no phone number")
	}

	b, err := s.bookingRepo.GetNextUpcomingByCustomer(ctx, orgID, customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.Validation("customer has no upcoming booking")
		}
		return apperrors.Internal("failed to get upcoming booking", err)
	}

	return s.notifService.QueueBookingConfirmationSMS(ctx, orgID, b, correlationID)
}

func (s *customerService) SendEmail(ctx context.Context, orgID, customerID, correlationID uuid.UUID) error {
	c, err := s.GetByID(ctx, orgID, customerID)
	if err != nil {
		return err
	}
	if c.Email == "" {
		return apperrors.Validation("customer has no email address")
	}

	b, err := s.bookingRepo.GetNextUpcomingByCustomer(ctx, orgID, customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.Validation("customer has no upcoming booking")
		}
		return apperrors.Internal("failed to get upcoming booking", err)
	}

	return s.notifService.QueueBookingConfirmationEmail(ctx, orgID, b, correlationID)
}
