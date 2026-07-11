package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/backend/internal/schedule"
)

type Repository interface {
	SetWorkingHours(ctx context.Context, orgID uuid.UUID, employeeID *uuid.UUID, hours []schedule.WorkingHours) error
	GetWorkingHours(ctx context.Context, orgID uuid.UUID, employeeID uuid.UUID, dayOfWeek int) ([]schedule.WorkingHours, error)
	GetBreaks(ctx context.Context, orgID uuid.UUID, employeeID uuid.UUID, dayOfWeek int) ([]schedule.Break, error)
	CreateHoliday(ctx context.Context, h *schedule.Holiday) error
	ListHolidays(ctx context.Context, orgID uuid.UUID, employeeID *uuid.UUID, from, to time.Time) ([]schedule.Holiday, error)
	IsHoliday(ctx context.Context, orgID uuid.UUID, employeeID uuid.UUID, date time.Time) (bool, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) SetWorkingHours(ctx context.Context, orgID uuid.UUID, employeeID *uuid.UUID, hours []schedule.WorkingHours) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Where("organization_id = ?", orgID)
		if employeeID != nil {
			query = query.Where("employee_id = ?", *employeeID)
		} else {
			query = query.Where("employee_id IS NULL")
		}
		if err := query.Delete(&schedule.WorkingHours{}).Error; err != nil {
			return err
		}
		for i := range hours {
			if err := tx.Create(&hours[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *gormRepository) GetWorkingHours(ctx context.Context, orgID uuid.UUID, employeeID uuid.UUID, dayOfWeek int) ([]schedule.WorkingHours, error) {
	var hours []schedule.WorkingHours

	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND employee_id = ? AND day_of_week = ? AND is_active = true", orgID, employeeID, dayOfWeek).
		Find(&hours).Error
	if err != nil {
		return nil, err
	}
	if len(hours) > 0 {
		return hours, nil
	}

	err = r.db.WithContext(ctx).
		Where("organization_id = ? AND employee_id IS NULL AND day_of_week = ? AND is_active = true", orgID, dayOfWeek).
		Find(&hours).Error
	return hours, err
}

func (r *gormRepository) GetBreaks(ctx context.Context, orgID uuid.UUID, employeeID uuid.UUID, dayOfWeek int) ([]schedule.Break, error) {
	var breaks []schedule.Break
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND (employee_id = ? OR employee_id IS NULL) AND day_of_week = ?", orgID, employeeID, dayOfWeek).
		Find(&breaks).Error
	return breaks, err
}

func (r *gormRepository) CreateHoliday(ctx context.Context, h *schedule.Holiday) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *gormRepository) ListHolidays(ctx context.Context, orgID uuid.UUID, employeeID *uuid.UUID, from, to time.Time) ([]schedule.Holiday, error) {
	var holidays []schedule.Holiday
	query := r.db.WithContext(ctx).Where("organization_id = ? AND date >= ? AND date <= ?", orgID, from, to)
	if employeeID != nil {
		query = query.Where("employee_id = ? OR employee_id IS NULL", *employeeID)
	}
	err := query.Order("date ASC").Find(&holidays).Error
	return holidays, err
}

func (r *gormRepository) IsHoliday(ctx context.Context, orgID uuid.UUID, employeeID uuid.UUID, date time.Time) (bool, error) {
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	var count int64
	err := r.db.WithContext(ctx).Model(&schedule.Holiday{}).
		Where("organization_id = ? AND (employee_id = ? OR employee_id IS NULL) AND date = ?", orgID, employeeID, dateOnly).
		Count(&count).Error
	return count > 0, err
}
