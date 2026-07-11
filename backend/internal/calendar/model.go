package calendar

import (
	"time"

	"github.com/google/uuid"
)

// SyncRequest represents a calendar sync event payload published to RabbitMQ.
type SyncRequest struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	EmployeeID     uuid.UUID `json:"employee_id"`
	BookingID      uuid.UUID `json:"booking_id"`
	Action         string    `json:"action"` // create, update, delete
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
}

// Supported external calendar providers for future integration.
const (
	ProviderGoogle  = "google"
	ProviderOutlook = "outlook"
	ProviderApple   = "apple"
)
