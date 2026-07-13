package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	redisclient "github.com/meetoria/meetoria/backend/internal/common/redis"
	"github.com/meetoria/meetoria/backend/internal/analytics"
	analyticsrepo "github.com/meetoria/meetoria/backend/internal/analytics/repository"
)

type Service interface {
	GetDashboard(ctx context.Context, orgID uuid.UUID, branchID *uuid.UUID, from, to time.Time) (*analytics.DashboardResponse, error)
	GetEmployeeAnalytics(ctx context.Context, orgID, employeeID uuid.UUID, from, to time.Time) ([]analytics.EmployeeStats, error)
	GetCustomerAnalytics(ctx context.Context, orgID, customerID uuid.UUID) (*analytics.CustomerStats, error)
}

type analyticsService struct {
	repo  analyticsrepo.Repository
	redis *redisclient.Client
}

func NewService(repo analyticsrepo.Repository, redis *redisclient.Client) Service {
	return &analyticsService{repo: repo, redis: redis}
}

func (s *analyticsService) GetDashboard(ctx context.Context, orgID uuid.UUID, branchID *uuid.UUID, from, to time.Time) (*analytics.DashboardResponse, error) {
	cacheKey := "dashboard:" + orgID.String() + ":" + from.Format("2006-01-02") + ":" + to.Format("2006-01-02")
	if branchID != nil {
		cacheKey += ":" + branchID.String()
	}

	summary, err := s.repo.GetLiveDashboardSummary(ctx, orgID, branchID, from, to)
	if err != nil {
		return nil, apperrors.Internal("failed to get dashboard summary", err)
	}

	prevFrom, prevTo := previousMonthRange(from)
	prevSummary, err := s.repo.GetLiveDashboardSummary(ctx, orgID, branchID, prevFrom, prevTo)
	if err != nil {
		return nil, apperrors.Internal("failed to get previous dashboard summary", err)
	}

	dashboard := &analytics.DashboardResponse{
		Scope:             "organization",
		TotalBookings:     summary.TotalBookings,
		CompletedBookings: summary.CompletedBookings,
		CancelledBookings: summary.CancelledBookings,
		NoShowBookings:    summary.NoShowBookings,
		Revenue:           summary.Revenue,
		NewCustomers:      summary.NewCustomers,
		Trends: analytics.DashboardTrends{
			TotalBookings:     buildMetricTrend(float64(summary.TotalBookings), float64(prevSummary.TotalBookings)),
			CompletedBookings: buildMetricTrend(float64(summary.CompletedBookings), float64(prevSummary.CompletedBookings)),
			Revenue:           buildMetricTrend(summary.Revenue, prevSummary.Revenue),
			NewCustomers:      buildMetricTrend(float64(summary.NewCustomers), float64(prevSummary.NewCustomers)),
		},
	}
	if branchID != nil {
		dashboard.Scope = "branch"
		dashboard.BranchID = branchID
	}

	popular, err := s.repo.GetPopularServices(ctx, orgID, branchID, from, to, 5)
	if err == nil {
		dashboard.PopularServices = popular
	}

	busyDays, err := s.repo.GetBusiestDays(ctx, orgID, branchID, from, to)
	if err == nil {
		dashboard.BusiestDays = busyDays
	}

	busyHours, err := s.repo.GetBusiestHours(ctx, orgID, branchID, from, to)
	if err == nil {
		dashboard.BusiestHours = busyHours
	}

	heatmap, err := s.repo.GetHourlyHeatmap(ctx, orgID, branchID, from, to)
	if err == nil {
		dashboard.HourlyHeatmap = heatmap
	}

	_ = s.redis.Set(ctx, cacheKey, "cached", 5*time.Minute)

	return dashboard, nil
}

func (s *analyticsService) GetEmployeeAnalytics(ctx context.Context, orgID, employeeID uuid.UUID, from, to time.Time) ([]analytics.EmployeeStats, error) {
	stats, err := s.repo.GetEmployeeStats(ctx, orgID, employeeID, from, to)
	if err != nil {
		return nil, apperrors.Internal("failed to get employee stats", err)
	}
	return stats, nil
}

func (s *analyticsService) GetCustomerAnalytics(ctx context.Context, orgID, customerID uuid.UUID) (*analytics.CustomerStats, error) {
	stats, err := s.repo.GetCustomerStats(ctx, orgID, customerID)
	if err != nil {
		return nil, apperrors.NotFound("customer analytics not found")
	}
	return stats, nil
}

func previousMonthRange(from time.Time) (time.Time, time.Time) {
	prevFrom := time.Date(from.Year(), from.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	prevTo := time.Date(from.Year(), from.Month(), 0, 23, 59, 59, 999999999, time.UTC)
	return prevFrom, prevTo
}

func buildMetricTrend(current, previous float64) analytics.MetricTrend {
	change := current - previous
	trend := analytics.MetricTrend{
		Previous: previous,
		Change:   change,
	}
	if previous != 0 {
		pct := (change / previous) * 100
		trend.ChangePct = &pct
	}
	return trend
}
