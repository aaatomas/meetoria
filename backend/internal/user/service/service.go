package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	"github.com/meetoria/meetoria/backend/internal/user"
	"github.com/meetoria/meetoria/backend/internal/user/repository"
)

type UserContext struct {
	ID         uuid.UUID
	KeycloakID uuid.UUID
	Email      string
	FirstName  string
	LastName   string
}

type Service interface {
	GetOrCreateByKeycloak(ctx context.Context, keycloakID uuid.UUID, email string) (*UserContext, error)
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	Update(ctx context.Context, id uuid.UUID, req user.UpdateUserRequest) (*user.User, error)
	SyncFromKeycloak(ctx context.Context, keycloakID uuid.UUID, req user.SyncUserRequest) (*user.User, error)
}

type userService struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &userService{repo: repo}
}

func (s *userService) GetOrCreateByKeycloak(ctx context.Context, keycloakID uuid.UUID, email string) (*UserContext, error) {
	u, err := s.repo.GetByKeycloakID(ctx, keycloakID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u = &user.User{
				KeycloakID: keycloakID,
				Email:      email,
				FirstName:  "User",
				LastName:   "",
			}
			if err := s.repo.Create(ctx, u); err != nil {
				return nil, apperrors.Internal("failed to create user", err)
			}
		} else {
			return nil, apperrors.Internal("failed to get user", err)
		}
	}

	return &UserContext{
		ID:         u.ID,
		KeycloakID: u.KeycloakID,
		Email:      u.Email,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
	}, nil
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("user not found")
		}
		return nil, apperrors.Internal("failed to get user", err)
	}
	return u, nil
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, req user.UpdateUserRequest) (*user.User, error) {
	u, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Phone != nil {
		u.Phone = *req.Phone
	}
	if req.FirstName != nil {
		u.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		u.LastName = *req.LastName
	}

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, apperrors.Internal("failed to update user", err)
	}
	return u, nil
}

func (s *userService) SyncFromKeycloak(ctx context.Context, keycloakID uuid.UUID, req user.SyncUserRequest) (*user.User, error) {
	u, err := s.repo.GetByKeycloakID(ctx, keycloakID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u = &user.User{
				KeycloakID: keycloakID,
				Email:      req.Email,
				FirstName:  req.FirstName,
				LastName:   req.LastName,
				Phone:      req.Phone,
			}
			if err := s.repo.Create(ctx, u); err != nil {
				return nil, apperrors.Internal("failed to create user", err)
			}
			return u, nil
		}
		return nil, apperrors.Internal("failed to get user", err)
	}

	u.Email = req.Email
	u.FirstName = req.FirstName
	u.LastName = req.LastName
	if req.Phone != "" {
		u.Phone = req.Phone
	}

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, apperrors.Internal("failed to sync user", err)
	}
	return u, nil
}
