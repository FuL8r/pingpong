package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tuna/pingpong/internal/config"
	"github.com/Tuna/pingpong/internal/pingpong"
	"github.com/Tuna/pingpong/internal/server"
	"github.com/Tuna/pingpong/internal/transport/httpapi"
)

func main() {
	if err := run(); err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("fatal", "err", err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	svc := pingpong.NewService()
	router := httpapi.NewRouter(svc, httpapi.Options{
		MaxBodyBytes: cfg.MaxBodyBytes,
		MaxInFlight:  cfg.MaxInFlight,
		TLSEnabled:   cfg.TLSEnabled(),
	}, logger)

	srv := server.New(cfg, router, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("starting pingpong", "addr", cfg.Addr, "tls", cfg.TLSEnabled())

	if err := srv.Run(ctx); err != nil {
		return fmt.Errorf("server: %w", err)
	}
	logger.Info("stopped cleanly")
	return nil
}
