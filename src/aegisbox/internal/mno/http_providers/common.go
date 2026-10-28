package http_providers

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// HTTPProvider interface defines the contract for all HTTP-based SMS providers
type HTTPProvider interface {
	SendSMS(ctx context.Context, from, to, message string) (string, error)
	CheckStatus(ctx context.Context, messageID string) (string, error)
	ParseDLR(ctx context.Context, req *http.Request) (messageID, status string, err error)
	Name() string
	SupportsDLR() bool
	GetDefaultSenderID() string
	SupportsAutoDLR() bool // Returns true if provider auto-sends DLR on 200 OK
}

// HTTPProviderConfig contains configuration for HTTP providers
type HTTPProviderConfig struct {
	Name           string
	BaseURL        string
	APIKey         string
	Username       string
	Password       string
	SecretKey      string
	DefaultSender  string
	RateLimit      int
	Timeout        time.Duration
	ProviderConfig map[string]interface{}
}

// Status constants for normalized DLR statuses
const (
	StatusDelivered   = "delivered"
	StatusSent        = "sent"
	StatusPending     = "pending"
	StatusFailed      = "failed"
	StatusExpired     = "expired"
	StatusUnreachable = "unreachable"
	StatusDNDBlocked  = "dnd_blocked"
	StatusUnknown     = "unknown"
)

// NewHTTPProvider creates an HTTPProvider based on the provider name
func NewHTTPProvider(name string, config HTTPProviderConfig, logger *slog.Logger) HTTPProvider {
	switch strings.ToUpper(name) {
	case "TERMII":
		return NewTermiiProvider(config, logger)
	case "AFRICATALKING":
		return NewAfricaTalkingProvider(config, logger)
	case "MULTITEXTER":
		return NewMultiTexterProvider(config, logger)
	case "SIMPU":
		return NewSimpuProvider(config, logger)
	case "SIMPU_OTP":
		return NewSimpuOTPProvider(config, logger)
	case "EBULKSMS":
		return NewEBulkSMSProvider(config, logger)
	case "AKUUK":
		return NewAkuukProvider(config, logger)
	default:
		logger.Warn("Unknown HTTP provider requested", "name", name)
		return nil
	}
}

// normalizeStatus converts provider-specific status strings to our internal status constants
func NormalizeStatus(provider, status string) string {
	status = strings.ToLower(strings.TrimSpace(status))

	switch strings.ToUpper(provider) {
	case "TERMII":
		switch status {
		case "delivered":
			return StatusDelivered
		case "sent":
			return StatusSent
		case "failed":
			return StatusFailed
		case "expired":
			return StatusExpired
		default:
			return StatusUnknown
		}

	case "AFRICATALKING":
		switch status {
		case "success", "delivered":
			return StatusDelivered
		case "sent":
			return StatusSent
		case "failed":
			return StatusFailed
		case "rejected":
			return StatusFailed
		default:
			return StatusUnknown
		}

	case "MULTITEXTER":
		switch status {
		case "deliveredpp", "delivered":
			return StatusDelivered
		case "failed":
			return StatusFailed
		case "pending":
			return StatusPending
		case "sent":
			return StatusSent
		default:
			return StatusUnknown
		}

	case "SIMPU":
		switch status {
		case "success", "delivered", "sent":
			// Simpu's "sent" status means delivered (confirmed by user)
			return StatusDelivered
		case "failed", "undelivered":
			return StatusFailed
		case "pending":
			return StatusPending
		default:
			return StatusUnknown
		}

	case "SIMPU_OTP":
		switch status {
		case "success", "delivered", "sent":
			// Simpu OTP's "sent" status means delivered
			return StatusDelivered
		case "failed", "undelivered":
			return StatusFailed
		case "pending":
			return StatusPending
		default:
			return StatusUnknown
		}

	case "EBULKSMS":
		switch status {
		case "delivered":
			return StatusDelivered
		case "sent":
			return StatusSent
		case "failed", "rejected":
			return StatusFailed
		case "buffered":
			return StatusPending
		case "expired":
			return StatusExpired
		default:
			return StatusUnknown
		}

	case "AKUUK":
		switch status {
		case "delivered", "success":
			return StatusDelivered
		case "sent", "submitted":
			return StatusSent
		case "pending":
			return StatusPending
		case "failed", "failure":
			return StatusFailed
		case "expired":
			return StatusExpired
		case "undelivered", "unreachable":
			return StatusUnreachable
		default:
			return StatusUnknown
		}

	default:
		return StatusUnknown
	}
}
