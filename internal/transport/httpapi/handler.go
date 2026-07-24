// Package httpapi is the delivery layer, adapts HTTP requests to the domain Service
package httpapi

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/Tuna/pingpong/internal/pingpong"
)

const (
	notBadHeader = "NotBad"
	notBadValue  = "true"
)

type Options struct {
	MaxBodyBytes int64
	MaxInFlight  int
	TLSEnabled   bool
}

type Handler struct {
	svc          pingpong.Service
	maxBodyBytes int64
}

// NewHandler constructs a Handler for the given service and request-body limit
func NewHandler(svc pingpong.Service, maxBodyBytes int64) *Handler {
	return &Handler{svc: svc, maxBodyBytes: maxBodyBytes}
}

// Routes returns the strict router
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /{$}", h.handlePost)
	return mux
}

// handlePost безопасно принимает данные от клиента, проверяет размер и передает в бизнес слой
func (h *Handler) handlePost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodyBytes)
	if _, err := io.Copy(io.Discard, r.Body); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "payload too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	cmd := pingpong.Command{NotBad: r.Header.Get(notBadHeader) == notBadValue}
	resp := h.svc.Respond(cmd)

	switch resp.Decision {
	case pingpong.Allowed:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, resp.Body)
	default:
		http.Error(w, "forbidden", http.StatusForbidden)
	}
}

// NewRouter builds the fully wired HTTP handler
func NewRouter(svc pingpong.Service, opts Options, logger *slog.Logger) http.Handler {
	h := NewHandler(svc, opts.MaxBodyBytes)
	return chain(h.Routes(),
		logging(logger),
		recoverer(logger),
		securityHeaders(opts.TLSEnabled),
		limitInFlight(opts.MaxInFlight),
	)
}
