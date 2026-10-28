package notification

import (
	"context"
	"log"
	"time"
)

// Notifier defines the interface for sending notifications.
type Notifier interface {
	Send(ctx context.Context, recipient, subject, body string) error
}

// LogNotifier is a simple implementation that just logs notifications.
type LogNotifier struct{}

func NewLogNotifier() *LogNotifier {
	return &LogNotifier{}
}

// Send logs the notification details. Replace with actual email/SMS sending logic.
func (n *LogNotifier) Send(ctx context.Context, recipient, subject, body string) error {
	// Simulate network delay and potential context cancellation
	select {
	case <-time.After(50 * time.Millisecond):
		log.Printf("NOTIFICATION to %s: Subject='%s', Body='%s'", recipient, subject, body)
		return nil // Assume success for logging
	case <-ctx.Done():
		log.Printf("Notification sending cancelled for recipient %s", recipient)
		return ctx.Err()
	}
}

// Compile-time check to ensure LogNotifier implements Notifier
var _ Notifier = (*LogNotifier)(nil)
