package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/backend/internal/notification"
)

type Repository interface {
	Create(ctx context.Context, n *notification.Notification) error
	Update(ctx context.Context, n *notification.Notification) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*notification.Notification, error)
	GetByMessageID(ctx context.Context, orgID, messageID uuid.UUID) (*notification.Notification, error)
	ListByBooking(ctx context.Context, orgID, bookingID uuid.UUID) ([]notification.Notification, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, n *notification.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *gormRepository) Update(ctx context.Context, n *notification.Notification) error {
	return r.db.WithContext(ctx).Save(n).Error
}

func (r *gormRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (*notification.Notification, error) {
	var n notification.Notification
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		First(&n).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *gormRepository) GetByMessageID(ctx context.Context, orgID, messageID uuid.UUID) (*notification.Notification, error) {
	var n notification.Notification
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND message_id = ?", orgID, messageID).
		First(&n).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *gormRepository) ListByBooking(ctx context.Context, orgID, bookingID uuid.UUID) ([]notification.Notification, error) {
	var notifications []notification.Notification
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND booking_id = ?", orgID, bookingID).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}
