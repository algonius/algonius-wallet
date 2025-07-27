// Package errors provides tests for the error handling package.
package errors

import (
	"errors"
	"testing"
)

func TestErrorCreation(t *testing.T) {
	// Test basic error creation
	err := New(ErrInvalidParameter, "test error")
	if err.Code != ErrInvalidParameter {
		t.Errorf("Expected code %s, got %s", ErrInvalidParameter, err.Code)
	}
	if err.Message != "test error" {
		t.Errorf("Expected message 'test error', got '%s'", err.Message)
	}
}

func TestErrorWithDetails(t *testing.T) {
	err := New(ErrInvalidParameter, "test error").WithDetails("detailed info")
	if err.Details != "detailed info" {
		t.Errorf("Expected details 'detailed info', got '%s'", err.Details)
	}
}

func TestErrorWithSuggestion(t *testing.T) {
	err := New(ErrInvalidParameter, "test error").WithSuggestion("try this")
	if err.Suggestion != "try this" {
		t.Errorf("Expected suggestion 'try this', got '%s'", err.Suggestion)
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	err := Wrap(originalErr, ErrInternal, "wrapped error")
	if err.Code != ErrInternal {
		t.Errorf("Expected code %s, got %s", ErrInternal, err.Code)
	}
	if err.Details != "original error" {
		t.Errorf("Expected details 'original error', got '%s'", err.Details)
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError("field", "invalid format")
	if err.Code != ErrInvalidParameter {
		t.Errorf("Expected code %s, got %s", ErrInvalidParameter, err.Code)
	}
	if err.Suggestion == "" {
		t.Error("Expected suggestion to be set")
	}
}

func TestMissingRequiredFieldError(t *testing.T) {
	err := MissingRequiredFieldError("field")
	if err.Code != ErrMissingRequiredField {
		t.Errorf("Expected code %s, got %s", ErrMissingRequiredField, err.Code)
	}
	if err.Suggestion == "" {
		t.Error("Expected suggestion to be set")
	}
}