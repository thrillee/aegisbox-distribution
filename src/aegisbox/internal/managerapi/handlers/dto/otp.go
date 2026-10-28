package dto

import (
	"time"
)

// --- OTP Alternative Sender ---
type CreateOtpAlternativeSenderRequest struct {
	SenderIDString     string  `json:"sender_id_string" binding:"required,max=16"`
	MnoID              *int32  `json:"mno_id"` // Pointer for optional
	Status             string  `json:"status" binding:"omitempty,oneof=active inactive depleted_awaiting_reset"`
	MaxUsageCount      int32   `json:"max_usage_count" binding:"required,gt=0"`
	ResetIntervalHours *int32  `json:"reset_interval_hours"`
	Notes              *string `json:"notes"`
	ServiceProviderID  *int32  `json:"service_provider_id"` // Added based on previous update
}

type UpdateOtpAlternativeSenderRequest struct {
	SenderIDString         *string    `json:"sender_id_string,omitempty" binding:"omitempty,max=16"`
	MnoID                  *int32     `json:"mno_id"`
	Status                 *string    `json:"status,omitempty" binding:"omitempty,oneof=active inactive depleted_awaiting_reset"`
	MaxUsageCount          *int32     `json:"max_usage_count,omitempty" binding:"omitempty,gt=0"`
	CurrentUsageCount      *int32     `json:"current_usage_count,omitempty" binding:"omitempty,gte=0"` // For manual adjustment
	ResetIntervalHours     *int32     `json:"reset_interval_hours"`
	LastResetAt            *time.Time `json:"last_reset_at"` // For manual adjustment
	Notes                  *string    `json:"notes"`
	ServiceProviderID      *int32     `json:"service_provider_id"`
	ClearServiceProviderID *bool      `json:"clear_service_provider_id"` // To set service_provider_id to NULL
	ClearMnoID             *bool      `json:"clear_mno_id"`              // To set mno_id to NULL
}

type OtpAlternativeSenderResponse struct {
	ID                  int32      `json:"id"`
	SenderIDString      string     `json:"sender_id_string"`
	MnoID               *int32     `json:"mno_id,omitempty"`
	MnoName             *string    `json:"mno_name,omitempty"` // If joined in query
	ServiceProviderID   *int32     `json:"service_provider_id,omitempty"`
	ServiceProviderName *string    `json:"service_provider_name,omitempty"` // If joined
	Status              string     `json:"status"`
	CurrentUsageCount   int32      `json:"current_usage_count"`
	MaxUsageCount       int32      `json:"max_usage_count"`
	ResetIntervalHours  *int32     `json:"reset_interval_hours,omitempty"`
	LastResetAt         *time.Time `json:"last_reset_at,omitempty"`
	LastUsedAt          *time.Time `json:"last_used_at,omitempty"`
	Notes               *string    `json:"notes,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// --- OTP Message Template ---
type CreateOtpMessageTemplateRequest struct {
	Name             string  `json:"name" binding:"required,max=255"`
	ContentTemplate  string  `json:"content_template" binding:"required"`
	DefaultBrandName *string `json:"default_brand_name,omitempty" binding:"omitempty,max=100"`
	Status           string  `json:"status" binding:"omitempty,oneof=active inactive"`
}

type UpdateOtpMessageTemplateRequest struct {
	Name             *string `json:"name,omitempty" binding:"omitempty,max=255"`
	ContentTemplate  *string `json:"content_template,omitempty"`
	DefaultBrandName *string `json:"default_brand_name,omitempty" binding:"omitempty,max=100"`
	Status           *string `json:"status,omitempty" binding:"omitempty,oneof=active inactive"`
}

type OtpMessageTemplateResponse struct {
	ID               int32     `json:"id"`
	Name             string    `json:"name"`
	ContentTemplate  string    `json:"content_template"`
	DefaultBrandName *string   `json:"default_brand_name,omitempty"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// --- OTP Sender Template Assignment ---
type CreateOtpSenderTemplateAssignmentRequest struct {
	OtpAlternativeSenderID int32  `json:"otp_alternative_sender_id" binding:"required"`
	OtpMessageTemplateID   int32  `json:"otp_message_template_id" binding:"required"`
	MnoID                  *int32 `json:"mno_id"`
	Priority               int32  `json:"priority"` // Defaults to 0 in DB if not provided / can be omitted
	Status                 string `json:"status" binding:"omitempty,oneof=active inactive"`
}

type UpdateOtpSenderTemplateAssignmentRequest struct {
	OtpAlternativeSenderID *int32  `json:"otp_alternative_sender_id,omitempty"`
	OtpMessageTemplateID   *int32  `json:"otp_message_template_id,omitempty"`
	MnoID                  *int32  `json:"mno_id"`
	ClearMnoID             *bool   `json:"clear_mno_id"` // To set mno_id to NULL
	Priority               *int32  `json:"priority,omitempty"`
	Status                 *string `json:"status,omitempty" binding:"omitempty,oneof=active inactive"`
}

type OtpSenderTemplateAssignmentResponse struct {
	ID                      int32     `json:"id"`
	OtpAlternativeSenderID  int32     `json:"otp_alternative_sender_id"`
	AlternativeSenderString *string   `json:"alternative_sender_string,omitempty"` // from JOIN
	OtpMessageTemplateID    int32     `json:"otp_message_template_id"`
	TemplateName            *string   `json:"template_name,omitempty"` // from JOIN
	MnoID                   *int32    `json:"mno_id,omitempty"`
	MnoName                 *string   `json:"mno_name,omitempty"` // from JOIN
	Priority                int32     `json:"priority"`
	Status                  string    `json:"status"`
	AssignmentUsageCount    int64     `json:"assignment_usage_count"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// --- Special Listing Endpoint DTOs ---
type AssignedTemplateInfo struct {
	TemplateID         int32   `json:"template_id"`
	TemplateName       string  `json:"template_name"`
	AssignmentID       int32   `json:"assignment_id"`
	AssignmentStatus   string  `json:"assignment_status"`
	AssignmentPriority int32   `json:"assignment_priority"`
	MnoID              *int32  `json:"mno_id,omitempty"` // MNO specific assignment
	MnoName            *string `json:"mno_name,omitempty"`
}

type SenderWithTemplatesResponse struct {
	AlternativeSenderID     int32                  `json:"alternative_sender_id"`
	SenderIDString          string                 `json:"sender_id_string"`
	AlternativeSenderStatus string                 `json:"alternative_sender_status"`
	MnoID                   *int32                 `json:"mno_id,omitempty"` // MNO of the sender itself
	MnoName                 *string                `json:"mno_name,omitempty"`
	ServiceProviderID       *int32                 `json:"service_provider_id,omitempty"`
	ServiceProviderName     *string                `json:"service_provider_name,omitempty"`
	AssignedTemplates       []AssignedTemplateInfo `json:"assigned_templates"`
}
