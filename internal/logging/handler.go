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
	MNOConnIDKey contextKey = "mno_conn_id"
	MNOMsgIDKey  contextKey = "mno_msg_id"
	SegMsgIDKey  contextKey = "seg_msg_id"
	MSISDNKey    contextKey = "msisdn"
	WalletIDKey  contextKey = "wallet_id"
	CallBackKey  contextKey = "callback_url"
	WorkerID     contextKey = "worker_id"
	JobIDKey     contextKey = "job_id"
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

func ContextWithMNOConnID(ctx context.Context, connID int32) context.Context {
	return context.WithValue(ctx, MNOConnIDKey, connID)
}

func ContextWithMNOMsgID(ctx context.Context, msgID string) context.Context {
	return context.WithValue(ctx, MNOMsgIDKey, msgID)
}

func ContextWithSegmentID(ctx context.Context, segmentID int64) context.Context {
	return context.WithValue(ctx, SegMsgIDKey, segmentID)
}

func ContextWithMSISDN(ctx context.Context, msisdn string) context.Context {
	return context.WithValue(ctx, MSISDNKey, msisdn)
}

func ContextWithWalletID(ctx context.Context, walletID int32) context.Context {
	return context.WithValue(ctx, WalletIDKey, walletID)
}

func ContextWithCallbackURL(ctx context.Context, callbackURL string) context.Context {
	return context.WithValue(ctx, CallBackKey, callbackURL)
}

func ContextWithWorkerID(ctx context.Context, workerID string) context.Context {
	return context.WithValue(ctx, CallBackKey, workerID)
}

func ContextWithJobID(ctx context.Context, jobID int64) context.Context {
	return context.WithValue(ctx, JobIDKey, jobID)
}
