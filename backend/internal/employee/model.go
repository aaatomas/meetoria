package employee

import (
	"github.com/google/uuid"

	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
)

type Employee struct {
	commonmodel.OrganizationScoped
	BranchID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"branch_id"`
	UserID    *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	FirstName string     `gorm:"size:100;not null" json:"first_name"`
	LastName  string     `gorm:"size:100;not null" json:"last_name"`
	Email     string     `gorm:"size:255" json:"email"`
	Phone     string     `gorm:"size:20" json:"phone"`
	Title     string     `gorm:"size:100" json:"title"`
	Bio       string     `json:"bio"`
	AvatarURL string     `json:"avatar_url"`
	Color     string     `gorm:"size:7;default:#1976d2" json:"color"`
	IsActive  bool       `gorm:"not null;default:true" json:"is_active"`
}

type CreateEmployeeRequest struct {
	BranchID   *uuid.UUID  `json:"branch_id"`
	FirstName  string      `json:"first_name" binding:"required,min=1,max=100"`
	LastName   string      `json:"last_name" binding:"required,min=1,max=100"`
	Email      string      `json:"email" binding:"omitempty,email"`
	Phone      string      `json:"phone" binding:"omitempty,e164"`
	Title      string      `json:"title"`
	Bio        string      `json:"bio"`
	UserID     *uuid.UUID  `json:"user_id"`
	ServiceIDs []uuid.UUID `json:"service_ids"`
}

type UpdateEmployeeRequest struct {
	BranchID  *uuid.UUID `json:"branch_id"`
	FirstName *string    `json:"first_name" binding:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name" binding:"omitempty,min=1,max=100"`
	Email     *string `json:"email" binding:"omitempty,email"`
	Phone     *string `json:"phone" binding:"omitempty,e164"`
	Title     *string `json:"title"`
	Bio       *string `json:"bio"`
	Color     *string `json:"color"`
	IsActive  *bool   `json:"is_active"`
}

type EmployeeService struct {
	commonmodel.OrganizationJunction
	EmployeeID uuid.UUID `gorm:"type:uuid;not null" json:"employee_id"`
	ServiceID  uuid.UUID `gorm:"type:uuid;not null" json:"service_id"`
}

func (EmployeeService) TableName() string {
	return "employee_services"
}

type PublicEmployee struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Title     string    `json:"title,omitempty"`
	AvatarURL string    `json:"avatar_url,omitempty"`
}
