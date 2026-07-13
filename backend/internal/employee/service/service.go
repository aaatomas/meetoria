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
	"github.com/meetoria/meetoria/backend/internal/common/storage"
	"github.com/meetoria/meetoria/backend/internal/employee"
	"github.com/meetoria/meetoria/backend/internal/employee/repository"
)

type Service interface {
	Create(ctx context.Context, orgID uuid.UUID, req employee.CreateEmployeeRequest) (*employee.Employee, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*employee.Employee, error)
	Update(ctx context.Context, orgID, id uuid.UUID, req employee.UpdateEmployeeRequest) (*employee.Employee, error)
	UpdateAvatar(ctx context.Context, orgID, id uuid.UUID, avatarURL string) (*employee.Employee, error)
	CheckDeletion(ctx context.Context, orgID, id uuid.UUID) (commonmodel.DeletionCheck, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID, params commonmodel.PaginationParams, activeOnly bool) (commonmodel.PaginatedResponse[employee.Employee], error)
}

type employeeService struct {
	repo        repository.Repository
	bookingRepo bookingrepo.Repository
	storage     *storage.LocalStorage
}

func NewService(repo repository.Repository, bookingRepo bookingrepo.Repository, fileStorage *storage.LocalStorage) Service {
	return &employeeService{repo: repo, bookingRepo: bookingRepo, storage: fileStorage}
}

func (s *employeeService) Create(ctx context.Context, orgID uuid.UUID, req employee.CreateEmployeeRequest) (*employee.Employee, error) {
	e := &employee.Employee{
		OrganizationScoped: commonmodel.OrganizationScoped{OrganizationID: orgID},
		UserID:             req.UserID,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Email:              req.Email,
		Phone:              req.Phone,
		Title:              req.Title,
		Bio:                req.Bio,
		IsActive:           true,
	}

	if err := s.repo.Create(ctx, e); err != nil {
		return nil, apperrors.Internal("failed to create employee", err)
	}

	if len(req.ServiceIDs) > 0 {
		if err := s.repo.SetServices(ctx, orgID, e.ID, req.ServiceIDs); err != nil {
			return nil, apperrors.Internal("failed to assign services", err)
		}
	}

	return e, nil
}

func (s *employeeService) GetByID(ctx context.Context, orgID, id uuid.UUID) (*employee.Employee, error) {
	e, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("employee not found")
		}
		return nil, apperrors.Internal("failed to get employee", err)
	}
	return e, nil
}

func (s *employeeService) Update(ctx context.Context, orgID, id uuid.UUID, req employee.UpdateEmployeeRequest) (*employee.Employee, error) {
	e, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	if req.FirstName != nil {
		e.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		e.LastName = *req.LastName
	}
	if req.Email != nil {
		e.Email = *req.Email
	}
	if req.Phone != nil {
		e.Phone = *req.Phone
	}
	if req.Title != nil {
		e.Title = *req.Title
	}
	if req.Bio != nil {
		e.Bio = *req.Bio
	}
	if req.Color != nil {
		e.Color = *req.Color
	}
	if req.IsActive != nil {
		e.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, e); err != nil {
		return nil, apperrors.Internal("failed to update employee", err)
	}
	return e, nil
}

func (s *employeeService) UpdateAvatar(ctx context.Context, orgID, id uuid.UUID, avatarURL string) (*employee.Employee, error) {
	e, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	oldURL := e.AvatarURL
	e.AvatarURL = avatarURL

	if err := s.repo.Update(ctx, e); err != nil {
		return nil, apperrors.Internal("failed to update employee avatar", err)
	}

	if s.storage != nil && oldURL != "" && oldURL != avatarURL {
		_ = s.storage.DeleteByURL(oldURL)
	}

	return e, nil
}

func (s *employeeService) CheckDeletion(ctx context.Context, orgID, id uuid.UUID) (commonmodel.DeletionCheck, error) {
	if _, err := s.GetByID(ctx, orgID, id); err != nil {
		return commonmodel.DeletionCheck{}, err
	}

	count, err := s.bookingRepo.CountByEmployeeID(ctx, orgID, id)
	if err != nil {
		return commonmodel.DeletionCheck{}, apperrors.Internal("failed to check employee bookings", err)
	}

	check := commonmodel.DeletionCheck{
		CanDelete:     count == 0,
		BookingsCount: count,
	}
	if count > 0 {
		check.Message = fmt.Sprintf("Cannot delete employee with %d booking(s). Deactivate them instead.", count)
	}
	return check, nil
}

func (s *employeeService) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	check, err := s.CheckDeletion(ctx, orgID, id)
	if err != nil {
		return err
	}
	if !check.CanDelete {
		return apperrors.Conflict(check.Message)
	}

	e, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteEmployeeServiceLinks(ctx, orgID, id); err != nil {
		return apperrors.Internal("failed to remove employee service links", err)
	}

	if s.storage != nil && e.AvatarURL != "" {
		_ = s.storage.DeleteByURL(e.AvatarURL)
	}

	return s.repo.Delete(ctx, orgID, id)
}

func (s *employeeService) List(ctx context.Context, orgID uuid.UUID, params commonmodel.PaginationParams, activeOnly bool) (commonmodel.PaginatedResponse[employee.Employee], error) {
	employees, total, err := s.repo.List(ctx, orgID, params.Offset(), params.Limit, activeOnly)
	if err != nil {
		return commonmodel.PaginatedResponse[employee.Employee]{}, apperrors.Internal("failed to list employees", err)
	}
	return commonmodel.NewPaginatedResponse(employees, total, params.Page, params.Limit), nil
}
