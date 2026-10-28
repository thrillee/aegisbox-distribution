package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/thrillee/aegisbox/internal/app"
	"github.com/thrillee/aegisbox/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	app.SetupLogging(cfg.LogLevel)

	if err := app.RunGateway(ctx, cfg); err != nil {
		slog.Error("SMPP gateway stopped with an error", slog.Any("error", err))
		os.Exit(1)
	}
}
