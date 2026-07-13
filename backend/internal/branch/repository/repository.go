package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	"github.com/meetoria/meetoria/backend/internal/branch"
)

type Repository interface {
	Create(ctx context.Context, b *branch.Branch) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*branch.Branch, error)
	GetDefault(ctx context.Context, orgID uuid.UUID) (*branch.Branch, error)
	Update(ctx context.Context, b *branch.Branch) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID, offset, limit int, activeOnly bool) ([]branch.Branch, int64, error)
	CountBookings(ctx context.Context, orgID, branchID uuid.UUID) (int64, error)
	CountEmployees(ctx context.Context, orgID, branchID uuid.UUID) (int64, error)
	SetDefault(ctx context.Context, orgID, branchID uuid.UUID) error
	SetServices(ctx context.Context, orgID, branchID uuid.UUID, serviceIDs []uuid.UUID) error
	AddService(ctx context.Context, orgID, branchID, serviceID uuid.UUID) error
	RemoveServiceLinks(ctx context.Context, orgID, serviceID uuid.UUID) error
	GetServiceIDs(ctx context.Context, orgID, branchID uuid.UUID) ([]uuid.UUID, error)
	ListServiceIDsByBranch(ctx context.Context, orgID, branchID uuid.UUID) ([]uuid.UUID, error)
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

func (r *gormRepository) Create(ctx context.Context, b *branch.Branch) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *gormRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (*branch.Branch, error) {
	var b branch.Branch
	err := r.scoped(ctx, orgID).First(&b, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormRepository) GetDefault(ctx context.Context, orgID uuid.UUID) (*branch.Branch, error) {
	var b branch.Branch
	err := r.scoped(ctx, orgID).Where("is_default = true AND is_active = true").First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormRepository) Update(ctx context.Context, b *branch.Branch) error {
	return r.db.WithContext(ctx).Save(b).Error
}

func (r *gormRepository) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return r.scoped(ctx, orgID).Delete(&branch.Branch{}, "id = ?", id).Error
}

func (r *gormRepository) List(ctx context.Context, orgID uuid.UUID, offset, limit int, activeOnly bool) ([]branch.Branch, int64, error) {
	var branches []branch.Branch
	var total int64
	query := r.scoped(ctx, orgID).Model(&branch.Branch{})
	if activeOnly {
		query = query.Where("is_active = true")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("is_default DESC, name ASC").Offset(offset).Limit(limit).Find(&branches).Error
	return branches, total, err
}

func (r *gormRepository) CountBookings(ctx context.Context, orgID, branchID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("bookings").
		Where("organization_id = ? AND branch_id = ? AND deleted_at IS NULL", orgID, branchID).
		Count(&count).Error
	return count, err
}

func (r *gormRepository) CountEmployees(ctx context.Context, orgID, branchID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("employees").
		Where("organization_id = ? AND branch_id = ? AND deleted_at IS NULL", orgID, branchID).
		Count(&count).Error
	return count, err
}

func (r *gormRepository) SetDefault(ctx context.Context, orgID, branchID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&branch.Branch{}).
			Where("organization_id = ?", orgID).
			Update("is_default", false).Error; err != nil {
			return err
		}
		result := tx.Model(&branch.Branch{}).
			Where("organization_id = ? AND id = ?", orgID, branchID).
			Update("is_default", true)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *gormRepository) SetServices(ctx context.Context, orgID, branchID uuid.UUID, serviceIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("organization_id = ? AND branch_id = ?", orgID, branchID).
			Delete(&branch.BranchService{}).Error; err != nil {
			return err
		}
		for _, sid := range serviceIDs {
			bs := &branch.BranchService{
				OrganizationJunction: commonmodel.OrganizationJunction{OrganizationID: orgID},
				BranchID:           branchID,
				ServiceID:          sid,
			}
			if err := tx.Create(bs).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *gormRepository) AddService(ctx context.Context, orgID, branchID, serviceID uuid.UUID) error {
	bs := &branch.BranchService{
		OrganizationJunction: commonmodel.OrganizationJunction{OrganizationID: orgID},
		BranchID:           branchID,
		ServiceID:          serviceID,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "branch_id"}, {Name: "service_id"}},
		DoNothing: true,
	}).Create(bs).Error
}

func (r *gormRepository) RemoveServiceLinks(ctx context.Context, orgID, serviceID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("organization_id = ? AND service_id = ?", orgID, serviceID).
		Delete(&branch.BranchService{}).Error
}

func (r *gormRepository) GetServiceIDs(ctx context.Context, orgID, branchID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&branch.BranchService{}).
		Where("organization_id = ? AND branch_id = ?", orgID, branchID).
		Pluck("service_id", &ids).Error
	return ids, err
}

func (r *gormRepository) ListServiceIDsByBranch(ctx context.Context, orgID, branchID uuid.UUID) ([]uuid.UUID, error) {
	return r.GetServiceIDs(ctx, orgID, branchID)
}
