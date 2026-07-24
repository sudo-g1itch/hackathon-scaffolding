// Package middleware holds the cross-cutting concerns of the HTTP layer.
//
// Chain order: Recovery → RequestID → CORS → Logger → [route-specific]
package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/config"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/ctxkey"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/response"
)

const RequestIDHeader = "X-Request-Id"

// Recovery converts a panic into a sanitized 500.
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				ctxkey.Logger(c).Error("panic recovered",
					zap.Any("panic", r),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.Stack("stack"),
				)
				response.Error(c, apperr.Internal(errFromPanic(r)))
			}
		}()
		c.Next()
	}
}

func errFromPanic(r any) error {
	if err, ok := r.(error); ok {
		return err
	}
	return apperr.Newf(apperr.CodeInternal, "panic: %v", r)
}

// RequestID assigns or adopts a correlation id.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" || len(id) > 128 {
			id = newID()
		}
		ctxkey.SetRequestID(c, id)
		c.Header(RequestIDHeader, id)
		c.Next()
	}
}

func newID() string {
	if id, err := uuid.NewV7(); err == nil {
		return id.String()
	}
	return uuid.NewString()
}

// CORS applies the browser access policy from configuration.
func CORS(cfg config.CORS) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     cfg.AllowedMethods,
		AllowHeaders:     cfg.AllowedHeaders,
		ExposeHeaders:    cfg.ExposedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	})
}

// Logger attaches a request-scoped logger and emits one structured line per request.
func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		reqLog := log.With(
			zap.String("request_id", ctxkey.RequestID(c)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		ctxkey.SetLogger(c, reqLog)

		c.Next()

		summaryLog := reqLog.WithOptions(zap.AddStacktrace(zapcore.PanicLevel))

		status := c.Writer.Status()
		fields := []zap.Field{
			zap.Int("status", status),
			zap.Duration("duration", time.Since(start)),
			zap.Int("bytes", c.Writer.Size()),
			zap.String("ip", c.ClientIP()),
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		switch {
		case status >= 500:
			summaryLog.Error("request", fields...)
		case status >= 400:
			summaryLog.Warn("request", fields...)
		default:
			summaryLog.Info("request", fields...)
		}
	}
}

// NotFound answers unmatched routes with the standard error envelope.
func NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, apperr.NotFound("endpoint"))
	}
}

// MethodNotAllowed answers a known path with the wrong verb.
func MethodNotAllowed() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, apperr.New(apperr.CodeNotFound, "That method is not allowed on this endpoint."))
	}
}
