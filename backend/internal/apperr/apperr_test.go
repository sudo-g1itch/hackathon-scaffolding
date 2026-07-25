package apperr

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(CodeValidation, "validation failed")
	if err.Code != CodeValidation {
		t.Errorf("expected code %s, got %s", CodeValidation, err.Code)
	}
	if err.Message != "validation failed" {
		t.Errorf("expected message 'validation failed', got %s", err.Message)
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound("user")
	if err.Code != CodeNotFound {
		t.Errorf("expected code %s, got %s", CodeNotFound, err.Code)
	}
	if err.Message != "user not found" {
		t.Errorf("expected message 'user not found', got %s", err.Message)
	}
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("please login")
	if err.Code != CodeUnauthorized {
		t.Errorf("expected code %s, got %s", CodeUnauthorized, err.Code)
	}
	if err.Message != "please login" {
		t.Errorf("expected message 'please login', got %s", err.Message)
	}
}

func TestIs(t *testing.T) {
	err1 := NotFound("user")
	err2 := Unauthorized("login required")
	
	if !err1.Is(ErrNotFound) {
		t.Errorf("expected err1 to match ErrNotFound")
	}
	if err2.Is(ErrNotFound) {
		t.Errorf("expected err2 to not match ErrNotFound")
	}
}

func TestWrap(t *testing.T) {
	baseErr := errors.New("underlying error")
	appErr := Internal(baseErr)
	
	if !errors.Is(appErr, baseErr) {
		t.Errorf("expected appErr to wrap baseErr")
	}
}
