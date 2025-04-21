package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/config"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/notification"
	"github.com/thrillee/aegisbox/internal/sms"
	"github.com/thrillee/aegisbox/internal/workers"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	dbpool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()
	if err := dbpool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection pool established")

	dbQueries := database.New(dbpool) // sqlc concrete implementation

	logLevel := slog.LevelInfo // Default
	if cfg.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: logLevel}
	// Choose base handler (Text or JSON)
	// baseHandler := slog.NewTextHandler(os.Stdout, opts)
	baseHandler := slog.NewJSONHandler(os.Stdout, opts) // JSON is better for parsing

	// Wrap with context handler
	contextHandler := logging.NewContextHandler(baseHandler)

	// Set as default logger
	logger := slog.New(contextHandler)
	slog.SetDefault(logger)

	slog.Info("Logging initialized", "level", logLevel.String())

	// --- Initialize Services ---
	mnoClientManager := smppclient.NewManager(cfg.MNOClientConfig, dbQueries) // Needs implementation
	go mnoClientManager.Start(ctx)                                            // Needs implementation

	smsProcessor := sms.NewProcessor(dbpool, dbQueries, mnoClientManager) // Create the SMS Processor

	notifier := notification.NewLogNotifier() // Choose your notifier implementation

	workerManager := workers.NewManager(dbpool, dbQueries, smsProcessor, notifier, cfg.WorkerConfig) // Pass processor & notifier

	// --- Start Components ---
	go workerManager.StartSMSProcessing(ctx)      // Start SMS worker loops
	go workerManager.StartLowBalanceNotifier(ctx) // Start low balance checks

	server := smppserver.NewServer(cfg.ServerConfig, dbpool, dbQueries) // Pass pool and queries
	go func() {
		log.Printf("Starting SMPP Server on %s", cfg.ServerConfig.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("SMPP Server error: %v", err)
			cancel()
		}
	}()

	// --- Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown signal received, shutting down gracefully...")
	cancel()

	_, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	// Add actual shutdown calls here if implemented
	// server.Shutdown(shutdownCtx)
	// mnoClientManager.Shutdown(shutdownCtx)

	log.Println("Closing database pool...")
	dbpool.Close()
	log.Println("Server gracefully stopped")
}
