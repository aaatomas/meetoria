package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/backend/internal/user"
)

type Repository interface {
	Create(ctx context.Context, u *user.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	GetByKeycloakID(ctx context.Context, keycloakID uuid.UUID) (*user.User, error)
	Update(ctx context.Context, u *user.User) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *gormRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *gormRepository) GetByKeycloakID(ctx context.Context, keycloakID uuid.UUID) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).First(&u, "keycloak_id = ?", keycloakID).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *gormRepository) Update(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}
