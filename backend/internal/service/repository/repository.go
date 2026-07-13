package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/backend/internal/employee"
	"github.com/meetoria/meetoria/backend/internal/service"
)

type Repository interface {
	Create(ctx context.Context, s *service.Service) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*service.Service, error)
	Update(ctx context.Context, s *service.Service) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID, offset, limit int, activeOnly bool) ([]service.Service, int64, error)
	ListByBranch(ctx context.Context, orgID, branchID uuid.UUID, offset, limit int, activeOnly bool) ([]service.Service, int64, error)
	UpdateCurrencyByOrg(ctx context.Context, orgID uuid.UUID, currency string) error
	DeleteEmployeeServiceLinks(ctx context.Context, orgID, serviceID uuid.UUID) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) scoped(ctx context.Context, orgID uuid.UUID) *gorm.DB {
	return r.db.WithContext(ctx).Where("organization_id = ?", orgID)
}

func (r *gormRepository) Create(ctx context.Context, s *service.Service) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *gormRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (*service.Service, error) {
	var svc service.Service
	err := r.scoped(ctx, orgID).First(&svc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

func (r *gormRepository) Update(ctx context.Context, s *service.Service) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *gormRepository) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return r.scoped(ctx, orgID).Delete(&service.Service{}, "id = ?", id).Error
}

func (r *gormRepository) List(ctx context.Context, orgID uuid.UUID, offset, limit int, activeOnly bool) ([]service.Service, int64, error) {
	var services []service.Service
	var total int64

	query := r.scoped(ctx, orgID).Model(&service.Service{})
	if activeOnly {
		query = query.Where("is_active = true")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("name ASC").Offset(offset).Limit(limit).Find(&services).Error
	return services, total, err
}

func (r *gormRepository) ListByBranch(ctx context.Context, orgID, branchID uuid.UUID, offset, limit int, activeOnly bool) ([]service.Service, int64, error) {
	var services []service.Service
	var total int64

	query := r.db.WithContext(ctx).
		Model(&service.Service{}).
		Joins("JOIN branch_services bs ON bs.service_id = services.id AND bs.organization_id = services.organization_id").
		Where("services.organization_id = ? AND bs.branch_id = ?", orgID, branchID)
	if activeOnly {
		query = query.Where("services.is_active = true")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("services.name ASC").Offset(offset).Limit(limit).Find(&services).Error
	return services, total, err
}

func (r *gormRepository) UpdateCurrencyByOrg(ctx context.Context, orgID uuid.UUID, currency string) error {
	return r.scoped(ctx, orgID).Model(&service.Service{}).Update("currency", currency).Error
}

func (r *gormRepository) DeleteEmployeeServiceLinks(ctx context.Context, orgID, serviceID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("organization_id = ? AND service_id = ?", orgID, serviceID).
		Delete(&employee.EmployeeService{}).Error
}
