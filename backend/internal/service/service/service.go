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
	svc "github.com/meetoria/meetoria/backend/internal/service"
	"github.com/meetoria/meetoria/backend/internal/service/repository"
)

type Service interface {
	Create(ctx context.Context, orgID uuid.UUID, req svc.CreateServiceRequest, currency string) (*svc.Service, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*svc.Service, error)
	Update(ctx context.Context, orgID, id uuid.UUID, req svc.UpdateServiceRequest) (*svc.Service, error)
	CheckDeletion(ctx context.Context, orgID, id uuid.UUID) (commonmodel.DeletionCheck, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID, params commonmodel.PaginationParams, activeOnly bool) (commonmodel.PaginatedResponse[svc.Service], error)
}

type serviceLayer struct {
	repo        repository.Repository
	bookingRepo bookingrepo.Repository
}

func NewService(repo repository.Repository, bookingRepo bookingrepo.Repository) Service {
	return &serviceLayer{repo: repo, bookingRepo: bookingRepo}
}

func (s *serviceLayer) Create(ctx context.Context, orgID uuid.UUID, req svc.CreateServiceRequest, currency string) (*svc.Service, error) {
	if currency == "" {
		currency = "EUR"
	}
	svcModel := &svc.Service{
		OrganizationScoped: commonmodel.OrganizationScoped{OrganizationID: orgID},
		Name:               req.Name,
		Description:        req.Description,
		DurationMinutes:    req.DurationMinutes,
		Price:              req.Price,
		Currency:           currency,
		Category:           req.Category,
		Color:              svc.NormalizeColor(req.Color),
		IsActive:           true,
	}

	if err := s.repo.Create(ctx, svcModel); err != nil {
		return nil, apperrors.Internal("failed to create service", err)
	}
	return svcModel, nil
}

func (s *serviceLayer) GetByID(ctx context.Context, orgID, id uuid.UUID) (*svc.Service, error) {
	svcModel, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("service not found")
		}
		return nil, apperrors.Internal("failed to get service", err)
	}
	return svcModel, nil
}

func (s *serviceLayer) Update(ctx context.Context, orgID, id uuid.UUID, req svc.UpdateServiceRequest) (*svc.Service, error) {
	svcModel, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		svcModel.Name = *req.Name
	}
	if req.Description != nil {
		svcModel.Description = *req.Description
	}
	if req.DurationMinutes != nil {
		svcModel.DurationMinutes = *req.DurationMinutes
	}
	if req.Price != nil {
		svcModel.Price = *req.Price
	}
	if req.Category != nil {
		svcModel.Category = *req.Category
	}
	if req.Color != nil {
		svcModel.Color = svc.NormalizeColor(*req.Color)
	}
	if req.IsActive != nil {
		svcModel.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, svcModel); err != nil {
		return nil, apperrors.Internal("failed to update service", err)
	}
	return svcModel, nil
}

func (s *serviceLayer) CheckDeletion(ctx context.Context, orgID, id uuid.UUID) (commonmodel.DeletionCheck, error) {
	if _, err := s.GetByID(ctx, orgID, id); err != nil {
		return commonmodel.DeletionCheck{}, err
	}

	count, err := s.bookingRepo.CountByServiceID(ctx, orgID, id)
	if err != nil {
		return commonmodel.DeletionCheck{}, apperrors.Internal("failed to check service bookings", err)
	}

	check := commonmodel.DeletionCheck{
		CanDelete:     count == 0,
		BookingsCount: count,
	}
	if count > 0 {
		check.Message = fmt.Sprintf("Cannot delete service with %d booking(s). Deactivate it instead.", count)
	}
	return check, nil
}

func (s *serviceLayer) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	check, err := s.CheckDeletion(ctx, orgID, id)
	if err != nil {
		return err
	}
	if !check.CanDelete {
		return apperrors.Conflict(check.Message)
	}

	if err := s.repo.DeleteEmployeeServiceLinks(ctx, orgID, id); err != nil {
		return apperrors.Internal("failed to remove employee service links", err)
	}
	return s.repo.Delete(ctx, orgID, id)
}

func (s *serviceLayer) List(ctx context.Context, orgID uuid.UUID, params commonmodel.PaginationParams, activeOnly bool) (commonmodel.PaginatedResponse[svc.Service], error) {
	services, total, err := s.repo.List(ctx, orgID, params.Offset(), params.Limit, activeOnly)
	if err != nil {
		return commonmodel.PaginatedResponse[svc.Service]{}, apperrors.Internal("failed to list services", err)
	}
	return commonmodel.NewPaginatedResponse(services, total, params.Page, params.Limit), nil
}
