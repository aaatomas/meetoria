package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/workers/sms-worker/internal/provider"
)

const (
	exchangeName = "meetoria.events"
	queueName    = "sms.notifications"
	routingKey   = "notification.sms"
)

type SMSMessage struct {
	Event          string            `json:"event"`
	MessageID      uuid.UUID         `json:"message_id"`
	CorrelationID  uuid.UUID         `json:"correlation_id"`
	OrganizationID uuid.UUID         `json:"organization_id"`
	BookingID      uuid.UUID         `json:"booking_id"`
	Recipient      struct {
		Phone string `json:"phone"`
	} `json:"recipient"`
	Template  string            `json:"template"`
	Variables map[string]string `json:"variables"`
	Timestamp time.Time         `json:"timestamp"`
	Source    string            `json:"source"`
}

type SMSRecord struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	MessageID         uuid.UUID `gorm:"type:uuid;uniqueIndex"`
	CorrelationID     uuid.UUID `gorm:"type:uuid;index"`
	OrganizationID    uuid.UUID `gorm:"type:uuid"`
	BookingID         uuid.UUID `gorm:"type:uuid"`
	RecipientPhone    string
	Template          string
	Variables         string `gorm:"type:jsonb"`
	Provider          string
	ProviderMessageID string
	Status            string
	RetryCount        int
	SentAt            *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (SMSRecord) TableName() string {
	return "sms_messages"
}

type Consumer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	db       *gorm.DB
	provider provider.Provider
}

func NewConsumer(rabbitURL string, db *gorm.DB, smsProvider provider.Provider) (*Consumer, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil); err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	if err := ch.QueueBind(q.Name, routingKey, exchangeName, false, nil); err != nil {
		return nil, err
	}

	if err := ch.Qos(10, 0, false); err != nil {
		return nil, err
	}

	return &Consumer{conn: conn, channel: ch, db: db, provider: smsProvider}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(queueName, "sms-worker", false, false, false, false, nil)
	if err != nil {
		return err
	}

	slog.Info("sms worker started", "queue", queueName)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}
			if err := c.handleMessage(ctx, msg); err != nil {
				slog.Error("handle message failed", "error", err)
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
			}
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event SMSMessage
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return err
	}

	slog.Info("processing sms",
		"message_id", event.MessageID,
		"correlation_id", event.CorrelationID,
		"phone", event.Recipient.Phone,
	)

	variablesJSON, _ := json.Marshal(event.Variables)
	record := SMSRecord{
		ID:             uuid.New(),
		MessageID:      event.MessageID,
		CorrelationID:  event.CorrelationID,
		OrganizationID: event.OrganizationID,
		BookingID:      event.BookingID,
		RecipientPhone: event.Recipient.Phone,
		Template:       event.Template,
		Variables:      string(variablesJSON),
		Provider:       c.provider.Name(),
		Status:         "processing",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := c.db.WithContext(ctx).Create(&record).Error; err != nil {
		return err
	}

	body := provider.RenderTemplate(event.Template, event.Variables)
	result, err := c.provider.Send(ctx, event.Recipient.Phone, body)
	if err != nil {
		record.Status = "failed"
		record.RetryCount++
		c.db.Save(&record)
		return err
	}

	now := time.Now().UTC()
	record.Status = "sent"
	record.ProviderMessageID = result.ProviderMessageID
	record.SentAt = &now
	return c.db.WithContext(ctx).Save(&record).Error
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
