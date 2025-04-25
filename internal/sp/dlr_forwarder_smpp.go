package sp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/pkg/errormapper"
)

// ServerSessionManager defines the dependency needed to get a writer for an SP connection.
// The SMPP server implementation needs to provide this.
type ServerSessionManager interface {
	GetBoundClientConnection(systemID string) (*gosmpp.Connection, error)
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
	conn, err := f.sessionMgr.GetBoundClientConnection(systemID) // Use new method
	if err != nil {
		slog.WarnContext(logCtx, "Cannot forward DLR: SP session not found or connection unavailable", slog.Any("error", err))
		return fmt.Errorf("SP session %s not found or connection unavailable: %w", systemID, err)
	}

	// 2. Construct deliver_sm PDU
	deliverSM := pdu.NewDeliverSM() // Assume this returns *pdu.DeliverSM

	// --- Set Addresses ---
	// Create Address objects first
	sourceAddress := pdu.NewAddress()
	sourceAddress.SetTon(1) // National? International? Use appropriate default/logic
	sourceAddress.SetNpi(1) // E.164
	if err := sourceAddress.SetAddress(dlr.SourceAddr); err != nil {
		slog.ErrorContext(logCtx, "Failed to set DLR source address", slog.String("address", dlr.SourceAddr), slog.Any("error", err))
		return fmt.Errorf("invalid DLR source address '%s': %w", dlr.SourceAddr, err)
	}

	deliverSM.SourceAddr = sourceAddress

	destAddress := pdu.NewAddress()
	// TODO: Determine TON/NPI based on dlr.DestAddr content or SP config?
	destAddress.SetTon(5) // Alphanumeric? Use appropriate default/logic
	destAddress.SetNpi(0) // Unknown
	if err := destAddress.SetAddress(dlr.DestAddr); err != nil {
		slog.ErrorContext(logCtx, "Failed to set DLR destination address", slog.String("address", dlr.DestAddr), slog.Any("error", err))
		return fmt.Errorf("invalid DLR destination address '%s': %w", dlr.DestAddr, err)
	}
	deliverSM.DestAddr = destAddress // Assign Address struct

	// Set Flags
	deliverSM.EsmClass = 0x04        // Set ESM class bit 2 for Delivery Receipt
	deliverSM.RegisteredDelivery = 0 // Do not request DLR for a DLR

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
	err = deliverSM.Message.SetMessageWithEncoding(dlrText, data.GSM7BIT) // Assuming GSM7 is default/suitable
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to set DLR short message content", slog.Any("error", err))
		return fmt.Errorf("failed to set DLR text: %w", err)
	}
	deliverSM.DataCoding = data.GSM7BIT
	// deliverSM.SetDataCoding(0x00) // Default encoding

	// 4. Write PDU to the specific client connection
	err = conn.Write(deliverSM) // Use the Write method on gosmpp.Connection
	if err != nil {
		slog.ErrorContext(logCtx, "Failed to write forwarded DLR PDU to SP connection", slog.Any("error", err))
		// This indicates a connection problem, session manager mi
		return fmt.Errorf("failed writing DLR to SP %s connection: %w", systemID, err)
	}

	slog.InfoContext(logCtx, "Successfully forwarded DLR via SMPP", slog.String("dlr_status", dlr.Status))
	return nil
}
