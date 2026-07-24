// Package database opens and owns the PostgreSQL connection.
//
// It exposes a *gorm.DB, which by convention only the repository layer touches.
package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/config"
)

// Open connects to PostgreSQL, configures the pool, and pings.
func Open(cfg config.Database, log *zap.Logger) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger:                 newGormLogger(log, cfg),
		SkipDefaultTransaction: true,
		NowFunc:                func() time.Time { return time.Now().UTC() },
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("database: connecting to %s: %w", cfg.Redacted(), err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database: obtaining underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database: pinging %s: %w", cfg.Redacted(), err)
	}

	log.Info("database connected",
		zap.String("dsn", cfg.Redacted()),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
	)
	return db, nil
}

const pingTimeout = 5 * time.Second

// Ping verifies the connection is usable. Backs the /readyz probe.
func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Close drains and closes the pool.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Stats exposes pool statistics for health output.
func Stats(db *gorm.DB) (map[string]any, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	s := sqlDB.Stats()
	return map[string]any{
		"open_connections": s.OpenConnections,
		"in_use":           s.InUse,
		"idle":             s.Idle,
		"wait_count":       s.WaitCount,
		"wait_duration_ms": s.WaitDuration.Milliseconds(),
	}, nil
}

// gormZapLogger adapts GORM's logger onto zap.
type gormZapLogger struct {
	log           *zap.Logger
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

func newGormLogger(log *zap.Logger, cfg config.Database) gormlogger.Interface {
	return &gormZapLogger{
		log:           log.Named("gorm").WithOptions(zap.AddCallerSkip(3)),
		level:         parseGormLevel(cfg.LogLevel),
		slowThreshold: cfg.SlowQueryThreshold,
	}
}

func parseGormLevel(s string) gormlogger.LogLevel {
	switch s {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "info":
		return gormlogger.Info
	default:
		return gormlogger.Warn
	}
}

func (l *gormZapLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.level = level
	return &clone
}

func (l *gormZapLogger) Info(_ context.Context, msg string, data ...any) {
	if l.level >= gormlogger.Info {
		l.log.Sugar().Infof(msg, data...)
	}
}

func (l *gormZapLogger) Warn(_ context.Context, msg string, data ...any) {
	if l.level >= gormlogger.Warn {
		l.log.Sugar().Warnf(msg, data...)
	}
}

func (l *gormZapLogger) Error(_ context.Context, msg string, data ...any) {
	if l.level >= gormlogger.Error {
		l.log.Sugar().Errorf(msg, data...)
	}
}

func (l *gormZapLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Duration("elapsed", elapsed),
	}

	switch {
	case err != nil && l.level >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		l.log.Error("query failed", append(fields, zap.Error(err))...)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		l.log.Warn("slow query", append(fields, zap.Duration("threshold", l.slowThreshold))...)
	case l.level >= gormlogger.Info:
		l.log.Debug("query", fields...)
	}
}
