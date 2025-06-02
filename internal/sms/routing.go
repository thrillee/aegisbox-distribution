package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/mno"
)

// DefaultRouter implements the Router interface using the new flexible routing tables.
type DefaultRouter struct {
	dbQueries              database.Querier
	defaultRoutingGroupRef *string // Reference for a system-wide default routing group (e.g., "SYS_DEFAULT_ROUTE")
}

// NewDefaultRouter creates a new DefaultRouter.
// defaultRoutingGroupRef can be an empty string if no system default is configured.
func NewDefaultRouter(q database.Querier, defaultRoutingGroupRef string) *DefaultRouter {
	nsRef := defaultRoutingGroupRef
	return &DefaultRouter{
		dbQueries:              q,
		defaultRoutingGroupRef: &nsRef,
	}
}

// MnoRoutingDetails contains the resolved MNO ID and potentially specific connection details.
// This extends your mno.RoutingResult or could replace parts of it if more details are fetched.
type MnoRoutingDetails struct {
	MNOID           int32
	MnoName         *string
	MnoConnectionID *int32  // Specific connection to use for this MNO
	MnoSystemID     *string // SystemID of the specific MNO connection
}

func (r *DefaultRouter) Route(
	ctx context.Context,
	spCredentialID int32,
	msisdn string,
) (*mno.RoutingResult, error) {
	logCtx := logging.ContextWithMSISDN(ctx, msisdn)
	// You might want a logging helper to add spCredentialID to context too
	logCtx = logging.ContextWithSPCredentialID(logCtx, spCredentialID)

	slog.DebugContext(logCtx, "Routing attempt initiated",
		slog.Int("sp_credential_id", int(spCredentialID)),
		slog.String("msisdn", msisdn))

	outputResult := &mno.RoutingResult{ShouldUse: false} // Default to no route

	normalizedMSISDN := strings.TrimPrefix(msisdn, "+")
	if normalizedMSISDN == "" {
		outputResult.Error = fmt.Errorf(
			"MSISDN is empty or invalid after normalization: %s",
			msisdn,
		)
		slog.WarnContext(logCtx, outputResult.Error.Error())
		return outputResult, nil // Not an internal error, but a data error
	}

	// 1. Determine the Routing Group ID to use
	var currentRoutingGroupID *int32

	// Try to get routing_group_id from sp_credentials
	spCredRoutingInfo, err := r.dbQueries.GetSpCredentialRoutingGroupId(logCtx, spCredentialID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(
				logCtx,
				"SP Credential not found, cannot determine routing group",
				slog.Int("sp_credential_id", int(spCredentialID)),
			)
			outputResult.Error = fmt.Errorf(
				"service provider credential %d not found",
				spCredentialID,
			)
			return outputResult, nil // Or a different error type if SP cred must exist
		}
		slog.ErrorContext(
			logCtx,
			"Failed to get routing group ID for SP credential",
			slog.Any("error", err),
			slog.Int("sp_credential_id", int(spCredentialID)),
		)
		outputResult.Error = fmt.Errorf(
			"database error fetching SP credential routing info: %w",
			err,
		)
		return outputResult, err // Internal DB error
	}
	currentRoutingGroupID = spCredRoutingInfo // This is routing_group_id which can be NULL (sql.NullInt32)

	// If sp_credential has no specific routing_group_id, try to use the system default
	if currentRoutingGroupID == nil && r.defaultRoutingGroupRef == nil {
		slog.DebugContext(
			logCtx,
			"SP credential has no routing group, attempting system default",
			slog.Any("default_ref", r.defaultRoutingGroupRef),
		)
		defaultGroup, err := r.dbQueries.GetRoutingGroupByReference(
			logCtx,
			r.defaultRoutingGroupRef,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				slog.WarnContext(
					logCtx,
					"System default routing group reference not found in DB",
					slog.Any("ref", r.defaultRoutingGroupRef),
				)
				// Proceed without a routing group, or error, depending on desired behavior
			} else {
				slog.ErrorContext(logCtx, "Failed to get system default routing group by reference", slog.Any("error", err))
				// Potentially an internal error, but routing might still be possible if rules don't strictly need a group (not in this design)
			}
		} else {
			currentRoutingGroupID = &defaultGroup.ID
			slog.DebugContext(logCtx, "Using system default routing group", slog.Int("id", int(defaultGroup.ID)))
		}
	}

	if currentRoutingGroupID == nil {
		slog.WarnContext(
			logCtx,
			"No applicable routing group found (SP specific or system default)",
			slog.Int("sp_credential_id", int(spCredentialID)),
		)
		outputResult.Error = fmt.Errorf(
			"no routing group configured for credential %d and no system default applicable",
			spCredentialID,
		)
		return outputResult, nil
	}

	// 2. Find the MSISDN Prefix Group for the given MSISDN
	slog.DebugContext(
		logCtx,
		"Finding MSISDN prefix group",
		slog.String("normalized_msisdn", normalizedMSISDN),
	)
	matchedPrefixInfo, err := r.dbQueries.FindMatchingPrefixGroupForMSISDN(logCtx, normalizedMSISDN)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(
				logCtx,
				"No MSISDN prefix entry matched",
				slog.String("normalized_msisdn", normalizedMSISDN),
			)
			outputResult.Error = fmt.Errorf(
				"no matching MSISDN prefix group for %s",
				normalizedMSISDN,
			)
			return outputResult, nil
		}
		slog.ErrorContext(
			logCtx,
			"Database error finding matching MSISDN prefix group",
			slog.Any("error", err),
		)
		outputResult.Error = fmt.Errorf("database error finding prefix group: %w", err)
		return outputResult, err // Internal DB error
	}
	slog.DebugContext(logCtx, "Matched MSISDN prefix group",
		slog.Int("prefix_group_id", int(matchedPrefixInfo.MsisdnPrefixGroupID)),
		slog.String("matched_prefix", matchedPrefixInfo.MatchedPrefix))

	// 3. Get the Applicable MNO using the routing_group_id and msisdn_prefix_group_id
	routingParams := database.GetApplicableMnoForRoutingParams{
		RoutingGroupID:      *currentRoutingGroupID,
		MsisdnPrefixGroupID: matchedPrefixInfo.MsisdnPrefixGroupID,
	}
	slog.DebugContext(logCtx, "Getting applicable MNO from routing assignments",
		slog.Int("routing_group_id", int(routingParams.RoutingGroupID)),
		slog.Int("msisdn_prefix_group_id", int(routingParams.MsisdnPrefixGroupID)))

	mnoDetails, err := r.dbQueries.GetApplicableMnoForRouting(logCtx, routingParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(logCtx, "No active routing assignment found",
				slog.Int("routing_group_id", int(routingParams.RoutingGroupID)),
				slog.Int("msisdn_prefix_group_id", int(routingParams.MsisdnPrefixGroupID)))
			outputResult.Error = fmt.Errorf(
				"no route assignment for prefix group %d in routing group %d",
				routingParams.MsisdnPrefixGroupID,
				routingParams.RoutingGroupID,
			)
			return outputResult, nil
		}
		slog.ErrorContext(
			logCtx,
			"Database error getting applicable MNO from routing assignments",
			slog.Any("error", err),
		)
		outputResult.Error = fmt.Errorf("database error in final route lookup: %w", err)
		return outputResult, err // Internal DB error
	}

	// Route found!
	outputResult.ShouldUse = true
	outputResult.MNOID = mnoDetails.MnoID

	slog.InfoContext(logCtx, "Route determined successfully",
		slog.Int("mno_id", int(mnoDetails.MnoID)),
		slog.Any("mno_name", mnoDetails.MnoName),
		slog.Any("mno_system_id", mnoDetails.MnoSystemID),
		slog.Any("mno_connection_id", mnoDetails.MnoConnectionID),
	)

	return outputResult, nil
}
