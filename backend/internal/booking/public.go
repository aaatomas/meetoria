package booking

import (
	"time"

	"github.com/google/uuid"
)

type PublicCustomerInfo struct {
	FirstName string `json:"first_name" binding:"required,min=1,max=100"`
	LastName  string `json:"last_name" binding:"required,min=1,max=100"`
	Phone     string `json:"phone" binding:"required"`
	Email     string `json:"email" binding:"omitempty,email"`
}

type PublicCreateBookingRequest struct {
	BranchID   uuid.UUID          `json:"branch_id" binding:"required"`
	ServiceID  uuid.UUID          `json:"service_id" binding:"required"`
	EmployeeID *uuid.UUID         `json:"employee_id"`
	StartTime  time.Time          `json:"start_time" binding:"required"`
	Notes      string             `json:"notes"`
	Customer   PublicCustomerInfo `json:"customer" binding:"required"`
}

type PublicAvailabilityRequest struct {
	BranchID   string `form:"branch_id" binding:"required"`
	ServiceID  string `form:"service_id" binding:"required"`
	EmployeeID string `form:"employee_id"`
	Date       string `form:"date" binding:"required"`
}

func (r PublicAvailabilityRequest) ParsedBranchID() (uuid.UUID, error) {
	return uuid.Parse(r.BranchID)
}

func (r PublicAvailabilityRequest) ParsedServiceID() (uuid.UUID, error) {
	return uuid.Parse(r.ServiceID)
}

func (r PublicAvailabilityRequest) ParsedEmployeeID() (*uuid.UUID, error) {
	if r.EmployeeID == "" {
		return nil, nil
	}
	id, err := uuid.Parse(r.EmployeeID)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

type PublicTimeSlot struct {
	StartTime   time.Time   `json:"start_time"`
	EndTime     time.Time   `json:"end_time"`
	Available   bool        `json:"available"`
	EmployeeIDs []uuid.UUID `json:"employee_ids,omitempty"`
}

type PublicOrganizationProfile struct {
	Name               string `json:"name"`
	Slug               string `json:"slug"`
	Timezone           string `json:"timezone"`
	LogoURL            string `json:"logo_url,omitempty"`
	Address            string `json:"address,omitempty"`
	Phone              string `json:"phone,omitempty"`
	Email              string `json:"email,omitempty"`
	CancellationPolicy string `json:"cancellation_policy,omitempty"`
	ReschedulingPolicy string `json:"rescheduling_policy,omitempty"`
	EmailRequired      bool   `json:"email_required"`
	Currency           string `json:"currency"`
	TimeFormat         string `json:"time_format"`
}
