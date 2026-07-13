package analytics

import (
	"time"

	"github.com/google/uuid"

	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
)

type OrganizationStats struct {
	commonmodel.BaseModel
	OrganizationID    uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	PeriodDate        time.Time `gorm:"type:date;not null" json:"period_date"`
	TotalBookings     int       `gorm:"not null;default:0" json:"total_bookings"`
	CompletedBookings int       `gorm:"not null;default:0" json:"completed_bookings"`
	CancelledBookings int       `gorm:"not null;default:0" json:"cancelled_bookings"`
	NoShowBookings    int       `gorm:"not null;default:0" json:"no_show_bookings"`
	Revenue           float64   `gorm:"type:decimal(12,2);not null;default:0" json:"revenue"`
	NewCustomers      int       `gorm:"not null;default:0" json:"new_customers"`
}

func (OrganizationStats) TableName() string {
	return "analytics_organization_stats"
}

type EmployeeStats struct {
	commonmodel.BaseModel
	OrganizationID       uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	EmployeeID           uuid.UUID `gorm:"type:uuid;not null;index" json:"employee_id"`
	PeriodDate           time.Time `gorm:"type:date;not null" json:"period_date"`
	TotalAppointments    int       `gorm:"not null;default:0" json:"total_appointments"`
	CompletedAppointments int      `gorm:"not null;default:0" json:"completed_appointments"`
	Revenue              float64   `gorm:"type:decimal(12,2);not null;default:0" json:"revenue"`
	UtilizationPercent   float64   `gorm:"type:decimal(5,2);not null;default:0" json:"utilization_percent"`
}

func (EmployeeStats) TableName() string {
	return "analytics_employee_stats"
}

type CustomerStats struct {
	commonmodel.BaseModel
	OrganizationID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	CustomerID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"customer_id"`
	VisitCount         int        `gorm:"not null;default:0" json:"visit_count"`
	LastVisit          *time.Time `json:"last_visit,omitempty"`
	TotalSpending      float64    `gorm:"type:decimal(12,2);not null;default:0" json:"total_spending"`
	CancellationCount  int        `gorm:"not null;default:0" json:"cancellation_count"`
	NoShowCount        int        `gorm:"not null;default:0" json:"no_show_count"`
	FavoriteEmployeeID *uuid.UUID `gorm:"type:uuid" json:"favorite_employee_id,omitempty"`
	FavoriteServiceID  *uuid.UUID `gorm:"type:uuid" json:"favorite_service_id,omitempty"`
}

func (CustomerStats) TableName() string {
	return "analytics_customer_stats"
}

type DashboardResponse struct {
	Scope             string             `json:"scope"`
	BranchID          *uuid.UUID         `json:"branch_id,omitempty"`
	BranchName        string             `json:"branch_name,omitempty"`
	TotalBookings     int                `json:"total_bookings"`
	CompletedBookings int                `json:"completed_bookings"`
	CancelledBookings int                `json:"cancelled_bookings"`
	NoShowBookings    int                `json:"no_show_bookings"`
	Revenue           float64            `json:"revenue"`
	NewCustomers      int                `json:"new_customers"`
	Trends            DashboardTrends    `json:"trends"`
	PopularServices   []PopularService   `json:"popular_services"`
	BusiestDays       []DayCount         `json:"busiest_days"`
	BusiestHours      []HourCount        `json:"busiest_hours"`
	HourlyHeatmap     [][]HeatmapCell    `json:"hourly_heatmap"`
}

type DashboardTrends struct {
	TotalBookings     MetricTrend `json:"total_bookings"`
	CompletedBookings MetricTrend `json:"completed_bookings"`
	Revenue           MetricTrend `json:"revenue"`
	NewCustomers      MetricTrend `json:"new_customers"`
}

type MetricTrend struct {
	Previous  float64  `json:"previous"`
	Change    float64  `json:"change"`
	ChangePct *float64 `json:"change_pct"`
}

type LiveDashboardSummary struct {
	TotalBookings     int     `gorm:"column:total_bookings"`
	CompletedBookings int     `gorm:"column:completed_bookings"`
	CancelledBookings int     `gorm:"column:cancelled_bookings"`
	NoShowBookings    int     `gorm:"column:no_show_bookings"`
	Revenue           float64 `gorm:"column:revenue"`
	NewCustomers      int     `gorm:"column:new_customers"`
}

type PopularService struct {
	ServiceID   uuid.UUID `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Color       string    `json:"color,omitempty"`
	Count       int       `json:"count"`
	Revenue     float64   `json:"revenue"`
}

type HeatmapCell struct {
	Count int `json:"count"`
}

type DayCount struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

type HourCount struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

type DateRangeQuery struct {
	From string `form:"from" binding:"required"`
	To   string `form:"to" binding:"required"`
}
