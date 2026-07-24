// Package apperr defines the error vocabulary shared by every layer of the API.
//
// Services describe what went wrong in domain terms; handlers translate that
// into an HTTP status. Services never import net/http or gin, and handlers
// never guess a status from a string.
package apperr

import (
	"errors"
	"fmt"
)

// Code is the machine-readable error code carried in the response envelope.
type Code string

const (
	CodeValidation    Code = "VALIDATION_ERROR"
	CodeUnprocessable Code = "UNPROCESSABLE_ENTITY"
	CodeNotFound      Code = "NOT_FOUND"
	CodeConflict      Code = "CONFLICT"
	CodeUnauthorized  Code = "UNAUTHORIZED"
	CodeForbidden     Code = "FORBIDDEN"
	CodeInternal      Code = "INTERNAL_ERROR"
)

// Fields carries per-field validation messages, keyed by JSON field name.
type Fields map[string][]string

// Error is the single error type crossing layer boundaries.
type Error struct {
	Code    Code
	Message string
	Fields  Fields
	err     error
}

func (e *Error) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.err }

func (e *Error) Is(target error) bool {
	var t *Error
	if !errors.As(target, &t) {
		return false
	}
	return e.Code == t.Code
}

func (e *Error) WithFields(f Fields) *Error {
	e.Fields = f
	return e
}

func (e *Error) Wrap(err error) *Error {
	e.err = err
	return e
}

// Sentinels for errors.Is comparisons.
var (
	ErrValidation    = &Error{Code: CodeValidation}
	ErrUnprocessable = &Error{Code: CodeUnprocessable}
	ErrNotFound      = &Error{Code: CodeNotFound}
	ErrConflict      = &Error{Code: CodeConflict}
	ErrUnauthorized  = &Error{Code: CodeUnauthorized}
	ErrForbidden     = &Error{Code: CodeForbidden}
	ErrInternal      = &Error{Code: CodeInternal}
)

func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func Newf(code Code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

func NotFound(resource string) *Error {
	return Newf(CodeNotFound, "%s not found", resource)
}

func Conflict(format string, args ...any) *Error {
	return Newf(CodeConflict, format, args...)
}

func Unprocessable(format string, args ...any) *Error {
	return Newf(CodeUnprocessable, format, args...)
}

func Validation(fields Fields) *Error {
	return New(CodeValidation, "The request contains invalid fields.").WithFields(fields)
}

func Unauthorized(message string) *Error {
	return New(CodeUnauthorized, message)
}

func Forbidden(message string) *Error {
	return New(CodeForbidden, message)
}

func Internal(err error) *Error {
	return (&Error{Code: CodeInternal, Message: "internal error"}).Wrap(err)
}

func From(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return Internal(err)
}
