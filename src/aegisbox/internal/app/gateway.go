package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/config"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/httpserver"
	"github.com/thrillee/aegisbox/internal/mno"
	"github.com/thrillee/aegisbox/internal/mno/http_providers"
	"github.com/thrillee/aegisbox/internal/notification"
	"github.com/thrillee/aegisbox/internal/smppserver"
	"github.com/thrillee/aegisbox/internal/sms"
	"github.com/thrillee/aegisbox/internal/sp"
	"github.com/thrillee/aegisbox/internal/wallet"
	"github.com/thrillee/aegisbox/pkg/segmenter"
)

// RunGateway runs the SMPP gateway (SMPP server, SP HTTP API, MNO DLR server and
// background workers) until ctx is cancelled or one of its servers fails.
func RunGateway(ctx context.Context, cfg *config.Config) error {
	appCtx, rootCancel := context.WithCancel(ctx)
	defer rootCancel()
	logger := slog.Default()

	// A server failure stops the whole gateway and is returned to the caller.
	var runErr error
	var failOnce sync.Once
	fail := func(err error) {
		failOnce.Do(func() { runErr = err })
		rootCancel()
	}

	// --- Database ---
	slog.Info("Connecting to database...")
	dbpool, err := pgxpool.New(appCtx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Unable to connect to database", slog.Any("error", err))
		return fmt.Errorf("smpp gateway startup: %w", err)
	}
	defer dbpool.Close()
	if err := dbpool.Ping(appCtx); err != nil {
		slog.Error("Failed to ping database", slog.Any("error", err))
		return fmt.Errorf("smpp gateway startup: %w", err)
	}
	slog.Info("Database connection pool established")
	dbQueries := database.New(dbpool) // sqlc concrete implementation

	// --- Initialize Core Services & Dependencies ---
	slog.Info("Initializing services...")
	mainSegmenter := segmenter.NewDefaultSegmenter()
	walletSvc := wallet.NewService(dbpool, dbQueries) // Use concrete impl for now
	notifier := notification.NewLogNotifier()

	otpPreprocessor := sms.NewOtpScopePreprocessor(dbQueries, cfg.OtpScopePreprocessorConfig)

	incomingMessageHandler := sms.NewDefaultIncomingMessageHandler(
		dbQueries,
		[]sp.MessagePreprocessor{
			otpPreprocessor,
		},
	)

	// --- Initialize Health Trackers for Intelligent Routing ---
	slog.Info("Initializing health trackers...")
	senderHealthTracker := sms.NewSenderHealthTracker(dbQueries, logger)
	routeHealthTracker := sms.NewRouteHealthTracker(dbQueries, logger)

	// Create processor dependencies - forwarder factory creation is deferred
	processorDeps := sms.ProcessorDependencies{
		DBPool:              dbpool,
		DBQueries:           dbQueries,
		Router:              sms.NewIntelligentRouter(dbQueries, cfg.DefaultRoutingGroupRef, senderHealthTracker),
		Validator:           sms.NewDefaultValidator(dbQueries),
		Pricer:              sms.NewDefaultPricer(dbpool),
		WalletService:       walletSvc,
		DLRForwarder:        nil, // Will be set on the factory used by the worker
		SenderHealthTracker: senderHealthTracker,
		RouteHealthTracker:  routeHealthTracker,
		// QueueClient: // Not using queue client in this version
	}
	smsProcessor := sms.NewProcessor(processorDeps) // smsProcessor now holds core logic

	httpClient := &http.Client{Timeout: 20 * time.Second}
	httpForwarder := sp.NewHTTPSPForwarder(
		sp.HTTPForwarderConfig{Timeout: 15 * time.Second},
		httpClient,
	)

	// --- Initialize MNO Connection Manager ---
	mnoManager := mno.NewManager(
		dbpool,
		dbQueries,
		mainSegmenter,
		smsProcessor.UpdateSegmentDLRStatus,
		cfg.MNOClientConfig.SyncInterval,
		logger)

	// --- Initialize SMS Processor (Needs initialized dependencies) ---
	smsProcessor.SetSender(sms.NewDefaultSender(dbQueries, mnoManager, mainSegmenter))

	// --- Initialize SMPP Server (Implements Session Manager) ---
	smppServer := smppserver.NewServer(
		appCtx,
		cfg.ServerConfig,
		dbQueries,
		incomingMessageHandler,
	) // Pass core msg handler

	// --- Initialize DLR Forwarders & Factory ---
	var forwarderFactory sms.DLRForwarderFactory
	smppForwarder := smppserver.NewSMPSPForwarder(smppServer)
	forwarderFactoryImpl, err := sms.NewMapDLRForwarderFactory(smppForwarder, httpForwarder)
	if err != nil {
		slog.Error("Failed to create DLR Forwarder Factory", slog.Any("error", err))
		return fmt.Errorf("smpp gateway startup: %w", err)
	}
	forwarderFactory = forwarderFactoryImpl // Assign to the interface variable used by worker
	dlrForwarderWorker := sms.NewDLRWorker(dbQueries, forwarderFactory)

	// --- Initialize OTP Sender Reset Worker ---
	otpSenderResetWorker := sms.NewOtpSenderResetWorker(dbQueries)

	// --- Initialize Worker Manager ---
	// Pass the *factory* to the worker manager, which passes it to the DLR worker
	workerManager := sms.NewManager(
		dbpool,
		dbQueries,
		smsProcessor, // Processor reference may not be needed if workers use DB queue/services
		dlrForwarderWorker,
		otpSenderResetWorker,
		notifier,
		cfg.WorkerConfig,
	)

	// --- Initialize Intelligent Routing Workers ---
	slog.Info("Initializing intelligent routing workers...")

	// DLR Timeout Worker - marks segments with missing DLRs as "no_dlr"
	dlrTimeoutWorker := sms.NewDLRTimeoutWorker(
		dbQueries,
		senderHealthTracker,
		logger,
		sms.DLRTimeoutWorkerConfig{
			Interval:       sms.DLRTimeoutCheckInterval,  // Run every minute
			TimeoutSeconds: sms.DefaultDLRTimeoutSeconds, // Fallback; the global routing config's DLR timeout wins
			BatchSize:      1000,            // Process 1000 segments per batch
		},
	)

	// Auto DLR Worker - completes messages of SPs with auto DLR enabled when the MNO sends no DLR
	autoDLRService := sms.NewDLRService(sms.DLRServiceDependencies{
		DBQueries:     dbQueries,
		WalletService: walletSvc,
	})
	autoDLRWorker := sms.NewAutoDLRWorker(
		dbQueries,
		autoDLRService,
		logger,
		sms.AutoDLRWorkerConfig{
			Interval:  1 * time.Minute,    // Run every minute
			BatchSize: 500,                // Fetch 500 messages per batch
			MaxAge:    7 * 24 * time.Hour, // Ignore segments sent more than 7 days ago
		},
	)

	// Sender Auto-Unblock Worker - automatically unblocks inactive senders after 7 days
	senderAutoUnblockWorker := sms.NewSenderAutoUnblockWorker(
		dbQueries,
		logger,
		sms.SenderAutoUnblockWorkerConfig{
			Interval: sms.SenderAutoUnblockCheckInterval, // Run every 6 hours
		},
	)

	// Ring Buffer Cleanup Worker - cleans up old delivery attempt records
	ringBufferCleanupWorker := sms.NewRingBufferCleanupWorker(
		dbQueries,
		logger,
		sms.RingBufferCleanupWorkerConfig{
			Interval:      sms.RingBufferCleanupInterval, // Run daily
			RetentionDays: sms.RingBufferRetentionDays,   // Keep 90 days of history
		},
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

	// Start Intelligent Routing Workers
	wg.Add(1)
	go func() {
		defer wg.Done()
		dlrTimeoutWorker.Start(appCtx)
		slog.Info("DLR Timeout Worker stopped.")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		autoDLRWorker.Start(appCtx)
		slog.Info("Auto DLR Worker stopped.")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		senderAutoUnblockWorker.Start(appCtx)
		slog.Info("Sender Auto-Unblock Worker stopped.")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ringBufferCleanupWorker.Start(appCtx)
		slog.Info("Ring Buffer Cleanup Worker stopped.")
	}()

	// Start SMPP Server (listens for SPs)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := smppServer.ListenAndServe(); err != nil {
			slog.Error("SMPP Server failed", slog.Any("error", err))
			fail(err) // Signal other components to stop
		}
		slog.Info("SMPP Server stopped.")
	}()

	// Start HTTP Server (listens for SPs)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP Server failed", slog.Any("error", err))
			fail(err) // Signal other components to stop
		}
		slog.Info("HTTP Server stopped.")
	}()

	// Start DLR Server (listens for SPs)
	var dlrServer *mno.DLRServer
	wg.Add(1)
	go func() {
		defer wg.Done()
		if dlrServer, err = mno.StartDLRServer(smsProcessor.UpdateSegmentDLRStatus, cfg.MNOClientConfig.DLRAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("DLR HTTP Server failed", slog.Any("error", err))
			fail(err) // Signal other components to stop
		}
		slog.Info("DLR HTTP Server stopped.")
	}()

	// --- Wait for Shutdown Signal ---
	<-appCtx.Done() // Wait for shutdown (SIGINT/SIGTERM) or a server failure
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
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		if dlrServer == nil {
			return // The DLR server never started
		}
		slog.Info("Shutting down HTTP DLR Server...")
		if err := dlrServer.Shutdown(shutdownCtx); err != nil {
			slog.Warn("Error during HTTP DLR Server shutdown", slog.Any("error", err))
		} else {
			slog.Info("HTTP DLR Server shutdown complete.")
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
	slog.Info("SMPP gateway gracefully stopped.")
	return runErr
}

func getHTTPProviderConfig(dbQueries database.Querier, providerName string) *http_providers.HTTPProviderConfig {
	return nil
}
