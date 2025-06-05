package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5" // For pgx.ErrNoRows
	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/sp" // For MessagePreprocessor interface & IncomingSPMessage
)

var (
	// ErrNoRouteFound can be returned by the preprocessor when routing fails logically.
	ErrNoRouteFound = errors.New("no route found for destination")
	// ErrRoutingConfiguration is for issues like missing default groups.
	ErrRoutingConfiguration = errors.New("routing configuration error")
)

// Compile-time check
var _ sp.MessagePreprocessor = (*RoutingPreprocessor)(nil)

// RoutingPreprocessor determines the MNO for an incoming message.
type RoutingPreprocessor struct {
	dbQueries              database.Querier
	defaultRoutingGroupRef string // Reference for a system-wide default routing group (e.g., "SYS_DEFAULT_ROUTE_GROUP")
}

// NewRoutingPreprocessor creates a new RoutingPreprocessor.
// defaultRoutingGroupRef can be an empty string if no system default is configured.
func NewRoutingPreprocessor(
	q database.Querier,
	defaultRoutingGroupRef string,
) *RoutingPreprocessor {
	return &RoutingPreprocessor{
		dbQueries:              q,
		defaultRoutingGroupRef: defaultRoutingGroupRef,
	}
}

func (p *RoutingPreprocessor) Name() string {
	return "RoutingPreprocessor"
}

// Process implements the sp.MessagePreprocessor interface.
// It determines the MNO ID for the message and populates msg.MnoID.
// Returns true if MnoID was set, and an error if routing fails critically or logically.
func (p *RoutingPreprocessor) Process(
	ctx context.Context,
	msg *sp.IncomingSPMessage,
	spInfo database.ServiceProvider, // May not be directly used here but part of interface
	credInfo database.SpCredential, // Used to get routing_group_id
) (modified bool, err error) {
	logCtx := logging.ContextWithMSISDN(ctx, msg.DestinationMSISDN)
	logCtx = logging.ContextWithSPCredentialID(logCtx, msg.CredentialID)
	logCtx = logging.ContextWithService(logCtx, p.Name()) // Add preprocessor name

	slog.DebugContext(logCtx, "Routing preprocessing attempt initiated")

	normalizedMSISDN := strings.TrimPrefix(msg.DestinationMSISDN, "+")
	if normalizedMSISDN == "" {
		slog.WarnContext(
			logCtx,
			"MSISDN is empty or invalid after normalization",
			slog.String("original_msisdn", msg.DestinationMSISDN),
		)
		return false, fmt.Errorf(
			"MSISDN is empty or invalid: %s",
			msg.DestinationMSISDN,
		) // Logical error
	}

	// 1. Determine the Routing Group ID to use from sp_credentials or system default
	var currentRoutingGroupID_pg *int32

	// Use credInfo.RoutingGroupID directly if your GetSpCredentialByID populates it.
	// Assuming credInfo.RoutingGroupID is pgtype.Int4
	if credInfo.RoutingGroupID != nil {
		currentRoutingGroupID_pg = credInfo.RoutingGroupID
		slog.DebugContext(
			logCtx,
			"Using routing group from SP credential",
			slog.Int("id", int(*currentRoutingGroupID_pg)),
		)
	} else if p.defaultRoutingGroupRef != "" {
		slog.DebugContext(logCtx, "SP credential has no routing group, attempting system default", slog.String("default_ref", p.defaultRoutingGroupRef))
		defaultGroup, dbErr := p.dbQueries.GetRoutingGroupByReference(logCtx, &p.defaultRoutingGroupRef) // Assuming GetRoutingGroupByReference takes *string for nullable ref
		if dbErr != nil {
			if errors.Is(dbErr, pgx.ErrNoRows) {
				slog.WarnContext(logCtx, "System default routing group reference not found in DB", slog.String("ref", p.defaultRoutingGroupRef))
				// This could be a configuration error. Decide whether to error out or proceed without a group.
				// For routing, a group is usually essential with the new design.
				return false, fmt.Errorf("%w: system default routing group '%s' not found", ErrRoutingConfiguration, p.defaultRoutingGroupRef)
			}
			slog.ErrorContext(logCtx, "Failed to get system default routing group by reference", slog.Any("error", dbErr))
			return false, fmt.Errorf("database error fetching system default routing group: %w", dbErr) // Internal DB error
		}
		currentRoutingGroupID_pg = &defaultGroup.ID
		slog.DebugContext(logCtx, "Using system default routing group", slog.Int("id", int(defaultGroup.ID)))
	}

	if currentRoutingGroupID_pg == nil {
		slog.WarnContext(
			logCtx,
			"No applicable routing group found (SP specific or system default)",
		)
		return false, fmt.Errorf(
			"%w: no routing group for credential %d and no system default applicable",
			ErrRoutingConfiguration,
			msg.CredentialID,
		)
	}

	// 2. Find the MSISDN Prefix Group for the given MSISDN
	slog.DebugContext(
		logCtx,
		"Finding MSISDN prefix group",
		slog.String("normalized_msisdn", normalizedMSISDN),
	)
	matchedPrefixInfo, dbErr := p.dbQueries.FindMatchingPrefixGroupForMSISDN(
		logCtx,
		normalizedMSISDN,
	)
	if dbErr != nil {
		if errors.Is(dbErr, pgx.ErrNoRows) {
			slog.InfoContext(
				logCtx,
				"No MSISDN prefix entry matched",
				slog.String("normalized_msisdn", normalizedMSISDN),
			)
			return false, fmt.Errorf(
				"%w: no matching MSISDN prefix for %s",
				ErrNoRouteFound,
				normalizedMSISDN,
			)
		}
		slog.ErrorContext(
			logCtx,
			"Database error finding matching MSISDN prefix group",
			slog.Any("error", dbErr),
		)
		return false, fmt.Errorf(
			"database error finding prefix group: %w",
			dbErr,
		) // Internal DB error
	}
	slog.DebugContext(logCtx, "Matched MSISDN prefix group",
		slog.Int("prefix_group_id", int(matchedPrefixInfo.MsisdnPrefixGroupID)),
		slog.String("matched_prefix", matchedPrefixInfo.MatchedPrefix))

	// 3. Get the Applicable MNO using the routing_group_id and msisdn_prefix_group_id
	routingParams := database.GetApplicableMnoForRoutingParams{
		RoutingGroupID:      *currentRoutingGroupID_pg, // Use .Int32 as it's Valid now
		MsisdnPrefixGroupID: matchedPrefixInfo.MsisdnPrefixGroupID,
	}
	slog.DebugContext(logCtx, "Getting applicable MNO from routing assignments",
		slog.Int("routing_group_id", int(routingParams.RoutingGroupID)),
		slog.Int("msisdn_prefix_group_id", int(routingParams.MsisdnPrefixGroupID)))

	mnoDetails, dbErr := p.dbQueries.GetApplicableMnoForRouting(logCtx, routingParams)
	if dbErr != nil {
		if errors.Is(dbErr, pgx.ErrNoRows) {
			slog.InfoContext(logCtx, "No active routing assignment found for prefix in group",
				slog.Int("routing_group_id", int(routingParams.RoutingGroupID)),
				slog.Int("msisdn_prefix_group_id", int(routingParams.MsisdnPrefixGroupID)))
			return false, fmt.Errorf(
				"%w: no route assignment for prefix group %d in routing group %d",
				ErrNoRouteFound,
				routingParams.MsisdnPrefixGroupID,
				routingParams.RoutingGroupID,
			)
		}
		slog.ErrorContext(
			logCtx,
			"Database error getting applicable MNO from routing assignments",
			slog.Any("error", dbErr),
		)
		return false, fmt.Errorf(
			"database error in final route lookup: %w",
			dbErr,
		) // Internal DB error
	}

	// Route found! Populate msg.MnoID
	msg.MnoID = &mnoDetails.MnoID // mnoDetails.MnoID is int32
	msg.ProcessingStatus = "routed"
	modified = true

	slog.InfoContext(
		logCtx,
		"Routing preprocessor determined MNO successfully",
		slog.Int("mno_id", int(*msg.MnoID)),
		slog.Any(
			"mno_name",
			mnoDetails.MnoName,
		), // Assuming GetApplicableMnoForRoutingRow has MnoName
		slog.Any("mno_system_id", mnoDetails.MnoSystemID), // And MnoSystemID etc.
		slog.Any("mno_connection_id", mnoDetails.MnoConnectionID),
	)

	return modified, nil
}
