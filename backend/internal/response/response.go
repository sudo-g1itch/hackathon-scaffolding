// Package response owns the HTTP response envelope.
//
// Every endpoint answers in one of two shapes:
//
//	{ "success": true,  "data": …, "meta": { "pagination": { … } } }
//	{ "success": false, "error": { "code": "…", "message": "…", "fields": { … } } }
package response

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/ctxkey"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/pagination"
)

const genericInternalMessage = "An unexpected error occurred. Please try again or contact support with the request id."

// Envelope is the success shape.
type Envelope struct {
	Success bool  `json:"success"`
	Data    any   `json:"data"`
	Meta    *Meta `json:"meta,omitempty"`
}

type Meta struct {
	Pagination *pagination.Meta `json:"pagination,omitempty"`
}

// ErrorEnvelope is the failure shape.
type ErrorEnvelope struct {
	Success bool `json:"success"`
	Error   Body `json:"error"`
}

type Body struct {
	Code      apperr.Code   `json:"code"`
	Message   string        `json:"message"`
	Fields    apperr.Fields `json:"fields,omitempty"`
	RequestID string        `json:"request_id,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

func Paginated[T any](c *gin.Context, page pagination.Page[T]) {
	meta := page.Meta
	c.JSON(http.StatusOK, Envelope{
		Success: true,
		Data:    page.Items,
		Meta:    &Meta{Pagination: &meta},
	})
}

// Status maps a domain error to its HTTP status.
func Status(code apperr.Code) int {
	switch code {
	case apperr.CodeValidation, apperr.CodeUnprocessable:
		return http.StatusUnprocessableEntity
	case apperr.CodeNotFound:
		return http.StatusNotFound
	case apperr.CodeConflict:
		return http.StatusConflict
	case apperr.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperr.CodeForbidden:
		return http.StatusForbidden
	case apperr.CodeUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// Error writes the failure envelope for any error.
func Error(c *gin.Context, err error) {
	appErr := apperr.From(err)
	status := Status(appErr.Code)

	message := appErr.Message

	// Every 5xx is logged server-side with the request id. Only CodeInternal —
	// the genuinely unexpected — has its message sanitized; a CodeUnavailable
	// message is written for the user ("AI is not configured") and leaks
	// nothing, so it is passed through.
	if status >= http.StatusInternalServerError {
		ctxkey.Logger(c).Error("request failed",
			zap.Error(err),
			zap.String("path", c.FullPath()),
			zap.String("method", c.Request.Method),
		)
		if appErr.Code == apperr.CodeInternal {
			message = genericInternalMessage
		}
	}

	c.AbortWithStatusJSON(status, ErrorEnvelope{
		Success: false,
		Error: Body{
			Code:      appErr.Code,
			Message:   message,
			Fields:    appErr.Fields,
			RequestID: ctxkey.RequestID(c),
		},
	})
}

// BindJSON binds and validates a JSON body. Returns false when the response
// has already been written.
func BindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		Error(c, bindingError(err))
		return false
	}
	return true
}

// BindQuery binds and validates query parameters.
func BindQuery(c *gin.Context, target any) bool {
	if err := c.ShouldBindQuery(target); err != nil {
		Error(c, bindingError(err))
		return false
	}
	return true
}

func bindingError(err error) error {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make(apperr.Fields, len(ve))
		for _, fe := range ve {
			name := fieldName(fe)
			fields[name] = append(fields[name], describe(fe))
		}
		return apperr.Validation(fields)
	}
	return apperr.New(apperr.CodeValidation, "The request body could not be parsed: "+err.Error())
}

func fieldName(fe validator.FieldError) string {
	ns := fe.Namespace()
	if i := strings.Index(ns, "."); i >= 0 {
		ns = ns[i+1:]
	}
	parts := strings.Split(ns, ".")
	for i, p := range parts {
		parts[i] = toSnake(p)
	}
	return strings.Join(parts, ".")
}

func toSnake(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 4)
	runes := []rune(s)
	for i, r := range runes {
		isUpper := r >= 'A' && r <= 'Z'
		if isUpper && i > 0 {
			prevLower := runes[i-1] >= 'a' && runes[i-1] <= 'z'
			nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
			if prevLower || nextLower {
				b.WriteByte('_')
			}
		}
		if isUpper {
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}

func describe(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", strings.ReplaceAll(fe.Param(), " ", ", "))
	case "uuid":
		return "must be a valid UUID"
	case "url":
		return "must be a valid URL"
	default:
		return "is invalid"
	}
}
