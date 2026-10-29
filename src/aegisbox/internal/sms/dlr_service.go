package sms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/mno"
	"github.com/thrillee/aegisbox/internal/sp"
	"github.com/thrillee/aegisbox/pkg/codes"
)

type DLRService struct {
	dbQueries           database.Querier
	walletService       WalletService
	senderHealthTracker *SenderHealthTracker
	routeHealthTracker  *RouteHealthTracker
	batchHealthWorker   *BatchHealthWorker
}

type DLRServiceDependencies struct {
	DBQueries           database.Querier
	WalletService       WalletService
	SenderHealthTracker *SenderHealthTracker
	RouteHealthTracker  *RouteHealthTracker
	BatchHealthWorker   *BatchHealthWorker
}

func NewDLRService(deps DLRServiceDependencies) *DLRService {
	return &DLRService{
		dbQueries:           deps.DBQueries,
		walletService:       deps.WalletService,
		senderHealthTracker: deps.SenderHealthTracker,
		routeHealthTracker:  deps.RouteHealthTracker,
		batchHealthWorker:   deps.BatchHealthWorker,
	}
}

func (s *DLRService) UpdateSegmentDLRStatus(
	ctx context.Context,
	connectorID int32,
	dlrInfo mno.DLRInfo,
) error {
	logCtx := ctx
	logCtx = logging.ContextWithMNOConnID(logCtx, connectorID)
	logCtx = logging.ContextWithMNOMsgID(logCtx, dlrInfo.MnoMessageID)

	slog.InfoContext(logCtx, "Processing received DLR",
		slog.String("mno_msg_id", dlrInfo.MnoMessageID),
		slog.Int64("og_msg_id", dlrInfo.OriginalMsgID),
		slog.String("dlr_status", dlrInfo.Status),
		slog.Any("dlr_error_code", dlrInfo.ErrorCode),
		slog.Time("dlr_timestamp", dlrInfo.Timestamp),
	)

	segmentInfo, err := s.dbQueries.FindSegmentByMnoMessageID(logCtx, &dlrInfo.MnoMessageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(logCtx, "Received DLR for unknown/unmatched MNO message ID", "messageID", dlrInfo.MnoMessageID)
			return nil
		}
		slog.ErrorContext(logCtx, "DB error finding segment for DLR", slog.Any("error", err))
		return fmt.Errorf("db error finding segment for DLR (MNO ID %s): %w", dlrInfo.MnoMessageID, err)
	}

	logCtx = logging.ContextWithMessageID(logCtx, segmentInfo.MessageID)
	logCtx = logging.ContextWithSegmentID(logCtx, segmentInfo.SegmentID)

	err = s.dbQueries.UpdateSegmentDLR(logCtx, database.UpdateSegmentDLRParams{
		ID:        segmentInfo.SegmentID,
		DlrStatus: &dlrInfo.Status,
		ErrorCode: &dlrInfo.ErrorCode,
	})
	if err != nil {
		slog.ErrorContext(
			logCtx,
			"CRITICAL: Failed to update segment DLR status in DB",
			slog.Any("error", err),
		)
		return fmt.Errorf("failed to update segment DLR status: %w", err)
	}
	slog.InfoContext(logCtx, "Updated segment DLR status in DB",
		slog.Int("segment_seqn", int(segmentInfo.SegmentSeqn)),
		slog.String("dlr_status", dlrInfo.Status),
	)

	if s.batchHealthWorker != nil && segmentInfo.ApprovedSenderID != nil && segmentInfo.RoutedMnoID != nil {
		dlrStatus := strings.ToUpper(dlrInfo.Status)
		var deliveryStatus string
		switch dlrStatus {
		case "DELIVRD":
			deliveryStatus = "delivered"
		case "EXPIRED", "DELETED", "UNDELIV", "REJECTD":
			deliveryStatus = "failed"
		default:
			deliveryStatus = "unknown"
		}

		if deliveryStatus != "unknown" {
			sentAt := time.Now()
			if segmentInfo.SentToMnoAt.Valid {
				sentAt = segmentInfo.SentToMnoAt.Time
			}

			var errorCodePtr *string
			if dlrInfo.ErrorCode != "" {
				errorCodePtr = &dlrInfo.ErrorCode
			}

			record := HealthRecord{
				SenderID:         *segmentInfo.ApprovedSenderID,
				MNOID:            *segmentInfo.RoutedMnoID,
				MessageSegmentID: &segmentInfo.SegmentID,
				Status:           deliveryStatus,
				ErrorCode:        errorCodePtr,
				SentAt:           sentAt,
				DLRReceivedAt:    &dlrInfo.Timestamp,
			}
			s.batchHealthWorker.RecordHealth(record)
		}
	} else if s.senderHealthTracker != nil && segmentInfo.ApprovedSenderID != nil && segmentInfo.RoutedMnoID != nil {
		dlrStatus := strings.ToUpper(dlrInfo.Status)
		var deliveryStatus string
		switch dlrStatus {
		case "DELIVRD":
			deliveryStatus = "delivered"
		case "EXPIRED", "DELETED", "UNDELIV", "REJECTD":
			deliveryStatus = "failed"
		default:
			deliveryStatus = "unknown"
		}

		if deliveryStatus != "unknown" {
			sentAt := time.Now()
			if segmentInfo.SentToMnoAt.Valid {
				sentAt = segmentInfo.SentToMnoAt.Time
			}

			err = s.senderHealthTracker.RecordDeliveryResult(
				logCtx,
				*segmentInfo.ApprovedSenderID,
				*segmentInfo.RoutedMnoID,
				segmentInfo.SegmentID,
				deliveryStatus,
				&dlrInfo.ErrorCode,
				sentAt,
				&dlrInfo.Timestamp,
			)
			if err != nil {
				slog.WarnContext(logCtx, "Failed to record delivery result in health tracker",
					slog.Int("sender_id", int(*segmentInfo.ApprovedSenderID)),
					slog.Int("mno_id", int(*segmentInfo.RoutedMnoID)),
					slog.Any("error", err))
			}
		}
	}

	if s.routeHealthTracker != nil && segmentInfo.RoutingAssignmentID != nil {
		dlrStatus := strings.ToUpper(dlrInfo.Status)
		routeSuccess := dlrStatus == "DELIVRD"

		err = s.routeHealthTracker.RecordDeliveryOutcome(
			logCtx,
			*segmentInfo.RoutingAssignmentID,
			&segmentInfo.MessageID,
			routeSuccess,
			&dlrInfo.ErrorCode,
		)
		if err != nil {
			slog.WarnContext(logCtx, "Failed to record route health",
				slog.Int("routing_assignment_id", int(*segmentInfo.RoutingAssignmentID)),
				slog.Any("error", err))
		}
	}

	go s.aggregateMessageStatus(logCtx, segmentInfo.MessageID)
	go s.enqueueDLRForForwarding(
		logging.ContextWithMessageID(context.Background(), segmentInfo.MessageID),
		segmentInfo.MessageID,
		dlrInfo,
	)

	return nil
}

func (s *DLRService) aggregateMessageStatus(ctx context.Context, messageID int64) {
	slog.DebugContext(ctx, "Starting message status aggregation")

	maxRetries := 5
	baseDelay := 50 * time.Millisecond

	var segments []database.GetSegmentStatusesForMessageRow
	var totalSegmentsDB int32

	for attempt := 0; attempt < maxRetries; attempt++ {
		var err error
		segments, err = s.dbQueries.GetSegmentStatusesForMessage(ctx, messageID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			slog.ErrorContext(
				ctx,
				"Error fetching segments for status aggregation",
				slog.Any("error", err),
			)
			return
		}

		totalSegmentsDB, err = s.dbQueries.GetMessageTotalSegments(ctx, messageID)
		if err != nil {
			slog.ErrorContext(
				ctx,
				"Error fetching total segment count for status aggregation",
				slog.Any("error", err),
			)
			return
		}

		if len(segments) >= int(totalSegmentsDB) {
			break
		}

		if attempt < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<attempt)
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
		}
	}

	var deliveredCount, failedCount, pendingCount int
	finalStatus := "unknown"

	for _, seg := range segments {
		status := ""
		if seg.DlrStatus != nil {
			status = strings.ToUpper(*seg.DlrStatus)
		}
		switch status {
		case "DELIVRD":
			deliveredCount++
		case "EXPIRED", "DELETED", "UNDELIV", "REJECTD":
			failedCount++
		case "":
			pendingCount++
		default:
			pendingCount++
		}
	}

	if pendingCount > 0 {
		finalStatus = codes.MsgFinalStatusProcessing
	} else if failedCount > 0 {
		finalStatus = codes.MsgFinalStatusFailed
	} else if deliveredCount > 0 && deliveredCount >= int(totalSegmentsDB) {
		finalStatus = codes.MsgFinalStatusDelivered
	}

	// A failed DLR does not refund the message: it reached the operator and is billed. Refunds
	// for failed deliveries are made by hand from the wallet.

	if finalStatus == codes.MsgFinalStatusDelivered || finalStatus == codes.MsgFinalStatusFailed {
		currentStatusRec, err := s.dbQueries.GetMessageFinalStatus(ctx, messageID)
		statusChanged := true
		if err == nil && currentStatusRec != "" && currentStatusRec == finalStatus {
			statusChanged = false
		} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(ctx, "Could not get current final status before aggregation update", slog.Any("error", err))
		}

		if statusChanged {
			slog.InfoContext(
				ctx,
				"Attempting to update final message status",
				slog.String("new_status", finalStatus),
			)
			aggErrCode := ""
			aggErrMsg := ""
			err = s.dbQueries.UpdateMessageFinalStatus(ctx, database.UpdateMessageFinalStatusParams{
				ID:               messageID,
				FinalStatus:      finalStatus,
				ErrorCode:        &aggErrCode,
				ErrorDescription: &aggErrMsg,
			})
			if err != nil {
				slog.ErrorContext(
					ctx,
					"Error updating final message status",
					slog.String("attempted_status", finalStatus),
					slog.Any("error", err),
				)
			} else {
				slog.InfoContext(ctx, "Aggregated final message status updated successfully", slog.String("final_status", finalStatus))
			}
		} else {
			slog.DebugContext(ctx, "Final status already matches target, skipping update", slog.String("status", finalStatus))
		}
	} else {
		slog.DebugContext(ctx, "Aggregation determined non-terminal status, no update performed", slog.String("status", finalStatus))
	}
}

// buildDLRForwardingPayload builds the dlr_forwarding_queue payload for a message's DLR.
// It also returns the SP details so callers can check the DLR can actually be forwarded.
func (s *DLRService) buildDLRForwardingPayload(
	ctx context.Context,
	messageID int64,
	dlrInfo mno.DLRInfo,
) ([]byte, sp.SPDetails, error) {
	spMsgInfo, err := s.dbQueries.GetSPMessageInfoForDLR(ctx, messageID)
	if err != nil {
		return nil, sp.SPDetails{}, fmt.Errorf("failed to get SP/message info for DLR: %w", err)
	}

	spDetails := sp.SPDetails{
		ServiceProviderID: spMsgInfo.ServiceProviderID,
		Protocol:          spMsgInfo.Protocol,
	}

	if spMsgInfo.SystemID != nil {
		spDetails.SMPPSystemID = *spMsgInfo.SystemID
	}

	callbackUrl := sp.GetHttpCallBackURL(spMsgInfo.HttpConfig)
	authConfig := sp.GetHttpCallBackKey(spMsgInfo.HttpConfig)

	if callbackUrl != nil {
		spDetails.HTTPCallbackURL = *callbackUrl
	}

	if authConfig != nil {
		spDetails.HTTPAuthConfig = *authConfig
	}

	// An empty error code must be sent as no error code: the forwarders map any
	// non-nil code, and "" maps to their error default (HTTP 500 / SMPP 008), which
	// would tell the SP a delivered message failed.
	var errorCode *string
	if dlrInfo.ErrorCode != "" {
		errorCode = &dlrInfo.ErrorCode
	}

	forwardedDLR := sp.ForwardedDLRInfo{
		InternalMessageID: messageID,
		ClientMessageRef:  spMsgInfo.ClientRef,
		ClientMessageID:   spMsgInfo.ClientMessageID,
		SourceAddr:        spMsgInfo.OriginalDestinationAddr,
		DestAddr:          spMsgInfo.OriginalSourceAddr,
		SubmitDate:        spMsgInfo.SubmittedAt.Time,
		DoneDate:          dlrInfo.Timestamp,
		Status:            dlrInfo.Status,
		ErrorCode:         errorCode,
		TotalSegments:     spMsgInfo.TotalSegments,
	}

	payloadData := map[string]interface{}{
		"sp_details": spDetails,
		"dlr_info":   forwardedDLR,
	}
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		return nil, spDetails, fmt.Errorf("failed to marshal DLR forwarding payload: %w", err)
	}
	return payloadBytes, spDetails, nil
}

func (s *DLRService) enqueueDLRForForwarding(
	ctx context.Context,
	messageID int64,
	dlrInfo mno.DLRInfo,
) {
	logCtx := logging.ContextWithMessageID(ctx, messageID)
	slog.DebugContext(
		logCtx,
		"Attempting to enqueue success/failure DLR",
		slog.String("dlr_status", dlrInfo.Status),
	)

	payloadBytes, _, err := s.buildDLRForwardingPayload(logCtx, messageID, dlrInfo)
	if err != nil {
		slog.ErrorContext(
			logCtx,
			"CRITICAL: Cannot forward DLR: failed to build forwarding payload",
			slog.Any("error", err),
		)
		return
	}

	err = s.dbQueries.EnqueueDLRForForwarding(logCtx, database.EnqueueDLRForForwardingParams{
		MessageID:   messageID,
		Payload:     payloadBytes,
		MaxAttempts: DLRMaxRetries,
	})
	if err != nil {
		slog.ErrorContext(
			logCtx,
			"Failed to enqueue DLR forwarding job to DB",
			slog.Any("error", err),
		)
	} else {
		slog.InfoContext(logCtx, "Successfully enqueued DLR forwarding job to DB")
	}
}

func (s *DLRService) EnqueueFailureDLR(
	ctx context.Context,
	messageID int64,
	failureStatus string,
	failureReason error,
	errorCode *string,
) {
	logCtx := logging.ContextWithMessageID(ctx, messageID)
	slog.DebugContext(
		logCtx,
		"Attempting to enqueue internal failure DLR",
		slog.String("failure_status", failureStatus),
		slog.String("failure_reason", failureReason.Error()),
	)

	spMsgInfo, err := s.dbQueries.GetSPMessageInfoForDLR(logCtx, messageID)
	if err != nil {
		slog.ErrorContext(
			logCtx,
			"Cannot forward DLR: Failed to get SP/Message info",
			slog.Any("error", err),
		)
		return
	}

	httpCallbackURL := sp.GetHttpCallBackURL(spMsgInfo.HttpConfig)
	httpCallbackKey := sp.GetHttpCallBackKey(spMsgInfo.HttpConfig)
	var callbackURL string
	var callbackKey string
	if httpCallbackURL != nil {
		callbackURL = *httpCallbackURL
	} else {
		callbackURL = ""
	}

	if httpCallbackKey != nil {
		callbackKey = *httpCallbackKey
	} else {
		callbackKey = ""
	}

	spDetails := sp.SPDetails{
		ServiceProviderID: spMsgInfo.ServiceProviderID,
		Protocol:          spMsgInfo.Protocol,
		SMPPSystemID:      *spMsgInfo.SystemID,
		HTTPCallbackURL:   callbackURL,
		HTTPAuthConfig:    callbackKey,
	}

	dlrStatusString := "UNDELIV"
	if failureStatus == codes.MsgStatusFailedValidation || failureStatus == codes.MsgStatusFailedRouting {
		dlrStatusString = "REJECTD"
	}

	forwardedDLR := sp.ForwardedDLRInfo{
		InternalMessageID: messageID,
		ClientMessageRef:  spMsgInfo.ClientRef,
		ClientMessageID:   spMsgInfo.ClientMessageID,
		SourceAddr:        spMsgInfo.OriginalDestinationAddr,
		DestAddr:          spMsgInfo.OriginalSourceAddr,
		SubmitDate:        spMsgInfo.SubmittedAt.Time,
		DoneDate:          time.Now(),
		Status:            dlrStatusString,
		ErrorCode:         errorCode,
		TotalSegments:     spMsgInfo.TotalSegments,
	}

	payloadData := map[string]interface{}{
		"sp_details": spDetails,
		"dlr_info":   forwardedDLR,
	}
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to marshal failure DLR payload", slog.Any("error", err))
		return
	}

	err = s.dbQueries.EnqueueDLRForForwarding(logCtx, database.EnqueueDLRForForwardingParams{
		MessageID:   messageID,
		Payload:     payloadBytes,
		MaxAttempts: DLRMaxRetries,
	})
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to enqueue failure DLR", slog.Any("error", err))
	} else {
		slog.InfoContext(logCtx, "Successfully enqueued failure DLR")
	}
}
