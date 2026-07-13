package model

type DeletionCheck struct {
	CanDelete     bool   `json:"can_delete"`
	BookingsCount int64  `json:"bookings_count"`
	Message       string `json:"message,omitempty"`
}
