package notification

import (
	"time"

	"github.com/google/uuid"
)

type Channel string

const (
	ChannelSMS   Channel = "sms"
	ChannelEmail Channel = "email"
)

type Status string

const (
	StatusCreated   Status = "created"
	StatusQueued    Status = "queued"
	StatusSent      Status = "sent"
	StatusDelivered Status = "delivered"
	StatusFailed    Status = "failed"
)

type Notification struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	BookingID      *uuid.UUID `gorm:"type:uuid;index" json:"booking_id,omitempty"`
	MessageID      *uuid.UUID `gorm:"type:uuid;index" json:"message_id,omitempty"`
	Channel        Channel    `gorm:"type:notification_channel;not null" json:"channel"`
	Template       string     `gorm:"size:100;not null" json:"template"`
	Recipient      string     `gorm:"size:255;not null" json:"recipient"`
	Status         Status     `gorm:"type:notification_status;not null;default:created" json:"status"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
