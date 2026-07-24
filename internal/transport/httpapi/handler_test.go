package httpapi_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Tuna/pingpong/internal/pingpong"
	"github.com/Tuna/pingpong/internal/transport/httpapi"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testRouter(opts httpapi.Options) http.Handler {
	return httpapi.NewRouter(pingpong.NewService(), opts, discardLogger())
}

func defaultOpts() httpapi.Options {
	return httpapi.Options{MaxBodyBytes: 4096, MaxInFlight: 256, TLSEnabled: false}
}

func TestRouter_PostWithValidHeader_ReturnsReallyNotBad(t *testing.T) {
	router := testRouter(defaultOpts())

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("NotBad", "true")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != "ReallyNotBad" {
		t.Errorf("body = %q, want %q", got, "ReallyNotBad")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/plain; charset=utf-8", ct)
	}
}

func TestRouter_PostWithoutHeader_IsForbidden(t *testing.T) {
	router := testRouter(defaultOpts())

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "ReallyNotBad") {
		t.Errorf("denied response must not leak the allowed body, got %q", rec.Body.String())
	}
}

func TestRouter_PostWithWrongHeaderValue_IsForbidden(t *testing.T) {
	router := testRouter(defaultOpts())

	for _, val := range []string{"false", "TRUE", "True", "1", " true", ""} {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("NotBad", val)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("NotBad=%q: status = %d, want 403", val, rec.Code)
		}
	}
}

func TestRouter_NonPostMethod_IsMethodNotAllowed(t *testing.T) {
	router := testRouter(defaultOpts())

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		req := httptest.NewRequest(method, "/", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s /: status = %d, want 405", method, rec.Code)
		}
		if allow := rec.Header().Get("Allow"); !strings.Contains(allow, http.MethodPost) {
			t.Errorf("%s /: Allow = %q, want to contain POST", method, allow)
		}
	}
}

func TestRouter_UnknownPath_IsNotFound(t *testing.T) {
	router := testRouter(defaultOpts())

	req := httptest.NewRequest(http.MethodPost, "/anything", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestRouter_OversizedBody_IsRequestEntityTooLarge(t *testing.T) {
	opts := defaultOpts()
	opts.MaxBodyBytes = 16
	router := testRouter(opts)

	body := strings.NewReader(strings.Repeat("A", 1024))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("NotBad", "true")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "ReallyNotBad") {
		t.Errorf("oversized request must not be answered with the allowed body")
	}
}

func TestRouter_SecurityHeaders_Present(t *testing.T) {
	router := testRouter(defaultOpts())

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("NotBad", "true")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	want := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"Content-Security-Policy": "default-src 'none'",
		"Referrer-Policy":         "no-referrer",
		"Cache-Control":           "no-store",
	}
	for h, v := range want {
		if got := rec.Header().Get(h); got != v {
			t.Errorf("header %s = %q, want %q", h, got, v)
		}
	}
	if hsts := rec.Header().Get("Strict-Transport-Security"); hsts != "" {
		t.Errorf("HSTS must be absent without TLS, got %q", hsts)
	}
	if srv := rec.Header().Get("Server"); srv != "" {
		t.Errorf("Server header must not leak, got %q", srv)
	}
}

func TestRouter_HSTS_PresentWhenTLSEnabled(t *testing.T) {
	opts := defaultOpts()
	opts.TLSEnabled = true
	router := testRouter(opts)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("NotBad", "true")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if hsts := rec.Header().Get("Strict-Transport-Security"); hsts == "" {
		t.Error("HSTS must be present when TLS is enabled")
	}
}
