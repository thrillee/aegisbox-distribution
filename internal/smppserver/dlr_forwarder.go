package smppserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/sp"
	"github.com/thrillee/aegisbox/pkg/errormapper"
	// Assume needed internal packages
)

// RawPDUForwarder defines the interface needed by SMPSPForwarder for this raw implementation.
type RawPDUForwarder interface {
	ForwardRawPDU(ctx context.Context, systemID string, pduBytes []byte) error
}

// SMPSPForwarder implements DLRForwarder using the raw PDU forwarding mechanism.
type SMPSPForwarder struct {
	rawForwarder RawPDUForwarder // Use the raw PDU forwarder interface
}

// NewSMPSPForwarder creates a new SMPP DLR forwarder.
func NewSMPSPForwarder(forwarder RawPDUForwarder) *SMPSPForwarder {
	if forwarder == nil {
		panic("RawPDUForwarder cannot be nil")
	}
	return &SMPSPForwarder{rawForwarder: forwarder}
}

// Helper to write CString (value + NUL byte) to a buffer
func writeCString(buf *bytes.Buffer, s string) {
	buf.WriteString(s)
	buf.WriteByte(0x00)
}

// ForwardDLR implements the DLRForwarder interface for SMPP using raw PDU construction.
func (f *SMPSPForwarder) ForwardDLR(ctx context.Context, spDetails sp.SPDetails, dlr sp.ForwardedDLRInfo) error {
	if spDetails.SMPPSystemID == "" {
		return errors.New("cannot forward DLR via SMPP: missing SystemID in SPDetails")
	}

	systemID := spDetails.SMPPSystemID
	logCtx := logging.ContextWithMessageID(ctx, dlr.InternalMessageID)
	logCtx = logging.ContextWithSystemID(logCtx, systemID)
	logCtx = logging.ContextWithSPID(logCtx, spDetails.ServiceProviderID)
	logCtx = logging.ContextWithClientMessageID(logCtx, *dlr.ClientMessageID)

	slog.InfoContext(logCtx, "Constructing raw DeliverSM PDU for DLR forwarding", slog.String("dlr_status", dlr.Status))

	// 1. Format Short Message Text (Appendix B)
	var smppErrCode string
	if dlr.ErrorCode != nil {
		smppErrCode = errormapper.MapErrorCode(*dlr.ErrorCode, "smpp")
	} else {
		smppErrCode = "000"
	}
	smppErrCode = "000"
	// idField := fmt.Sprintf("%d", dlr.ClientMessageID)
	// if dlr.ClientMessageRef != nil && *dlr.ClientMessageRef != "" {
	// 	ref := *dlr.ClientMessageRef
	// 	if len(ref) <= 10 { // Max 10 chars suggested by Appendix B for 'id'
	// 		idField = ref
	// 	} else {
	// 		slog.WarnContext(logCtx, "Client reference too long for SMPP DLR 'id' field, using internal ID instead", slog.String("client_ref", ref))
	// 	}
	// }
	submitDateStr := dlr.SubmitDate.UTC().Format("0601021504") // Use UTC? Or local? Check spec/expectations. Using UTC.
	doneDateStr := dlr.DoneDate.UTC().Format("0601021504")
	sub := "001"
	dlvrd := "001" // Simplified
	dlrStat := strings.ToUpper(dlr.Status)
	if len(dlrStat) > 7 {
		dlrStat = dlrStat[:7]
	}
	text := "000" // Keep short message text minimal for DLR
	dlrText := fmt.Sprintf("id:%s sub:%s dlvrd:%s submit date:%s done date:%s stat:%s err:%s text:%s",
		*dlr.ClientMessageID, sub, dlvrd, submitDateStr, doneDateStr, dlrStat, smppErrCode, text)

	slog.InfoContext(logCtx, "Forwarded DLR Text",
		slog.String("client_message_id", *dlr.ClientMessageID),
		slog.String("dlr_text", dlrText),
		slog.String("destAddr", dlr.DestAddr),
		slog.String("srcAddr", dlr.SourceAddr),
	)

	p := pdu.NewSubmitSM().(*pdu.SubmitSM)

	// Set Addresses
	srcAddr := pdu.NewAddress()
	srcAddr.SetTon(1) // Use configured defaults
	srcAddr.SetNpi(1)
	if err := srcAddr.SetAddress(dlr.SourceAddr); err != nil {
		return fmt.Errorf("invalid source address (SenderID) '%s': %w", dlr.SourceAddr, err)
	}
	p.SourceAddr = srcAddr

	destAddr := pdu.NewAddress()
	destAddr.SetTon(5)
	destAddr.SetNpi(0)
	if err := destAddr.SetAddress(dlr.DestAddr); err != nil {
		return fmt.Errorf("invalid destination address '%s': %w", dlr.DestAddr, err)
	}
	p.DestAddr = destAddr

	// Set Message Content and Encoding
	coding := data.GSM7BIT

	err := p.Message.SetMessageWithEncoding(dlrText, coding)
	if err != nil {
		return fmt.Errorf("failed to set message content (encoding: %d): %w", coding, err)
	}

	// Set Flags and Options
	p.ProtocolID = 0           // Usually 0
	p.RegisteredDelivery = 0   // No DLR requested
	p.ReplaceIfPresentFlag = 0 // Standard behavior

	// ESM Class: Handle UDH and Message Type (e.g., Flash)
	p.EsmClass = 0x04 // smpp.MSG_MODE_DEFAULT | smpp.MSG_TYPE_DEFAULT // Default

	dlrBuf := pdu.NewBuffer(nil)
	p.Marshal(dlrBuf)
	marshaledDlrPduBytes := dlrBuf.Bytes()

	// 5. Forward using the RawPDUForwarder interface
	err = f.rawForwarder.ForwardRawPDU(logCtx, systemID, marshaledDlrPduBytes)
	if err != nil {
		// Error already logged by ForwardRawPDU if needed
		return fmt.Errorf("failed forwarding raw DLR PDU to SP %s: %w", systemID, err)
	}

	slog.InfoContext(logCtx, "Successfully forwarded raw DLR PDU via SMPP", slog.String("dlr_status", dlr.Status))
	return nil
}
