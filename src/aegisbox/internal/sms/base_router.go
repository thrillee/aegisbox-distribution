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

type RoutingContext struct {
	SPCredentialID   int32
	MSISDN           string
	SenderID         string
	IsOTP            bool
	NormalizedMSISDN string
	RoutingGroupID   *int32
	PrefixGroupID    int32
	MatchedPrefix    string
	LogCtx           context.Context
}

type BaseRouter struct {
	dbQueries              database.Querier
	defaultRoutingGroupRef *string
}

func NewBaseRouter(q database.Querier, defaultRoutingGroupRef string) *BaseRouter {
	nsRef := defaultRoutingGroupRef
	return &BaseRouter{
		dbQueries:              q,
		defaultRoutingGroupRef: &nsRef,
	}
}

func (r *BaseRouter) InitRoutingContext(ctx context.Context, req RoutingRequest) (*RoutingContext, error) {
	logCtx := logging.ContextWithMSISDN(ctx, req.MSISDN)
	logCtx = logging.ContextWithSPCredentialID(logCtx, req.SPCredentialID)

	slog.DebugContext(logCtx, "Routing attempt initiated",
		slog.Int("sp_credential_id", int(req.SPCredentialID)),
		slog.String("msisdn", req.MSISDN),
		slog.String("sender_id", req.SenderID),
		slog.Bool("is_otp", req.IsOTP),
	)

	normalizedMSISDN := strings.TrimPrefix(req.MSISDN, "+")
	if normalizedMSISDN == "" {
		return nil, fmt.Errorf("MSISDN is empty or invalid after normalization: %s", req.MSISDN)
	}

	return &RoutingContext{
		SPCredentialID:   req.SPCredentialID,
		MSISDN:           req.MSISDN,
		SenderID:         req.SenderID,
		IsOTP:            req.IsOTP,
		NormalizedMSISDN: normalizedMSISDN,
		RoutingGroupID:   nil,
		PrefixGroupID:    0,
		MatchedPrefix:    "",
		LogCtx:           logCtx,
	}, nil
}

func (r *BaseRouter) DetermineRoutingGroup(ctx *RoutingContext) error {
	logCtx := ctx.LogCtx

	spCredRoutingInfo, err := r.dbQueries.GetSpCredentialRoutingGroupId(logCtx, ctx.SPCredentialID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(logCtx, "SP Credential not found, cannot determine routing group",
				slog.Int("sp_credential_id", int(ctx.SPCredentialID)),
			)
			return fmt.Errorf("service provider credential %d not found", ctx.SPCredentialID)
		}
		slog.ErrorContext(logCtx, "Failed to get routing group ID for SP credential", slog.Any("error", err))
		return fmt.Errorf("database error fetching SP credential routing info: %w", err)
	}
	ctx.RoutingGroupID = spCredRoutingInfo

	if ctx.RoutingGroupID == nil && r.defaultRoutingGroupRef != nil {
		slog.DebugContext(logCtx, "SP credential has no routing group, attempting system default",
			slog.Any("default_ref", r.defaultRoutingGroupRef),
		)
		defaultGroup, err := r.dbQueries.GetRoutingGroupByReference(logCtx, r.defaultRoutingGroupRef)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				slog.ErrorContext(logCtx, "Failed to get system default routing group", slog.Any("error", err))
			}
		} else {
			ctx.RoutingGroupID = &defaultGroup.ID
			slog.DebugContext(logCtx, "Using system default routing group", slog.Int("id", int(defaultGroup.ID)))
		}
	}

	if ctx.RoutingGroupID == nil {
		slog.WarnContext(logCtx, "No applicable routing group found",
			slog.Int("sp_credential_id", int(ctx.SPCredentialID)),
		)
		return fmt.Errorf("no routing group configured for credential %d and no system default", ctx.SPCredentialID)
	}

	return nil
}

func (r *BaseRouter) FindPrefixGroup(ctx *RoutingContext) error {
	logCtx := ctx.LogCtx

	slog.DebugContext(logCtx, "Finding MSISDN prefix group", slog.String("normalized_msisdn", ctx.NormalizedMSISDN))
	matchedPrefixInfo, err := r.dbQueries.FindMatchingPrefixGroupForMSISDN(logCtx, ctx.NormalizedMSISDN)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(logCtx, "No MSISDN prefix entry matched", slog.String("normalized_msisdn", ctx.NormalizedMSISDN))
			return fmt.Errorf("no matching MSISDN prefix group for %s", ctx.NormalizedMSISDN)
		}
		slog.ErrorContext(logCtx, "Database error finding matching MSISDN prefix group", slog.Any("error", err))
		return fmt.Errorf("database error finding prefix group: %w", err)
	}

	ctx.PrefixGroupID = matchedPrefixInfo.MsisdnPrefixGroupID
	ctx.MatchedPrefix = matchedPrefixInfo.MatchedPrefix

	slog.DebugContext(logCtx, "Matched MSISDN prefix group",
		slog.Int("prefix_group_id", int(ctx.PrefixGroupID)),
		slog.String("matched_prefix", ctx.MatchedPrefix),
	)

	return nil
}

func (r *BaseRouter) GetSingleRoute(ctx *RoutingContext) (*mno.RoutingResult, error) {
	logCtx := ctx.LogCtx

	routingParams := database.GetApplicableMnoForRoutingParams{
		RoutingGroupID:      *ctx.RoutingGroupID,
		MsisdnPrefixGroupID: ctx.PrefixGroupID,
	}

	slog.DebugContext(logCtx, "Getting applicable MNO from routing assignments",
		slog.Int("routing_group_id", int(routingParams.RoutingGroupID)),
		slog.Int("msisdn_prefix_group_id", int(routingParams.MsisdnPrefixGroupID)),
	)

	mnoDetails, err := r.dbQueries.GetApplicableMnoForRouting(logCtx, routingParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(logCtx, "No active routing assignment found",
				slog.Int("routing_group_id", int(routingParams.RoutingGroupID)),
				slog.Int("msisdn_prefix_group_id", int(routingParams.MsisdnPrefixGroupID)),
			)
			return nil, fmt.Errorf("no route assignment for prefix group %d in routing group %d",
				routingParams.MsisdnPrefixGroupID, routingParams.RoutingGroupID)
		}
		slog.ErrorContext(logCtx, "Database error getting applicable MNO", slog.Any("error", err))
		return nil, fmt.Errorf("database error in final route lookup: %w", err)
	}

	result := &mno.RoutingResult{
		ShouldUse:           true,
		MNOID:               mnoDetails.MnoID,
		RoutingAssignmentID: &mnoDetails.RoutingAssignmentID,
		ConnectionID:        mnoDetails.MnoConnectionID,
	}
	if mnoDetails.MnoProtocol != nil {
		result.Protocol = *mnoDetails.MnoProtocol
	}

	slog.InfoContext(logCtx, "Route determined successfully",
		slog.Int("routing_assignment_id", int(mnoDetails.RoutingAssignmentID)),
		slog.Int("mno_id", int(mnoDetails.MnoID)),
		slog.String("mno_name", mnoDetails.MnoName),
		slog.Int("mno_connection_id", int(int32PtrToInt32(mnoDetails.MnoConnectionID, 0))),
		slog.String("protocol", result.Protocol),
	)

	return result, nil
}
