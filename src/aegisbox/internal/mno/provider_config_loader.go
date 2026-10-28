package mno

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/mno/http_providers"
)

// ProviderConfigLoader loads HTTP provider configurations from the database
type ProviderConfigLoader struct {
	dbQueries database.Querier
	logger    *slog.Logger
}

// NewProviderConfigLoader creates a new provider config loader
func NewProviderConfigLoader(q database.Querier, logger *slog.Logger) *ProviderConfigLoader {
	return &ProviderConfigLoader{
		dbQueries: q,
		logger:    logger,
	}
}

// GetProviderByName loads an HTTP provider instance by provider name from database
// It looks up the mno_connection record where protocol='http' and matches the provider name
func (l *ProviderConfigLoader) GetProviderByName(ctx context.Context, providerName string) (http_providers.HTTPProvider, error) {
	// Get all active HTTP connections (protocol='http')
	connections, err := l.dbQueries.GetActiveHTTPConnections(ctx)
	if err != nil {
		l.logger.ErrorContext(ctx, "Failed to get active HTTP connections",
			"error", err)
		return nil, err
	}

	// Find connection matching the provider name
	for _, conn := range connections {
		// Parse provider_config JSON to get the provider name
		var providerConfig map[string]interface{}
		if conn.ProviderConfig != nil {
			if err := json.Unmarshal(conn.ProviderConfig, &providerConfig); err != nil {
				l.logger.WarnContext(ctx, "Failed to parse provider_config JSON",
					"conn_id", conn.ID,
					"error", err)
				continue
			}
		}

		// Check if provider name matches (case-insensitive)
		configProviderName, _ := providerConfig["provider_name"].(string)
		if !strings.EqualFold(configProviderName, providerName) {
			continue
		}

		// Build HTTPProviderConfig from database record
		config := l.buildProviderConfig(conn, providerConfig)

		// Create and return the provider instance
		provider := http_providers.NewHTTPProvider(providerName, config, l.logger)
		if provider == nil {
			return nil, fmt.Errorf("failed to create provider instance for: %s", providerName)
		}

		l.logger.InfoContext(ctx, "Loaded provider from database",
			"provider", providerName,
			"conn_id", conn.ID,
			"mno_id", conn.MnoID)

		return provider, nil
	}

	return nil, &UnknownProviderError{Name: providerName}
}

// GetProviderByConnectionID loads an HTTP provider instance by mno_connection ID
func (l *ProviderConfigLoader) GetProviderByConnectionID(ctx context.Context, connID int32) (http_providers.HTTPProvider, error) {
	conn, err := l.dbQueries.GetMNOConnectionByID(ctx, connID)
	if err != nil {
		l.logger.ErrorContext(ctx, "Failed to get MNO connection",
			"conn_id", connID,
			"error", err)
		return nil, err
	}

	if conn.Protocol != "http" {
		return nil, fmt.Errorf("connection %d is not HTTP protocol: %s", connID, conn.Protocol)
	}

	// Parse provider_config JSON
	var providerConfig map[string]interface{}
	if conn.ProviderConfig != nil {
		if err := json.Unmarshal(conn.ProviderConfig, &providerConfig); err != nil {
			return nil, fmt.Errorf("failed to parse provider_config: %w", err)
		}
	}

	// Get provider name from config
	providerName, ok := providerConfig["provider_name"].(string)
	if !ok || providerName == "" {
		return nil, fmt.Errorf("provider_name not found in provider_config for connection %d", connID)
	}

	// Build config and create provider
	config := l.buildProviderConfig(conn, providerConfig)
	provider := http_providers.NewHTTPProvider(providerName, config, l.logger)
	if provider == nil {
		return nil, fmt.Errorf("failed to create provider instance for: %s", providerName)
	}

	l.logger.InfoContext(ctx, "Loaded provider from database",
		"provider", providerName,
		"conn_id", connID,
		"mno_id", conn.MnoID)

	return provider, nil
}

// buildProviderConfig constructs an HTTPProviderConfig from database record
func (l *ProviderConfigLoader) buildProviderConfig(conn database.MnoConnection, providerConfig map[string]interface{}) http_providers.HTTPProviderConfig {
	// Get provider name from config
	providerName, _ := providerConfig["provider_name"].(string)

	// Parse timeout
	timeout := 30 * time.Second
	if conn.TimeoutSecs != nil {
		timeout = time.Duration(*conn.TimeoutSecs) * time.Second
	}

	// Get API key (prioritize api_key field, fallback to conn.ApiKey)
	apiKey := ""
	if configAPIKey, ok := providerConfig["api_key"].(string); ok && configAPIKey != "" {
		apiKey = configAPIKey
	} else if conn.ApiKey != nil {
		apiKey = *conn.ApiKey
	}

	// Get secret key
	secretKey := ""
	if configSecretKey, ok := providerConfig["secret_key"].(string); ok && configSecretKey != "" {
		secretKey = configSecretKey
	} else if conn.SecretKey != nil {
		secretKey = *conn.SecretKey
	}

	// Get username
	username := ""
	if conn.Username != nil {
		username = *conn.Username
	}

	// Get password
	password := ""
	if conn.HttpPassword != nil {
		password = *conn.HttpPassword
	}

	// Get default sender
	defaultSender := ""
	if conn.DefaultSender != nil {
		defaultSender = *conn.DefaultSender
	}

	// Get rate limit
	rateLimit := 100 // Default
	if conn.RateLimit != nil {
		rateLimit = int(*conn.RateLimit)
	}

	// Build base URL
	baseURL := ""
	if conn.BaseUrl != "" {
		baseURL = conn.BaseUrl
	}

	return http_providers.HTTPProviderConfig{
		Name:           providerName,
		BaseURL:        baseURL,
		APIKey:         apiKey,
		Username:       username,
		Password:       password,
		SecretKey:      secretKey,
		DefaultSender:  defaultSender,
		RateLimit:      rateLimit,
		Timeout:        timeout,
		ProviderConfig: providerConfig,
	}
}

// GetAllHTTPProviders returns all active HTTP providers from database
func (l *ProviderConfigLoader) GetAllHTTPProviders(ctx context.Context) ([]http_providers.HTTPProvider, error) {
	connections, err := l.dbQueries.GetActiveMNOConnections(ctx)
	if err != nil {
		return nil, err
	}

	var providers []http_providers.HTTPProvider
	for _, conn := range connections {
		if conn.Protocol != "http" {
			continue
		}

		// Parse provider_config
		var providerConfig map[string]interface{}
		if conn.ProviderConfig != nil {
			if err := json.Unmarshal(conn.ProviderConfig, &providerConfig); err != nil {
				l.logger.WarnContext(ctx, "Failed to parse provider_config, skipping",
					"conn_id", conn.ID,
					"error", err)
				continue
			}
		}

		providerName, ok := providerConfig["provider_name"].(string)
		if !ok || providerName == "" {
			l.logger.WarnContext(ctx, "Provider name not found in config, skipping",
				"conn_id", conn.ID)
			continue
		}

		config := l.buildProviderConfig(conn, providerConfig)
		provider := http_providers.NewHTTPProvider(providerName, config, l.logger)
		if provider != nil {
			providers = append(providers, provider)
		}
	}

	return providers, nil
}
