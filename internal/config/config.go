//config file, read params from env variables

package config

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr              string
	TLSCert           string
	TLSKey            string
	MaxBodyBytes      int64
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxInFlight       int
	LogLevel          slog.Level
}

// TLSEnabled check enabled TLS
func (c Config) TLSEnabled() bool {
	return c.TLSCert != "" && c.TLSKey != ""
}

// Get values from env
type getenv func(key string) (value string, ok bool)

// Load params from config to memory
func Load(get getenv) (Config, error) {
	cfg := Config{
		Addr:              lookupString(get, "PINGPONG_ADDR", ":8089"),
		TLSCert:           lookupString(get, "TLS_CERT", ""),
		TLSKey:            lookupString(get, "TLS_KEY", ""),
		MaxBodyBytes:      4096,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		MaxInFlight:       256,
		LogLevel:          slog.LevelInfo,
	}

	var err error
	if cfg.MaxBodyBytes, err = lookupInt64(get, "PINGPONG_MAX_BODY_BYTES", cfg.MaxBodyBytes); err != nil {
		return Config{}, err
	}
	if cfg.MaxInFlight, err = lookupInt(get, "PINGPONG_MAX_INFLIGHT", cfg.MaxInFlight); err != nil {
		return Config{}, err
	}
	for _, d := range []struct {
		key string
		dst *time.Duration
	}{
		{"PINGPONG_READ_HEADER_TIMEOUT", &cfg.ReadHeaderTimeout},
		{"PINGPONG_READ_TIMEOUT", &cfg.ReadTimeout},
		{"PINGPONG_WRITE_TIMEOUT", &cfg.WriteTimeout},
		{"PINGPONG_IDLE_TIMEOUT", &cfg.IdleTimeout},
		{"PINGPONG_SHUTDOWN_TIMEOUT", &cfg.ShutdownTimeout},
	} {
		if *d.dst, err = lookupDuration(get, d.key, *d.dst); err != nil {
			return Config{}, err
		}
	}
	if cfg.LogLevel, err = lookupLevel(get, "PINGPONG_LOG_LEVEL", cfg.LogLevel); err != nil {
		return Config{}, err
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// validate config params
func (c Config) validate() error {
	if strings.TrimSpace(c.Addr) == "" {
		return fmt.Errorf("config: PINGPONG_ADDR must not be empty")
	}
	if c.MaxBodyBytes <= 0 {
		return fmt.Errorf("config: PINGPONG_MAX_BODY_BYTES must be > 0, got %d", c.MaxBodyBytes)
	}
	if c.MaxInFlight <= 0 {
		return fmt.Errorf("config: PINGPONG_MAX_INFLIGHT must be > 0, got %d", c.MaxInFlight)
	}
	if (c.TLSCert == "") != (c.TLSKey == "") {
		return fmt.Errorf("config: TLS_CERT and TLS_KEY must be set together")
	}
	return nil
}

// lookupString get data and return in string type
func lookupString(get getenv, key, def string) string {
	if v, ok := get(key); ok {
		return v
	}
	return def
}

// lookupInt64 get data and return in Int64 type
func lookupInt64(get getenv, key string, def int64) (int64, error) {
	v, ok := get(key)
	if !ok {
		return def, nil
	}
	n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("config: %s: invalid integer %q: %w", key, v, err)
	}
	return n, nil
}

// lookupInt get data and return in Int type
func lookupInt(get getenv, key string, def int) (int, error) {
	n, err := lookupInt64(get, key, int64(def))
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// lookupDuration get data and return in time.Duration  >= 0
func lookupDuration(get getenv, key string, def time.Duration) (time.Duration, error) {
	v, ok := get(key)
	if !ok {
		return def, nil
	}
	d, err := time.ParseDuration(strings.TrimSpace(v))
	if err != nil {
		return 0, fmt.Errorf("config: %s: invalid duration %q: %w", key, v, err)
	}
	if d < 0 {
		return 0, fmt.Errorf("config: %s: duration must not be negative, got %v", key, d)
	}
	return d, nil
}

// lookupLevel get data and return log severity
func lookupLevel(get getenv, key string, def slog.Level) (slog.Level, error) {
	v, ok := get(key)
	if !ok {
		return def, nil
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("config: %s: unknown log level %q", key, v)
	}
}
