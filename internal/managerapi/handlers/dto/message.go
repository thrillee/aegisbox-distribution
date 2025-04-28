package dto

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// MessageListResponse defines the structure for message data in list views (Basic).
type MessageListResponse struct {
	// ... (defined previously) ...
}

// MessageDetailResponse provides comprehensive message info, excluding content.
type MessageDetailResponse struct {
	ID                      int64          `json:"id"`
	ServiceProviderID       int32          `json:"service_provider_id"`
	ServiceProviderName     string         `json:"service_provider_name"` // Joined
	SpCredentialID          int32          `json:"sp_credential_id"`
	CredentialIdentifier    *string        `json:"credential_identifier"` // Joined (SystemID or API Key Identifier)
	Protocol                string         `json:"protocol"`              // Joined from credential
	ClientMessageID         *string        `json:"client_message_id"`
	ClientRef               *string        `json:"client_ref"`
	OriginalSourceAddr      string         `json:"sender_id"`
	OriginalDestinationAddr string         `json:"destination"`
	TotalSegments           int32          `json:"total_segments"`
	SubmittedAt             time.Time      `json:"submitted_at"`
	ReceivedAt              time.Time      `json:"received_at"`
	RoutedMnoID             *int32         `json:"routed_mno_id"`
	MnoName                 *string        `json:"mno_name"` // Joined
	CurrencyCode            string         `json:"currency_code"`
	Cost                    pgtype.Numeric `json:"cost"`
	ProcessingStatus        string         `json:"processing_status"`
	FinalStatus             string         `json:"final_status"`
	ErrorCode               *string        `json:"error_code"`
	ErrorDescription        *string        `json:"error_description"`
	CompletedAt             time.Time      `json:"completed_at"`
	MnoMessageIDs           *string        `json:"mno_message_ids"` // Comma-separated list from aggregation
}
