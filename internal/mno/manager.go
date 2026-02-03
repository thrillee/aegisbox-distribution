package mno

import (
	"context"
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

var _ ConnectorProvider = (*Manager)(nil)

type Manager struct {
	pool         *pgxpool.Pool
	dbQueries    database.Querier
	segmenter    segmenter.Segmenter
	dlrHandler   DLRHandlerFunc
	syncInterval time.Duration
	logger       *slog.Logger

	connectors      sync.Map
	mnoConnectors   sync.Map
	circuitBreakers sync.Map // map[int32]*CircuitBreaker per MNO ID
	mu              sync.RWMutex
	stopCh          chan struct{}
	wg              sync.WaitGroup
}

func NewManager(pool *pgxpool.Pool, q database.Querier, seg segmenter.Segmenter, dlrHandler DLRHandlerFunc, syncInterval time.Duration, logger *slog.Logger) *Manager {
	if syncInterval <= 0 {
		syncInterval = 1 * time.Minute
	}
	return &Manager{
		pool:            pool,
		dbQueries:       q,
		segmenter:       seg,
		dlrHandler:      dlrHandler,
		syncInterval:    syncInterval,
		stopCh:          make(chan struct{}),
		circuitBreakers: sync.Map{},
		logger:          logger,
	}
}

func (m *Manager) GetDLRHandler() DLRHandlerFunc {
	return m.dlrHandler
}

func (m *Manager) Start(ctx context.Context) {
	slog.InfoContext(ctx, "Starting MNO Connection Manager", slog.Duration("sync_interval", m.syncInterval))
	m.wg.Add(1)
	go m.runSyncLoop(ctx)
}

func (m *Manager) runSyncLoop(ctx context.Context) {
	defer m.wg.Done()
	slog.InfoContext(ctx, "MNO Manager sync loop started.")

	if err := m.loadAndSyncConnectors(ctx); err != nil {
		slog.ErrorContext(ctx, "Initial MNO connector sync failed", slog.Any("error", err))
	}

	m.wg.Add(1)
	go m.runReconnectionLoop(ctx)

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

func (m *Manager) runReconnectionLoop(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.reconnectDisconnected(ctx)
		}
	}
}

func (m *Manager) reconnectDisconnected(ctx context.Context) {
	m.connectors.Range(func(key, value any) bool {
		connID := key.(int32)
		conn := value.(Connector)

		status := conn.Status()
		if status == codes.StatusDisconnected || status == codes.StatusError || status == codes.StatusBindingFailed {
			logCtx := logging.ContextWithMNOConnID(ctx, connID)
			logCtx = logging.ContextWithMNOID(logCtx, conn.MnoID())

			slog.InfoContext(logCtx, "Attempting to reconnect disconnected connector")

			if smppConn, ok := conn.(*SMPPMNOConnector); ok {
				m.wg.Add(1)
				go func() {
					defer m.wg.Done()
					m.reconnectWithBackoff(logCtx, smppConn)
				}()
			} else if httpConn, ok := conn.(*HTTPMNOConnector); ok {
				m.wg.Add(1)
				go func() {
					defer m.wg.Done()
					m.reconnectHTTP(logCtx, httpConn)
				}()
			}
		}
		return true
	})
}

func (m *Manager) reconnectWithBackoff(ctx context.Context, conn *SMPPMNOConnector) {
	backoff := time.Second * 5
	maxBackoff := time.Minute * 5

	for i := 0; i < 10; i++ {
		select {
		case <-m.stopCh:
			return
		default:
		}

		slog.InfoContext(ctx, "Reconnection attempt", slog.Int("attempt", i+1), slog.Duration("backoff", backoff))

		if err := conn.ConnectAndBind(ctx); err != nil {
			slog.ErrorContext(ctx, "Reconnection failed", slog.Any("error", err), slog.Duration("retry_delay", backoff))

			select {
			case <-m.stopCh:
				return
			case <-time.After(backoff):
			}

			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		slog.InfoContext(ctx, "Reconnection successful")
		return
	}

	slog.ErrorContext(ctx, "Max reconnection attempts reached, will retry in next cycle")
}

func (m *Manager) reconnectHTTP(ctx context.Context, conn *HTTPMNOConnector) {
	backoff := time.Second * 5
	maxBackoff := time.Minute * 5

	for i := 0; i < 10; i++ {
		select {
		case <-m.stopCh:
			return
		default:
		}

		slog.InfoContext(ctx, "HTTP connector health check/reconnect attempt", slog.Int("attempt", i+1))

		if err := conn.HealthCheck(ctx); err != nil {
			slog.ErrorContext(ctx, "HTTP connector health check failed", slog.Any("error", err))

			select {
			case <-m.stopCh:
				return
			case <-time.After(backoff):
			}

			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		slog.InfoContext(ctx, "HTTP connector health check passed")
		return
	}
}

func (m *Manager) GetCircuitBreaker(mnoID int32) *CircuitBreaker {
	if cb, ok := m.circuitBreakers.Load(mnoID); ok {
		return cb.(*CircuitBreaker)
	}

	newCB := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          30 * time.Second,
		RequestTimeout:   10 * time.Second,
		VolumeThreshold:  10,
		Logger:           m.logger,
		MNOID:            mnoID,
	})

	actual, _ := m.circuitBreakers.LoadOrStore(mnoID, newCB)
	return actual.(*CircuitBreaker)
}

func (m *Manager) loadAndSyncConnectors(ctx context.Context) error {
	slog.DebugContext(ctx, "Loading active MNO connections from database...")
	dbConfigs, err := m.dbQueries.GetActiveMNOConnections(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch active MNO connections: %w", err)
	}
	slog.DebugContext(ctx, "Fetched active MNO connection configs", slog.Int("count", len(dbConfigs)))

	m.mu.Lock()
	defer m.mu.Unlock()

	activeDBConnIDs := make(map[int32]bool)
	newMnoMap := make(map[int32][]int32)

	for _, dbConn := range dbConfigs {
		connID := dbConn.ID
		mnoID := dbConn.MnoID
		activeDBConnIDs[connID] = true

		logCtx := logging.ContextWithMNOConnID(ctx, connID)
		logCtx = logging.ContextWithMNOID(logCtx, mnoID)

		if existingVal, ok := m.connectors.Load(connID); ok {
			existingConn := existingVal.(Connector)
			if existingConn.MnoID() == mnoID {
				newMnoMap[mnoID] = append(newMnoMap[mnoID], connID)
			} else {
				slog.WarnContext(logCtx, "Existing connector MNO ID mismatch, scheduling shutdown.", slog.Int("old_mno_id", int(existingConn.MnoID())))
				go func(conn Connector) {
					shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel()
					_ = conn.Shutdown(shutdownCtx)
				}(existingConn)
				m.connectors.Delete(connID)
			}
			continue
		}

		slog.InfoContext(logCtx, "Creating new MNO connector", slog.String("protocol", dbConn.Protocol))
		var newConn Connector
		var creationErr error

		switch strings.ToLower(dbConn.Protocol) {
		case "smpp":
			smppCfg, err := NewSMPPConfigFromDB(dbConn)
			if err != nil {
				creationErr = fmt.Errorf("invalid SMPP config for conn %d: %w", connID, err)
			} else {
				newConn, creationErr = NewSMPPMNOConnector(smppCfg, m.segmenter, m.dbQueries)
			}
		case "http":
			httpCfg, err := NewHTTPConfigFromDB(dbConn)
			if err == nil {
				newConn = NewHTTPMNOConnector(httpCfg)
			} else {
				creationErr = err
			}
		default:
			creationErr = fmt.Errorf("unknown protocol '%s' for conn %d", dbConn.Protocol, connID)
		}

		if creationErr != nil {
			slog.ErrorContext(logCtx, "Failed to create MNO connector", slog.Any("error", creationErr))
			continue
		}

		newConn.RegisterDLRHandler(m.dlrHandler)
		m.connectors.Store(connID, newConn)
		newMnoMap[mnoID] = append(newMnoMap[mnoID], connID)

		go func(conn Connector) {
			connCtx := logging.ContextWithMNOConnID(context.Background(), conn.ConnectionID())
			connCtx = logging.ContextWithMNOID(connCtx, conn.MnoID())
			if smppConn, ok := conn.(*SMPPMNOConnector); ok {
				backoff := time.Second * 5
				for {
					slog.InfoContext(connCtx, "Attempting initial connect and bind...")
					err := smppConn.ConnectAndBind(connCtx)
					if err == nil {
						slog.InfoContext(connCtx, "ConnectAndBind succeeded.")
						break
					}
					slog.ErrorContext(connCtx, "ConnectAndBind failed, will retry.", slog.Any("error", err), slog.Duration("retry_delay", backoff))

					select {
					case <-m.stopCh:
						slog.InfoContext(connCtx, "Manager stopping, aborting connect retry.")
						return
					case <-time.After(backoff):
						if backoff < time.Minute {
							backoff *= 2
						}
					}
				}
			} else if httpConn, ok := conn.(*HTTPMNOConnector); ok {
				_ = httpConn
			}
		}(newConn)

	}

	m.connectors.Range(func(key, value any) bool {
		connID := key.(int32)
		if !activeDBConnIDs[connID] {
			slog.InfoContext(ctx, "Removing connector no longer active in database", slog.Int("conn_id", int(connID)))
			conn := value.(Connector)
			go func(c Connector) {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				_ = c.Shutdown(shutdownCtx)
			}(conn)
			m.connectors.Delete(connID)
		}
		return true
	})

	m.mnoConnectors = sync.Map{}
	for mnoID, connIDs := range newMnoMap {
		m.mnoConnectors.Store(mnoID, connIDs)
	}
	slog.DebugContext(ctx, "Finished MNO connector sync")

	return nil
}

func (m *Manager) GetConnector(ctx context.Context, mnoID int32) (Connector, error) {
	return m.GetConnectorWithProtocol(ctx, mnoID, "")
}

func (m *Manager) GetConnectorWithProtocol(ctx context.Context, mnoID int32, preferredProtocol string) (Connector, error) {
	logCtx := logging.ContextWithMNOID(ctx, mnoID)

	cb := m.GetCircuitBreaker(mnoID)
	if !cb.AllowRequest() {
		slog.WarnContext(logCtx, "Circuit breaker is open for MNO, rejecting request", slog.String("state", cb.State().String()))
		return nil, fmt.Errorf("circuit breaker open for MNO ID %d", mnoID)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	connIDsVal, ok := m.mnoConnectors.Load(mnoID)
	if !ok {
		cb.RecordFailure()
		return nil, fmt.Errorf("no connectors configured for MNO ID %d", mnoID)
	}
	connIDs := connIDsVal.([]int32)
	if len(connIDs) == 0 {
		cb.RecordFailure()
		return nil, fmt.Errorf("no connectors available for MNO ID %d", mnoID)
	}

	var preferredConnectors []Connector
	var fallbackConnectors []Connector

	for _, connID := range connIDs {
		connVal, ok := m.connectors.Load(connID)
		if !ok {
			continue
		}
		conn := connVal.(Connector)
		status := conn.Status()
		if status != codes.StatusBound && status != codes.StatusHttpOk {
			continue
		}

		connProtocol := m.getConnectorProtocol(conn)
		if preferredProtocol != "" && strings.EqualFold(connProtocol, preferredProtocol) {
			preferredConnectors = append(preferredConnectors, conn)
		} else {
			fallbackConnectors = append(fallbackConnectors, conn)
		}
	}

	var selected []Connector
	if len(preferredConnectors) > 0 {
		selected = preferredConnectors
		slog.DebugContext(logCtx, "Using preferred protocol connectors", slog.String("protocol", preferredProtocol))
	} else if len(fallbackConnectors) > 0 {
		selected = fallbackConnectors
		slog.DebugContext(logCtx, "No preferred protocol connectors available, using fallback")
	}

	if len(selected) == 0 {
		slog.WarnContext(logCtx, "No active connectors found for MNO",
			slog.Int("total_configured", len(connIDs)),
			slog.String("preferred_protocol", preferredProtocol))
		cb.RecordFailure()
		return nil, fmt.Errorf("no active connectors available for MNO ID %d", mnoID)
	}

	selectedConnector := selected[rand.Intn(len(selected))]
	slog.DebugContext(logCtx, "Selected active connector",
		slog.Int("conn_id", int(selectedConnector.ConnectionID())),
		slog.String("status", selectedConnector.Status()),
		slog.String("protocol", m.getConnectorProtocol(selectedConnector)))

	cb.RecordSuccess()
	return selectedConnector, nil
}

func (m *Manager) getConnectorProtocol(conn Connector) string {
	switch conn.(type) {
	case *SMPPMNOConnector:
		return "smpp"
	case *HTTPMNOConnector:
		return "http"
	default:
		return ""
	}
}

func (m *Manager) RecordConnectionFailure(mnoID int32) {
	cb := m.GetCircuitBreaker(mnoID)
	cb.RecordFailure()
}

func (m *Manager) RecordConnectionSuccess(mnoID int32) {
	cb := m.GetCircuitBreaker(mnoID)
	cb.RecordSuccess()
}

func (m *Manager) Shutdown(ctx context.Context) error {
	slog.InfoContext(ctx, "Shutting down MNO Connection Manager...")
	close(m.stopCh)

	m.wg.Wait()

	slog.InfoContext(ctx, "Shutting down individual MNO connectors...")
	var shutdownWg sync.WaitGroup
	var errors []error
	var errMu sync.Mutex

	m.connectors.Range(func(key, value any) bool {
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
		return true
	})

	shutdownWg.Wait()
	slog.InfoContext(ctx, "MNO Connection Manager shutdown complete.")
	if len(errors) > 0 {
		return fmt.Errorf("encountered %d error(s) during connector shutdown: %v", len(errors), errors)
	}
	return nil
}
