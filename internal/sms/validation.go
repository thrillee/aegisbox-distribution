package sms

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/thrillee/aegisbox/internal/database"
)

// Compile-time check
var _ Validator = (*DefaultValidator)(nil)

type DefaultValidator struct {
	dbQueries database.Querier
}

func NewDefaultValidator(q database.Querier) *DefaultValidator {
	return &DefaultValidator{dbQueries: q}
}

// Validate performs SenderID validation (add Template validation if needed).
func (v *DefaultValidator) Validate(ctx context.Context, spID int32, senderID string) (*ValidationResult, error) {
	result := &ValidationResult{IsValid: false} // Default to invalid

	// --- SenderID Validation ---
	sidRecordID, err := v.dbQueries.ValidateSenderID(ctx, database.ValidateSenderIDParams{
		ServiceProviderID: spID,
		SenderIDString:    senderID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			result.Error = fmt.Errorf("sender ID '%s' not approved for SP %d", senderID, spID)
			log.Printf("Validation Failed: %v", result.Error)
			return result, nil // Logical failure, not internal error
		}
		// DB or other internal error
		result.Error = fmt.Errorf("db error validating sender ID: %w", err)
		return result, err // Return internal error
	}

	// SenderID is valid
	result.IsValid = true
	result.ApprovedSenderID = sidRecordID

	// --- Template Validation (Optional Example) ---
	// if templateRequired {
	//     content, err := v.dbQueries.GetTemplateContent(...)
	//     if err != nil { ... handle error ... result.IsValid = false; result.Error = ...; return result, err }
	//     // Compare content, etc.
	//     result.TemplateID = ...
	// }

	log.Printf("Validation Passed: SP=%d, SenderID=%s (RecordID: %d)", spID, senderID, sidRecordID)
	return result, nil
}
