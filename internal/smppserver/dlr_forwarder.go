package smppserver

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"strings"

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

	slog.InfoContext(logCtx, "Constructing raw DeliverSM PDU for DLR forwarding", slog.String("dlr_status", dlr.Status))

	// 1. Format Short Message Text (Appendix B)
	var smppErrCode string
	if dlr.ErrorCode != nil {
		smppErrCode = errormapper.MapErrorCode(*dlr.ErrorCode, "smpp")
	} else {
		smppErrCode = "000"
	}
	idField := fmt.Sprintf("%d", dlr.InternalMessageID)
	if dlr.ClientMessageRef != nil && *dlr.ClientMessageRef != "" {
		ref := *dlr.ClientMessageRef
		if len(ref) <= 10 { // Max 10 chars suggested by Appendix B for 'id'
			idField = ref
		} else {
			slog.WarnContext(logCtx, "Client reference too long for SMPP DLR 'id' field, using internal ID instead", slog.String("client_ref", ref))
		}
	}
	submitDateStr := dlr.SubmitDate.UTC().Format("0601021504") // Use UTC? Or local? Check spec/expectations. Using UTC.
	doneDateStr := dlr.DoneDate.UTC().Format("0601021504")
	sub := "001"
	dlvrd := "001" // Simplified
	dlrStat := strings.ToUpper(dlr.Status)
	if len(dlrStat) > 7 {
		dlrStat = dlrStat[:7]
	}
	text := "" // Keep short message text minimal for DLR
	dlrText := fmt.Sprintf("id:%s sub:%s dlvrd:%s submit date:%s done date:%s stat:%s err:%s text:%s",
		idField, sub, dlvrd, submitDateStr, doneDateStr, dlrStat, smppErrCode, text)

	// Assume default GSM7 encoding for the text payload for now
	dlrTextBytes := []byte(dlrText)
	dataCoding := byte(0x00) // Default alphabet

	// 2. Construct DeliverSM PDU Body Bytes manually
	// service_type<NUL> - Empty for DLR usually
	// source_addr_ton, source_addr_npi, source_addr<NUL> - Use original Destination
	// dest_addr_ton, dest_addr_npi, destination_addr<NUL> - Use original SenderID
	// esm_class, protocol_id, priority_flag
	// schedule_delivery_time<NUL> - Empty
	// validity_period<NUL> - Empty
	// registered_delivery, replace_if_present_flag, data_coding, sm_default_msg_id
	// sm_length, short_message
	// Optional TLVs... none for now

	var bodyBuf bytes.Buffer
	writeCString(&bodyBuf, "") // service_type

	// Source Address (Original Destination)
	bodyBuf.WriteByte(1) // TON (e.g., 1=International) - TODO: Determine from original dest?
	bodyBuf.WriteByte(1) // NPI (e.g., 1=E.164)
	writeCString(&bodyBuf, dlr.SourceAddr)

	// Destination Address (Original SenderID)
	bodyBuf.WriteByte(5) // TON (e.g., 5=Alphanumeric) - TODO: Determine from original sender?
	bodyBuf.WriteByte(0) // NPI (e.g., 0=Unknown)
	writeCString(&bodyBuf, dlr.DestAddr)

	// Flags
	bodyBuf.WriteByte(0x04) // esm_class (bit 2 set for DLR)
	bodyBuf.WriteByte(0)    // protocol_id
	bodyBuf.WriteByte(0)    // priority_flag

	// Empty optional fields
	writeCString(&bodyBuf, "") // schedule_delivery_time
	writeCString(&bodyBuf, "") // validity_period

	// More flags
	bodyBuf.WriteByte(0)          // registered_delivery (0 for DLR itself)
	bodyBuf.WriteByte(0)          // replace_if_present_flag
	bodyBuf.WriteByte(dataCoding) // data_coding
	bodyBuf.WriteByte(0)          // sm_default_msg_id

	// Short Message
	smLength := byte(len(dlrTextBytes))
	bodyBuf.WriteByte(smLength)
	bodyBuf.Write(dlrTextBytes)

	// TLVs - None for now

	bodyBytes := bodyBuf.Bytes()

	// 3. Construct Header
	hdr := PDUHeader{
		CommandID:      CommandDeliverSM,
		CommandStatus:  StatusOk,
		SequenceNumber: 0, // SMSC initiated, sequence number can be 0 or managed by server
		Length:         uint32(16 + len(bodyBytes)),
	}

	// 4. Combine Header and Body into raw PDU bytes
	pduBytes := make([]byte, hdr.Length)
	binary.BigEndian.PutUint32(pduBytes[0:], hdr.Length)
	binary.BigEndian.PutUint32(pduBytes[4:], hdr.CommandID)
	binary.BigEndian.PutUint32(pduBytes[8:], hdr.CommandStatus)
	binary.BigEndian.PutUint32(pduBytes[12:], hdr.SequenceNumber)
	copy(pduBytes[16:], bodyBytes)

	// 5. Forward using the RawPDUForwarder interface
	err := f.rawForwarder.ForwardRawPDU(logCtx, systemID, pduBytes)
	if err != nil {
		// Error already logged by ForwardRawPDU if needed
		return fmt.Errorf("failed forwarding raw DLR PDU to SP %s: %w", systemID, err)
	}

	slog.InfoContext(logCtx, "Successfully forwarded raw DLR PDU via SMPP", slog.String("dlr_status", dlr.Status))
	return nil
}
