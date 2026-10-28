package sms

import "time"

// Fixed schedules of the routing health workers. The manager API reports these as
// read-only values next to the editable routing configuration.
const (
	DLRTimeoutCheckInterval        = 1 * time.Minute
	DefaultDLRTimeoutSeconds       = 300 // used when the global routing config has no DLR timeout
	SenderAutoUnblockCheckInterval = 6 * time.Hour
	SenderAutoUnblockAfterInactive = 7 * 24 * time.Hour // see GetBlockedSendersForAutoUnblock
	RingBufferCleanupInterval      = 24 * time.Hour
	RingBufferRetentionDays        = 90
)
