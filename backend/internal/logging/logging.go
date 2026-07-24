// Package logging builds the application's zap logger.
//
// One logger is constructed in main and injected downward. Requests get a
// child logger tagged with their request id.
package logging

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/config"
)

// New builds the root logger.
func New(cfg config.Log, appCfg config.App) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("logging: invalid log.level %q: %w", cfg.Level, err)
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.EncodeDuration = zapcore.MillisDurationEncoder

	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encCfg)
	}

	core := zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), level)
	if appCfg.IsProduction() {
		core = zapcore.NewSamplerWithOptions(core, samplingTick, samplingFirst, samplingThereafter)
	}

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("service", appCfg.Name),
			zap.String("env", appCfg.Env),
		),
	)
	return logger, nil
}

const (
	samplingTick       = time.Second
	samplingFirst      = 100
	samplingThereafter = 100
)

// NewNop returns a logger that discards everything. Useful in tests.
func NewNop() *zap.Logger { return zap.NewNop() }
