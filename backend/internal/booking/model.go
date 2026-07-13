package booking

import (
	"time"

	"github.com/google/uuid"

	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
)

type BookingStatus string

const (
	StatusPending    BookingStatus = "pending"
	StatusConfirmed  BookingStatus = "confirmed"
	StatusInProgress BookingStatus = "in_progress"
	StatusCompleted  BookingStatus = "completed"
	StatusCancelled  BookingStatus = "cancelled"
	StatusNoShow     BookingStatus = "no_show"
)

type Booking struct {
	commonmodel.OrganizationScoped
	BranchID           uuid.UUID     `gorm:"type:uuid;not null;index" json:"branch_id"`
	CustomerID         uuid.UUID     `gorm:"type:uuid;not null;index" json:"customer_id"`
	EmployeeID         uuid.UUID     `gorm:"type:uuid;not null;index" json:"employee_id"`
	ServiceID          uuid.UUID     `gorm:"type:uuid;not null;index" json:"service_id"`
	StartTime          time.Time     `gorm:"not null;index" json:"start_time"`
	EndTime            time.Time     `gorm:"not null" json:"end_time"`
	Status             BookingStatus `gorm:"type:booking_status;not null;default:pending" json:"status"`
	Notes              string        `json:"notes"`
	CancellationReason string        `json:"cancellation_reason,omitempty"`
	Price              float64       `gorm:"type:decimal(10,2);not null" json:"price"`
	Currency           string        `gorm:"size:3;not null;default:EUR" json:"currency"`
}

type CreateBookingRequest struct {
	BranchID   *uuid.UUID `json:"branch_id"`
	CustomerID uuid.UUID  `json:"customer_id" binding:"required"`
	EmployeeID uuid.UUID `json:"employee_id" binding:"required"`
	ServiceID  uuid.UUID `json:"service_id" binding:"required"`
	StartTime  time.Time `json:"start_time" binding:"required"`
	Notes      string    `json:"notes"`
}

type UpdateBookingRequest struct {
	CustomerID *uuid.UUID     `json:"customer_id"`
	StartTime  *time.Time     `json:"start_time"`
	EmployeeID *uuid.UUID     `json:"employee_id"`
	ServiceID  *uuid.UUID     `json:"service_id"`
	Notes      *string        `json:"notes"`
	Status     *BookingStatus `json:"status"`
}

type CancelBookingRequest struct {
	Reason string `json:"reason" binding:"required,min=1"`
}

type AvailabilityRequest struct {
	BranchID   *uuid.UUID `form:"branch_id"`
	EmployeeID uuid.UUID  `form:"employee_id" binding:"required"`
	ServiceID  uuid.UUID `form:"service_id" binding:"required"`
	Date       string    `form:"date" binding:"required"`
}

type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Available bool      `json:"available"`
}
