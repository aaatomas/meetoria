package user

import (
	"github.com/google/uuid"

	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
)

type User struct {
	commonmodel.BaseModel
	KeycloakID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"keycloak_id"`
	Email      string    `gorm:"size:255;not null" json:"email"`
	Phone      string    `gorm:"size:20" json:"phone"`
	FirstName  string    `gorm:"size:100;not null" json:"first_name"`
	LastName   string    `gorm:"size:100;not null" json:"last_name"`
}

type UpdateUserRequest struct {
	Phone     *string `json:"phone" binding:"omitempty,e164"`
	FirstName *string `json:"first_name" binding:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name" binding:"omitempty,min=1,max=100"`
}

type SyncUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required,min=1,max=100"`
	LastName  string `json:"last_name" binding:"required,min=1,max=100"`
	Phone     string `json:"phone" binding:"omitempty,e164"`
}
