package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/backend/internal/organization"
)

type Repository interface {
	Create(ctx context.Context, org *organization.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*organization.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*organization.Organization, error)
	Update(ctx context.Context, org *organization.Organization) error
	List(ctx context.Context, userID uuid.UUID, offset, limit int) ([]organization.Organization, int64, error)
	AddMember(ctx context.Context, member *organization.OrganizationUser) error
	GetMember(ctx context.Context, orgID, userID uuid.UUID) (*organization.OrganizationUser, error)
	ListMembers(ctx context.Context, orgID uuid.UUID, offset, limit int) ([]organization.OrganizationUser, int64, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, org *organization.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

func (r *gormRepository) GetByID(ctx context.Context, id uuid.UUID) (*organization.Organization, error) {
	var org organization.Organization
	err := r.db.WithContext(ctx).First(&org, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *gormRepository) GetBySlug(ctx context.Context, slug string) (*organization.Organization, error) {
	var org organization.Organization
	err := r.db.WithContext(ctx).First(&org, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *gormRepository) Update(ctx context.Context, org *organization.Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

func (r *gormRepository) List(ctx context.Context, userID uuid.UUID, offset, limit int) ([]organization.Organization, int64, error) {
	var orgs []organization.Organization
	var total int64

	query := r.db.WithContext(ctx).
		Model(&organization.Organization{}).
		Joins("JOIN organization_users ou ON ou.organization_id = organizations.id").
		Where("ou.user_id = ? AND ou.is_active = true", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&orgs).Error
	return orgs, total, err
}

func (r *gormRepository) AddMember(ctx context.Context, member *organization.OrganizationUser) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *gormRepository) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*organization.OrganizationUser, error) {
	var member organization.OrganizationUser
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *gormRepository) ListMembers(ctx context.Context, orgID uuid.UUID, offset, limit int) ([]organization.OrganizationUser, int64, error) {
	var members []organization.OrganizationUser
	var total int64

	query := r.db.WithContext(ctx).Model(&organization.OrganizationUser{}).
		Where("organization_id = ? AND is_active = true", orgID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&members).Error
	return members, total, err
}
