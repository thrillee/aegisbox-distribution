package sp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/pkg/errormapper"
	// smpp "your-module-name/pkg/smpp/pdu" // Placeholder for constants like MES_TYPE_DEL_RCPT
)

// ServerSessionManager defines the dependency needed to get a writer for an SP connection.
// The SMPP server implementation needs to provide this.
type ServerSessionManager interface {
	GetBoundClientWriter(systemID string) (pdu.Writer, error)
}

// SMPSPForwarder implements DLRForwarder for SMPP clients.
type SMPSPForwarder struct {
	sessionMgr ServerSessionManager
}

// NewSMPSPForwarder creates a new SMPP DLR forwarder.
func NewSMPSPForwarder(mgr ServerSessionManager) *SMPSPForwarder {
	return &SMPSPForwarder{sessionMgr: mgr}
}

// ForwardDLR implements the DLRForwarder interface for SMPP.
func (f *SMPSPForwarder) ForwardDLR(ctx context.Context, spDetails SPDetails, dlr ForwardedDLRInfo) error {
	if spDetails.SMPPSystemID == "" {
		return errors.New("cannot forward DLR via SMPP: missing SystemID in SPDetails")
	}
	systemID := spDetails.SMPPSystemID
	logCtx := logging.ContextWithMessageID(ctx, dlr.InternalMessageID)
	logCtx = logging.ContextWithSystemID(logCtx, systemID)
	logCtx = logging.ContextWithSPID(logCtx, spDetails.ServiceProviderID)

	slog.InfoContext(logCtx, "Attempting to forward DLR via SMPP", slog.String("dlr_status", dlr.Status))

	// 1. Get SP Connection Writer
	writer, err := f.sessionMgr.GetBoundClientWriter(systemID)
	if err != nil {
		// SP is likely not connected or session not found
		slog.WarnContext(logCtx, "Cannot forward DLR: SP session not found or writer unavailable", slog.Any("error", err))
		// This might be temporary, returning error could trigger retry if using a queue worker later
		return fmt.Errorf("SP session %s not found or writer unavailable: %w", systemID, err)
	}

	// 2. Construct deliver_sm PDU
	deliverSM := pdu.NewDeliverSM()

	// Set Addresses
	deliverSM.SetSourceAddr(dlr.SourceAddr) // Original Destination is Source for DLR PDU
	deliverSM.SetDestAddr(dlr.DestAddr)     // Original SenderID is Destination for DLR PDU

	// Set Flags
	deliverSM.SetEsmClass(smpp.MES_TYPE_DEL_RCPT) // Set ESM class bit for Delivery Receipt
	deliverSM.SetRegisteredDelivery(0)            // Do not request DLR for a DLR

	// 3. Format Short Message according to SMPP Appendix B
	// id:IIIIIIIIII sub:SSS dlvrd:DDD submit date:YYMMDDhhmm done date:YYMMDDhhmm stat:DDDDDDD err:E network:NNN text: ...
	var smppErrCode string
	if dlr.ErrorCode != nil {
		smppErrCode = errormapper.MapErrorCode(*dlr.ErrorCode, "smpp")
	} else {
		smppErrCode = "000" // Default success
	}

	// Use Client's Reference ID if available
	idField := fmt.Sprintf("%d", dlr.InternalMessageID) // Default to our ID
	if dlr.ClientMessageRef != nil && *dlr.ClientMessageRef != "" {
		idField = *dlr.ClientMessageRef // Use client's ref if present
		// SMPP Spec Appendix B states ID should be up to 10 chars decimal representation?
		// Let's truncate client ref if needed, or use ours if client's is too long/invalid?
		if len(idField) > 10 { // Adjust max length as needed
			slog.WarnContext(logCtx, "Client reference too long for SMPP DLR 'id' field, using internal ID instead", slog.String("client_ref", idField))
			idField = fmt.Sprintf("%d", dlr.InternalMessageID)
		}
	}

	// Format dates: YYMMDDhhmm
	submitDateStr := dlr.SubmitDate.Format("0601021504") // Go time formatting constants
	doneDateStr := dlr.DoneDate.Format("0601021504")

	// Submitter and Delivered counts (often 001 unless handling multipart DLRs specifically)
	sub := "001"
	dlvrd := "001" // If status is DELIVRD, otherwise 000? Let's use 001 for simplicity now.

	// Status mapping - Ensure it matches SMPP spec values
	dlrStat := strings.ToUpper(dlr.Status)
	if len(dlrStat) > 7 { // Max 7 chars for stat field
		dlrStat = dlrStat[:7]
	}

	// Text field - often omitted or short message snippet
	text := "" // Keep short

	dlrText := fmt.Sprintf("id:%s sub:%s dlvrd:%s submit date:%s done date:%s stat:%s err:%s text:%s",
		idField,
		sub,
		dlvrd,
		submitDateStr,
		doneDateStr,
		dlrStat,
		smppErrCode,
		text,
	)

	// TODO: Handle encoding of dlrText properly (should be default GSM-7 or Latin-1 usually for DLRs)
	// Check if library handles encoding or set appropriate data_coding
	deliverSM.SetShortMessage(dlrText)
	// deliverSM.SetDataCoding(0x00) // Default encoding

	// 4. Write PDU to the specific client connection
	err = writer.Write(deliverSM)
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to write forwarded DLR PDU to SP", slog.Any("error", err))
		// This indicates a connection problem, session manager might need to handle cleanup
		return fmt.Errorf("failed writing DLR to SP %s: %w", systemID, err)
	}

	slog.InfoContext(logCtx, "Successfully forwarded DLR via SMPP", slog.String("dlr_status", dlr.Status))
	return nil
}
