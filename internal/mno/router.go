package mno

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/thrillee/aegisbox/internal/database"
)

type NetworkRouter struct {
	dbQueries    database.Querier
	providers    map[string]HTTPProvider
	senderConfig map[string]SenderConfig
	logger       *slog.Logger
	mu           sync.RWMutex
}

type SenderConfig struct {
	SenderID    string
	Provider    string
	RateLimit   int
	OTPTemplate string
	Networks    []string
}

func NewNetworkRouter(dbQueries database.Querier, logger *slog.Logger) *NetworkRouter {
	return &NetworkRouter{
		dbQueries:    dbQueries,
		providers:    make(map[string]HTTPProvider),
		senderConfig: make(map[string]SenderConfig),
		logger:       logger,
	}
}

func (r *NetworkRouter) RegisterProvider(name string, provider HTTPProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[strings.ToUpper(name)] = provider
}

func (r *NetworkRouter) SetSenderConfig(senderID string, config SenderConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.senderConfig[senderID] = config
}

func (r *NetworkRouter) LoadFromDatabase(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger.InfoContext(ctx, "Loading sender configurations from database")
	return nil
}

func (r *NetworkRouter) GetNetworkFromMSISDN(msisdn string) string {
	msisdn = strings.TrimPrefix(msisdn, "+")
	msisdn = strings.TrimPrefix(msisdn, "00")

	prefixes := map[string]string{
		"234": "NIGERIA",
		"233": "GHANA",
		"20":  "EGYPT",
	}

	if len(msisdn) >= 3 {
		countryPrefix := msisdn[:3]
		if country, ok := prefixes[countryPrefix]; ok {
			switch country {
			case "NIGERIA":
				return r.getNigerianNetwork(msisdn)
			case "GHANA":
				return r.getGhanaianNetwork(msisdn)
			case "EGYPT":
				return r.getEgyptianNetwork(msisdn)
			}
		}
	}

	return "UNKNOWN"
}

func (r *NetworkRouter) getNigerianNetwork(msisdn string) string {
	if len(msisdn) < 6 {
		return "UNKNOWN"
	}

	prefixes := map[string]string{
		"703":  "MTN",
		"704":  "MTN",
		"706":  "MTN",
		"803":  "MTN",
		"806":  "MTN",
		"810":  "MTN",
		"813":  "MTN",
		"814":  "MTN",
		"816":  "MTN",
		"903":  "MTN",
		"906":  "MTN",
		"913":  "MTN",
		"916":  "MTN",
		"7025": "MTN",
		"7026": "MTN",
		"705":  "GLO",
		"805":  "GLO",
		"807":  "GLO",
		"811":  "GLO",
		"815":  "GLO",
		"905":  "GLO",
		"915":  "GLO",
		"701":  "AIRTEL",
		"708":  "AIRTEL",
		"802":  "AIRTEL",
		"808":  "AIRTEL",
		"812":  "AIRTEL",
		"901":  "AIRTEL",
		"902":  "AIRTEL",
		"904":  "AIRTEL",
		"907":  "AIRTEL",
		"912":  "AIRTEL",
		"809":  "9MOBILE",
		"817":  "9MOBILE",
		"818":  "9MOBILE",
		"909":  "9MOBILE",
		"908":  "9MOBILE",
	}

	for prefix, network := range prefixes {
		if strings.HasPrefix(msisdn, prefix) {
			return network
		}
	}

	return "UNKNOWN"
}

func (r *NetworkRouter) getGhanaianNetwork(msisdn string) string {
	if len(msisdn) < 5 {
		return "UNKNOWN"
	}
	prefix := msisdn[3:5]

	switch prefix {
	case "24", "53", "54", "55", "59", "25":
		return "MTN"
	case "20", "50":
		return "VODAFONE"
	case "27", "57", "26", "56":
		return "AIRTELTIGO"
	}
	return "UNKNOWN"
}

func (r *NetworkRouter) getEgyptianNetwork(msisdn string) string {
	if len(msisdn) < 4 {
		return "UNKNOWN"
	}
	prefix := msisdn[2:4]

	switch prefix {
	case "10":
		return "VODAFONE"
	case "11":
		return "ETISALAT"
	case "12":
		return "ORANGE"
	case "15":
		return "WE"
	}
	return "UNKNOWN"
}

func (r *NetworkRouter) IsNetworkAllowed(senderID, network string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cfg, exists := r.senderConfig[senderID]
	if !exists {
		return false
	}

	if len(cfg.Networks) == 0 {
		return true
	}

	for _, n := range cfg.Networks {
		if strings.EqualFold(n, network) {
			return true
		}
	}
	return false
}

func (r *NetworkRouter) SelectProvider(ctx context.Context, senderID, msisdn string) (HTTPProvider, error) {
	r.mu.RLock()
	providerName := r.senderConfig[senderID].Provider
	provider, exists := r.providers[strings.ToUpper(providerName)]
	r.mu.RUnlock()

	if exists {
		return provider, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.providers {
		return p, nil
	}

	return nil, fmt.Errorf("no providers available")
}

func (r *NetworkRouter) GetOptimalSenderID(ctx context.Context, msisdn string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.senderConfig) == 0 {
		return "", fmt.Errorf("no sender IDs configured")
	}

	network := r.GetNetworkFromMSISDN(msisdn)
	r.logger.DebugContext(ctx, "Detected network", "network", network)

	var eligibleSenders []string
	for senderID := range r.senderConfig {
		if !r.IsNetworkAllowed(senderID, network) {
			continue
		}
		eligibleSenders = append(eligibleSenders, senderID)
	}

	if len(eligibleSenders) == 0 {
		return "", fmt.Errorf("no sender IDs available for network %s", network)
	}

	for _, senderID := range eligibleSenders {
		senderCfg := r.senderConfig[senderID]
		if senderCfg.Provider == "" {
			continue
		}
		if _, exists := r.providers[strings.ToUpper(senderCfg.Provider)]; exists {
			return senderID, nil
		}
	}

	return eligibleSenders[0], nil
}

func (r *NetworkRouter) GetSenderConfig(senderID string) (SenderConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cfg, exists := r.senderConfig[senderID]
	return cfg, exists
}

func (r *NetworkRouter) GetHTTPProvider(name string) (HTTPProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[strings.ToUpper(name)]
	return p, ok
}

func (r *NetworkRouter) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
