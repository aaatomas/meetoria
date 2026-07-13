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

	"github.com/meetoria/meetoria/workers/email-worker/internal/provider"
)

const (
	backendExchangeName = "meetoria-backend"
	workerExchangeName  = "meetoria-email-worker"
	workerQueueName     = "meetoria-email-worker"
	inboundRoutingKey     = "notification.email"
	processingRoutingKey  = "notification.email.processing"
	sentRoutingKey        = "notification.email.sent"
	failedRoutingKey      = "notification.email.failed"
	exchangeType        = "topic"
)

type EmailMessage struct {
	Event          string            `json:"event"`
	MessageID      uuid.UUID         `json:"message_id"`
	CorrelationID  uuid.UUID         `json:"correlation_id"`
	OrganizationID uuid.UUID         `json:"organization_id"`
	BookingID      uuid.UUID         `json:"booking_id"`
	Recipient      struct {
		Email string `json:"email"`
	} `json:"recipient"`
	Template  string            `json:"template"`
	Variables map[string]string `json:"variables"`
	Timestamp time.Time         `json:"timestamp"`
	Source    string            `json:"source"`
}

type notificationResultEvent struct {
	Event               string    `json:"event"`
	MessageID           uuid.UUID `json:"message_id"`
	CorrelationID       uuid.UUID `json:"correlation_id"`
	Timestamp           time.Time `json:"timestamp"`
	Source              string    `json:"source"`
	OrganizationID      uuid.UUID `json:"organization_id"`
	BookingID           uuid.UUID `json:"booking_id"`
	Channel             string    `json:"channel"`
	Status              string    `json:"status"`
	Provider            string    `json:"provider,omitempty"`
	ProviderMessageID   string    `json:"provider_message_id,omitempty"`
	Error               string    `json:"error,omitempty"`
}

type EmailRecord struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	MessageID         uuid.UUID `gorm:"type:uuid;uniqueIndex"`
	CorrelationID     uuid.UUID `gorm:"type:uuid;index"`
	OrganizationID    uuid.UUID `gorm:"type:uuid;index"`
	BookingID         uuid.UUID `gorm:"type:uuid;index"`
	RecipientEmail    string
	Template          string
	Variables         string `gorm:"type:jsonb"`
	Subject           string
	Provider          string
	ProviderMessageID string
	Status            string
	RetryCount        int
	SentAt            *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (EmailRecord) TableName() string {
	return "email_messages"
}

type Consumer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	db       *gorm.DB
	provider provider.Provider
}

func NewConsumer(rabbitURL string, db *gorm.DB, emailProvider provider.Provider) (*Consumer, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := declareWorkerTopology(ch); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	if err := ch.Qos(10, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &Consumer{conn: conn, channel: ch, db: db, provider: emailProvider}, nil
}

func declareWorkerTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(backendExchangeName, exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare backend exchange: %w", err)
	}

	if err := ch.ExchangeDeclare(workerExchangeName, exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare worker exchange: %w", err)
	}

	q, err := ch.QueueDeclare(workerQueueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare worker queue: %w", err)
	}

	if err := ch.QueueBind(q.Name, inboundRoutingKey, backendExchangeName, false, nil); err != nil {
		return fmt.Errorf("bind worker queue: %w", err)
	}

	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(workerQueueName, "email-worker", false, false, false, false, nil)
	if err != nil {
		return err
	}

	slog.Info("email worker started",
		"queue", workerQueueName,
		"inbound_exchange", backendExchangeName,
		"outbound_exchange", workerExchangeName,
	)

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
	var event EmailMessage
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return err
	}

	slog.Info("processing email",
		"message_id", event.MessageID,
		"correlation_id", event.CorrelationID,
		"email", event.Recipient.Email,
	)

	variablesJSON, _ := json.Marshal(event.Variables)
	subject, htmlBody := provider.RenderEmail(event.Template, event.Variables)

	record := EmailRecord{
		ID:             uuid.New(),
		MessageID:      event.MessageID,
		CorrelationID:  event.CorrelationID,
		OrganizationID: event.OrganizationID,
		BookingID:      event.BookingID,
		RecipientEmail: event.Recipient.Email,
		Template:       event.Template,
		Variables:      string(variablesJSON),
		Subject:        subject,
		Provider:       c.provider.Name(),
		Status:         "processing",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := c.db.WithContext(ctx).Create(&record).Error; err != nil {
		return err
	}

	if err := c.publishStatus(ctx, event, processingRoutingKey, "processing", "", ""); err != nil {
		return err
	}

	result, err := c.provider.Send(ctx, event.Recipient.Email, subject, htmlBody)
	if err != nil {
		record.Status = "failed"
		record.RetryCount++
		c.db.Save(&record)
		_ = c.publishStatus(ctx, event, failedRoutingKey, "failed", c.provider.Name(), err.Error())
		return err
	}

	now := time.Now().UTC()
	record.Status = "sent"
	record.ProviderMessageID = result.ProviderMessageID
	record.SentAt = &now
	if err := c.db.WithContext(ctx).Save(&record).Error; err != nil {
		return err
	}

	return c.publishStatus(ctx, event, sentRoutingKey, "sent", c.provider.Name(), result.ProviderMessageID)
}

func (c *Consumer) publishStatus(ctx context.Context, event EmailMessage, routingKey, status, provider, detail string) error {
	result := notificationResultEvent{
		Event:          routingKey,
		MessageID:      event.MessageID,
		CorrelationID:  event.CorrelationID,
		Timestamp:      time.Now().UTC(),
		Source:         "meetoria-email-worker",
		OrganizationID: event.OrganizationID,
		BookingID:      event.BookingID,
		Channel:        "email",
		Status:         status,
	}

	switch status {
	case "sent":
		result.Provider = provider
		result.ProviderMessageID = detail
	case "failed":
		result.Provider = provider
		result.Error = detail
	}

	body, err := json.Marshal(result)
	if err != nil {
		return err
	}

	if err := c.channel.PublishWithContext(ctx, workerExchangeName, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
	}); err != nil {
		return err
	}

	slog.Info("email status published",
		"exchange", workerExchangeName,
		"routing_key", routingKey,
		"status", status,
		"message_id", event.MessageID,
		"correlation_id", event.CorrelationID,
	)

	return nil
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
