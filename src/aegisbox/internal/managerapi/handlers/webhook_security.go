package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thrillee/aegisbox/internal/database"
)

// WebhookSecurityConfig holds security configuration for webhooks
type WebhookSecurityConfig struct {
	// Rate limiting
	MaxRequestsPerMinute int
	MaxRequestsPerHour   int

	// IP whitelisting (optional)
	AllowedIPRanges []string

	// Signature verification
	EnableSignatureVerification bool
}

// DefaultWebhookSecurityConfig returns default security configuration
func DefaultWebhookSecurityConfig() *WebhookSecurityConfig {
	return &WebhookSecurityConfig{
		MaxRequestsPerMinute:        100,        // 100 requests per minute per IP
		MaxRequestsPerHour:          1000,       // 1000 requests per hour per IP
		EnableSignatureVerification: false,      // Disabled by default for backward compatibility
		AllowedIPRanges:             []string{}, // Empty = allow all IPs
	}
}

// WebhookSecurityMiddleware provides security features for webhook endpoints
type WebhookSecurityMiddleware struct {
	config      *WebhookSecurityConfig
	rateLimiter *rateLimiter
	allowedIPs  []*net.IPNet
	logger      *slog.Logger
}

// NewWebhookSecurityMiddleware creates a new security middleware instance
func NewWebhookSecurityMiddleware(config *WebhookSecurityConfig) *WebhookSecurityMiddleware {
	middleware := &WebhookSecurityMiddleware{
		config:      config,
		rateLimiter: newRateLimiter(config.MaxRequestsPerMinute, config.MaxRequestsPerHour),
		logger:      slog.Default(),
	}

	// Parse allowed IP ranges
	if len(config.AllowedIPRanges) > 0 {
		for _, ipRange := range config.AllowedIPRanges {
			_, ipNet, err := net.ParseCIDR(ipRange)
			if err != nil {
				middleware.logger.Warn("Failed to parse IP range", "range", ipRange, "error", err)
				continue
			}
			middleware.allowedIPs = append(middleware.allowedIPs, ipNet)
		}
	}

	return middleware
}

// Middleware returns a Gin middleware function
func (m *WebhookSecurityMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		ctx := c.Request.Context()

		// 1. IP Whitelisting
		if len(m.allowedIPs) > 0 && !m.isIPAllowed(clientIP) {
			m.logger.WarnContext(ctx, "Webhook request from disallowed IP blocked",
				"ip", clientIP,
				"path", c.Request.URL.Path)
			c.JSON(http.StatusForbidden, gin.H{"error": "IP not allowed"})
			c.Abort()
			return
		}

		// 2. Rate Limiting
		if !m.rateLimiter.allow(clientIP) {
			m.logger.WarnContext(ctx, "Webhook rate limit exceeded",
				"ip", clientIP,
				"path", c.Request.URL.Path)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		// 3. Signature Verification (provider-specific, handled separately)
		// This is implemented per-provider in their ParseDLR methods

		c.Next()
	}
}

// isIPAllowed checks if an IP is in the allowed ranges
func (m *WebhookSecurityMiddleware) isIPAllowed(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, ipNet := range m.allowedIPs {
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}

// VerifyTermiiSignature verifies Termii webhook signature
// Termii uses HMAC-SHA256 signature verification
func VerifyTermiiSignature(secretKey, payload, signature string) bool {
	if secretKey == "" || signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// GetTermiiSecretKey retrieves the secret key for a Termii connection
func GetTermiiSecretKey(ctx any, dbQueries database.Querier, connID int32) (string, error) {
	// This would query the database for the connection's secret key
	// For now, return empty string to disable verification
	// TODO: Implement when connection ID is properly passed to webhook handlers
	return "", nil
}

// rateLimiter implements a simple rate limiter
type rateLimiter struct {
	maxPerMinute int
	maxPerHour   int
	requests     map[string]*ipRequests
	mu           sync.RWMutex
	cleanupDone  chan struct{}
}

type ipRequests struct {
	minuteCount int
	hourCount   int
	minuteReset time.Time
	hourReset   time.Time
}

func newRateLimiter(maxPerMinute, maxPerHour int) *rateLimiter {
	rl := &rateLimiter{
		maxPerMinute: maxPerMinute,
		maxPerHour:   maxPerHour,
		requests:     make(map[string]*ipRequests),
		cleanupDone:  make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// allow checks if a request from the given IP is allowed
func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	req, exists := rl.requests[ip]

	if !exists {
		rl.requests[ip] = &ipRequests{
			minuteCount: 1,
			hourCount:   1,
			minuteReset: now.Add(time.Minute),
			hourReset:   now.Add(time.Hour),
		}
		return true
	}

	// Reset counters if needed
	if now.After(req.minuteReset) {
		req.minuteCount = 0
		req.minuteReset = now.Add(time.Minute)
	}
	if now.After(req.hourReset) {
		req.hourCount = 0
		req.hourReset = now.Add(time.Hour)
	}

	// Check limits
	if req.minuteCount >= rl.maxPerMinute {
		return false
	}
	if req.hourCount >= rl.maxPerHour {
		return false
	}

	// Increment counters
	req.minuteCount++
	req.hourCount++

	return true
}

// cleanup removes old entries from the rate limiter
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, req := range rl.requests {
				// Remove entries that haven't been used in the last hour
				if now.After(req.hourReset) {
					delete(rl.requests, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.cleanupDone:
			return
		}
	}
}

// Close stops the cleanup goroutine
func (rl *rateLimiter) Close() {
	close(rl.cleanupDone)
}

// Provider-specific IP ranges (documented)
var ProviderIPRanges = map[string][]string{
	"TERMII": {
		// Add Termii IP ranges here when available
		// Example: "52.x.x.x/24"
	},
	"AFRICATALKING": {
		// Add AfricaTalking IP ranges here when available
	},
	"SIMPU": {
		// Add Simpu IP ranges here when available
	},
	"EBULKSMS": {
		// Add eBulkSMS IP ranges here when available
	},
	"MULTITEXTER": {
		// Add MultiTexter IP ranges here when available
	},
}

// GetProviderIPRanges returns the known IP ranges for a provider
func GetProviderIPRanges(provider string) []string {
	return ProviderIPRanges[strings.ToUpper(provider)]
}
