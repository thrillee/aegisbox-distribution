package sms

import (
	"context"
	"fmt"
	"log"

	"github.com/thrillee/aegisbox/internal/database"
	// "github.com/thrillee/gosmpp/pdu" // If needed for parameter mapping
)

// Compile-time check
var _ Sender = (*DefaultSender)(nil)

type DefaultSender struct {
	dbQueries    database.Querier
	mnoClientMgr *smppclient.Manager // Your manager for MNO connections
}

func NewDefaultSender(q database.Querier, mgr *smppclient.Manager) *DefaultSender {
	return &DefaultSender{dbQueries: q, mnoClientMgr: mgr}
}

// Send attempts to send the message via the appropriate MNO client.
func (s *DefaultSender) Send(ctx context.Context, mnoID int32, msg database.GetMessageDetailsForSendingRow) (*SendingResult, error) {
	result := &SendingResult{Sent: false}

	// Get active client connection for this MNO
	mnoClient, err := s.mnoClientMgr.GetClient(ctx, mnoID)
	if err != nil {
		// Error getting client (e.g., config issue, manager error)
		result.Error = fmt.Errorf("failed to get MNO client for MNO ID %d: %w", mnoID, err)
		log.Printf("Sender Error: %v", result.Error)
		return result, err // Return internal error
	}
	if mnoClient == nil {
		// No active connection available currently
		result.Error = fmt.Errorf("no active SMPP client connection for MNO ID %d", mnoID)
		log.Printf("Sender Info: %v", result.Error)
		// This might be a temporary state, could retry later.
		// For now, treat as failure to send.
		return result, nil // Return logical error (no client available)
	}

	// --- Map Database Model to SMPP SubmitSM Parameters ---
	// This requires knowing the specific fields in your GetMessageDetailsForSendingRow
	// and mapping them to gosmpp PDU fields. Placeholder example:
	submitParams := smppclient.SubmitSMParams{
		SourceAddr:      msg.SenderIDUsed,
		DestinationAddr: msg.DestinationMsisdn,
		ShortMessage:    msg.ShortMessage,
		// DataCoding: // Set based on message content or config
		// ESMClass:   // Set if needed
		// RegisteredDelivery: // Request DLRs
		// Add other optional parameters (TLVs) if needed
	}
	// --- End Mapping ---

	// Submit via the MNO client's method
	mnoRespMsgID, submitErr := mnoClient.SubmitSM(ctx, submitParams)

	if submitErr != nil {
		// SMPP submit failed
		result.Error = fmt.Errorf("smpp submit failed for MNO ID %d: %w", mnoID, submitErr)
		log.Printf("Sender Error: Failed to send msg %d via MNO %d: %v", msg.ID, mnoID, submitErr)
		return result, nil // Return logical error (submit failed)
	}

	// Success!
	result.Sent = true
	result.MNOMessageID = mnoRespMsgID
	result.MNOConnectionID = mnoClient.ConnectionID // Get Conn ID from client

	log.Printf("Sender Success: Msg %d sent via MNO %d. MNO MsgID: %s", msg.ID, mnoID, mnoRespMsgID)
	return result, nil
}
