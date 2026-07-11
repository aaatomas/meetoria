package errors_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name       string
		input      error
		wantCode   string
		wantStatus int
	}{
		{"not found", apperrors.ErrNotFound, "NOT_FOUND", 404},
		{"double booking", apperrors.ErrDoubleBooking, "CONFLICT", 409},
		{"tenant isolation", apperrors.ErrTenantIsolation, "FORBIDDEN", 403},
		{"validation", apperrors.ErrValidation, "VALIDATION_ERROR", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := apperrors.MapError(tt.input)
			assert.Equal(t, tt.wantCode, result.Code)
			assert.Equal(t, tt.wantStatus, result.Status)
		})
	}
}

func TestAppErrorUnwrap(t *testing.T) {
	inner := apperrors.ErrNotFound
	appErr := apperrors.NotFound("test not found")
	assert.True(t, errors.Is(appErr, inner))
}
