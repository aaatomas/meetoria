package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/backend/internal/analytics"
)

type Repository interface {
	GetOrganizationStats(ctx context.Context, orgID uuid.UUID, from, to time.Time) ([]analytics.OrganizationStats, error)
	GetLiveDashboardSummary(ctx context.Context, orgID uuid.UUID, from, to time.Time) (*analytics.LiveDashboardSummary, error)
	GetEmployeeStats(ctx context.Context, orgID, employeeID uuid.UUID, from, to time.Time) ([]analytics.EmployeeStats, error)
	GetCustomerStats(ctx context.Context, orgID, customerID uuid.UUID) (*analytics.CustomerStats, error)
	UpsertOrganizationStats(ctx context.Context, stats *analytics.OrganizationStats) error
	GetPopularServices(ctx context.Context, orgID uuid.UUID, from, to time.Time, limit int) ([]analytics.PopularService, error)
	GetHourlyHeatmap(ctx context.Context, orgID uuid.UUID, from, to time.Time) ([][]analytics.HeatmapCell, error)
	GetBusiestDays(ctx context.Context, orgID uuid.UUID, from, to time.Time) ([]analytics.DayCount, error)
	GetBusiestHours(ctx context.Context, orgID uuid.UUID, from, to time.Time) ([]analytics.HourCount, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) GetOrganizationStats(ctx context.Context, orgID uuid.UUID, from, to time.Time) ([]analytics.OrganizationStats, error) {
	var stats []analytics.OrganizationStats
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND period_date >= ? AND period_date <= ?", orgID, from, to).
		Order("period_date ASC").
		Find(&stats).Error
	return stats, err
}

func (r *gormRepository) GetLiveDashboardSummary(ctx context.Context, orgID uuid.UUID, from, to time.Time) (*analytics.LiveDashboardSummary, error) {
	var summary analytics.LiveDashboardSummary

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) FILTER (WHERE status NOT IN ('cancelled')) AS total_bookings,
			COUNT(*) FILTER (WHERE status = 'completed') AS completed_bookings,
			COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled_bookings,
			COUNT(*) FILTER (WHERE status = 'no_show') AS no_show_bookings,
			COALESCE(SUM(price) FILTER (WHERE status = 'completed'), 0) AS revenue
		FROM bookings
		WHERE organization_id = ? AND start_time >= ? AND start_time <= ?
		  AND deleted_at IS NULL
	`, orgID, from, to).Scan(&summary).Error
	if err != nil {
		return nil, err
	}

	err = r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) AS new_customers
		FROM customers
		WHERE organization_id = ? AND created_at >= ? AND created_at <= ?
		  AND deleted_at IS NULL
	`, orgID, from, to).Scan(&summary.NewCustomers).Error
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

func (r *gormRepository) GetEmployeeStats(ctx context.Context, orgID, employeeID uuid.UUID, from, to time.Time) ([]analytics.EmployeeStats, error) {
	var stats []analytics.EmployeeStats
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND employee_id = ? AND period_date >= ? AND period_date <= ?", orgID, employeeID, from, to).
		Order("period_date ASC").
		Find(&stats).Error
	return stats, err
}

func (r *gormRepository) GetCustomerStats(ctx context.Context, orgID, customerID uuid.UUID) (*analytics.CustomerStats, error) {
	var stats analytics.CustomerStats
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND customer_id = ?", orgID, customerID).
		First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *gormRepository) UpsertOrganizationStats(ctx context.Context, stats *analytics.OrganizationStats) error {
	return r.db.WithContext(ctx).
		Where("organization_id = ? AND period_date = ?", stats.OrganizationID, stats.PeriodDate).
		Assign(stats).
		FirstOrCreate(stats).Error
}

func (r *gormRepository) GetPopularServices(ctx context.Context, orgID uuid.UUID, from, to time.Time, limit int) ([]analytics.PopularService, error) {
	var results []analytics.PopularService
	err := r.db.WithContext(ctx).Raw(`
		SELECT b.service_id, s.name as service_name, s.color as color, COUNT(*) as count, SUM(b.price) as revenue
		FROM bookings b
		JOIN services s ON s.id = b.service_id
		WHERE b.organization_id = ? AND b.start_time >= ? AND b.start_time <= ?
		  AND b.status NOT IN ('cancelled') AND b.deleted_at IS NULL
		GROUP BY b.service_id, s.name, s.color
		ORDER BY count DESC
		LIMIT ?
	`, orgID, from, to, limit).Scan(&results).Error
	return results, err
}

func (r *gormRepository) GetHourlyHeatmap(ctx context.Context, orgID uuid.UUID, from, to time.Time) ([][]analytics.HeatmapCell, error) {
	type heatmapRow struct {
		Weekday int
		Slot    int
		Count   int
	}

	var rows []heatmapRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			((EXTRACT(DOW FROM start_time)::int + 6) % 7) AS weekday,
			(EXTRACT(HOUR FROM start_time)::int / 2) AS slot,
			COUNT(*)::int AS count
		FROM bookings
		WHERE organization_id = ? AND start_time >= ? AND start_time <= ?
		  AND status NOT IN ('cancelled') AND deleted_at IS NULL
		GROUP BY weekday, slot
	`, orgID, from, to).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	const weekdays = 7
	const slotsPerDay = 12
	heatmap := make([][]analytics.HeatmapCell, weekdays)
	for weekday := range heatmap {
		heatmap[weekday] = make([]analytics.HeatmapCell, slotsPerDay)
	}

	for _, row := range rows {
		if row.Weekday < 0 || row.Weekday >= weekdays || row.Slot < 0 || row.Slot >= slotsPerDay {
			continue
		}
		heatmap[row.Weekday][row.Slot].Count = row.Count
	}

	return heatmap, nil
}

func (r *gormRepository) GetBusiestDays(ctx context.Context, orgID uuid.UUID, from, to time.Time) ([]analytics.DayCount, error) {
	var results []analytics.DayCount
	err := r.db.WithContext(ctx).Raw(`
		SELECT TO_CHAR(start_time, 'Day') as day, COUNT(*) as count
		FROM bookings
		WHERE organization_id = ? AND start_time >= ? AND start_time <= ?
		  AND status NOT IN ('cancelled') AND deleted_at IS NULL
		GROUP BY TO_CHAR(start_time, 'Day'), EXTRACT(DOW FROM start_time)
		ORDER BY count DESC
	`, orgID, from, to).Scan(&results).Error
	return results, err
}

func (r *gormRepository) GetBusiestHours(ctx context.Context, orgID uuid.UUID, from, to time.Time) ([]analytics.HourCount, error) {
	var results []analytics.HourCount
	err := r.db.WithContext(ctx).Raw(`
		SELECT EXTRACT(HOUR FROM start_time)::int as hour, COUNT(*) as count
		FROM bookings
		WHERE organization_id = ? AND start_time >= ? AND start_time <= ?
		  AND status NOT IN ('cancelled') AND deleted_at IS NULL
		GROUP BY EXTRACT(HOUR FROM start_time)
		ORDER BY hour ASC
	`, orgID, from, to).Scan(&results).Error
	return results, err
}
