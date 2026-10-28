package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/mno"
	"github.com/thrillee/aegisbox/pkg/codes"
)

type SendingService struct {
	dbQueries           database.Querier
	sender              Sender
	walletService       WalletService
	routeHealthTracker  *RouteHealthTracker
	batchReversalWorker *BatchReversalWorker
}

type SendingServiceDependencies struct {
	DBQueries           database.Querier
	Sender              Sender
	WalletService       WalletService
	RouteHealthTracker  *RouteHealthTracker
	BatchReversalWorker *BatchReversalWorker
}

func NewSendingService(deps SendingServiceDependencies) *SendingService {
	return &SendingService{
		dbQueries:           deps.DBQueries,
		sender:              deps.Sender,
		walletService:       deps.WalletService,
		routeHealthTracker:  deps.RouteHealthTracker,
		batchReversalWorker: deps.BatchReversalWorker,
	}
}

// SetSender allows updating the sender after initialization (needed for circular dependency resolution)
func (s *SendingService) SetSender(sender Sender) {
	s.sender = sender
}

func (s *SendingService) SendMessage(
	ctx context.Context,
	msg database.GetMessagesByStatusRow,
	router Router,
) (bool, error) {
	logCtx := logging.ContextWithMessageID(ctx, msg.ID)
	logCtx = logging.ContextWithSPID(logCtx, msg.ServiceProviderID)

	fullMsg, dbErr := s.dbQueries.GetMessageDetailsForSending(logCtx, msg.ID)
	if dbErr != nil {
		slog.ErrorContext(
			logCtx,
			"CRITICAL: Error fetching full details for sending message",
			slog.Any("error", dbErr),
		)
		sendErr := fmt.Errorf("failed to fetch details: %w", dbErr)
		s.attemptReversal(logCtx, msg.ID, sendErr.Error())
		s.markAsFailed(logCtx, msg.ID, sendErr, "SYS_ERR")
		return false, nil
	}

	if fullMsg.RoutedMnoID == nil {
		sendErr := errors.New("missing Routed MNO ID")
		slog.ErrorContext(logCtx, "Cannot send message", slog.Any("reason", sendErr))
		s.attemptReversal(logCtx, msg.ID, sendErr.Error())
		s.markAsFailed(logCtx, msg.ID, sendErr, "ROUTE_ERR")
		return false, nil
	}
	logCtx = logging.ContextWithMNOID(logCtx, *fullMsg.RoutedMnoID)

	var routingCandidates []RouteCandidate
	var routingAssignmentID *int32
	var maxRetries int32 = 0

	if intelligentRouter, ok := router.(*IntelligentRouter); ok {
		routingResult, routeErr := intelligentRouter.RouteWithCandidates(logCtx, RoutingRequest{
			SPCredentialID: msg.SpCredentialID,
			MSISDN:         fullMsg.OriginalDestinationAddr,
			SenderID:       fullMsg.OriginalSourceAddr,
		})

		if routeErr == nil && routingResult != nil && routingResult.RoutingResult.ShouldUse {
			routingCandidates = routingResult.Candidates
			if routingResult.RoutingConfig != nil && routingResult.RoutingConfig.MaxRouteRetries != nil {
				maxRetries = *routingResult.RoutingConfig.MaxRouteRetries
			}
			slog.DebugContext(logCtx, "Retrieved routing candidates for failover",
				slog.Int("total_candidates", len(routingCandidates)),
				slog.Int("max_retries", int(maxRetries)),
			)
		}
	}

	var submitResult *mno.SubmitResult
	var sendErr error
	attemptCount := 0
	maxAttempts := int(maxRetries) + 1

	for attemptCount < maxAttempts && attemptCount < len(routingCandidates)+1 {
		var currentRouteCandidate *RouteCandidate

		if attemptCount < len(routingCandidates) {
			currentRouteCandidate = &routingCandidates[attemptCount]
			routingAssignmentID = &currentRouteCandidate.RoutingAssignmentID

			if attemptCount > 0 {
				slog.InfoContext(logCtx, "Attempting failover to alternate route",
					slog.Int("attempt", attemptCount+1),
					slog.Int("routing_assignment_id", int(currentRouteCandidate.RoutingAssignmentID)),
					slog.Int("mno_id", int(currentRouteCandidate.MNOID)),
					slog.String("mno_name", currentRouteCandidate.MNOName),
				)

				fullMsg.RoutedMnoID = &currentRouteCandidate.MNOID
				logCtx = logging.ContextWithMNOID(logCtx, currentRouteCandidate.MNOID)
			}
		}
		submitResult, sendErr = s.sender.Send(logCtx, fullMsg)
		submitResult.MnoID = currentRouteCandidate.MNOID

		if s.routeHealthTracker != nil && routingAssignmentID != nil {
			routeSuccess := (sendErr == nil && submitResult != nil && submitResult.WasSubmitted && submitResult.Error == nil)
			var errorCode *string
			var errorDesc *string

			if !routeSuccess {
				if sendErr != nil {
					errStr := sendErr.Error()
					errorDesc = &errStr
				} else if submitResult != nil && submitResult.Error != nil {
					errStr := submitResult.Error.Error()
					errorDesc = &errStr
				}
				errCodeStr := "SEND_FAIL"
				errorCode = &errCodeStr
			}

			err := s.routeHealthTracker.RecordRouteAttempt(
				logCtx,
				*routingAssignmentID,
				fullMsg.ApprovedSenderID,
				&msg.ID,
				routeSuccess,
				errorCode,
				errorDesc,
			)
			if err != nil {
				slog.WarnContext(logCtx, "Failed to record route health attempt",
					slog.Int("routing_assignment_id", int(*routingAssignmentID)),
					slog.Bool("success", routeSuccess),
					slog.Any("error", err))
			}
		}

		slog.InfoContext(
			logCtx,
			"Send attempt completed",
			slog.Int("attempt", attemptCount+1),
			slog.Bool("was_submitted", submitResult != nil && submitResult.WasSubmitted),
			slog.Any("send_error", sendErr),
		)

		if sendErr == nil && submitResult != nil && submitResult.WasSubmitted && submitResult.Error == nil {
			slog.InfoContext(logCtx, "Message sent successfully",
				slog.Int("attempt", attemptCount+1),
			)
			break
		}

		attemptCount++
		if attemptCount < maxAttempts && attemptCount < len(routingCandidates) {
			slog.WarnContext(logCtx, "Send attempt failed, trying alternate route",
				slog.Int("failed_attempt", attemptCount),
				slog.Int("remaining_attempts", maxAttempts-attemptCount),
				slog.Any("failure_reason", sendErr),
			)
		} else {
			slog.ErrorContext(logCtx, "All send attempts exhausted",
				slog.Int("total_attempts", attemptCount),
			)
			break
		}
	}

	if sendErr != nil {
		slog.ErrorContext(
			logCtx,
			"Internal error during send attempt (all routes failed)",
			slog.Int("total_attempts", attemptCount),
			slog.Any("error", sendErr),
		)
		s.attemptReversal(logCtx, msg.ID, fmt.Sprintf("Internal sender error: %v", sendErr))
		s.markAsFailed(logCtx, msg.ID, sendErr, "SYS_ERR")
		return false, nil
	}

	if submitResult == nil || !submitResult.WasSubmitted || submitResult.Error != nil {
		failureReason := errors.New("MNO submission failed or was not attempted")
		if submitResult != nil && submitResult.Error != nil {
			failureReason = submitResult.Error
		}
		errorCode := "MNO_FAIL"
		slog.WarnContext(logCtx, "MNO submission failed (all routes)",
			slog.Int("total_attempts", attemptCount),
			slog.Any("reason", failureReason))
		s.attemptReversal(
			logCtx,
			msg.ID,
			fmt.Sprintf("MNO submission failed: %v", failureReason),
		)
		s.markAsFailed(logCtx, msg.ID, failureReason, errorCode)
		s.markSendAttempted(ctx, msg.ID, submitResult.MnoID)
		return false, nil
	}

	updateErr := s.markSendAttempted(logCtx, msg.ID, submitResult.MnoID)
	if updateErr != nil {
		slog.ErrorContext(
			logCtx,
			"Failed to update message status to send_attempted",
			slog.Any("error", updateErr),
		)
	} else {
		slog.InfoContext(logCtx, "Message successfully processed for sending",
			slog.Int("retry_attempts", attemptCount),
		)
		return true, nil
	}

	return false, nil
}

func (s *SendingService) attemptReversal(ctx context.Context, msgID int64, reason string) {
	slog.WarnContext(ctx, "Attempting wallet debit reversal", slog.String("reason", reason))

	if s.batchReversalWorker != nil {
		s.batchReversalWorker.QueueReversal(msgID, reason)
		slog.InfoContext(ctx, "Reversal queued for batch processing")
	} else {
		reverseErr := s.walletService.ReverseDebit(ctx, msgID, reason)
		if reverseErr != nil {
			slog.ErrorContext(ctx, "CRITICAL: Debit reversal FAILED",
				slog.Any("reversal_error", reverseErr),
				slog.String("reason", reason),
			)
		} else {
			slog.InfoContext(ctx, "Debit reversal successful")
		}
	}
}

func (s *SendingService) markAsFailed(
	ctx context.Context,
	msgID int64,
	failureReason error,
	errorCode string,
) error {
	errMsg := "processing failed"
	if failureReason != nil {
		errMsg = failureReason.Error()
	}
	dbErrMsg := errMsg
	dbErrCode := errorCode

	err := s.dbQueries.UpdateMessageFinalStatus(ctx, database.UpdateMessageFinalStatusParams{
		ID:               msgID,
		FinalStatus:      codes.MsgFinalStatusFailed,
		ErrorCode:        &dbErrCode,
		ErrorDescription: &dbErrMsg,
	})
	if err == nil {
		slog.WarnContext(
			ctx,
			"Message marked as failed",
			slog.String("reason", errMsg),
			slog.String("error_code", errorCode),
		)
	} else {
		slog.ErrorContext(ctx, "Failed to update message status to failed", slog.Any("db_error", err))
	}
	return err
}

func (s *SendingService) markSendAttempted(ctx context.Context, msgID int64, mnoID int32) error {
	err := s.dbQueries.UpdateMessageSendAttempted(ctx, database.UpdateMessageSendAttemptedParams{ID: msgID, RoutedMnoID: &mnoID})
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Failed to update message status to send_attempted",
			slog.Any("db_error", err),
		)
	}
	return err
}
