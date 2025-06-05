package main

import (
	"context"
	"errors"
	"log"
	"log/slog" // Use slog
	"net/http" // Needed for http.Client injection maybe
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/config"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/httpserver"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/mno"
	"github.com/thrillee/aegisbox/internal/notification"
	"github.com/thrillee/aegisbox/internal/smppserver"
	"github.com/thrillee/aegisbox/internal/sms"
	"github.com/thrillee/aegisbox/internal/sp"
	"github.com/thrillee/aegisbox/internal/wallet"
	"github.com/thrillee/aegisbox/pkg/segmenter"
)

func main() {
	// --- Context and Basic Setup ---
	appCtx, rootCancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer rootCancel()

	// --- Configuration ---
	cfg, err := config.Load()
	if err != nil {
		// Use standard log before slog is configured
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// --- Setup Logging ---
	logLevel := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: logLevel <= slog.LevelDebug,
	} // Add source location for debug
	baseHandler := slog.NewJSONHandler(os.Stdout, opts)
	contextHandler := logging.NewContextHandler(baseHandler)
	logger := slog.New(contextHandler)
	slog.SetDefault(logger)
	slog.Info("Logging initialized", "level", logLevel.String())

	// --- Database ---
	slog.Info("Connecting to database...")
	dbpool, err := pgxpool.New(appCtx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Unable to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer dbpool.Close()
	if err := dbpool.Ping(appCtx); err != nil {
		slog.Error("Failed to ping database", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Database connection pool established")
	dbQueries := database.New(dbpool) // sqlc concrete implementation

	// --- Initialize Core Services & Dependencies ---
	slog.Info("Initializing services...")
	mainSegmenter := segmenter.NewDefaultSegmenter()
	walletSvc := wallet.NewService(dbpool, dbQueries) // Use concrete impl for now
	notifier := notification.NewLogNotifier()

	routingPreprocessor := sms.NewRoutingPreprocessor(dbQueries, cfg.DefaultRoutingGroupRef)
	otpPreprocessor := sms.NewOtpScopePreprocessor(dbQueries, cfg.OtpScopePreprocessorConfig)

	incomingMessageHandler := sms.NewDefaultIncomingMessageHandler(
		dbQueries,
		[]sp.MessagePreprocessor{
			routingPreprocessor,
			otpPreprocessor,
		},
	)

	// Create processor dependencies - forwarder factory creation is deferred
	processorDeps := sms.ProcessorDependencies{
		DBPool:        dbpool,
		DBQueries:     dbQueries,
		Router:        sms.NewDefaultRouter(dbQueries, cfg.DefaultRoutingGroupRef),
		Validator:     sms.NewDefaultValidator(dbQueries),
		Pricer:        sms.NewDefaultPricer(dbpool),
		WalletService: walletSvc,
		DLRForwarder:  nil, // Will be set on the factory used by the worker
		// QueueClient: // Not using queue client in this version
	}
	smsProcessor := sms.NewProcessor(processorDeps) // smsProcessor now holds core logic

	httpClient := &http.Client{Timeout: 20 * time.Second}
	httpForwarder := sp.NewHTTPSPForwarder(
		sp.HTTPForwarderConfig{Timeout: 15 * time.Second},
		httpClient,
	)

	// --- DLR Forwarder Factory (Needs initialized forwarders) ---
	// We defer SMPSPForwarder initialization slightly
	var forwarderFactory sms.DLRForwarderFactory

	// --- Initialize MNO Connection Manager ---
	mnoManager := mno.NewManager(
		dbpool,
		dbQueries,
		mainSegmenter,
		smsProcessor.UpdateSegmentDLRStatus,
		cfg.MNOClientConfig.SyncInterval) // Pass DLR handler func from sms.Processor

	// --- Initialize SMS Processor (Needs initialized dependencies) ---
	smsProcessor.SetSender(sms.NewDefaultSender(dbQueries, mnoManager, mainSegmenter))

	// --- Initialize SMPP Server (Implements Session Manager) ---
	smppServer := smppserver.NewServer(
		appCtx,
		cfg.ServerConfig,
		dbQueries,
		incomingMessageHandler,
	) // Pass core msg handler

	// --- Initialize DLR Forwarders & Factory (Now that Session Manager exists) ---
	smppForwarder := smppserver.NewSMPSPForwarder(smppServer)
	forwarderFactoryImpl, err := sms.NewMapDLRForwarderFactory(smppForwarder, httpForwarder)
	if err != nil {
		slog.Error("Failed to create DLR Forwarder Factory", slog.Any("error", err))
		os.Exit(1)
	}
	forwarderFactory = forwarderFactoryImpl // Assign to the interface variable used by worker
	dlrForwarderWorker := sms.NewDLRWorker(dbQueries, forwarderFactory)

	// --- Initialize Worker Manager ---
	// Pass the *factory* to the worker manager, which passes it to the DLR worker
	workerManager := sms.NewManager(
		dbpool,
		dbQueries,
		smsProcessor, // Processor reference may not be needed if workers use DB queue/services
		dlrForwarderWorker,
		notifier,
		cfg.WorkerConfig,
	)

	// --- Initialize HTTP Server ---
	httpServer := httpserver.NewServer(
		cfg.HttpConfig,
		dbQueries,
		incomingMessageHandler,
	) // Pass core msg handler

	// --- Start Components Concurrently ---
	var wg sync.WaitGroup // Use waitgroup for graceful shutdown

	slog.Info("Starting application components...")

	// Start MNO Client Manager (connects to MNOs)
	wg.Add(1)
	go func() {
		defer wg.Done()
		mnoManager.Start(appCtx) // Manages connections based on DB
		slog.Info("MNO Manager stopped.")
	}()

	// Start Worker Loops (DLR forwarding, low balance etc.)
	wg.Add(1)
	go func() {
		defer wg.Done()
		workerManager.StartSMSProcessing(appCtx)
		slog.Info("Worker Manager stopped.")
	}()

	// Start SMPP Server (listens for SPs)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := smppServer.ListenAndServe(); err != nil {
			slog.Error("SMPP Server failed", slog.Any("error", err))
			rootCancel() // Signal other components to stop
		}
		slog.Info("SMPP Server stopped.")
	}()

	// Start HTTP Server (listens for SPs)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP Server failed", slog.Any("error", err))
			rootCancel() // Signal other components to stop
		}
		slog.Info("HTTP Server stopped.")
	}()

	// --- Wait for Shutdown Signal ---
	<-appCtx.Done() // Wait for context cancellation (SIGINT/SIGTERM)
	slog.Info("Shutdown signal received, initiating graceful shutdown...")

	// --- Graceful Shutdown Sequence ---
	// Create a context with timeout for shutdown operations
	shutdownTimeout := 20 * time.Second
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Stop accepting new connections first (Servers)
	var shutdownWg sync.WaitGroup

	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		slog.Info("Shutting down HTTP Server...")
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Warn("Error during HTTP Server shutdown", slog.Any("error", err))
		} else {
			slog.Info("HTTP Server shutdown complete.")
		}
	}()

	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		slog.Info("Shutting down SMPP Server...")
		if err := smppServer.Shutdown(shutdownCtx); err != nil {
			slog.Warn("Error during SMPP Server shutdown", slog.Any("error", err))
		} else {
			slog.Info("SMPP Server shutdown complete.")
		}
	}()

	// Wait for servers to stop accepting new requests
	shutdownWg.Wait()
	slog.Info("Servers stopped accepting new connections.")

	// Shutdown other components (Workers, MNO Manager)
	// Worker manager shutdown should happen before MNO manager if workers depend on MNO connections
	slog.Info("Shutting down Worker Manager...")
	// Add workerManager.Shutdown(shutdownCtx) if implemented

	slog.Info("Shutting down MNO Client Manager...")
	if err := mnoManager.Shutdown(shutdownCtx); err != nil {
		slog.Warn("Error during MNO Client Manager shutdown", slog.Any("error", err))
	} else {
		slog.Info("MNO Client Manager shutdown complete.")
	}

	// Wait for background goroutines managed by main waitgroup
	slog.Info("Waiting for main application goroutines to stop...")
	wg.Wait()

	slog.Info("Closing database pool...")
	dbpool.Close()
	slog.Info("Application gracefully stopped.")
}
