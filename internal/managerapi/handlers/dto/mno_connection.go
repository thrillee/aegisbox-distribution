package dto

import (
	"encoding/json"
	"time"
)

// MNOConnectionResponse defines the structure for MNO Connection data returned by the API.
// Sensitive info like password/http auth details are omitted.
type MNOConnectionResponse struct {
	ID       int32  `json:"id"`
	MnoID    int32  `json:"mno_id"`
	Protocol string `json:"protocol"`
	Status   string `json:"status"` // The status stored in DB ('active', 'inactive')
	// RuntimeStatus string `json:"runtime_status"` // Optional: Live status from manager? Requires more complex state sharing.
	// SMPP Fields
	SystemID                *string `json:"system_id,omitempty"`
	Host                    *string `json:"host,omitempty"`
	Port                    *int32  `json:"port,omitempty"`    // Use *int32 for nullable int
	UseTLS                  *bool   `json:"use_tls,omitempty"` // Use *bool for nullable boolean
	BindType                *string `json:"bind_type,omitempty"`
	SystemType              *string `json:"system_type,omitempty"`
	EnquireLinkIntervalSecs *int32  `json:"enquire_link_interval_secs,omitempty"`
	RequestTimeoutSecs      *int32  `json:"request_timeout_secs,omitempty"`
	ConnectRetryDelaySecs   *int32  `json:"connect_retry_delay_secs,omitempty"`
	MaxWindowSize           *int32  `json:"max_window_size,omitempty"`
	DefaultDataCoding       *int32  `json:"default_data_coding,omitempty"`
	SourceAddrTON           *int32  `json:"source_addr_ton,omitempty"`
	SourceAddrNPI           *int32  `json:"source_addr_npi,omitempty"`
	DestAddrTON             *int32  `json:"dest_addr_ton,omitempty"`
	DestAddrNPI             *int32  `json:"dest_addr_npi,omitempty"`
	// HTTP Fields
	HTTPConfig json.RawMessage `json:"http_config,omitempty"` // Return relevant (non-sensitive) parts or full config? Be careful.
	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateMNOConnectionRequest defines the body for POST /mno-connections or /mnos/:mno_id/connections
type CreateMNOConnectionRequest struct {
	MnoID    int32   `json:"mno_id" binding:"required"` // Required if using direct endpoint, optional if nested
	Protocol string  `json:"protocol" binding:"required,oneof=smpp http"`
	Status   *string `json:"status" binding:"omitempty,oneof=active inactive"`
	// SMPP Fields (Required if protocol=smpp)
	SystemID                *string `json:"system_id" binding:"required_if=Protocol smpp"`
	Password                *string `json:"password" binding:"required_if=Protocol smpp"` // Sent on create only
	Host                    *string `json:"host" binding:"required_if=Protocol smpp"`
	Port                    *int32  `json:"port" binding:"required_if=Protocol smpp"`
	UseTLS                  *bool   `json:"use_tls"` // Defaults to false
	BindType                *string `json:"bind_type" binding:"required_if=Protocol smpp,omitempty,oneof=trx tx rx"`
	SystemType              *string `json:"system_type"`
	EnquireLinkIntervalSecs *int32  `json:"enquire_link_interval_secs"`
	RequestTimeoutSecs      *int32  `json:"request_timeout_secs"`
	ConnectRetryDelaySecs   *int32  `json:"connect_retry_delay_secs"`
	MaxWindowSize           *int32  `json:"max_window_size"`
	DefaultDataCoding       *int32  `json:"default_data_coding"`
	SourceAddrTON           *int32  `json:"source_addr_ton"`
	SourceAddrNPI           *int32  `json:"source_addr_npi"`
	DestAddrTON             *int32  `json:"dest_addr_ton"`
	DestAddrNPI             *int32  `json:"dest_addr_npi"`
	// HTTP Fields (Required if protocol=http)
	HTTPConfig json.RawMessage `json:"http_config" binding:"required_if=Protocol http"` // Requires valid JSON object
}

type HTTPMnoConfigRetryPolicy struct {
	MaxRetries     int64 `json:"max_retries"`
	BackoffSeconds int64 `json:"backoff_seconds"`
}

type HTTPMnoConfig struct {
	Endpoint    string                   `json:"endpoint"`
	AuthToken   string                   `json:"auth_token"`
	TimeoutSecs string                   `json:"timeout_secs"`
	RetryPolicy HTTPMnoConfigRetryPolicy `json:"retry_policy"`
}

// UpdateMNOConnectionRequest defines the body for PUT /mno-connections/:conn_id
// Cannot change MNO ID or Protocol easily.
type UpdateMNOConnectionRequest struct {
	Status *string `json:"status" binding:"omitempty,oneof=active inactive"`
	// SMPP Fields (Allow updating most fields)
	SystemID                *string `json:"system_id"`
	Password                *string `json:"password"` // Allow updating password
	Host                    *string `json:"host"`
	Port                    *int32  `json:"port"`
	UseTLS                  *bool   `json:"use_tls"`
	BindType                *string `json:"bind_type" binding:"omitempty,oneof=trx tx rx"`
	SystemType              *string `json:"system_type"`
	EnquireLinkIntervalSecs *int32  `json:"enquire_link_interval_secs"`
	RequestTimeoutSecs      *int32  `json:"request_timeout_secs"`
	ConnectRetryDelaySecs   *int32  `json:"connect_retry_delay_secs"`
	MaxWindowSize           *int32  `json:"max_window_size"`
	DefaultDataCoding       *int32  `json:"default_data_coding"`
	SourceAddrTON           *int32  `json:"source_addr_ton"`
	SourceAddrNPI           *int32  `json:"source_addr_npi"`
	DestAddrTON             *int32  `json:"dest_addr_ton"`
	DestAddrNPI             *int32  `json:"dest_addr_npi"`
	// HTTP Fields
	HTTPConfig json.RawMessage `json:"http_config"` // Allow updating HTTP config JSON
}
