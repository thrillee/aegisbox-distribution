package sms

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/config"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/notification"
	"github.com/thrillee/aegisbox/internal/workers"
)

// WorkerConfig holds configuration for worker intervals and batch sizes.
type WorkerConfig struct {
	RoutingInterval        time.Duration
	PricingInterval        time.Duration
	SendingInterval        time.Duration
	LowBalanceInterval     time.Duration
	RoutingBatchSize       int
	PricingBatchSize       int
	SendingBatchSize       int
	DLRForwardingBatchSize int
	DLRForwarderInterval   time.Duration
}

// Manager orchestrates the background worker loops.
type Manager struct {
	dbpool             *pgxpool.Pool
	dbQueries          database.Querier
	smsProcessor       *Processor            // Holds the SMS processing logic
	dlrForwarderWorker *DLRWorker            // For forwarding DLRs
	notifier           notification.Notifier // For low balance notifications
	workerConfig       config.WorkerConfig
}

func NewManager(pool *pgxpool.Pool, queries database.Querier, processor *Processor, dlrForwarderWorker *DLRWorker, notifier notification.Notifier, cfg config.WorkerConfig) *Manager {
	return &Manager{
		dbpool:             pool,
		dbQueries:          queries, // Keep queries for low balance check for now
		smsProcessor:       processor,
		notifier:           notifier,
		workerConfig:       cfg,
		dlrForwarderWorker: dlrForwarderWorker,
	}
}

// StartSMSProcessing launches worker loops for different SMS processing stages.
func (m *Manager) StartSMSProcessing(ctx context.Context) {
	log.Println("Starting SMS processing workers...")
	go workers.RunWorkerLoop(ctx, "SMS-Routing", m.workerConfig.RoutingInterval, m.workerConfig.RoutingBatchSize, m.smsProcessor.ProcessRoutingStep)
	go workers.RunWorkerLoop(ctx, "SMS-Pricing", m.workerConfig.PricingInterval, m.workerConfig.PricingBatchSize, m.smsProcessor.ProcessPricingStep)
	go workers.RunWorkerLoop(ctx, "SMS-Sending", m.workerConfig.SendingInterval, m.workerConfig.SendingBatchSize, m.smsProcessor.ProcessSendingStep)
	go workers.RunWorkerLoop(ctx, "SMS-DLR-FORWARDER",
		m.workerConfig.DLRForwarderInterval,
		m.workerConfig.DLRForwardingBatchSize,
		m.dlrForwarderWorker.ProcessDLRForwardingBatch)
}

// StartLowBalanceNotifier launches the worker loop for checking low balances.
func (m *Manager) StartLowBalanceNotifier(ctx context.Context) {
	log.Println("Starting Low Balance Notifier worker...")
	go workers.RunWorkerLoop(ctx, "LowBalanceNotifier", m.workerConfig.LowBalanceInterval, 100, m.checkLowBalances) // Batch size not really applicable here
}

// checkLowBalances is the WorkerFunc for the low balance notifier.
func (m *Manager) checkLowBalances(ctx context.Context, _ int) (int, error) {
	wallets, err := m.dbQueries.GetWalletsBelowThreshold(ctx)
	if err != nil {
		return 0, err // Propagate error (runWorkerLoop handles ErrNoRows)
	}

	processedCount := 0
	for _, wallet := range wallets {
		balanceStr := "N/A"
		thresholdStr := "N/A"
		emailStr := "N/A"

		if wallet.Balance.IsPositive() {
			balanceStr = fmt.Sprintf("%v", wallet.Balance)
		}
		if wallet.LowBalanceThreshold.IsPositive() {
			thresholdStr = fmt.Sprintf("%v", wallet.LowBalanceThreshold)
		}
		if wallet.Email != "" {
			emailStr = wallet.Email
		}

		log.Printf("Low balance detected for SP %d (Email: %s), Wallet %d (%s): Balance=%s, Threshold=%s",
			wallet.ServiceProviderID, emailStr, wallet.WalletID, wallet.CurrencyCode, balanceStr, thresholdStr)

		notificationSubject := fmt.Sprintf("Low Balance Alert - %s Wallet", wallet.CurrencyCode)
		notificationBody := fmt.Sprintf("Your %s wallet balance (%s) is below the threshold (%s). Please top up.",
			wallet.CurrencyCode, balanceStr, thresholdStr)

		// Use the notifier interface
		err = m.notifier.Send(ctx, emailStr, notificationSubject, notificationBody)
		if err != nil {
			log.Printf("Failed to send low balance notification to %s for wallet %d: %v", emailStr, wallet.WalletID, err)
			continue // Don't update timestamp if notification failed
		}

		// Update Notification Timestamp
		err = m.dbQueries.UpdateLowBalanceNotifiedAt(ctx, wallet.WalletID)
		if err != nil {
			log.Printf("Failed to update low balance notified timestamp for wallet %d: %v", wallet.WalletID, err)
			// Continue processing other wallets even if this update fails
		} else {
			processedCount++ // Count successful notifications sent + timestamp updated
			log.Printf("Low balance notification sent successfully for wallet %d", wallet.WalletID)
		}
	}
	return processedCount, nil // Return number successfully notified
}

// TODO: Add Shutdown method for graceful termination if needed
// func (m *Manager) Shutdown(ctx context.Context) error { ... }
