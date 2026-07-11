package customer

import (
	"github.com/google/uuid"

	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
)

type Customer struct {
	commonmodel.OrganizationScoped
	UserID    *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	FirstName string     `gorm:"size:100;not null" json:"first_name"`
	LastName  string     `gorm:"size:100;not null" json:"last_name"`
	Email     string     `gorm:"size:255" json:"email"`
	Phone     string     `gorm:"size:20" json:"phone"`
	Notes     string     `json:"notes"`
}

type CreateCustomerRequest struct {
	FirstName string     `json:"first_name" binding:"required,min=1,max=100"`
	LastName  string     `json:"last_name" binding:"required,min=1,max=100"`
	Email     string     `json:"email" binding:"omitempty,email"`
	Phone     string     `json:"phone" binding:"omitempty,e164"`
	Notes     string     `json:"notes"`
	UserID    *uuid.UUID `json:"user_id"`
}

type UpdateCustomerRequest struct {
	FirstName *string `json:"first_name" binding:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name" binding:"omitempty,min=1,max=100"`
	Email     *string `json:"email" binding:"omitempty,email"`
	Phone     *string `json:"phone" binding:"omitempty,e164"`
	Notes     *string `json:"notes"`
}

type ListItem struct {
	Customer
	BookingsCount      int64 `json:"bookings_count"`
	CancellationsCount int64 `json:"cancellations_count"`
}

type BookingStats struct {
	BookingsCount      int64
	CancellationsCount int64
}
