package config_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/Tuna/pingpong/internal/config"
)

func envFrom(m map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

func TestLoad_Defaults(t *testing.T) {
	cfg, err := config.Load(envFrom(map[string]string{}))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Addr != ":8089" {
		t.Errorf("Addr = %q, want %q", cfg.Addr, ":8089")
	}
	if cfg.MaxBodyBytes != 4096 {
		t.Errorf("MaxBodyBytes = %d, want 4096", cfg.MaxBodyBytes)
	}
	if cfg.ReadHeaderTimeout != 5*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want 5s", cfg.ReadHeaderTimeout)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 10s", cfg.ShutdownTimeout)
	}
	if cfg.MaxInFlight != 256 {
		t.Errorf("MaxInFlight = %d, want 256", cfg.MaxInFlight)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want info", cfg.LogLevel)
	}
	if cfg.TLSEnabled() {
		t.Error("TLSEnabled() = true, want false when no cert/key set")
	}
}

func TestLoad_Overrides(t *testing.T) {
	cfg, err := config.Load(envFrom(map[string]string{
		"PINGPONG_ADDR":                "127.0.0.1:9000",
		"TLS_CERT":                     "/tmp/cert.pem",
		"TLS_KEY":                      "/tmp/key.pem",
		"PINGPONG_MAX_BODY_BYTES":      "8192",
		"PINGPONG_READ_HEADER_TIMEOUT": "3s",
		"PINGPONG_MAX_INFLIGHT":        "10",
		"PINGPONG_LOG_LEVEL":           "debug",
	}))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Addr != "127.0.0.1:9000" {
		t.Errorf("Addr = %q, want 127.0.0.1:9000", cfg.Addr)
	}
	if !cfg.TLSEnabled() {
		t.Error("TLSEnabled() = false, want true when cert and key set")
	}
	if cfg.MaxBodyBytes != 8192 {
		t.Errorf("MaxBodyBytes = %d, want 8192", cfg.MaxBodyBytes)
	}
	if cfg.ReadHeaderTimeout != 3*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want 3s", cfg.ReadHeaderTimeout)
	}
	if cfg.MaxInFlight != 10 {
		t.Errorf("MaxInFlight = %d, want 10", cfg.MaxInFlight)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}
}

func TestLoad_Errors(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{"cert without key", map[string]string{"TLS_CERT": "/tmp/cert.pem"}},
		{"key without cert", map[string]string{"TLS_KEY": "/tmp/key.pem"}},
		{"non-numeric body limit", map[string]string{"PINGPONG_MAX_BODY_BYTES": "big"}},
		{"zero body limit", map[string]string{"PINGPONG_MAX_BODY_BYTES": "0"}},
		{"negative body limit", map[string]string{"PINGPONG_MAX_BODY_BYTES": "-1"}},
		{"invalid duration", map[string]string{"PINGPONG_READ_TIMEOUT": "soon"}},
		{"zero in-flight", map[string]string{"PINGPONG_MAX_INFLIGHT": "0"}},
		{"invalid log level", map[string]string{"PINGPONG_LOG_LEVEL": "screaming"}},
		{"empty addr", map[string]string{"PINGPONG_ADDR": ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := tt.env
			get := func(key string) (string, bool) {
				v, ok := env[key]
				return v, ok
			}
			if _, err := config.Load(get); err == nil {
				t.Fatalf("Load(%v) = nil error, want error", tt.env)
			}
		})
	}
}
