package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/thrillee/aegisbox/internal/config"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/sp"
)

// Compile-time check
var _ sp.MessagePreprocessor = (*OtpScopePreprocessor)(nil)

const OtpScope = "OTP"

type OtpScopePreprocessor struct {
	dbQueries database.Querier
	router    Router // Use your existing sms.Router interface
	config    config.OtpScopePreprocessorConfig
}

// NewOtpScopePreprocessor creates a new OtpScopePreprocessor.
func NewOtpScopePreprocessor(
	q database.Querier,
	cfg config.OtpScopePreprocessorConfig,
) *OtpScopePreprocessor {
	if cfg.OtpExtractionRegex == "" {
		cfg.OtpExtractionRegex = `\b(\d{4,8})\b` // Default Regex
	}
	return &OtpScopePreprocessor{
		dbQueries: q,
		config:    cfg,
	}
}

func (p *OtpScopePreprocessor) Name() string {
	return "OtpScopePreprocessor"
}

func (p *OtpScopePreprocessor) Process(
	ctx context.Context,
	msg *sp.IncomingSPMessage,
	spInfo database.ServiceProvider,
	credInfo database.SpCredential,
) (bool, error) {
	logCtx := logging.ContextWithService(ctx, p.Name())

	if credInfo.Scope != OtpScope || credInfo.Scope == "" {
		return false, nil // Not an OTP scope
	}

	slog.InfoContext(logCtx, "OTP scope detected, attempting preprocessing",
		slog.String("original_sender", msg.SenderID),
		slog.Int("sp_cred_id", int(msg.CredentialID)),
		slog.Int("service_provider_id", int(spInfo.ID))) // Log the SP ID

	// 1. Extract OTP (remains the same)
	otpRegex, err := regexp.Compile(p.config.OtpExtractionRegex)
	if err != nil {
		slog.ErrorContext(logCtx, "Invalid OTP extraction regex", slog.Any("error", err))
		return false, fmt.Errorf("internal config error: invalid otp regex: %w", err)
	}
	matches := otpRegex.FindStringSubmatch(msg.MessageContent)
	var extractedOtp string
	if len(matches) > 1 {
		extractedOtp = matches[1]
	} else {
		slog.WarnContext(logCtx, "Could not extract OTP from message content", slog.String("content", msg.MessageContent))
		return false, nil // Or an error if OTP must be present.
	}

	// 3. Select an Alternative Sender ID - Updated Logic
	var selectedSender database.OtpAlternativeSender
	var foundSender bool
	fetchLimit := int32(5) // How many candidates to consider

	// Try SP-specific senders first
	slog.DebugContext(
		logCtx,
		"Attempting to find SP-specific OTP alternative sender",
		slog.Int("sp_id", int(spInfo.ID)),
	)
	spSenders, errSp := p.dbQueries.ListSpSpecificOtpAlternativeSenders(
		ctx,
		database.ListSpSpecificOtpAlternativeSendersParams{
			ServiceProviderID: &spInfo.ID, // Pass the actual SP ID
			MnoID:             msg.MnoID,  // pgtype.Int4, can be invalid for MNO-agnostic
			Limit:             fetchLimit,
		},
	)

	if errSp != nil && !errors.Is(errSp, pgx.ErrNoRows) {
		slog.ErrorContext(
			logCtx,
			"Error listing SP-specific OTP alternative senders",
			slog.Any("error", errSp),
		)
		// Depending on policy, you might return an error or try to proceed with global
	}

	if len(spSenders) > 0 {
		selectedSender = spSenders[0] // Pick the first one (e.g., least recently used based on query ORDER BY)
		foundSender = true
		slog.InfoContext(logCtx, "Selected SP-specific OTP alternative sender",
			slog.String("sender_id_string", selectedSender.SenderIDString),
			slog.Int("db_sender_id", int(selectedSender.ID)))
	} else {
		// Fallback to global senders (service_provider_id IS NULL)
		slog.DebugContext(logCtx, "No SP-specific sender found, attempting global OTP alternative sender")
		globalSenders, errGlobal := p.dbQueries.ListGlobalOtpAlternativeSenders(ctx, database.ListGlobalOtpAlternativeSendersParams{
			MnoID: msg.MnoID, // pgtype.Int4, can be invalid for MNO-agnostic
			Limit: fetchLimit,
		})
		if errGlobal != nil && !errors.Is(errGlobal, pgx.ErrNoRows) {
			slog.ErrorContext(logCtx, "Error listing global OTP alternative senders", slog.Any("error", errGlobal))
			// Depending on policy, you might return an error
		}
		if len(globalSenders) > 0 {
			selectedSender = globalSenders[0]
			foundSender = true
			slog.InfoContext(logCtx, "Selected global OTP alternative sender",
				slog.String("sender_id_string", selectedSender.SenderIDString),
				slog.Int("db_sender_id", int(selectedSender.ID)))
		}
	}

	if !foundSender {
		slog.WarnContext(
			logCtx,
			"No available OTP alternative sender found (neither SP-specific nor global)",
			slog.Int("sp_id", int(spInfo.ID)),
			slog.Any("target_mno_id", msg.MnoID),
		) // Log target_mno_id if mnoID.Valid
		return false, nil // No sender found, do not modify original message.
	}

	// 4. Select a Template for the chosen Sender (remains the same, uses selectedSender.ID)
	assignment, err := p.dbQueries.GetActiveOtpTemplateAssignment(
		ctx,
		database.GetActiveOtpTemplateAssignmentParams{
			OtpAlternativeSenderID: selectedSender.ID,
			MnoID:                  msg.MnoID, // Use the same MNO context for template assignment
		},
	)
	if err != nil {
		slog.WarnContext(logCtx, "No active template assignment found for selected OTP sender",
			slog.Int("selected_alt_sender_db_id", int(selectedSender.ID)), slog.Any("error", err))
		return false, nil // No template, do not modify.
	}

	// 5. Construct the new message (remains the same)
	newContent := assignment.ContentTemplate
	newContent = strings.ReplaceAll(newContent, "[OTP]", extractedOtp)

	brandName := spInfo.Name // Default to SP name from service_providers table
	if assignment.DefaultBrandName != nil && *assignment.DefaultBrandName != "" {
		brandName = *assignment.DefaultBrandName
	}
	newContent = strings.ReplaceAll(newContent, "[BRAND_NAME]", brandName)
	newContent = strings.ReplaceAll(newContent, "[APP_NAME]", p.config.DefaultAppName)
	newContent = strings.ReplaceAll(newContent, "[VALIDITY_MINUTES]", p.config.DefaultValidityMin)

	// 6. Update message struct (remains the same)
	originalSenderID := msg.SenderID
	originalContent := msg.MessageContent
	msg.SenderID = selectedSender.SenderIDString
	msg.MessageContent = newContent

	slog.InfoContext(logCtx, "OTP message preprocessed successfully",
		slog.String("original_sender", originalSenderID),
		slog.String("new_sender", msg.SenderID),
		slog.String("original_content_preview", firstNChars(originalContent, 30)),
		slog.String("new_content_preview", firstNChars(msg.MessageContent, 30)),
		slog.Int("alt_sender_id_db", int(selectedSender.ID)),
		slog.Int("template_id_db", int(assignment.OtpMessageTemplateID)))

	// 7. Atomically Increment Usage Counts (remains the same)
	_, errIncSender := p.dbQueries.IncrementOtpAlternativeSenderUsage(ctx, selectedSender.ID)
	if errIncSender != nil {
		slog.ErrorContext(
			logCtx,
			"Failed to increment alternative sender usage count",
			slog.Any("error", errIncSender),
			slog.Int("sender_db_id", int(selectedSender.ID)),
		)
	}
	_, errIncAssign := p.dbQueries.IncrementOtpAssignmentUsage(ctx, assignment.ID)
	if errIncAssign != nil {
		slog.ErrorContext(
			logCtx,
			"Failed to increment template assignment usage count",
			slog.Any("error", errIncAssign),
			slog.Int("assignment_db_id", int(assignment.ID)),
		)
	}

	return true, nil // Message was modified
}

func firstNChars(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
