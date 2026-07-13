package branch

import (
	"github.com/google/uuid"

	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
)

type Branch struct {
	commonmodel.OrganizationScoped
	Name      string `gorm:"size:255;not null" json:"name"`
	Address   string `json:"address,omitempty"`
	City      string `gorm:"size:100" json:"city,omitempty"`
	Country   string `gorm:"size:100" json:"country,omitempty"`
	Timezone  string `gorm:"size:50" json:"timezone,omitempty"`
	Phone     string `gorm:"size:20" json:"phone,omitempty"`
	Email     string `gorm:"size:255" json:"email,omitempty"`
	IsActive  bool   `gorm:"not null;default:true" json:"is_active"`
	IsDefault bool   `gorm:"not null;default:false" json:"is_default"`
}

type BranchService struct {
	commonmodel.OrganizationJunction
	BranchID  uuid.UUID `gorm:"type:uuid;not null;index" json:"branch_id"`
	ServiceID uuid.UUID `gorm:"type:uuid;not null;index" json:"service_id"`
}

func (BranchService) TableName() string {
	return "branch_services"
}

type CreateBranchRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=255"`
	Address  string `json:"address"`
	City     string `json:"city"`
	Country  string `json:"country"`
	Timezone string `json:"timezone"`
	Phone    string `json:"phone" binding:"omitempty,e164"`
	Email    string `json:"email" binding:"omitempty,email"`
}

type UpdateBranchRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=1,max=255"`
	Address   *string `json:"address"`
	City      *string `json:"city"`
	Country   *string `json:"country"`
	Timezone  *string `json:"timezone"`
	Phone     *string `json:"phone" binding:"omitempty,e164"`
	Email     *string `json:"email" binding:"omitempty,email"`
	IsActive  *bool   `json:"is_active"`
	ServiceIDs []uuid.UUID `json:"service_ids"`
}

type PublicBranch struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Address string    `json:"address,omitempty"`
	City    string    `json:"city,omitempty"`
	Country string    `json:"country,omitempty"`
	Phone   string    `json:"phone,omitempty"`
}
