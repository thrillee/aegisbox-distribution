package mno

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	DLRKeyPrefix      = "dlr:callback:"
	DLRProviderKey    = "dlr:provider:"
	DLRRetryKey       = "dlr:retry_queue"
	HTTPClientTimeout = 10 * time.Second
)

type DLRForwarder struct {
	redisClient *redis.Client
	logger      *slog.Logger
	httpClient  *http.Client
}

func NewDLRForwarder(redisClient *redis.Client, logger *slog.Logger) *DLRForwarder {
	return &DLRForwarder{
		redisClient: redisClient,
		logger:      logger,
		httpClient: &http.Client{
			Timeout: HTTPClientTimeout,
		},
	}
}

func (d *DLRForwarder) StoreDLRMapping(ctx context.Context, messageID, callbackURL string) error {
	err := d.redisClient.Set(ctx, DLRKeyPrefix+messageID, callbackURL, 7*24*time.Hour).Err()
	if err != nil {
		d.logger.ErrorContext(ctx, "Failed to store DLR mapping",
			"message_id", messageID,
			"callback_url", callbackURL,
			"error", err,
		)
		return err
	}

	d.logger.DebugContext(ctx, "DLR mapping stored",
		"message_id", messageID,
		"callback_url", callbackURL,
	)
	return nil
}

func (d *DLRForwarder) GetCallbackURL(ctx context.Context, messageID string) (string, error) {
	callbackURL, err := d.redisClient.Get(ctx, DLRKeyPrefix+messageID).Result()
	if err != nil {
		d.logger.WarnContext(ctx, "DLR mapping not found",
			"message_id", messageID,
		)
		return "", err
	}

	d.logger.DebugContext(ctx, "DLR mapping retrieved",
		"message_id", messageID,
		"callback_url", callbackURL,
	)
	return callbackURL, nil
}

func (d *DLRForwarder) StoreProviderName(ctx context.Context, messageID, providerName string) error {
	return d.redisClient.Set(ctx, DLRProviderKey+messageID, providerName, 7*24*time.Hour).Err()
}

func (d *DLRForwarder) GetProviderName(ctx context.Context, messageID string) (string, error) {
	return d.redisClient.Get(ctx, DLRProviderKey+messageID).Result()
}

func (d *DLRForwarder) ForwardDLR(ctx context.Context, messageID, status, callbackURL string) error {
	payloadData := map[string]any{
		"message_id": messageID,
		"status":     mapStatusToProviderFormat(status),
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	payload, err := json.Marshal(payloadData)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", callbackURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Aegisbox-DLR/1.0")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, _ = io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return &DLRForwardingError{
		StatusCode: resp.StatusCode,
		Message:    string(payload),
	}
}

func (d *DLRForwarder) ForwardDLRWithRetry(ctx context.Context, messageID, status, callbackURL string) {
	go func() {
		logger := d.logger.With(
			"message_id", messageID,
			"callback_url", callbackURL,
			"status", status,
		)

		logger.InfoContext(ctx, "Starting DLR forwarding with retry")

		backoff := []time.Duration{1, 5, 15, 30, 60, 60, 120, 180}

		for attempt, delay := range backoff {
			time.Sleep(delay * time.Second)

			err := d.ForwardDLR(ctx, messageID, status, callbackURL)
			if err != nil {
				logger.WarnContext(ctx, "DLR forwarding attempt failed",
					"attempt", attempt+1,
					"delay_seconds", delay,
					"error", err,
				)
				continue
			}

			logger.InfoContext(ctx, "DLR forwarded successfully",
				"attempt", attempt+1,
			)
			return
		}

		logger.ErrorContext(ctx, "All DLR forwarding attempts failed")
	}()
}

func (d *DLRForwarder) ProcessDLRWebhook(ctx context.Context, providerName string, req *http.Request) error {
	provider := NewHTTPProvider(providerName, HTTPProviderConfig{}, d.logger)
	if provider == nil {
		return &UnknownProviderError{Name: providerName}
	}

	messageID, status, err := provider.ParseDLR(ctx, req)
	if err != nil {
		return err
	}

	callbackURL, err := d.GetCallbackURL(ctx, messageID)
	if err != nil {
		d.logger.WarnContext(ctx, "No callback URL found for DLR",
			"message_id", messageID,
		)
		return nil
	}

	d.ForwardDLRWithRetry(ctx, messageID, status, callbackURL)
	return nil
}

type DLRForwardingError struct {
	StatusCode int
	Message    string
}

func (e *DLRForwardingError) Error() string {
	return "DLR forwarding failed with status " + itoa(e.StatusCode) + ": " + e.Message
}

type UnknownProviderError struct {
	Name string
}

func (e *UnknownProviderError) Error() string {
	return "unknown DLR provider: " + e.Name
}

func mapStatusToProviderFormat(status string) string {
	statusMap := map[string]string{
		"delivered": "DELIVRD",
		"failed":    "UNDELIV",
		"sent":      "SENT",
		"pending":   "PENDING",
		"unknown":   "UNKNOWN",
	}

	if mapped, exists := statusMap[status]; exists {
		return mapped
	}
	return "UNKNOWN"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
