package mno

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/pkg/codes"
	"github.com/thrillee/aegisbox/pkg/segmenter"
)

// Compile-time check to ensure Manager implements ConnectorProvider
var _ ConnectorProvider = (*Manager)(nil)

// Manager handles the lifecycle and access to MNO Connector instances.
type Manager struct {
	pool         *pgxpool.Pool
	dbQueries    database.Querier
	segmenter    segmenter.Segmenter // Required by SMPPConnector
	dlrHandler   DLRHandlerFunc      // The main DLR handler callback from sms.Processor
	syncInterval time.Duration       // How often to check DB for config changes

	connectors    sync.Map      // map[int32]Connector - Key: mno_connections.id (DB PK)
	mnoConnectors sync.Map      // map[int32][]int32 - Key: mno.id, Value: Slice of connection IDs for that MNO
	mu            sync.RWMutex  // Protects internal maps during sync
	stopCh        chan struct{} // Signals the manager loop to stop
	wg            sync.WaitGroup
}

// NewManager creates a new MNO Connection Manager.
func NewManager(pool *pgxpool.Pool, q database.Querier, seg segmenter.Segmenter, dlrHandler DLRHandlerFunc, syncInterval time.Duration) *Manager {
	if syncInterval <= 0 {
		syncInterval = 1 * time.Minute // Default sync interval
	}
	return &Manager{
		pool:         pool,
		dbQueries:    q,
		segmenter:    seg,
		dlrHandler:   dlrHandler,
		syncInterval: syncInterval,
		stopCh:       make(chan struct{}),
	}
}

// Start initiates the manager's main loop for syncing connectors with DB config.
func (m *Manager) Start(ctx context.Context) {
	slog.InfoContext(ctx, "Starting MNO Connection Manager", slog.Duration("sync_interval", m.syncInterval))
	m.wg.Add(1)
	go m.runSyncLoop(ctx)
}

// runSyncLoop is the main background loop.
func (m *Manager) runSyncLoop(ctx context.Context) {
	defer m.wg.Done()
	slog.InfoContext(ctx, "MNO Manager sync loop started.")

	// Initial sync on startup
	if err := m.loadAndSyncConnectors(ctx); err != nil {
		slog.ErrorContext(ctx, "Initial MNO connector sync failed", slog.Any("error", err))
		// Depending on severity, might want to stop the application
	}

	ticker := time.NewTicker(m.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "MNO Manager sync loop context cancelled.")
			return
		case <-m.stopCh:
			slog.InfoContext(ctx, "MNO Manager sync loop received stop signal.")
			return
		case <-ticker.C:
			slog.DebugContext(ctx, "Running periodic MNO connector sync...")
			if err := m.loadAndSyncConnectors(ctx); err != nil {
				slog.ErrorContext(ctx, "Periodic MNO connector sync failed", slog.Any("error", err))
			}
		}
	}
}

// loadAndSyncConnectors fetches DB config and updates running connectors.
func (m *Manager) loadAndSyncConnectors(ctx context.Context) error {
	slog.DebugContext(ctx, "Loading active MNO connections from database...")
	dbConfigs, err := m.dbQueries.GetActiveMNOConnections(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch active MNO connections: %w", err)
	}
	slog.DebugContext(ctx, "Fetched active MNO connection configs", slog.Int("count", len(dbConfigs)))

	m.mu.Lock() // Lock maps for syncing
	defer m.mu.Unlock()

	activeDBConnIDs := make(map[int32]bool)
	newMnoMap := make(map[int32][]int32) // Build new map[mno_id][]conn_id

	// Iterate through DB configs to add/update connectors
	for _, dbConn := range dbConfigs {
		connID := dbConn.ID
		mnoID := dbConn.MnoID
		activeDBConnIDs[connID] = true // Mark this ID as active in DB

		logCtx := logging.ContextWithMNOConnID(ctx, connID)
		logCtx = logging.ContextWithMNOID(logCtx, mnoID)

		// Check if connector already exists
		if existingVal, ok := m.connectors.Load(connID); ok {
			// Exists - Check if config changed significantly (e.g., host/port/systemid/protocol)
			// For now, we assume config changes require restart and don't handle dynamic updates.
			// Just ensure it's added to the new MNO map.
			existingConn := existingVal.(Connector) // Type assert
			if existingConn.MnoID() == mnoID {      // Basic check
				newMnoMap[mnoID] = append(newMnoMap[mnoID], connID)
			} else {
				slog.WarnContext(logCtx, "Existing connector MNO ID mismatch, scheduling shutdown.", slog.Int("old_mno_id", int(existingConn.MnoID())))
				go func(conn Connector) { // Shutdown in background
					shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel()
					_ = conn.Shutdown(shutdownCtx)
				}(existingConn)
				m.connectors.Delete(connID) // Remove inconsistent connector
			}
			continue
		}

		// --- Create New Connector ---
		slog.InfoContext(logCtx, "Creating new MNO connector", slog.String("protocol", dbConn.Protocol))
		var newConn Connector
		var creationErr error

		switch strings.ToLower(dbConn.Protocol) {
		case "smpp":
			smppCfg, err := NewSMPPConfigFromDB(dbConn) // Use helper
			if err != nil {
				creationErr = fmt.Errorf("invalid SMPP config for conn %d: %w", connID, err)
			} else {
				newConn, creationErr = NewSMPPMNOConnector(smppCfg, m.segmenter, m.dbQueries)
			}
		case "http":
			// TODO: Implement HTTP Connector creation
			// httpCfg, err := NewHTTPConfigFromDB(dbConn)
			// if err == nil {
			//     newConn, creationErr = NewHTTPMNOConnector(httpCfg, m.dbQueries)
			// } else { creationErr = err }
			creationErr = errors.New("http MNO connector not yet implemented")
		default:
			creationErr = fmt.Errorf("unknown protocol '%s' for conn %d", dbConn.Protocol, connID)
		}

		if creationErr != nil {
			slog.ErrorContext(logCtx, "Failed to create MNO connector", slog.Any("error", creationErr))
			continue // Skip this connection
		}

		// Register the central DLR handler
		newConn.RegisterDLRHandler(m.dlrHandler)

		// Store the new connector
		m.connectors.Store(connID, newConn)
		newMnoMap[mnoID] = append(newMnoMap[mnoID], connID)

		// Start the connector's connection process in background
		go func(conn Connector) {
			connCtx := logging.ContextWithMNOConnID(context.Background(), conn.ConnectionID())
			connCtx = logging.ContextWithMNOID(connCtx, conn.MnoID())
			if smppConn, ok := conn.(*SMPPMNOConnector); ok {
				// TODO: Add retry logic here?
				backoff := time.Second * 5 // Initial backoff
				for {
					slog.InfoContext(connCtx, "Attempting initial connect and bind...")
					err := smppConn.ConnectAndBind(connCtx)
					if err == nil {
						slog.InfoContext(connCtx, "ConnectAndBind succeeded.")
						break // Successfully connected
					}
					slog.ErrorContext(connCtx, "ConnectAndBind failed, will retry.", slog.Any("error", err), slog.Duration("retry_delay", backoff))

					// Check if manager is stopping
					select {
					case <-m.stopCh:
						slog.InfoContext(connCtx, "Manager stopping, aborting connect retry.")
						return
					case <-time.After(backoff):
						// Increase backoff potentially (e.g., backoff *= 2) up to a max
						if backoff < time.Minute {
							backoff *= 2
						}
					}
				}
			} else if httpConn, ok := conn.(*HTTPMNOConnector); ok {
				// TODO: Start HTTP connector background tasks if any (e.g., health checks)
				_ = httpConn // Placeholder
			}
		}(newConn)

	} // End DB config loop

	// Remove connectors that are no longer active in the DB
	m.connectors.Range(func(key, value interface{}) bool {
		connID := key.(int32)
		if !activeDBConnIDs[connID] {
			slog.InfoContext(ctx, "Removing connector no longer active in database", slog.Int("conn_id", int(connID)))
			conn := value.(Connector)
			// Shutdown in background
			go func(c Connector) {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				_ = c.Shutdown(shutdownCtx)
			}(conn)
			m.connectors.Delete(connID)
		}
		return true // Continue iteration
	})

	// Update the mnoConnectors map (atomically if Go maps supported it, otherwise simple replace under lock)
	m.mnoConnectors = sync.Map{} // Clear and rebuild (or use newMnoMap directly if locking allows)
	for mnoID, connIDs := range newMnoMap {
		m.mnoConnectors.Store(mnoID, connIDs)
	}
	slog.DebugContext(ctx, "Finished MNO connector sync")

	return nil
}

// GetConnector implements the ConnectorProvider interface.
// It finds an *active* connector for the requested MNO ID.
func (m *Manager) GetConnector(ctx context.Context, mnoID int32) (Connector, error) {
	logCtx := logging.ContextWithMNOID(ctx, mnoID)
	m.mu.RLock() // Read lock while accessing map
	defer m.mu.RUnlock()

	connIDsVal, ok := m.mnoConnectors.Load(mnoID)
	if !ok {
		return nil, fmt.Errorf("no connectors configured for MNO ID %d", mnoID)
	}
	connIDs := connIDsVal.([]int32) // Type assert
	if len(connIDs) == 0 {
		return nil, fmt.Errorf("no connectors available for MNO ID %d", mnoID)
	}

	var activeConnectors []Connector
	for _, connID := range connIDs {
		connVal, ok := m.connectors.Load(connID)
		if !ok {
			continue
		} // Should not happen if maps are synced
		conn := connVal.(Connector)
		status := conn.Status()
		if status == codes.StatusBound || status == codes.StatusHttpOk {
			activeConnectors = append(activeConnectors, conn)
		}
	}

	if len(activeConnectors) == 0 {
		slog.WarnContext(logCtx, "No currently active connectors found for MNO", slog.Int("total_configured", len(connIDs)))
		return nil, fmt.Errorf("no active connectors available for MNO ID %d", mnoID)
	}

	// Simple load balancing: Pick randomly
	selected := activeConnectors[rand.Intn(len(activeConnectors))]
	slog.DebugContext(logCtx, "Selected active connector", slog.Int("conn_id", int(selected.ConnectionID())), slog.String("status", selected.Status()))
	return selected, nil
}

// Shutdown gracefully stops all managed connectors and the sync loop.
func (m *Manager) Shutdown(ctx context.Context) error {
	slog.InfoContext(ctx, "Shutting down MNO Connection Manager...")
	close(m.stopCh) // Signal sync loop to stop

	m.wg.Wait() // Wait for sync loop goroutine to finish

	slog.InfoContext(ctx, "Shutting down individual MNO connectors...")
	var shutdownWg sync.WaitGroup
	var errors []error
	var errMu sync.Mutex

	m.connectors.Range(func(key, value interface{}) bool {
		connID := key.(int32)
		conn := value.(Connector)
		shutdownWg.Add(1)
		go func(c Connector, id int32) {
			defer shutdownWg.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			logCtx := logging.ContextWithMNOConnID(ctx, id)
			slog.InfoContext(logCtx, "Requesting shutdown for connector")
			if err := c.Shutdown(shutdownCtx); err != nil {
				slog.ErrorContext(logCtx, "Error shutting down connector", slog.Any("error", err))
				errMu.Lock()
				errors = append(errors, fmt.Errorf("conn %d shutdown error: %w", id, err))
				errMu.Unlock()
			}
		}(conn, connID)
		return true // Continue iteration
	})

	shutdownWg.Wait() // Wait for all connectors to attempt shutdown
	slog.InfoContext(ctx, "MNO Connection Manager shutdown complete.")
	if len(errors) > 0 {
		return fmt.Errorf("encountered %d error(s) during connector shutdown: %v", len(errors), errors)
	}
	return nil
}

// --- Placeholder for HTTP Connector ---
type HTTPMNOConnector struct{}

func (c *HTTPMNOConnector) SubmitMessage(ctx context.Context, msg PreparedMessage) (SubmitResult, error) {
	panic("not implemented")
}
func (c *HTTPMNOConnector) RegisterDLRHandler(handler DLRHandlerFunc) { panic("not implemented") }
func (c *HTTPMNOConnector) Shutdown(ctx context.Context) error        { panic("not implemented") }
func (c *HTTPMNOConnector) Status() string                            { panic("not implemented") }
func (c *HTTPMNOConnector) ConnectionID() int32                       { panic("not implemented") }
func (c *HTTPMNOConnector) MnoID() int32                              { panic("not implemented") }
