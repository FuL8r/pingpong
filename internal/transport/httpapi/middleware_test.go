package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRecoverer_TurnsPanicInto500(t *testing.T) {
	panicky := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})
	h := recoverer(silentLogger())(panicky)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if rec.Body.String() == "boom" {
		t.Error("recoverer must not leak the panic value to the client")
	}
}

func TestLimitInFlight_RejectsOverCapacity(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{})
	blocking := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		entered <- struct{}{}
		<-release
	})
	h := limitInFlight(1)(blocking)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	}()
	<-entered

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("over-capacity status = %d, want 503", rec.Code)
	}

	close(release)
	wg.Wait()

	rec = httptest.NewRecorder()
	free := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	limitInFlight(1)(free).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("after release status = %d, want 200", rec.Code)
	}
}

func TestStatusRecorder_CapturesStatusAndBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, status: http.StatusOK}

	sr.WriteHeader(http.StatusTeapot)
	n, err := io.WriteString(sr, "hello")
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	if sr.status != http.StatusTeapot {
		t.Errorf("status = %d, want 418", sr.status)
	}
	if sr.bytes != n {
		t.Errorf("bytes = %d, want %d", sr.bytes, n)
	}
}
