// Package config is the only place in the codebase that reads the environment.
//
// Everything else receives a typed Config (or a sub-struct of it) through
// constructor injection.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

// Config is the complete runtime configuration of the API.
type Config struct {
	App      App      `mapstructure:"app"`
	HTTP     HTTP     `mapstructure:"http"`
	Database Database `mapstructure:"database"`
	Log      Log      `mapstructure:"log"`
	CORS     CORS     `mapstructure:"cors"`
	JWT      JWT      `mapstructure:"jwt"`
	AI       AI       `mapstructure:"ai"`
}

type App struct {
	Name        string `mapstructure:"name"`
	Env         string `mapstructure:"env"`
	AutoMigrate bool   `mapstructure:"auto_migrate"`
}

func (a App) IsProduction() bool  { return a.Env == EnvProduction }
func (a App) IsDevelopment() bool { return a.Env == EnvDevelopment }

type HTTP struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	TrustedProxies  []string      `mapstructure:"trusted_proxies"`
}

func (h HTTP) Addr() string { return fmt.Sprintf("%s:%d", h.Host, h.Port) }

type Database struct {
	Host               string        `mapstructure:"host"`
	Port               int           `mapstructure:"port"`
	User               string        `mapstructure:"user"`
	Password           string        `mapstructure:"password"`
	Name               string        `mapstructure:"name"`
	SSLMode            string        `mapstructure:"ssl_mode"`
	TimeZone           string        `mapstructure:"timezone"`
	MaxOpenConns       int           `mapstructure:"max_open_conns"`
	MaxIdleConns       int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime    time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime    time.Duration `mapstructure:"conn_max_idle_time"`
	LogLevel           string        `mapstructure:"log_level"`
	SlowQueryThreshold time.Duration `mapstructure:"slow_query_threshold"`
}

func (d Database) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode, d.TimeZone,
	)
}

func (d Database) Redacted() string {
	return fmt.Sprintf("postgres://%s:***@%s:%d/%s?sslmode=%s", d.User, d.Host, d.Port, d.Name, d.SSLMode)
}

type Log struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type CORS struct {
	AllowedOrigins   []string      `mapstructure:"allowed_origins"`
	AllowedMethods   []string      `mapstructure:"allowed_methods"`
	AllowedHeaders   []string      `mapstructure:"allowed_headers"`
	ExposedHeaders   []string      `mapstructure:"exposed_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
}

// AI holds the credentials and tuning for the two external providers:
// Gemini (reasoning) and Deepgram (speech-to-text / text-to-speech).
//
// Both are OPTIONAL. When a key is unset the corresponding feature reports
// itself as unavailable instead of preventing the API from booting.
type AI struct {
	GeminiAPIKey     string        `mapstructure:"gemini_api_key"`
	GeminiModel      string        `mapstructure:"gemini_model"`
	DeepgramAPIKey   string        `mapstructure:"deepgram_api_key"`
	DeepgramSTTModel string        `mapstructure:"deepgram_stt_model"`
	DeepgramTTSModel string        `mapstructure:"deepgram_tts_model"`
	RequestTimeout   time.Duration `mapstructure:"request_timeout"`
	MaxAudioBytes    int64         `mapstructure:"max_audio_bytes"`
}

// GeminiEnabled reports whether AI reasoning features can run.
func (a AI) GeminiEnabled() bool { return a.GeminiAPIKey != "" }

// VoiceEnabled reports whether speech transcription/synthesis can run.
func (a AI) VoiceEnabled() bool { return a.DeepgramAPIKey != "" }

// Load reads configuration from the environment and an optional .env file.
func Load(envPath string) (*Config, error) {
	if envPath != "" {
		if err := loadEnvFile(envPath); err != nil {
			return nil, err
		}
	}

	v := viper.New()
	setDefaults(v)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: parsing configuration: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("config: opening %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	values, err := gotenv.StrictParse(f)
	if err != nil {
		return fmt.Errorf("config: parsing %s: %w", path, err)
	}

	for k, val := range values {
		if _, exists := os.LookupEnv(k); exists {
			continue
		}
		if err := os.Setenv(k, val); err != nil {
			return fmt.Errorf("config: setting %s: %w", k, err)
		}
	}
	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "anchorOne")
	v.SetDefault("app.env", EnvDevelopment)
	v.SetDefault("app.auto_migrate", true)

	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 20080)
	// Generous read/write windows: check-in uploads carry audio, and a Gemini
	// round-trip (transcribe → reason) can legitimately take tens of seconds.
	v.SetDefault("http.read_timeout", 60*time.Second)
	v.SetDefault("http.write_timeout", 120*time.Second)
	v.SetDefault("http.idle_timeout", 60*time.Second)
	v.SetDefault("http.shutdown_timeout", 15*time.Second)
	v.SetDefault("http.trusted_proxies", []string{})

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 20543)
	v.SetDefault("database.user", "hackathon")
	v.SetDefault("database.password", "hackathon")
	v.SetDefault("database.name", "hackathon")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.timezone", "UTC")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", time.Hour)
	v.SetDefault("database.conn_max_idle_time", 10*time.Minute)
	v.SetDefault("database.log_level", "warn")
	v.SetDefault("database.slow_query_threshold", 200*time.Millisecond)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")

	v.SetDefault("cors.allowed_origins", []string{"http://localhost:20000"})
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowed_headers", []string{
		"Origin", "Content-Type", "Accept", "Authorization", "X-Request-Id",
	})
	v.SetDefault("cors.exposed_headers", []string{"X-Request-Id"})
	v.SetDefault("cors.allow_credentials", true)
	v.SetDefault("cors.max_age", 12*time.Hour)

	v.SetDefault("jwt.secret", "hackathon-jwt-secret-key-change-in-prod")
	v.SetDefault("jwt.access_token_ttl", 24*time.Hour)
	v.SetDefault("jwt.refresh_token_ttl", 7*24*time.Hour)
	// ---- ai (optional integrations; inert while their keys are unset) ----
	v.SetDefault("ai.gemini_api_key", "")
	v.SetDefault("ai.gemini_model", "gemini-3.5-flash")
	v.SetDefault("ai.deepgram_api_key", "")
	v.SetDefault("ai.deepgram_stt_model", "nova-2")
	v.SetDefault("ai.deepgram_tts_model", "aura-asteria-en")
	v.SetDefault("ai.request_timeout", 60*time.Second)
	v.SetDefault("ai.max_audio_bytes", 10<<20) // 10 MiB
}

func (c *Config) Validate() error {
	var problems []string

	switch c.App.Env {
	case EnvDevelopment, EnvProduction:
	default:
		problems = append(problems, fmt.Sprintf(
			"app.env must be one of %s, %s (got %q)",
			EnvDevelopment, EnvProduction, c.App.Env))
	}

	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		problems = append(problems, fmt.Sprintf("http.port must be between 1 and 65535 (got %d)", c.HTTP.Port))
	}
	if c.Database.Host == "" {
		problems = append(problems, "database.host is required")
	}
	if c.Database.Name == "" {
		problems = append(problems, "database.name is required")
	}
	// AI keys are intentionally NOT required — an unset key disables that
	// integration rather than blocking startup (see rule 3.8). Only the
	// tuning around them has to be sane.
	if c.AI.GeminiModel == "" {
		problems = append(problems, "ai.gemini_model is required")
	}
	if c.AI.DeepgramSTTModel == "" {
		problems = append(problems, "ai.deepgram_stt_model is required")
	}
	if c.AI.DeepgramTTSModel == "" {
		problems = append(problems, "ai.deepgram_tts_model is required")
	}
	if c.AI.RequestTimeout <= 0 {
		problems = append(problems, "ai.request_timeout must be greater than 0")
	}
	if c.AI.MaxAudioBytes <= 0 {
		problems = append(problems, "ai.max_audio_bytes must be greater than 0")
	}

	switch c.Log.Format {
	case "json", "console":
	default:
		problems = append(problems, fmt.Sprintf("log.format must be json or console (got %q)", c.Log.Format))
	}

	if c.App.IsProduction() {
		if c.App.AutoMigrate {
			problems = append(problems, "app.auto_migrate must be false in production — run migrations explicitly")
		}
		if c.Database.SSLMode == "disable" {
			problems = append(problems, "database.ssl_mode must not be disable in production")
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("config: invalid configuration:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}
