package sms

type RoutingOutcome struct {
	Status           string
	MNOID            int32
	ApprovedSenderID int32
	Error            error
	IsValid          bool
	IsLogicalFailure bool
}

type SendingOutcome struct {
	WasSent          bool
	Error            error
	IsLogicalFailure bool
}

type DLROutcome struct {
	WasProcessed     bool
	Error            error
	IsLogicalFailure bool
}
