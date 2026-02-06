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

type Router interface {
	Route(ctx context.Context, req RoutingRequest) (*mno.RoutingResult, error)
}

type RoutingRequest struct {
	SPCredentialID int32
	MSISDN         string
	SenderID       string
	IsOTP          bool
}

type RoutingResult struct {
	MNOID          int32
	ConnectionID   *int32
	Protocol       string
	ApprovedSender string
	ShouldUse      bool
	Error          error
}

type DefaultRouter struct {
	dbQueries              database.Querier
	defaultRoutingGroupRef *string
}

func NewDefaultRouter(q database.Querier, defaultRoutingGroupRef string) *DefaultRouter {
	nsRef := defaultRoutingGroupRef
	return &DefaultRouter{
		dbQueries:              q,
		defaultRoutingGroupRef: &nsRef,
	}
}

func (r *DefaultRouter) Route(ctx context.Context, req RoutingRequest) (*mno.RoutingResult, error) {
	logCtx := logging.ContextWithMSISDN(ctx, req.MSISDN)
	logCtx = logging.ContextWithSPCredentialID(logCtx, req.SPCredentialID)

	slog.DebugContext(logCtx, "Routing attempt initiated",
		slog.Int("sp_credential_id", int(req.SPCredentialID)),
		slog.String("msisdn", req.MSISDN),
		slog.String("sender_id", req.SenderID),
		slog.Bool("is_otp", req.IsOTP),
	)

	result := &mno.RoutingResult{ShouldUse: false}

	normalizedMSISDN := strings.TrimPrefix(req.MSISDN, "+")
	if normalizedMSISDN == "" {
		result.Error = fmt.Errorf("MSISDN is empty or invalid after normalization: %s", req.MSISDN)
		slog.WarnContext(logCtx, result.Error.Error())
		return result, nil
	}

	var currentRoutingGroupID *int32

	spCredRoutingInfo, err := r.dbQueries.GetSpCredentialRoutingGroupId(logCtx, req.SPCredentialID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(logCtx, "SP Credential not found, cannot determine routing group",
				slog.Int("sp_credential_id", int(req.SPCredentialID)),
			)
			result.Error = fmt.Errorf("service provider credential %d not found", req.SPCredentialID)
			return result, nil
		}
		slog.ErrorContext(logCtx, "Failed to get routing group ID for SP credential", slog.Any("error", err))
		result.Error = fmt.Errorf("database error fetching SP credential routing info: %w", err)
		return result, err
	}
	currentRoutingGroupID = spCredRoutingInfo

	if currentRoutingGroupID == nil && r.defaultRoutingGroupRef != nil {
		slog.DebugContext(logCtx, "SP credential has no routing group, attempting system default",
			slog.Any("default_ref", r.defaultRoutingGroupRef),
		)
		defaultGroup, err := r.dbQueries.GetRoutingGroupByReference(logCtx, r.defaultRoutingGroupRef)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				slog.ErrorContext(logCtx, "Failed to get system default routing group", slog.Any("error", err))
			}
		} else {
			currentRoutingGroupID = &defaultGroup.ID
			slog.DebugContext(logCtx, "Using system default routing group", slog.Int("id", int(defaultGroup.ID)))
		}
	}

	if currentRoutingGroupID == nil {
		slog.WarnContext(logCtx, "No applicable routing group found",
			slog.Int("sp_credential_id", int(req.SPCredentialID)),
		)
		result.Error = fmt.Errorf("no routing group configured for credential %d and no system default", req.SPCredentialID)
		return result, nil
	}

	slog.DebugContext(logCtx, "Finding MSISDN prefix group", slog.String("normalized_msisdn", normalizedMSISDN))
	matchedPrefixInfo, err := r.dbQueries.FindMatchingPrefixGroupForMSISDN(logCtx, normalizedMSISDN)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(logCtx, "No MSISDN prefix entry matched", slog.String("normalized_msisdn", normalizedMSISDN))
			result.Error = fmt.Errorf("no matching MSISDN prefix group for %s", normalizedMSISDN)
			return result, nil
		}
		slog.ErrorContext(logCtx, "Database error finding matching MSISDN prefix group", slog.Any("error", err))
		result.Error = fmt.Errorf("database error finding prefix group: %w", err)
		return result, err
	}

	slog.DebugContext(logCtx, "Matched MSISDN prefix group",
		slog.Int("prefix_group_id", int(matchedPrefixInfo.MsisdnPrefixGroupID)),
		slog.String("matched_prefix", matchedPrefixInfo.MatchedPrefix),
	)

	routingParams := database.GetApplicableMnoForRoutingParams{
		RoutingGroupID:      *currentRoutingGroupID,
		MsisdnPrefixGroupID: matchedPrefixInfo.MsisdnPrefixGroupID,
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
			result.Error = fmt.Errorf("no route assignment for prefix group %d in routing group %d",
				routingParams.MsisdnPrefixGroupID, routingParams.RoutingGroupID)
			return result, nil
		}
		slog.ErrorContext(logCtx, "Database error getting applicable MNO", slog.Any("error", err))
		result.Error = fmt.Errorf("database error in final route lookup: %w", err)
		return result, err
	}

	result.ShouldUse = true
	result.MNOID = mnoDetails.MnoID
	result.ConnectionID = mnoDetails.MnoConnectionID
	if mnoDetails.MnoProtocol != nil {
		result.Protocol = *mnoDetails.MnoProtocol
	} else {
		result.Protocol = ""
	}

	slog.InfoContext(logCtx, "Route determined successfully",
		slog.Int("mno_id", int(mnoDetails.MnoID)),
		slog.String("mno_name", mnoDetails.MnoName),
		slog.Int("mno_connection_id", intPtrToInt(mnoDetails.MnoConnectionID)),
		slog.String("protocol", result.Protocol),
	)

	return result, nil
}

func stringPtrToStr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func intPtrToInt(ptr *int32) int {
	if ptr == nil {
		return 0
	}
	return int(*ptr)
}
