package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/mno"
	"github.com/thrillee/aegisbox/pkg/segmenter"
)

// Compile-time check
var _ Sender = (*DefaultSender)(nil)

type DefaultSender struct {
	dbQueries            database.Querier
	mnoConnectorProvider mno.ConnectorProvider // Use the interface to get MNO connectors
	segmenter            segmenter.Segmenter   // Interface for message segmentation
}

// NewDefaultSender creates a sender instance.
func NewDefaultSender(q database.Querier, provider mno.ConnectorProvider, seg segmenter.Segmenter) *DefaultSender {
	return &DefaultSender{
		dbQueries:            q,
		mnoConnectorProvider: provider,
		segmenter:            seg,
	}
}

// Send handles segmentation, MNO connector interaction, and recording segment results.
func (s *DefaultSender) Send(ctx context.Context, msg database.GetMessageDetailsForSendingRow) (*mno.SubmitResult, error) {
	// Use context enriched by the caller (Processor) which includes msgID, spID, mnoID
	logCtx := ctx
	slog.DebugContext(logCtx, "Starting send process")

	// 1. Get MNO Connector
	if msg.RoutedMnoID == nil {
		// This should ideally be caught earlier, but double-check
		err := errors.New("cannot send message: Routed MNO ID is invalid or missing")
		slog.ErrorContext(logCtx, err.Error())
		// Return a SubmitResult indicating failure before any attempt
		return &mno.SubmitResult{WasSubmitted: false, Error: err}, nil // Return nil error for logical failure
	}
	mnoID := msg.RoutedMnoID

	connector, err := s.mnoConnectorProvider.GetConnector(logCtx, *mnoID)
	if err != nil {
		err = fmt.Errorf("failed to get active MNO connector for MNO ID %d: %w", mnoID, err)
		slog.ErrorContext(logCtx, err.Error())
		// Treat as potentially retryable infrastructure issue? For now, fail the submission.
		return &mno.SubmitResult{WasSubmitted: false, Error: err}, nil // Return nil error for logical failure
	}
	logCtx = logging.ContextWithMNOConnID(logCtx, connector.ConnectionID()) // Add Conn ID for tracing

	senderID, err := s.resolveSenderID(logCtx, msg, connector.ConnectionID())
	if err != nil {
		slog.WarnContext(logCtx, "No sender available for this connection", slog.Any("error", err))
		// Fail this attempt before creating segments, so failover can try the next route.
		return &mno.SubmitResult{WasSubmitted: false, Error: err}, nil
	}

	// 2. Prepare Message for Connector (Segmentation handled by Connector)
	// The Connector interface expects the full message content.
	preparedMsg := mno.PreparedMessage{
		InternalMessageID: msg.ID,
		TotalSegments:     msg.TotalSegments, // Inform connector about expected segments (for UDH?)
		SenderID:          senderID,
		DestinationMSISDN: msg.OriginalDestinationAddr,
		MessageContent:    msg.ShortMessage,
		RequestDLR:        true, // TODO: Make this configurable per SP/message?
		// Populate DataCoding, ESMClass, IsFlash based on msg fields if they exist
		// DataCoding: msg.DataCoding,
		// ESMClass: msg.EsmClass,
		// MNOInfo: // Fetch MNO details if needed by connector?
	}

	// 3. Create Segment Records in DB *before* submission attempt
	// The connector will update these with MNO Msg IDs / errors.
	// We store the segment IDs to link back the results.
	segmentIDs := make([]int64, msg.TotalSegments)
	for i := range msg.TotalSegments {
		segID, dbErr := s.dbQueries.CreateMessageSegment(logCtx, database.CreateMessageSegmentParams{
			MessageID:   msg.ID,
			SegmentSeqn: int32(i + 1),
		})
		if dbErr != nil {
			err = fmt.Errorf("failed to create DB record for segment %d/%d: %w", i+1, msg.TotalSegments, dbErr)
			slog.ErrorContext(logCtx, err.Error())
			// This is critical, fail the whole submission
			return &mno.SubmitResult{WasSubmitted: false, Error: err}, err // Return internal error
		}
		segmentIDs[i] = segID
		slog.InfoContext(logCtx, "Arranging Segmets", slog.Any("segID", segID))
	}
	preparedMsg.DBSeqs = segmentIDs
	slog.InfoContext(logCtx, "Created initial segment records in DB", slog.Int("count", len(segmentIDs)), slog.Any("DB Segements ID", segmentIDs))

	// 4. Submit Message via Connector
	submitResult, submitErr := connector.SubmitMessage(logCtx, preparedMsg)
	// submitErr represents catastrophic errors during the submission process itself (e.g., connection drop)
	// submitResult contains per-segment outcomes.

	errDesc := fmt.Sprintf("Submission error: %v", submitErr)
	errCode := "CONN_ERR"

	if submitErr != nil {
		slog.ErrorContext(logCtx, "Catastrophic error during MNO submission", slog.Any("error", submitErr))
		// Mark all created segments as failed?
		for i, segInfo := range submitResult.Segments {
			_ = s.dbQueries.UpdateSegmentSendFailed(logCtx, database.UpdateSegmentSendFailedParams{
				ID:               int64(segInfo.Seqn),
				ErrorDescription: &errDesc,
				ErrorCode:        &errCode,
			})
			slog.WarnContext(logCtx, "Marked segment as failed due to overall submission error", slog.Int64("segment_id", int64(segInfo.Seqn)), slog.Int("seqn", i+1))
		}
		// Return the overall error wrapped in SubmitResult
		submitResult.Error = submitErr
		submitResult.WasSubmitted = false // Indicate catastrophic failure
		return &submitResult, nil         // Return logical failure indication
	}

	// 5. Record Segment-Specific Results in DB
	slog.InfoContext(logCtx,
		"Processing MNO submission results for segments",
		slog.Any("segements", submitResult.Segments),
		slog.Int("segment_results_count", len(submitResult.Segments)))
	updateErrors := false
	for _, segResult := range submitResult.Segments {
		// Find the corresponding DB segment ID (using Seqn assumes order is preserved)
		// A map might be safer if order isn't guaranteed.
		var dbSegmentID int64
		if segResult.DBSegID == nil {
			// Map the 1-based sequence number to the segment rows created above.
			if segResult.Seqn <= 0 || int(segResult.Seqn) > len(segmentIDs) {
				slog.ErrorContext(logCtx, "Invalid segment sequence number received from connector result", slog.Int("seqn", int(segResult.Seqn)))
				updateErrors = true
				continue
			}

			dbSegmentID = segmentIDs[segResult.Seqn-1]
		} else {
			dbSegmentID = int64(*segResult.DBSegID)
		}
		segLogCtx := logging.ContextWithSegmentID(logCtx, int64(dbSegmentID)) // Add Segment ID

		mnoConnID := connector.ConnectionID()

		slog.InfoContext(segLogCtx, "Processing Segment",
			slog.Any("IsSuccess", segResult.IsSuccess),
			slog.Any("dbSegmentId", dbSegmentID),
			slog.Any("mnoConnID", mnoConnID))
		if segResult.IsSuccess {
			// Update segment as sent successfully
			err = s.dbQueries.UpdateSegmentSent(segLogCtx, database.UpdateSegmentSentParams{
				ID:              int64(dbSegmentID),
				MnoMessageID:    segResult.MnoMessageID,
				MnoConnectionID: &mnoConnID,
			})
			if err != nil {
				slog.ErrorContext(segLogCtx, "Failed to update DB for successfully sent segment", slog.Any("error", err))
				updateErrors = true // Log DB error but continue processing results
			} else {
				slog.InfoContext(segLogCtx, "Segment sent successfully to MNO", slog.Any("dbSegmentID", dbSegmentID), slog.Any("mno_msg_id", segResult.MnoMessageID))
			}
		} else {
			// Update segment as failed
			errMsg := "Segment submission failed"
			if segResult.Error != nil {
				errMsg = segResult.Error.Error()
			}
			err = s.dbQueries.UpdateSegmentSendFailed(segLogCtx, database.UpdateSegmentSendFailedParams{
				ID:               int64(dbSegmentID),
				ErrorDescription: &errMsg,
				ErrorCode:        segResult.ErrorCode, // Pass sql.NullString directly
			})
			if err != nil {
				slog.ErrorContext(segLogCtx, "Failed to update DB for failed segment submission", slog.Any("error", err), slog.Any("original_error", errMsg), slog.Any("mno_err_code", segResult.ErrorCode))
				updateErrors = true
			} else {
				slog.WarnContext(segLogCtx, "Segment submission failed at MNO", slog.String("reason", errMsg), slog.Any("mno_err_code", segResult.ErrorCode))
			}
		}
	}

	if updateErrors {
		// If DB updates failed, the overall SubmitResult might need adjustment or specific error flagging
		slog.ErrorContext(logCtx, "Encountered one or more errors updating segment statuses in DB after submission")
		// Set overall error? submitResult.Error = errors.New("failed to update segment statuses in DB")
	}

	// Return the result received from the connector (potentially modified if DB updates failed critically)
	return &submitResult, nil // Return logical outcome
}

// resolveSenderID returns the sender to submit with on the given connection: the provider's own
// sender, or for providers on the route sender pool, the next sender from that connection's pool.
func (s *DefaultSender) resolveSenderID(ctx context.Context, msg database.GetMessageDetailsForSendingRow, connID int32) (string, error) {
	usesPool, err := s.dbQueries.UsesRouteSenderPool(ctx, msg.ServiceProviderID)
	if err != nil {
		return "", fmt.Errorf("failed to check route sender pool for SP %d: %w", msg.ServiceProviderID, err)
	}
	if !usesPool {
		return msg.OriginalSourceAddr, nil
	}

	picked, err := s.dbQueries.PickRouteSender(ctx, connID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("no active sender in the route sender pool of connection %d", connID)
		}
		return "", fmt.Errorf("failed to pick route sender for connection %d: %w", connID, err)
	}

	if err := s.dbQueries.SetMessageSentSourceAddr(ctx, database.SetMessageSentSourceAddrParams{
		ID:             msg.ID,
		SentSourceAddr: &picked.SenderIDString,
	}); err != nil {
		// The send can go ahead; only the record of which sender was used is missing.
		slog.ErrorContext(ctx, "Failed to record route sender on message", slog.Any("error", err))
	}
	slog.InfoContext(ctx, "Using route pool sender",
		slog.String("original_sender", msg.OriginalSourceAddr),
		slog.String("route_sender", picked.SenderIDString),
		slog.Int("route_sender_id", int(picked.ID)))
	return picked.SenderIDString, nil
}

// Helper to check for unique violation errors (adapt based on actual pgx error type)
func isUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError // Use correct type from pgx driver
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
