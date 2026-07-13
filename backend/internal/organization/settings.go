package organization

import (
	"encoding/json"
)

type OrganizationSettings struct {
	Booking    BookingSettings `json:"booking"`
	TimeFormat string          `json:"time_format"`
}

const (
	TimeFormat24h = "24h"
	TimeFormat12h = "12h"
)

func NormalizeTimeFormat(value string) string {
	if value == TimeFormat12h {
		return TimeFormat12h
	}
	return TimeFormat24h
}

type BookingSettings struct {
	Enabled            bool   `json:"enabled"`
	BookingWindowDays  int    `json:"booking_window_days"`
	MinNoticeMinutes   int    `json:"min_notice_minutes"`
	MaxNoticeMinutes   *int   `json:"max_notice_minutes,omitempty"`
	EmailRequired      bool   `json:"email_required"`
	AutoConfirm        bool   `json:"auto_confirm"`
	ManualApproval     bool   `json:"manual_approval"`
	CancellationPolicy string `json:"cancellation_policy"`
	ReschedulingPolicy string `json:"rescheduling_policy"`
}

func DefaultBookingSettings() BookingSettings {
	return BookingSettings{
		Enabled:           true,
		BookingWindowDays: 30,
		MinNoticeMinutes:  60,
		AutoConfirm:       true,
	}
}

func DefaultOrganizationSettings() OrganizationSettings {
	return OrganizationSettings{Booking: DefaultBookingSettings()}.withDefaults()
}

func ParseSettings(raw string) OrganizationSettings {
	if raw == "" || raw == "{}" {
		return OrganizationSettings{Booking: DefaultBookingSettings()}.withDefaults()
	}
	var settings OrganizationSettings
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return OrganizationSettings{Booking: DefaultBookingSettings()}.withDefaults()
	}
	return settings.withDefaults()
}

func (s OrganizationSettings) withDefaults() OrganizationSettings {
	s.Booking = s.Booking.withDefaults()
	s.TimeFormat = NormalizeTimeFormat(s.TimeFormat)
	return s
}

func (b BookingSettings) WithDefaults() BookingSettings {
	defaults := DefaultBookingSettings()
	if b.BookingWindowDays <= 0 {
		b.BookingWindowDays = defaults.BookingWindowDays
	}
	if b.MinNoticeMinutes <= 0 {
		b.MinNoticeMinutes = defaults.MinNoticeMinutes
	}
	if b.MaxNoticeMinutes != nil && *b.MaxNoticeMinutes <= 0 {
		b.MaxNoticeMinutes = nil
	}
	return b
}

func (b BookingSettings) withDefaults() BookingSettings {
	return b.WithDefaults()
}

func (b BookingSettings) InitialBookingStatus() string {
	if b.ManualApproval {
		return "pending"
	}
	if b.AutoConfirm {
		return "confirmed"
	}
	return "pending"
}

func MarshalSettings(settings OrganizationSettings) (string, error) {
	data, err := json.Marshal(settings)
	if err != nil {
		return "{}", err
	}
	return string(data), nil
}
