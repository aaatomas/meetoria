package schedule

import (
	"time"

	"github.com/google/uuid"

	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
)

type WorkingHours struct {
	commonmodel.OrganizationScoped
	EmployeeID *uuid.UUID `gorm:"type:uuid;index" json:"employee_id,omitempty"`
	DayOfWeek  int        `gorm:"not null" json:"day_of_week"`
	StartTime  time.Time  `gorm:"type:time;not null" json:"start_time"`
	EndTime    time.Time  `gorm:"type:time;not null" json:"end_time"`
	IsActive   bool       `gorm:"not null;default:true" json:"is_active"`
}

type Break struct {
	commonmodel.OrganizationScoped
	EmployeeID *uuid.UUID `gorm:"type:uuid;index" json:"employee_id,omitempty"`
	DayOfWeek  int        `gorm:"not null" json:"day_of_week"`
	StartTime  time.Time  `gorm:"type:time;not null" json:"start_time"`
	EndTime    time.Time  `gorm:"type:time;not null" json:"end_time"`
}

type Holiday struct {
	commonmodel.OrganizationScoped
	EmployeeID  *uuid.UUID `gorm:"type:uuid;index" json:"employee_id,omitempty"`
	Name        string     `gorm:"size:255;not null" json:"name"`
	Date        time.Time  `gorm:"type:date;not null" json:"date"`
	IsRecurring bool       `gorm:"not null;default:false" json:"is_recurring"`
}

type SetWorkingHoursRequest struct {
	EmployeeID *uuid.UUID       `json:"employee_id"`
	Schedule   []DaySchedule    `json:"schedule" binding:"required,dive"`
}

type DaySchedule struct {
	DayOfWeek int         `json:"day_of_week" binding:"min=0,max=6"`
	Slots     []TimeRange `json:"slots" binding:"required,dive"`
}

type TimeRange struct {
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

type CreateHolidayRequest struct {
	EmployeeID  *uuid.UUID `json:"employee_id"`
	Name        string     `json:"name" binding:"required"`
	Date        string     `json:"date" binding:"required"`
	IsRecurring bool       `json:"is_recurring"`
}
