package smppserver

import "fmt"

// commandIDToString converts command ID to string for logging
func commandIDToString(cmdID uint32) string {
	switch cmdID {
	case CommandBindTransceiver:
		return "BindTransceiver"
	case CommandBindReceiver:
		return "BindReceiver"
	case CommandBindTransmitter:
		return "BindTransmitter"
	case CommandSubmitSM:
		return "SubmitSM"
	case CommandDeliverSM:
		return "DeliverSM"
	case CommandUnbind:
		return "Unbind"
	case CommandEnquireLink:
		return "EnquireLink"
	case CommandGenericNack:
		return "GenericNack"
	case CommandBindResp:
		return "BindResp"
	case CommandSubmitSMResp:
		return "SubmitSMResp"
	case CommandDeliverSMResp:
		return "DeliverSMResp"
	case CommandUnbindResp:
		return "UnbindResp"
	case CommandEnquireLinkResp:
		return "EnquireLinkResp"
	default:
		return fmt.Sprintf("Unknown(0x%X)", cmdID)
	}
}
