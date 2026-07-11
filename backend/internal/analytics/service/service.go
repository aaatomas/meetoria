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
	GetDashboard(ctx context.Context, orgID uuid.UUID, from, to time.Time) (*analytics.DashboardResponse, error)
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

func (s *analyticsService) GetDashboard(ctx context.Context, orgID uuid.UUID, from, to time.Time) (*analytics.DashboardResponse, error) {
	cacheKey := "dashboard:" + orgID.String() + ":" + from.Format("2006-01-02") + ":" + to.Format("2006-01-02")

	summary, err := s.repo.GetLiveDashboardSummary(ctx, orgID, from, to)
	if err != nil {
		return nil, apperrors.Internal("failed to get dashboard summary", err)
	}

	dashboard := &analytics.DashboardResponse{
		TotalBookings:     summary.TotalBookings,
		CompletedBookings: summary.CompletedBookings,
		CancelledBookings: summary.CancelledBookings,
		NoShowBookings:    summary.NoShowBookings,
		Revenue:           summary.Revenue,
		NewCustomers:      summary.NewCustomers,
	}

	popular, err := s.repo.GetPopularServices(ctx, orgID, from, to, 5)
	if err == nil {
		dashboard.PopularServices = popular
	}

	busyDays, err := s.repo.GetBusiestDays(ctx, orgID, from, to)
	if err == nil {
		dashboard.BusiestDays = busyDays
	}

	busyHours, err := s.repo.GetBusiestHours(ctx, orgID, from, to)
	if err == nil {
		dashboard.BusiestHours = busyHours
	}

	heatmap, err := s.repo.GetHourlyHeatmap(ctx, orgID, from, to)
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
