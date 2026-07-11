package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	"github.com/meetoria/meetoria/backend/internal/employee"
)

type Repository interface {
	Create(ctx context.Context, e *employee.Employee) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*employee.Employee, error)
	Update(ctx context.Context, e *employee.Employee) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID, offset, limit int, activeOnly bool) ([]employee.Employee, int64, error)
	SetServices(ctx context.Context, orgID, employeeID uuid.UUID, serviceIDs []uuid.UUID) error
	GetServiceIDs(ctx context.Context, orgID, employeeID uuid.UUID) ([]uuid.UUID, error)
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

func (r *gormRepository) Create(ctx context.Context, e *employee.Employee) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *gormRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (*employee.Employee, error) {
	var e employee.Employee
	err := r.scoped(ctx, orgID).First(&e, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *gormRepository) Update(ctx context.Context, e *employee.Employee) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *gormRepository) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return r.scoped(ctx, orgID).Delete(&employee.Employee{}, "id = ?", id).Error
}

func (r *gormRepository) List(ctx context.Context, orgID uuid.UUID, offset, limit int, activeOnly bool) ([]employee.Employee, int64, error) {
	var employees []employee.Employee
	var total int64

	query := r.scoped(ctx, orgID).Model(&employee.Employee{})
	if activeOnly {
		query = query.Where("is_active = true")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("first_name ASC").Offset(offset).Limit(limit).Find(&employees).Error
	return employees, total, err
}

func (r *gormRepository) SetServices(ctx context.Context, orgID, employeeID uuid.UUID, serviceIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("organization_id = ? AND employee_id = ?", orgID, employeeID).
			Delete(&employee.EmployeeService{}).Error; err != nil {
			return err
		}

		for _, sid := range serviceIDs {
			es := &employee.EmployeeService{
				OrganizationScoped: commonmodel.OrganizationScoped{OrganizationID: orgID},
				EmployeeID:         employeeID,
				ServiceID:          sid,
			}
			if err := tx.Create(es).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *gormRepository) GetServiceIDs(ctx context.Context, orgID, employeeID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&employee.EmployeeService{}).
		Where("organization_id = ? AND employee_id = ?", orgID, employeeID).
		Pluck("service_id", &ids).Error
	return ids, err
}
