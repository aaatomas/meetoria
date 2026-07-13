package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/backend/internal/common/rabbitmq"
	"github.com/meetoria/meetoria/backend/internal/notification"
	notifrepo "github.com/meetoria/meetoria/backend/internal/notification/repository"
	"github.com/meetoria/meetoria/backend/pkg/events"
)

type Handler struct {
	repo notifrepo.Repository
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{repo: notifrepo.NewRepository(db)}
}

func (h *Handler) Handle(ctx context.Context, routingKey string, body []byte) error {
	switch routingKey {
	case rabbitmq.RoutingNotificationSMSProcessing, rabbitmq.RoutingNotificationSMSSent, rabbitmq.RoutingNotificationSMSFailed,
		rabbitmq.RoutingNotificationEmailProcessing, rabbitmq.RoutingNotificationEmailSent, rabbitmq.RoutingNotificationEmailFailed:
		return h.handleNotificationResult(ctx, routingKey, body)
	default:
		slog.Warn("unhandled rabbitmq message", "routing_key", routingKey)
		return nil
	}
}

func (h *Handler) handleNotificationResult(ctx context.Context, routingKey string, body []byte) error {
	var event events.NotificationResultEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("unmarshal notification result: %w", err)
	}

	channel, status, err := resultStatusForRoutingKey(routingKey)
	if err != nil {
		return err
	}

	if status == "" {
		slog.Info("notification processing",
			"organization_id", event.OrganizationID,
			"booking_id", event.BookingID,
			"channel", event.Channel,
			"routing_key", routingKey,
			"correlation_id", event.CorrelationID,
		)
		return nil
	}

	notifications, err := h.repo.ListByBooking(ctx, event.OrganizationID, event.BookingID)
	if err != nil {
		return err
	}

	var target *notification.Notification
	if event.MessageID != uuid.Nil {
		if byMessage, err := h.repo.GetByMessageID(ctx, event.OrganizationID, event.MessageID); err == nil {
			target = byMessage
		}
	}

	if target == nil {
		for i := range notifications {
			n := &notifications[i]
			if n.Channel == channel && (n.Status == notification.StatusQueued || n.Status == notification.StatusCreated) {
				target = n
				break
			}
		}
	}

	if target == nil {
		slog.Warn("notification not found for worker result",
			"organization_id", event.OrganizationID,
			"booking_id", event.BookingID,
			"channel", channel,
			"routing_key", routingKey,
		)
		return nil
	}

	target.Status = status
	now := time.Now().UTC()
	if status == notification.StatusSent {
		target.SentAt = &now
	}

	if err := h.repo.Update(ctx, target); err != nil {
		return err
	}

	slog.Info("notification status updated from worker",
		"notification_id", target.ID,
		"status", status,
		"routing_key", routingKey,
		"correlation_id", event.CorrelationID,
	)

	return nil
}

func resultStatusForRoutingKey(routingKey string) (notification.Channel, notification.Status, error) {
	switch routingKey {
	case rabbitmq.RoutingNotificationSMSProcessing, rabbitmq.RoutingNotificationEmailProcessing:
		return "", "", nil
	case rabbitmq.RoutingNotificationSMSSent:
		return notification.ChannelSMS, notification.StatusSent, nil
	case rabbitmq.RoutingNotificationSMSFailed:
		return notification.ChannelSMS, notification.StatusFailed, nil
	case rabbitmq.RoutingNotificationEmailSent:
		return notification.ChannelEmail, notification.StatusSent, nil
	case rabbitmq.RoutingNotificationEmailFailed:
		return notification.ChannelEmail, notification.StatusFailed, nil
	default:
		return "", "", fmt.Errorf("unsupported routing key: %s", routingKey)
	}
}
