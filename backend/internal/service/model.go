package service

import (
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
)

type Service struct {
	commonmodel.OrganizationScoped
	Name            string  `gorm:"size:255;not null" json:"name"`
	Description     string  `json:"description"`
	DurationMinutes int     `gorm:"not null;default:30" json:"duration_minutes"`
	Price           float64 `gorm:"type:decimal(10,2);not null;default:0" json:"price"`
	Currency        string  `gorm:"size:3;not null;default:EUR" json:"currency"`
	Category        string  `gorm:"size:100" json:"category"`
	Color           string  `gorm:"size:20;not null;default:teal" json:"color"`
	IsActive        bool    `gorm:"not null;default:true" json:"is_active"`
}

type CreateServiceRequest struct {
	Name            string  `json:"name" binding:"required,min=1,max=255"`
	Description     string  `json:"description"`
	DurationMinutes int     `json:"duration_minutes" binding:"required,min=5,max=480"`
	Price           float64 `json:"price" binding:"min=0"`
	Currency        string  `json:"currency" binding:"required,len=3"`
	Category        string  `json:"category"`
	Color           string  `json:"color"`
}

type UpdateServiceRequest struct {
	Name            *string  `json:"name" binding:"omitempty,min=1,max=255"`
	Description     *string  `json:"description"`
	DurationMinutes *int     `json:"duration_minutes" binding:"omitempty,min=5,max=480"`
	Price           *float64 `json:"price" binding:"omitempty,min=0"`
	Category        *string  `json:"category"`
	Color           *string  `json:"color"`
	IsActive        *bool    `json:"is_active"`
}
