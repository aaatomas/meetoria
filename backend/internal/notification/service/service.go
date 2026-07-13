package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/meetoria/meetoria/backend/internal/booking"
	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	customerrepo "github.com/meetoria/meetoria/backend/internal/customer/repository"
	employeerepo "github.com/meetoria/meetoria/backend/internal/employee/repository"
	"github.com/meetoria/meetoria/backend/internal/common/rabbitmq"
	"github.com/meetoria/meetoria/backend/internal/notification"
	notifrepo "github.com/meetoria/meetoria/backend/internal/notification/repository"
	"github.com/meetoria/meetoria/backend/pkg/events"
)

type Service interface {
	QueueBookingConfirmation(ctx context.Context, orgID uuid.UUID, b *booking.Booking, correlationID uuid.UUID) error
	QueueBookingConfirmationSMS(ctx context.Context, orgID uuid.UUID, b *booking.Booking, correlationID uuid.UUID) error
	QueueBookingConfirmationEmail(ctx context.Context, orgID uuid.UUID, b *booking.Booking, correlationID uuid.UUID) error
	QueueBookingReminder(ctx context.Context, orgID uuid.UUID, b *booking.Booking, correlationID uuid.UUID) error
	UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status notification.Status) error
	ListByBooking(ctx context.Context, orgID, bookingID uuid.UUID) ([]notification.Notification, error)
}

type notificationService struct {
	repo         notifrepo.Repository
	customerRepo customerrepo.Repository
	employeeRepo employeerepo.Repository
	publisher    *rabbitmq.Publisher
}

func NewService(
	repo notifrepo.Repository,
	customerRepo customerrepo.Repository,
	employeeRepo employeerepo.Repository,
	publisher *rabbitmq.Publisher,
) Service {
	return &notificationService{
		repo:         repo,
		customerRepo: customerRepo,
		employeeRepo: employeeRepo,
		publisher:    publisher,
	}
}

func (s *notificationService) bookingConfirmationVariables(ctx context.Context, orgID uuid.UUID, b *booking.Booking) (map[string]string, error) {
	employee, err := s.employeeRepo.GetByID(ctx, orgID, b.EmployeeID)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"date":     b.StartTime.Format("2006-01-02"),
		"time":     b.StartTime.Format("15:04"),
		"employee": fmt.Sprintf("%s %s", employee.FirstName, employee.LastName),
	}, nil
}

func (s *notificationService) QueueBookingConfirmation(ctx context.Context, orgID uuid.UUID, b *booking.Booking, correlationID uuid.UUID) error {
	customer, err := s.customerRepo.GetByID(ctx, orgID, b.CustomerID)
	if err != nil {
		return err
	}

	variables, err := s.bookingConfirmationVariables(ctx, orgID, b)
	if err != nil {
		return err
	}

	if customer.Email != "" {
		if err := s.queueEmail(ctx, orgID, b.ID, customer.Email, "booking_confirmation", variables, correlationID); err != nil {
			return err
		}
	}

	if customer.Phone != "" {
		if err := s.queueSMS(ctx, orgID, b.ID, customer.Phone, "booking_confirmation", variables, correlationID); err != nil {
			return err
		}
	}

	return nil
}

func (s *notificationService) QueueBookingConfirmationSMS(ctx context.Context, orgID uuid.UUID, b *booking.Booking, correlationID uuid.UUID) error {
	customer, err := s.customerRepo.GetByID(ctx, orgID, b.CustomerID)
	if err != nil {
		return err
	}
	if customer.Phone == "" {
		return apperrors.Validation("customer has no phone number")
	}

	variables, err := s.bookingConfirmationVariables(ctx, orgID, b)
	if err != nil {
		return err
	}

	return s.queueSMS(ctx, orgID, b.ID, customer.Phone, "booking_confirmation", variables, correlationID)
}

func (s *notificationService) QueueBookingConfirmationEmail(ctx context.Context, orgID uuid.UUID, b *booking.Booking, correlationID uuid.UUID) error {
	customer, err := s.customerRepo.GetByID(ctx, orgID, b.CustomerID)
	if err != nil {
		return err
	}
	if customer.Email == "" {
		return apperrors.Validation("customer has no email address")
	}

	variables, err := s.bookingConfirmationVariables(ctx, orgID, b)
	if err != nil {
		return err
	}

	return s.queueEmail(ctx, orgID, b.ID, customer.Email, "booking_confirmation", variables, correlationID)
}

func (s *notificationService) QueueBookingReminder(ctx context.Context, orgID uuid.UUID, b *booking.Booking, correlationID uuid.UUID) error {
	customer, err := s.customerRepo.GetByID(ctx, orgID, b.CustomerID)
	if err != nil {
		return err
	}

	employee, err := s.employeeRepo.GetByID(ctx, orgID, b.EmployeeID)
	if err != nil {
		return err
	}

	variables := map[string]string{
		"date":     b.StartTime.Format("2006-01-02"),
		"time":     b.StartTime.Format("15:04"),
		"employee": fmt.Sprintf("%s %s", employee.FirstName, employee.LastName),
	}

	if customer.Phone != "" {
		return s.queueSMS(ctx, orgID, b.ID, customer.Phone, "booking_reminder", variables, correlationID)
	}
	return nil
}

func (s *notificationService) queueSMS(ctx context.Context, orgID, bookingID uuid.UUID, phone, template string, variables map[string]string, correlationID uuid.UUID) error {
	n := &notification.Notification{
		OrganizationID: orgID,
		BookingID:      &bookingID,
		Channel:        notification.ChannelSMS,
		Template:       template,
		Recipient:      phone,
		Status:         notification.StatusCreated,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return err
	}

	event := events.SMSNotificationEvent{
		BaseEvent:      events.NewBaseEvent(events.EventNotificationSMS, "meetoria-api", correlationID),
		OrganizationID: orgID,
		BookingID:      bookingID,
		Recipient:      events.SMSRecipient{Phone: phone},
		Template:       template,
		Variables:      variables,
	}

	n.MessageID = &event.MessageID

	body, err := events.MarshalEvent(event)
	if err != nil {
		return err
	}

	if err := s.publisher.Publish(ctx, events.EventNotificationSMS, body); err != nil {
		return err
	}

	n.Status = notification.StatusQueued
	return s.repo.Update(ctx, n)
}

func (s *notificationService) queueEmail(ctx context.Context, orgID, bookingID uuid.UUID, email, template string, variables map[string]string, correlationID uuid.UUID) error {
	n := &notification.Notification{
		OrganizationID: orgID,
		BookingID:      &bookingID,
		Channel:        notification.ChannelEmail,
		Template:       template,
		Recipient:      email,
		Status:         notification.StatusCreated,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return err
	}

	event := events.EmailNotificationEvent{
		BaseEvent:      events.NewBaseEvent(events.EventNotificationEmail, "meetoria-api", correlationID),
		OrganizationID: orgID,
		BookingID:      bookingID,
		Recipient:      events.EmailRecipient{Email: email},
		Template:       template,
		Variables:      variables,
	}

	n.MessageID = &event.MessageID

	body, err := events.MarshalEvent(event)
	if err != nil {
		return err
	}

	if err := s.publisher.Publish(ctx, events.EventNotificationEmail, body); err != nil {
		return err
	}

	n.Status = notification.StatusQueued
	return s.repo.Update(ctx, n)
}

func (s *notificationService) UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status notification.Status) error {
	n, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}
	n.Status = status
	now := time.Now().UTC()
	switch status {
	case notification.StatusSent:
		n.SentAt = &now
	case notification.StatusDelivered:
		n.DeliveredAt = &now
	}
	return s.repo.Update(ctx, n)
}

func (s *notificationService) ListByBooking(ctx context.Context, orgID, bookingID uuid.UUID) ([]notification.Notification, error) {
	return s.repo.ListByBooking(ctx, orgID, bookingID)
}
