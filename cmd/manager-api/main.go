package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	cfg "github.com/thrillee/aegisbox/internal/config"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	apihandlers "github.com/thrillee/aegisbox/internal/managerapi/handlers"
)

func main() {
	appCtx, rootCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer rootCancel()

	// --- Config & Logging (reuse from other main.go) ---
	config, err := cfg.Load() // Load config (ensure it has API server addr)
	if err != nil {
		log.Fatalf("Config load error: %v", err)
	}
	setupLogging(config.LogLevel) // Setup slog

	// --- Database ---
	slog.Info("Connecting to database...")
	dbpool, err := pgxpool.New(appCtx, config.DatabaseURL)
	if err != nil {
		slog.Error("DB connect error", slog.Any("error", err))
		os.Exit(1)
	}
	defer dbpool.Close()
	if err := dbpool.Ping(appCtx); err != nil {
		slog.Error("DB ping error", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Database connection established")
	dbQueries := database.New(dbpool)

	// --- Gin Router Setup ---
	// gin.SetMode(gin.ReleaseMode) // Set for production
	router := gin.Default() // Includes basic middleware (logger, recovery)

	// Health Check Endpoint
	router.GET("/health", func(c *gin.Context) {
		// Check DB connection?
		if err := dbpool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "db": "error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Setup API V1 Routes
	apiV1 := router.Group("/api/v1")                  // Group routes under /api/v1
	apihandlers.SetupRoutes(apiV1, dbQueries, dbpool) // Pass Engine group, querier, pool

	// --- HTTP Server ---
	// TODO: Get API server address from config
	apiAddr := config.ManagerAPI.Addr // Example: Assume config struct has ManagerAPI section
	if apiAddr == "" {
		apiAddr = ":8081"
	} // Default API port

	srv := &http.Server{
		Addr:    apiAddr,
		Handler: router,
		// Add timeouts from config
		ReadTimeout:  config.ManagerAPI.ReadTimeout,
		WriteTimeout: config.ManagerAPI.WriteTimeout,
		IdleTimeout:  config.ManagerAPI.IdleTimeout,
		ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelWarn),
	}

	// Start Server
	go func() {
		slog.Info("Starting Management API Server", slog.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Management API ListenAndServe error", slog.Any("error", err))
			rootCancel() // Trigger shutdown on server error
		}
	}()

	// --- Wait for Shutdown ---
	<-appCtx.Done()
	slog.Info("Shutdown signal received for Management API server.")

	// --- Graceful Shutdown ---
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Management API server forced to shutdown", slog.Any("error", err))
	}

	slog.Info("Management API server stopped.")
}

// setupLogging function (reuse from other main.go)
func setupLogging(logLevelStr string) {
	logLevel := slog.LevelInfo
	if logLevelStr == "debug" {
		logLevel = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: logLevel, AddSource: logLevel <= slog.LevelDebug}
	baseHandler := slog.NewJSONHandler(os.Stdout, opts)
	contextHandler := logging.NewContextHandler(baseHandler) // Assuming this exists
	logger := slog.New(contextHandler)
	slog.SetDefault(logger)
}
