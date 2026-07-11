package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/backend/internal/customer"
)

type Repository interface {
	Create(ctx context.Context, c *customer.Customer) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*customer.Customer, error)
	Update(ctx context.Context, c *customer.Customer) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID, offset, limit int, search string) ([]customer.Customer, int64, error)
	GetBookingStats(ctx context.Context, orgID uuid.UUID, customerIDs []uuid.UUID) (map[uuid.UUID]customer.BookingStats, error)
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

func (r *gormRepository) Create(ctx context.Context, c *customer.Customer) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *gormRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (*customer.Customer, error) {
	var c customer.Customer
	err := r.scoped(ctx, orgID).First(&c, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *gormRepository) Update(ctx context.Context, c *customer.Customer) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *gormRepository) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return r.scoped(ctx, orgID).Delete(&customer.Customer{}, "id = ?", id).Error
}

func (r *gormRepository) List(ctx context.Context, orgID uuid.UUID, offset, limit int, search string) ([]customer.Customer, int64, error) {
	var customers []customer.Customer
	var total int64

	query := r.scoped(ctx, orgID).Model(&customer.Customer{})
	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where(
			"first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR phone ILIKE ?",
			pattern, pattern, pattern, pattern,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&customers).Error
	return customers, total, err
}

func (r *gormRepository) GetBookingStats(ctx context.Context, orgID uuid.UUID, customerIDs []uuid.UUID) (map[uuid.UUID]customer.BookingStats, error) {
	stats := make(map[uuid.UUID]customer.BookingStats, len(customerIDs))
	if len(customerIDs) == 0 {
		return stats, nil
	}

	type statsRow struct {
		CustomerID         uuid.UUID
		BookingsCount      int64
		CancellationsCount int64
	}

	var rows []statsRow
	err := r.db.WithContext(ctx).
		Table("bookings").
		Select(`
			customer_id,
			COUNT(*) FILTER (WHERE status <> 'cancelled') AS bookings_count,
			COUNT(*) FILTER (WHERE status = 'cancelled') AS cancellations_count
		`).
		Where("organization_id = ? AND customer_id IN ? AND deleted_at IS NULL", orgID, customerIDs).
		Group("customer_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		stats[row.CustomerID] = customer.BookingStats{
			BookingsCount:      row.BookingsCount,
			CancellationsCount: row.CancellationsCount,
		}
	}

	return stats, nil
}
