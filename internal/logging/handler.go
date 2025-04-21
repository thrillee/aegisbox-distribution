package logging

import (
	"context"
	"log/slog"
)

type contextKey string

const (
	SPIDKey      contextKey = "sp_id"
	SystemIDKey  contextKey = "system_id"
	MessageIDKey contextKey = "msg_id"
	MNOIDKey     contextKey = "mno_id"
	// Add other keys as needed
)

// ContextHandler wraps another slog.Handler and adds attributes from context.
type ContextHandler struct {
	slog.Handler
}

// NewContextHandler creates a handler that extracts values from context.
func NewContextHandler(h slog.Handler) *ContextHandler {
	return &ContextHandler{Handler: h}
}

// Handle adds context attributes before calling the wrapped handler.
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if spID, ok := ctx.Value(SPIDKey).(int32); ok {
		r.AddAttrs(slog.Int("sp_id", int(spID)))
	}
	if sysID, ok := ctx.Value(SystemIDKey).(string); ok {
		r.AddAttrs(slog.String("system_id", sysID))
	}
	if msgID, ok := ctx.Value(MessageIDKey).(int64); ok {
		r.AddAttrs(slog.Int64("msg_id", msgID))
	}
	if mnoID, ok := ctx.Value(MNOIDKey).(int32); ok {
		r.AddAttrs(slog.Int("mno_id", int(mnoID)))
	}
	// Add more context value extractions here

	return h.Handler.Handle(ctx, r)
}

// Helper functions to add values to context
func ContextWithSPID(ctx context.Context, spID int32) context.Context {
	return context.WithValue(ctx, SPIDKey, spID)
}

func ContextWithSystemID(ctx context.Context, systemID string) context.Context {
	return context.WithValue(ctx, SystemIDKey, systemID)
}

func ContextWithMessageID(ctx context.Context, msgID int64) context.Context {
	return context.WithValue(ctx, MessageIDKey, msgID)
}

func ContextWithMNOID(ctx context.Context, mnoID int32) context.Context {
	return context.WithValue(ctx, MNOIDKey, mnoID)
}
