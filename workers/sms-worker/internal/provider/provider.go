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
	Send(ctx context.Context, phone, body string) (*SendResult, error)
}

type MockProvider struct{}

func NewMockProvider() Provider {
	return &MockProvider{}
}

func (p *MockProvider) Name() string {
	return "mock"
}

func (p *MockProvider) Send(ctx context.Context, phone, body string) (*SendResult, error) {
	slog.Info("mock sms sent", "phone", phone, "body", body)
	return &SendResult{
		ProviderMessageID: fmt.Sprintf("mock-%s", phone),
		Status:            "sent",
	}, nil
}

type TwilioProvider struct {
	accountSID string
	authToken  string
	fromNumber string
}

func NewTwilioProvider(accountSID, authToken, fromNumber string) Provider {
	return &TwilioProvider{accountSID: accountSID, authToken: authToken, fromNumber: fromNumber}
}

func (p *TwilioProvider) Name() string {
	return "twilio"
}

func (p *TwilioProvider) Send(ctx context.Context, phone, body string) (*SendResult, error) {
	// Twilio SDK integration placeholder - configure via env in production
	slog.Info("twilio sms", "phone", phone)
	return &SendResult{ProviderMessageID: "twilio-pending", Status: "sent"}, nil
}

func NewProvider(providerType string) Provider {
	switch providerType {
	case "twilio":
		return NewTwilioProvider("", "", "")
	default:
		return NewMockProvider()
	}
}

func RenderTemplate(template string, variables map[string]string) string {
	body := template
	for k, v := range variables {
		body = replaceAll(body, "{{"+k+"}}", v)
	}
	if template == "booking_confirmation" {
		return fmt.Sprintf("Your appointment is confirmed for %s at %s with %s.",
			variables["date"], variables["time"], variables["employee"])
	}
	if template == "booking_reminder" {
		return fmt.Sprintf("Reminder: appointment on %s at %s with %s.",
			variables["date"], variables["time"], variables["employee"])
	}
	return body
}

func replaceAll(s, old, new string) string {
	for i := 0; i < len(s); i++ {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			s = s[:i] + new + s[i+len(old):]
		}
	}
	return s
}
