package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	cfg "github.com/thrillee/aegisbox/internal/config"
	"github.com/thrillee/aegisbox/internal/database"
	apihandlers "github.com/thrillee/aegisbox/internal/managerapi/handlers"
)

// RunManagerAPI runs the management REST API until ctx is cancelled or the server fails.
func RunManagerAPI(ctx context.Context, config *cfg.Config) error {
	appCtx, rootCancel := context.WithCancel(ctx)
	defer rootCancel()

	// A server failure stops the API and is returned to the caller.
	var runErr error

	// --- Database ---
	slog.Info("Connecting to database...")
	dbpool, err := pgxpool.New(appCtx, config.DatabaseURL)
	if err != nil {
		slog.Error("DB connect error", slog.Any("error", err))
		return fmt.Errorf("manager api startup: %w", err)
	}
	defer dbpool.Close()
	if err := dbpool.Ping(appCtx); err != nil {
		slog.Error("DB ping error", slog.Any("error", err))
		return fmt.Errorf("manager api startup: %w", err)
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
	apiV1 := router.Group("/api/v1") // Group routes under /api/v1
	if config.ManagerAPI.APIKey == "" {
		slog.Warn("MANAGER_API_KEY is not set: the manager API accepts unauthenticated requests, " +
			"so anyone who can reach it can credit wallets and send DLRs. Set MANAGER_API_KEY.")
	}
	apiV1.Use(apihandlers.APIKeyMiddleware(config.ManagerAPI.APIKey))
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
			runErr = err
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
	return runErr
}
