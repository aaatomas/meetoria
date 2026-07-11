package organization

import (
	"github.com/google/uuid"

	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
)

type OrganizationRole string

const (
	RoleOrganizationOwner OrganizationRole = "organization_owner"
	RoleManager           OrganizationRole = "manager"
	RoleEmployee          OrganizationRole = "employee"
	RoleCustomer          OrganizationRole = "customer"
)

type Organization struct {
	commonmodel.BaseModel
	Name     string `gorm:"size:255;not null" json:"name"`
	Slug     string `gorm:"size:100;not null;uniqueIndex" json:"slug"`
	Timezone string `gorm:"size:50;not null;default:UTC" json:"timezone"`
	Email    string `gorm:"size:255" json:"email"`
	Phone    string `gorm:"size:20" json:"phone"`
	Address  string `json:"address"`
	LogoURL  string `json:"logo_url"`
	Settings string `gorm:"type:jsonb;default:'{}'" json:"settings"`
	IsActive bool   `gorm:"not null;default:true" json:"is_active"`
}

type OrganizationUser struct {
	commonmodel.BaseModel
	OrganizationID uuid.UUID        `gorm:"type:uuid;not null;index" json:"organization_id"`
	UserID         uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id"`
	Role           OrganizationRole `gorm:"type:organization_role;not null;default:employee" json:"role"`
	IsActive       bool             `gorm:"not null;default:true" json:"is_active"`
}

func (OrganizationUser) TableName() string {
	return "organization_users"
}

type CreateOrganizationRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=255"`
	Slug     string `json:"slug" binding:"required,min=2,max=100"`
	Timezone string `json:"timezone" binding:"required"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,e164"`
}

type UpdateOrganizationRequest struct {
	Name     *string `json:"name" binding:"omitempty,min=2,max=255"`
	Timezone *string `json:"timezone"`
	Email    *string `json:"email" binding:"omitempty,email"`
	Phone    *string `json:"phone" binding:"omitempty,e164"`
	Address  *string `json:"address"`
	LogoURL  *string `json:"logo_url"`
}
