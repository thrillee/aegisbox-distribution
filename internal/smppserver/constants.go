package smppserver

// SMPP Command IDs (subset needed)
const (
	CommandBindTransceiver uint32 = 0x00000009
	CommandBindReceiver    uint32 = 0x00000001
	CommandBindTransmitter uint32 = 0x00000002
	CommandSubmitSM        uint32 = 0x00000004
	CommandDeliverSM       uint32 = 0x00000005
	CommandUnbind          uint32 = 0x00000006
	CommandEnquireLink     uint32 = 0x00000015
	CommandGenericNack     uint32 = 0x80000000
	CommandBindResp        uint32 = 0x80000009 // Generic response for all bind types
	CommandSubmitSMResp    uint32 = 0x80000004
	CommandDeliverSMResp   uint32 = 0x80000005
	CommandUnbindResp      uint32 = 0x80000006
	CommandEnquireLinkResp uint32 = 0x80000015
)

// SMPP Command Status Codes (subset needed)
const (
	StatusOk          uint32 = 0x00000000 // ESME_ROK
	StatusInvMsgLen   uint32 = 0x00000001 // Invalid Message Length
	StatusInvCmdID    uint32 = 0x00000003 // Invalid Command ID
	StatusInvBndSts   uint32 = 0x00000004 // Invalid Bind Status (e.g., submit before bind)
	StatusBindFailed  uint32 = 0x0000000D // Bind Failed
	StatusInvPasswd   uint32 = 0x0000000E // Invalid Password
	StatusInvSysID    uint32 = 0x0000000F // Invalid System ID
	StatusSystemError uint32 = 0x00000008 // System Error
	StatusInvSrcAddr  uint32 = 0x0000000A // Invalid Source Address
	StatusInvDstAddr  uint32 = 0x0000000B // Invalid Destination Address
	StatusThrottled   uint32 = 0x00000058 // Throttling error (used for funds/queue issues)
	StatusInvDstTon   uint32 = 0x00000053 // Invalid Dest Addr TON
	StatusInvDstNpi   uint32 = 0x00000054 // Invalid Dest Addr NPI
	StatusInvSrcTon   uint32 = 0x00000063 // Invalid Source Addr TON
	StatusInvSrcNpi   uint32 = 0x00000064 // Invalid Source Addr NPI
)
