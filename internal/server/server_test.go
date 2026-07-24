package server

import (
	"context"
	"crypto/tls"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/Tuna/pingpong/internal/config"
)

func testConfig() config.Config {
	return config.Config{
		Addr:              "127.0.0.1:0",
		MaxBodyBytes:      4096,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   2 * time.Second,
		MaxInFlight:       16,
		LogLevel:          slog.LevelInfo,
	}
}

func silentLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func TestNew_AppliesHardenedTimeouts(t *testing.T) {
	cfg := testConfig()
	srv := New(cfg, http.NotFoundHandler(), silentLogger())

	if srv.http.ReadHeaderTimeout != cfg.ReadHeaderTimeout {
		t.Errorf("ReadHeaderTimeout = %v, want %v", srv.http.ReadHeaderTimeout, cfg.ReadHeaderTimeout)
	}
	if srv.http.ReadTimeout != cfg.ReadTimeout {
		t.Errorf("ReadTimeout = %v, want %v", srv.http.ReadTimeout, cfg.ReadTimeout)
	}
	if srv.http.WriteTimeout != cfg.WriteTimeout {
		t.Errorf("WriteTimeout = %v, want %v", srv.http.WriteTimeout, cfg.WriteTimeout)
	}
	if srv.http.IdleTimeout != cfg.IdleTimeout {
		t.Errorf("IdleTimeout = %v, want %v", srv.http.IdleTimeout, cfg.IdleTimeout)
	}
	if srv.http.MaxHeaderBytes != maxHeaderBytes {
		t.Errorf("MaxHeaderBytes = %d, want %d", srv.http.MaxHeaderBytes, maxHeaderBytes)
	}
	if srv.http.TLSConfig != nil {
		t.Error("TLSConfig must be nil when TLS is not configured")
	}
}

func TestNew_EnablesTLSConfigWhenConfigured(t *testing.T) {
	cfg := testConfig()
	cfg.TLSCert = "/tmp/cert.pem"
	cfg.TLSKey = "/tmp/key.pem"
	srv := New(cfg, http.NotFoundHandler(), silentLogger())

	if srv.http.TLSConfig == nil {
		t.Fatal("TLSConfig must be set when cert and key are configured")
	}
	if srv.http.TLSConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %x, want TLS 1.2", srv.http.TLSConfig.MinVersion)
	}
}

func TestTLSConfig_IsHardened(t *testing.T) {
	c := tlsConfig()
	if c.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %x, want TLS 1.2 (%x)", c.MinVersion, tls.VersionTLS12)
	}
	if len(c.CipherSuites) == 0 {
		t.Error("CipherSuites must be explicitly restricted")
	}
}

func TestServe_GracefulShutdownOnContextCancel(t *testing.T) {
	cfg := testConfig()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	})
	srv := New(cfg, handler, silentLogger())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.serve(ctx, ln) }()

	url := "http://" + ln.Addr().String() + "/"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("request to running server failed: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	http.DefaultClient.CloseIdleConnections()

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("graceful shutdown returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down within 5s")
	}
}
