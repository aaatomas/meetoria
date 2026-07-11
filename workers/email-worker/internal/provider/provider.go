package provider

import (
	"context"
	"fmt"
	"log/slog"
)

type SendResult struct {
	ProviderMessageID string
	Status            string
}

type Provider interface {
	Name() string
	Send(ctx context.Context, to, subject, htmlBody string) (*SendResult, error)
}

type MockProvider struct{}

func NewMockProvider() Provider {
	return &MockProvider{}
}

func (p *MockProvider) Name() string {
	return "mock"
}

func (p *MockProvider) Send(ctx context.Context, to, subject, htmlBody string) (*SendResult, error) {
	slog.Info("mock email sent", "to", to, "subject", subject)
	return &SendResult{ProviderMessageID: fmt.Sprintf("mock-email-%s", to), Status: "sent"}, nil
}

type SMTPProvider struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSMTPProvider(host string, port int, username, password, from string) Provider {
	return &SMTPProvider{host: host, port: port, username: username, password: password, from: from}
}

func (p *SMTPProvider) Name() string {
	return "smtp"
}

func (p *SMTPProvider) Send(ctx context.Context, to, subject, htmlBody string) (*SendResult, error) {
	slog.Info("smtp email", "to", to, "subject", subject)
	return &SendResult{ProviderMessageID: "smtp-pending", Status: "sent"}, nil
}

func NewProvider(providerType string) Provider {
	switch providerType {
	case "smtp":
		return NewSMTPProvider("localhost", 587, "", "", "noreply@meetoria.com")
	default:
		return NewMockProvider()
	}
}

func RenderEmail(template string, variables map[string]string) (subject, html string) {
	switch template {
	case "booking_confirmation":
		subject = "Appointment Confirmed"
		html = fmt.Sprintf("<h1>Booking Confirmed</h1><p>Hi %s, your appointment is confirmed for %s at %s with %s.</p>",
			variables["name"], variables["date"], variables["time"], variables["employee"])
	default:
		subject = "Meetoria Notification"
		html = fmt.Sprintf("<p>Notification: %s</p>", template)
	}
	return subject, html
}
