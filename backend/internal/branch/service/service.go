package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/backend/internal/branch"
	branchrepo "github.com/meetoria/meetoria/backend/internal/branch/repository"
	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	"github.com/meetoria/meetoria/backend/pkg/phone"
)

type Service interface {
	CreateDefault(ctx context.Context, orgID uuid.UUID, timezone string) (*branch.Branch, error)
	Create(ctx context.Context, orgID uuid.UUID, req branch.CreateBranchRequest) (*branch.Branch, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*branch.Branch, error)
	GetDefault(ctx context.Context, orgID uuid.UUID) (*branch.Branch, error)
	Update(ctx context.Context, orgID, id uuid.UUID, req branch.UpdateBranchRequest) (*branch.Branch, error)
	SetDefault(ctx context.Context, orgID, id uuid.UUID) (*branch.Branch, error)
	CheckDeletion(ctx context.Context, orgID, id uuid.UUID) (commonmodel.DeletionCheck, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID, params commonmodel.PaginationParams, activeOnly bool) (commonmodel.PaginatedResponse[branch.Branch], error)
	ResolveBranchID(ctx context.Context, orgID uuid.UUID, branchID *uuid.UUID) (uuid.UUID, error)
	GetServiceIDs(ctx context.Context, orgID, branchID uuid.UUID) ([]uuid.UUID, error)
	HasService(ctx context.Context, orgID, branchID, serviceID uuid.UUID) (bool, error)
	AddService(ctx context.Context, orgID, branchID, serviceID uuid.UUID) error
	RemoveServiceLinks(ctx context.Context, orgID, serviceID uuid.UUID) error
}

type branchService struct {
	repo branchrepo.Repository
}

func NewService(repo branchrepo.Repository) Service {
	return &branchService{repo: repo}
}

func (s *branchService) CreateDefault(ctx context.Context, orgID uuid.UUID, timezone string) (*branch.Branch, error) {
	b := &branch.Branch{
		OrganizationScoped: commonmodel.OrganizationScoped{OrganizationID: orgID},
		Name:               "Main location",
		Timezone:           timezone,
		IsActive:           true,
		IsDefault:          true,
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, apperrors.Internal("failed to create default branch", err)
	}
	return b, nil
}

func (s *branchService) Create(ctx context.Context, orgID uuid.UUID, req branch.CreateBranchRequest) (*branch.Branch, error) {
	phoneValue, err := phone.NormalizeOptional(req.Phone)
	if err != nil {
		return nil, err
	}

	b := &branch.Branch{
		OrganizationScoped: commonmodel.OrganizationScoped{OrganizationID: orgID},
		Name:               req.Name,
		Address:            req.Address,
		City:               req.City,
		Country:            req.Country,
		Timezone:           req.Timezone,
		Phone:              phoneValue,
		Email:              req.Email,
		IsActive:           true,
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, apperrors.Internal("failed to create branch", err)
	}
	return b, nil
}

func (s *branchService) GetByID(ctx context.Context, orgID, id uuid.UUID) (*branch.Branch, error) {
	b, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("branch not found")
		}
		return nil, apperrors.Internal("failed to get branch", err)
	}
	return b, nil
}

func (s *branchService) GetDefault(ctx context.Context, orgID uuid.UUID) (*branch.Branch, error) {
	b, err := s.repo.GetDefault(ctx, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("default branch not found")
		}
		return nil, apperrors.Internal("failed to get default branch", err)
	}
	return b, nil
}

func (s *branchService) Update(ctx context.Context, orgID, id uuid.UUID, req branch.UpdateBranchRequest) (*branch.Branch, error) {
	b, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		b.Name = *req.Name
	}
	if req.Address != nil {
		b.Address = *req.Address
	}
	if req.City != nil {
		b.City = *req.City
	}
	if req.Country != nil {
		b.Country = *req.Country
	}
	if req.Timezone != nil {
		b.Timezone = *req.Timezone
	}
	if req.Phone != nil {
		phoneValue, err := phone.NormalizeOptionalPtr(req.Phone)
		if err != nil {
			return nil, err
		}
		if phoneValue != nil {
			b.Phone = *phoneValue
		} else {
			b.Phone = ""
		}
	}
	if req.Email != nil {
		b.Email = *req.Email
	}
	if req.IsActive != nil {
		if b.IsDefault && !*req.IsActive {
			return nil, apperrors.Validation("cannot deactivate the default branch")
		}
		b.IsActive = *req.IsActive
	}
	if err := s.repo.Update(ctx, b); err != nil {
		return nil, apperrors.Internal("failed to update branch", err)
	}
	if req.ServiceIDs != nil {
		if err := s.repo.SetServices(ctx, orgID, id, req.ServiceIDs); err != nil {
			return nil, apperrors.Internal("failed to update branch services", err)
		}
	}
	return b, nil
}

func (s *branchService) SetDefault(ctx context.Context, orgID, id uuid.UUID) (*branch.Branch, error) {
	b, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	if !b.IsActive {
		return nil, apperrors.Validation("inactive branch cannot be set as default")
	}
	if b.IsDefault {
		return b, nil
	}
	if err := s.repo.SetDefault(ctx, orgID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("branch not found")
		}
		return nil, apperrors.Internal("failed to set default branch", err)
	}
	return s.GetByID(ctx, orgID, id)
}

func (s *branchService) CheckDeletion(ctx context.Context, orgID, id uuid.UUID) (commonmodel.DeletionCheck, error) {
	b, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return commonmodel.DeletionCheck{}, err
	}
	if b.IsDefault {
		return commonmodel.DeletionCheck{
			CanDelete: false,
			Message:   "Cannot delete the default location. Set another location as default first.",
		}, nil
	}

	bookingsCount, err := s.repo.CountBookings(ctx, orgID, id)
	if err != nil {
		return commonmodel.DeletionCheck{}, apperrors.Internal("failed to check branch bookings", err)
	}
	employeesCount, err := s.repo.CountEmployees(ctx, orgID, id)
	if err != nil {
		return commonmodel.DeletionCheck{}, apperrors.Internal("failed to check branch employees", err)
	}

	check := commonmodel.DeletionCheck{
		CanDelete:      bookingsCount == 0 && employeesCount == 0,
		BookingsCount:  bookingsCount,
		EmployeesCount: employeesCount,
	}
	switch {
	case bookingsCount > 0 && employeesCount > 0:
		check.Message = fmt.Sprintf(
			"Cannot delete location with %d booking(s) and %d employee(s). Reassign or remove them first.",
			bookingsCount, employeesCount,
		)
	case bookingsCount > 0:
		check.Message = fmt.Sprintf("Cannot delete location with %d booking(s).", bookingsCount)
	case employeesCount > 0:
		check.Message = fmt.Sprintf("Cannot delete location with %d employee(s). Reassign or remove them first.", employeesCount)
	}
	return check, nil
}

func (s *branchService) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	check, err := s.CheckDeletion(ctx, orgID, id)
	if err != nil {
		return err
	}
	if !check.CanDelete {
		return apperrors.Conflict(check.Message)
	}

	b, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}
	if b.IsDefault {
		return apperrors.Validation("cannot delete the default branch")
	}
	return s.repo.Delete(ctx, orgID, id)
}

func (s *branchService) List(ctx context.Context, orgID uuid.UUID, params commonmodel.PaginationParams, activeOnly bool) (commonmodel.PaginatedResponse[branch.Branch], error) {
	branches, total, err := s.repo.List(ctx, orgID, params.Offset(), params.Limit, activeOnly)
	if err != nil {
		return commonmodel.PaginatedResponse[branch.Branch]{}, apperrors.Internal("failed to list branches", err)
	}
	return commonmodel.NewPaginatedResponse(branches, total, params.Page, params.Limit), nil
}

func (s *branchService) ResolveBranchID(ctx context.Context, orgID uuid.UUID, branchID *uuid.UUID) (uuid.UUID, error) {
	if branchID != nil {
		if _, err := s.GetByID(ctx, orgID, *branchID); err != nil {
			return uuid.Nil, err
		}
		return *branchID, nil
	}
	def, err := s.GetDefault(ctx, orgID)
	if err != nil {
		return uuid.Nil, err
	}
	return def.ID, nil
}

func (s *branchService) GetServiceIDs(ctx context.Context, orgID, branchID uuid.UUID) ([]uuid.UUID, error) {
	return s.repo.GetServiceIDs(ctx, orgID, branchID)
}

func (s *branchService) HasService(ctx context.Context, orgID, branchID, serviceID uuid.UUID) (bool, error) {
	ids, err := s.repo.GetServiceIDs(ctx, orgID, branchID)
	if err != nil {
		return false, apperrors.Internal("failed to get branch services", err)
	}
	for _, id := range ids {
		if id == serviceID {
			return true, nil
		}
	}
	return false, nil
}

func (s *branchService) AddService(ctx context.Context, orgID, branchID, serviceID uuid.UUID) error {
	if _, err := s.GetByID(ctx, orgID, branchID); err != nil {
		return err
	}
	if err := s.repo.AddService(ctx, orgID, branchID, serviceID); err != nil {
		return apperrors.Internal("failed to link service to branch", err)
	}
	return nil
}

func (s *branchService) RemoveServiceLinks(ctx context.Context, orgID, serviceID uuid.UUID) error {
	if err := s.repo.RemoveServiceLinks(ctx, orgID, serviceID); err != nil {
		return apperrors.Internal("failed to remove branch service links", err)
	}
	return nil
}
