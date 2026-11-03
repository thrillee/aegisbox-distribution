package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/pkg/codes"
)

// LinkHoldRejecter rejects messages the link check stopped (link scanning Phase 2, block mode or
// an admin's rejection): the message is never sent, gets final status rejected and a REJECTD DLR,
// and its debit is reversed unless the customer is set to be charged for blocked messages.
type LinkHoldRejecter struct {
	q      database.Querier
	dlr    *DLRService
	wallet WalletService
}

// NewLinkHoldRejecter builds a rejecter outside the processor (manager API).
func NewLinkHoldRejecter(q database.Querier, dlr *DLRService, wallet WalletService) *LinkHoldRejecter {
	return &LinkHoldRejecter{q: q, dlr: dlr, wallet: wallet}
}

// LinkBlockedErrorCode is the error code recorded and forwarded for messages blocked by the link check.
const LinkBlockedErrorCode = "LINK_BLOCKED"

func (r *LinkHoldRejecter) RejectForLinks(ctx context.Context, messageID int64, reason string, refund bool) error {
	logCtx := logging.ContextWithMessageID(ctx, messageID)
	n, err := r.q.SetMessageProcessingStatus(logCtx, database.SetMessageProcessingStatusParams{
		ID: messageID, FromStatus: codes.MsgStatusHeldLinkScan, ToStatus: codes.MsgStatusBlockedLinkScan,
	})
	if err != nil {
		return fmt.Errorf("mark message blocked: %w", err)
	}
	if n == 0 {
		// Not held any more (sent or rejected meanwhile): leave it alone.
		return errors.New("message is no longer held")
	}
	code := LinkBlockedErrorCode
	desc := reason
	if err := r.q.UpdateMessageFinalStatus(logCtx, database.UpdateMessageFinalStatusParams{
		ID: messageID, FinalStatus: codes.MsgFinalStatusRejected, ErrorCode: &code, ErrorDescription: &desc,
	}); err != nil {
		slog.ErrorContext(logCtx, "Link block: could not set final status", slog.Any("error", err))
	}
	if r.dlr != nil {
		r.dlr.EnqueueFailureDLR(logCtx, messageID, codes.MsgStatusBlockedLinkScan, errors.New(reason), &code)
	}
	if refund && r.wallet != nil {
		if err := r.wallet.ReverseDebit(logCtx, messageID, reason); err != nil {
			slog.ErrorContext(logCtx, "Link block: debit reversal failed", slog.Any("error", err))
		}
	}
	return nil
}
