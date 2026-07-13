package phone

import (
	"regexp"
	"strings"

	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
)

var e164Pattern = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

func NormalizeE164(raw string) (string, error) {
	var b strings.Builder
	for _, r := range strings.TrimSpace(raw) {
		switch r {
		case ' ', '-', '(', ')':
			continue
		default:
			b.WriteRune(r)
		}
	}
	phone := b.String()
	if phone == "" {
		return "", apperrors.Validation("phone is required")
	}

	switch {
	case strings.HasPrefix(phone, "00"):
		phone = "+" + phone[2:]
	case !strings.HasPrefix(phone, "+") && strings.HasPrefix(phone, "0"):
		phone = "+370" + phone[1:]
	case !strings.HasPrefix(phone, "+"):
		phone = "+" + phone
	}

	if !e164Pattern.MatchString(phone) {
		return "", apperrors.Validation("phone must be in international format, e.g. +37060000000")
	}

	return phone, nil
}
