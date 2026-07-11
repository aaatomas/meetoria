package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	"github.com/meetoria/meetoria/backend/internal/organization"
	"github.com/meetoria/meetoria/backend/internal/organization/repository"
)

type Service interface {
	Create(ctx context.Context, req organization.CreateOrganizationRequest, ownerUserID uuid.UUID) (*organization.Organization, error)
	GetByID(ctx context.Context, id uuid.UUID) (*organization.Organization, error)
	Update(ctx context.Context, id uuid.UUID, req organization.UpdateOrganizationRequest) (*organization.Organization, error)
	List(ctx context.Context, userID uuid.UUID, params commonmodel.PaginationParams) (commonmodel.PaginatedResponse[organization.Organization], error)
	VerifyMembership(ctx context.Context, orgID, userID uuid.UUID, allowedRoles ...organization.OrganizationRole) error
}

type organizationService struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &organizationService{repo: repo}
}

func (s *organizationService) Create(ctx context.Context, req organization.CreateOrganizationRequest, ownerUserID uuid.UUID) (*organization.Organization, error) {
	existing, err := s.repo.GetBySlug(ctx, req.Slug)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.Internal("failed to check slug", err)
	}
	if existing != nil {
		return nil, apperrors.Conflict("organization slug already exists")
	}

	org := &organization.Organization{
		Name:     req.Name,
		Slug:     req.Slug,
		Timezone: req.Timezone,
		Email:    req.Email,
		Phone:    req.Phone,
		IsActive: true,
	}

	if err := s.repo.Create(ctx, org); err != nil {
		return nil, apperrors.Internal("failed to create organization", err)
	}

	member := &organization.OrganizationUser{
		OrganizationID: org.ID,
		UserID:         ownerUserID,
		Role:           organization.RoleOrganizationOwner,
		IsActive:       true,
	}
	if err := s.repo.AddMember(ctx, member); err != nil {
		return nil, apperrors.Internal("failed to add organization owner", err)
	}

	return org, nil
}

func (s *organizationService) GetByID(ctx context.Context, id uuid.UUID) (*organization.Organization, error) {
	org, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("organization not found")
		}
		return nil, apperrors.Internal("failed to get organization", err)
	}
	return org, nil
}

func (s *organizationService) Update(ctx context.Context, id uuid.UUID, req organization.UpdateOrganizationRequest) (*organization.Organization, error) {
	org, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.Timezone != nil {
		org.Timezone = *req.Timezone
	}
	if req.Email != nil {
		org.Email = *req.Email
	}
	if req.Phone != nil {
		org.Phone = *req.Phone
	}
	if req.Address != nil {
		org.Address = *req.Address
	}
	if req.LogoURL != nil {
		org.LogoURL = *req.LogoURL
	}

	if err := s.repo.Update(ctx, org); err != nil {
		return nil, apperrors.Internal("failed to update organization", err)
	}
	return org, nil
}

func (s *organizationService) List(ctx context.Context, userID uuid.UUID, params commonmodel.PaginationParams) (commonmodel.PaginatedResponse[organization.Organization], error) {
	orgs, total, err := s.repo.List(ctx, userID, params.Offset(), params.Limit)
	if err != nil {
		return commonmodel.PaginatedResponse[organization.Organization]{}, apperrors.Internal("failed to list organizations", err)
	}
	return commonmodel.NewPaginatedResponse(orgs, total, params.Page, params.Limit), nil
}

func (s *organizationService) VerifyMembership(ctx context.Context, orgID, userID uuid.UUID, allowedRoles ...organization.OrganizationRole) error {
	member, err := s.repo.GetMember(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrTenantIsolation
		}
		return apperrors.Internal("failed to verify membership", err)
	}

	if !member.IsActive {
		return apperrors.ErrTenantIsolation
	}

	if len(allowedRoles) == 0 {
		return nil
	}

	for _, role := range allowedRoles {
		if member.Role == role {
			return nil
		}
	}

	return apperrors.ErrForbidden
}
