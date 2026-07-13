package phone

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
)

const DisplayExample = "+370 123 12345"

var e164Pattern = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

func RegisterValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("e164", validatePhone)
	}
}

func validatePhone(fl validator.FieldLevel) bool {
	raw, ok := fieldString(fl.Field())
	if !ok || raw == "" {
		return true
	}
	_, err := NormalizeE164(raw)
	return err == nil
}

func fieldString(field reflect.Value) (string, bool) {
	switch field.Kind() {
	case reflect.String:
		return field.String(), true
	case reflect.Ptr:
		if field.IsNil() {
			return "", true
		}
		if field.Elem().Kind() == reflect.String {
			return field.Elem().String(), true
		}
	}
	return "", false
}

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
		return "", apperrors.Validation("phone must use international format, e.g. " + DisplayExample)
	}

	return phone, nil
}

func NormalizeOptional(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	return NormalizeE164(raw)
}

func NormalizeOptionalPtr(value *string) (*string, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return value, nil
	}
	normalized, err := NormalizeE164(*value)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}
