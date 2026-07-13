package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/backend/internal/booking"
)

type Repository interface {
	Create(ctx context.Context, tx *gorm.DB, b *booking.Booking) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*booking.Booking, error)
	Update(ctx context.Context, b *booking.Booking) error
	List(ctx context.Context, orgID uuid.UUID, filters ListFilters, offset, limit int) ([]booking.Booking, int64, error)
	HasOverlap(ctx context.Context, tx *gorm.DB, orgID, employeeID uuid.UUID, start, end time.Time, excludeID *uuid.UUID) (bool, error)
	GetByEmployeeAndDate(ctx context.Context, orgID, employeeID uuid.UUID, date time.Time) ([]booking.Booking, error)
	GetNextUpcomingByCustomer(ctx context.Context, orgID, customerID uuid.UUID) (*booking.Booking, error)
	CountByServiceID(ctx context.Context, orgID, serviceID uuid.UUID) (int64, error)
	CountByEmployeeID(ctx context.Context, orgID, employeeID uuid.UUID) (int64, error)
	CountByCustomerID(ctx context.Context, orgID, customerID uuid.UUID) (int64, error)
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type ListFilters struct {
	EmployeeID *uuid.UUID
	CustomerID *uuid.UUID
	Status     *booking.BookingStatus
	StartFrom  *time.Time
	StartTo    *time.Time
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

func (r *gormRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r *gormRepository) Create(ctx context.Context, tx *gorm.DB, b *booking.Booking) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(b).Error
}

func (r *gormRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (*booking.Booking, error) {
	var b booking.Booking
	err := r.scoped(ctx, orgID).First(&b, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormRepository) Update(ctx context.Context, b *booking.Booking) error {
	return r.db.WithContext(ctx).Save(b).Error
}

func (r *gormRepository) List(ctx context.Context, orgID uuid.UUID, filters ListFilters, offset, limit int) ([]booking.Booking, int64, error) {
	var bookings []booking.Booking
	var total int64

	query := r.scoped(ctx, orgID).Model(&booking.Booking{})
	if filters.EmployeeID != nil {
		query = query.Where("employee_id = ?", *filters.EmployeeID)
	}
	if filters.CustomerID != nil {
		query = query.Where("customer_id = ?", *filters.CustomerID)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.StartFrom != nil {
		query = query.Where("start_time >= ?", *filters.StartFrom)
	}
	if filters.StartTo != nil {
		query = query.Where("start_time <= ?", *filters.StartTo)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("start_time ASC").Offset(offset).Limit(limit).Find(&bookings).Error
	return bookings, total, err
}

func (r *gormRepository) HasOverlap(ctx context.Context, tx *gorm.DB, orgID, employeeID uuid.UUID, start, end time.Time, excludeID *uuid.UUID) (bool, error) {
	db := r.db
	if tx != nil {
		db = tx
	}

	query := db.WithContext(ctx).Model(&booking.Booking{}).
		Where("organization_id = ? AND employee_id = ?", orgID, employeeID).
		Where("status NOT IN ?", []booking.BookingStatus{booking.StatusCancelled}).
		Where("start_time < ? AND end_time > ?", end, start)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *gormRepository) GetByEmployeeAndDate(ctx context.Context, orgID, employeeID uuid.UUID, date time.Time) ([]booking.Booking, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var bookings []booking.Booking
	err := r.scoped(ctx, orgID).
		Where("employee_id = ? AND start_time >= ? AND start_time < ?", employeeID, startOfDay, endOfDay).
		Where("status NOT IN ?", []booking.BookingStatus{booking.StatusCancelled}).
		Order("start_time ASC").
		Find(&bookings).Error
	return bookings, err
}

func (r *gormRepository) GetNextUpcomingByCustomer(ctx context.Context, orgID, customerID uuid.UUID) (*booking.Booking, error) {
	var b booking.Booking
	err := r.scoped(ctx, orgID).
		Where("customer_id = ? AND start_time >= ? AND status NOT IN ?", customerID, time.Now().UTC(), []booking.BookingStatus{booking.StatusCancelled}).
		Order("start_time ASC").
		First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormRepository) CountByServiceID(ctx context.Context, orgID, serviceID uuid.UUID) (int64, error) {
	var count int64
	err := r.scoped(ctx, orgID).Model(&booking.Booking{}).
		Where("service_id = ?", serviceID).
		Count(&count).Error
	return count, err
}

func (r *gormRepository) CountByEmployeeID(ctx context.Context, orgID, employeeID uuid.UUID) (int64, error) {
	var count int64
	err := r.scoped(ctx, orgID).Model(&booking.Booking{}).
		Where("employee_id = ?", employeeID).
		Count(&count).Error
	return count, err
}

func (r *gormRepository) CountByCustomerID(ctx context.Context, orgID, customerID uuid.UUID) (int64, error) {
	var count int64
	err := r.scoped(ctx, orgID).Model(&booking.Booking{}).
		Where("customer_id = ?", customerID).
		Count(&count).Error
	return count, err
}
