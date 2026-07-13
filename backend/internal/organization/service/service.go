package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	"github.com/meetoria/meetoria/backend/internal/organization"
	"github.com/meetoria/meetoria/backend/internal/organization/repository"
	branchservice "github.com/meetoria/meetoria/backend/internal/branch/service"
	scheduleservice "github.com/meetoria/meetoria/backend/internal/schedule/service"
	servicerepo "github.com/meetoria/meetoria/backend/internal/service/repository"
	"github.com/meetoria/meetoria/backend/pkg/phone"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

type Service interface {
	Create(ctx context.Context, req organization.CreateOrganizationRequest, ownerUserID uuid.UUID) (*organization.Organization, error)
	GetByID(ctx context.Context, id uuid.UUID) (*organization.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*organization.Organization, error)
	Update(ctx context.Context, id uuid.UUID, req organization.UpdateOrganizationRequest) (*organization.Organization, error)
	List(ctx context.Context, userID uuid.UUID, params commonmodel.PaginationParams) (commonmodel.PaginatedResponse[organization.Organization], error)
	VerifyMembership(ctx context.Context, orgID, userID uuid.UUID, allowedRoles ...organization.OrganizationRole) error
}

type organizationService struct {
	repo            repository.Repository
	scheduleService scheduleservice.Service
	serviceRepo     servicerepo.Repository
	branchService   branchservice.Service
}

func NewService(repo repository.Repository, scheduleService scheduleservice.Service, serviceRepo servicerepo.Repository, branchService branchservice.Service) Service {
	return &organizationService{repo: repo, scheduleService: scheduleService, serviceRepo: serviceRepo, branchService: branchService}
}

func (s *organizationService) Create(ctx context.Context, req organization.CreateOrganizationRequest, ownerUserID uuid.UUID) (*organization.Organization, error) {
	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	if slug == "" {
		var err error
		slug, err = s.generateUniqueSlug(ctx, req.Name)
		if err != nil {
			return nil, err
		}
	} else if !slugPattern.MatchString(slug) {
		return nil, apperrors.Validation("slug must contain only lowercase letters, numbers, and hyphens")
	} else {
		existing, err := s.repo.GetBySlug(ctx, slug)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Internal("failed to check slug", err)
		}
		if existing != nil {
			return nil, apperrors.Conflict("organization slug already exists")
		}
	}

	timezone := strings.TrimSpace(req.Timezone)
	if timezone == "" {
		timezone = organization.DefaultTimezone
	}

	currency := req.Currency
	if currency == "" {
		currency = organization.DefaultCurrency
	}

	phoneValue, err := phone.NormalizeOptional(req.Phone)
	if err != nil {
		return nil, err
	}

	org := &organization.Organization{
		Name:     req.Name,
		Slug:     slug,
		Timezone: timezone,
		Currency: organization.NormalizeCurrency(currency),
		Email:    req.Email,
		Phone:    phoneValue,
		IsActive: true,
	}
	defaultSettings, _ := organization.MarshalSettings(organization.DefaultOrganizationSettings())
	org.Settings = defaultSettings

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

	if s.branchService != nil {
		defBranch, err := s.branchService.CreateDefault(ctx, org.ID, timezone)
		if err != nil {
			return nil, err
		}
		if s.scheduleService != nil {
			if err := s.scheduleService.SeedDefaultHours(ctx, org.ID, defBranch.ID); err != nil {
				return nil, apperrors.Internal("failed to seed default working hours", err)
			}
		}
	}

	return org, nil
}

func (s *organizationService) generateUniqueSlug(ctx context.Context, name string) (string, error) {
	base := organization.Slugify(name)
	if len(base) < 2 {
		base = "organization"
	}

	for i := 0; i < 100; i++ {
		slug := base
		if i > 0 {
			slug = fmt.Sprintf("%s-%d", base, i+1)
		}

		existing, err := s.repo.GetBySlug(ctx, slug)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", apperrors.Internal("failed to check slug", err)
		}
		if existing == nil {
			return slug, nil
		}
	}

	return "", apperrors.Internal("failed to generate unique slug", nil)
}

func (s *organizationService) GetByID(ctx context.Context, id uuid.UUID) (*organization.Organization, error) {
	org, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("organization not found")
		}
		return nil, apperrors.Internal("failed to get organization", err)
	}
	org.Currency = organization.NormalizeCurrency(org.Currency)
	return org, nil
}

func (s *organizationService) GetBySlug(ctx context.Context, slug string) (*organization.Organization, error) {
	org, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("organization not found")
		}
		return nil, apperrors.Internal("failed to get organization", err)
	}
	if !org.IsActive {
		return nil, apperrors.NotFound("organization not found")
	}
	org.Currency = organization.NormalizeCurrency(org.Currency)
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
	if req.Slug != nil {
		slug := strings.ToLower(strings.TrimSpace(*req.Slug))
		if !slugPattern.MatchString(slug) {
			return nil, apperrors.Validation("slug must contain only lowercase letters, numbers, and hyphens")
		}
		if slug != org.Slug {
			existing, err := s.repo.GetBySlug(ctx, slug)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.Internal("failed to check slug", err)
			}
			if existing != nil {
				return nil, apperrors.Conflict("organization slug already exists")
			}
			org.Slug = slug
		}
	}
	if req.Timezone != nil {
		org.Timezone = *req.Timezone
	}
	if req.Currency != nil {
		org.Currency = organization.NormalizeCurrency(*req.Currency)
	}
	if req.Email != nil {
		org.Email = *req.Email
	}
	if req.Phone != nil {
		phoneValue, err := phone.NormalizeOptionalPtr(req.Phone)
		if err != nil {
			return nil, err
		}
		if phoneValue != nil {
			org.Phone = *phoneValue
		} else {
			org.Phone = ""
		}
	}
	if req.Address != nil {
		org.Address = *req.Address
	}
	if req.LogoURL != nil {
		org.LogoURL = *req.LogoURL
	}
	settingsChanged := false
	settings := organization.ParseSettings(org.Settings)
	if req.TimeFormat != nil {
		settings.TimeFormat = organization.NormalizeTimeFormat(*req.TimeFormat)
		settingsChanged = true
	}
	if req.Booking != nil {
		settings.Booking = req.Booking.WithDefaults()
		settingsChanged = true
	}
	if settingsChanged {
		marshaled, err := organization.MarshalSettings(settings)
		if err != nil {
			return nil, apperrors.Internal("failed to marshal organization settings", err)
		}
		org.Settings = marshaled
	}

	if err := s.repo.Update(ctx, org); err != nil {
		return nil, apperrors.Internal("failed to update organization", err)
	}
	if req.Currency != nil && s.serviceRepo != nil {
		if err := s.serviceRepo.UpdateCurrencyByOrg(ctx, id, org.Currency); err != nil {
			return nil, apperrors.Internal("failed to update service currencies", err)
		}
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
