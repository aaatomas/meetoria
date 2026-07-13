package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	EventBookingCreated   = "booking.created"
	EventBookingUpdated   = "booking.updated"
	EventBookingCancelled = "booking.cancelled"
	EventNotificationSMS         = "notification.sms"
	EventNotificationEmail       = "notification.email"
	EventNotificationSMSSent       = "notification.sms.sent"
	EventNotificationSMSFailed     = "notification.sms.failed"
	EventNotificationSMSProcessing = "notification.sms.processing"
	EventNotificationEmailSent       = "notification.email.sent"
	EventNotificationEmailFailed     = "notification.email.failed"
	EventNotificationEmailProcessing = "notification.email.processing"
	EventPaymentCompleted = "payment.completed"
	EventCalendarSync     = "calendar.sync"
)

type BaseEvent struct {
	Event         string    `json:"event"`
	MessageID     uuid.UUID `json:"message_id"`
	CorrelationID uuid.UUID `json:"correlation_id"`
	Timestamp     time.Time `json:"timestamp"`
	Source        string    `json:"source"`
}

func NewBaseEvent(eventType, source string, correlationID uuid.UUID) BaseEvent {
	return BaseEvent{
		Event:         eventType,
		MessageID:     uuid.New(),
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC(),
		Source:        source,
	}
}

type BookingEvent struct {
	BaseEvent
	OrganizationID uuid.UUID `json:"organization_id"`
	BookingID      uuid.UUID `json:"booking_id"`
	CustomerID     uuid.UUID `json:"customer_id"`
	EmployeeID     uuid.UUID `json:"employee_id"`
	ServiceID      uuid.UUID `json:"service_id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Status         string    `json:"status"`
}

type SMSNotificationEvent struct {
	BaseEvent
	OrganizationID uuid.UUID         `json:"organization_id"`
	BookingID      uuid.UUID         `json:"booking_id"`
	Recipient      SMSRecipient      `json:"recipient"`
	Template       string            `json:"template"`
	Variables      map[string]string `json:"variables"`
}

type SMSRecipient struct {
	Phone string `json:"phone"`
}

type EmailNotificationEvent struct {
	BaseEvent
	OrganizationID uuid.UUID         `json:"organization_id"`
	BookingID      uuid.UUID         `json:"booking_id,omitempty"`
	Recipient      EmailRecipient    `json:"recipient"`
	Template       string            `json:"template"`
	Variables      map[string]string `json:"variables"`
}

type EmailRecipient struct {
	Email string `json:"email"`
}

type NotificationResultEvent struct {
	BaseEvent
	OrganizationID      uuid.UUID `json:"organization_id"`
	BookingID           uuid.UUID `json:"booking_id"`
	Channel             string    `json:"channel"`
	Status              string    `json:"status"`
	Provider            string    `json:"provider,omitempty"`
	ProviderMessageID   string    `json:"provider_message_id,omitempty"`
	Error               string    `json:"error,omitempty"`
}

func MarshalEvent(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal event: %w", err)
	}
	return data, nil
}
