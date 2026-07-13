package model

type DeletionCheck struct {
	CanDelete      bool   `json:"can_delete"`
	BookingsCount  int64  `json:"bookings_count"`
	EmployeesCount int64  `json:"employees_count,omitempty"`
	Message        string `json:"message,omitempty"`
}
