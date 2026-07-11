package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var (
	ErrNotFound            = errors.New("resource not found")
	ErrConflict            = errors.New("resource conflict")
	ErrForbidden           = errors.New("forbidden")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrValidation          = errors.New("validation error")
	ErrTenantIsolation     = errors.New("cross-tenant access denied")
	ErrDoubleBooking       = errors.New("time slot already booked")
	ErrInvalidTimeSlot     = errors.New("invalid time slot")
	ErrOutsideWorkingHours = errors.New("outside working hours")
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code, message string, status int, err error) *AppError {
	return &AppError{Code: code, Message: message, Status: status, Err: err}
}

func NotFound(message string) *AppError {
	return NewAppError("NOT_FOUND", message, http.StatusNotFound, ErrNotFound)
}

func Conflict(message string) *AppError {
	return NewAppError("CONFLICT", message, http.StatusConflict, ErrConflict)
}

func Forbidden(message string) *AppError {
	return NewAppError("FORBIDDEN", message, http.StatusForbidden, ErrForbidden)
}

func Unauthorized(message string) *AppError {
	return NewAppError("UNAUTHORIZED", message, http.StatusUnauthorized, ErrUnauthorized)
}

func Validation(message string) *AppError {
	return NewAppError("VALIDATION_ERROR", message, http.StatusBadRequest, ErrValidation)
}

func Internal(message string, err error) *AppError {
	return NewAppError("INTERNAL_ERROR", message, http.StatusInternalServerError, err)
}

func MapError(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	switch {
	case errors.Is(err, ErrNotFound):
		return NotFound(err.Error())
	case errors.Is(err, ErrConflict):
		return Conflict(err.Error())
	case errors.Is(err, ErrForbidden):
		return Forbidden(err.Error())
	case errors.Is(err, ErrUnauthorized):
		return Unauthorized(err.Error())
	case errors.Is(err, ErrValidation):
		return Validation(err.Error())
	case errors.Is(err, ErrTenantIsolation):
		return Forbidden("access denied to organization resource")
	case errors.Is(err, ErrDoubleBooking):
		return Conflict("time slot is already booked")
	case errors.Is(err, ErrInvalidTimeSlot):
		return Validation("invalid time slot")
	case errors.Is(err, ErrOutsideWorkingHours):
		return Validation("requested time is outside working hours")
	default:
		var validationErr validator.ValidationErrors
		if errors.As(err, &validationErr) {
			return Validation(validationErr.Error())
		}
		return Internal("an unexpected error occurred", err)
	}
}
