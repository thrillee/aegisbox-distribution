package sms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/internal/mno"
)

type RouteCandidate struct {
	RoutingAssignmentID int32
	MNOID               int32
	MNOName             string
	ConnectionID        *int32
	Protocol            string
	Priority            int32
	Weight              int32
	HealthScore         float64
	IsAvailable         bool
	ConsecutiveFailures int32
	MaxRetryAttempts    int32
}

type IntelligentRoutingResult struct {
	*mno.RoutingResult
	Candidates     []RouteCandidate
	RoutingGroupID int32
	PrefixGroupID  int32
	RoutingConfig  *database.RoutingConfig
}

type IntelligentRouter struct {
	*BaseRouter
	senderHealthTracker *SenderHealthTracker
}

func NewIntelligentRouter(
	q database.Querier,
	defaultRoutingGroupRef string,
	senderHealthTracker *SenderHealthTracker,
) *IntelligentRouter {
	return &IntelligentRouter{
		BaseRouter:          NewBaseRouter(q, defaultRoutingGroupRef),
		senderHealthTracker: senderHealthTracker,
	}
}

func (r *IntelligentRouter) Route(ctx context.Context, req RoutingRequest) (*mno.RoutingResult, error) {
	logCtx := logging.ContextWithMSISDN(ctx, req.MSISDN)
	logCtx = logging.ContextWithSPCredentialID(logCtx, req.SPCredentialID)

	slog.DebugContext(logCtx, "Intelligent routing attempt initiated",
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
			slog.WarnContext(logCtx, "SP Credential not found",
				slog.Int("sp_credential_id", int(req.SPCredentialID)),
			)
			result.Error = fmt.Errorf("service provider credential %d not found", req.SPCredentialID)
			return result, nil
		}
		slog.ErrorContext(logCtx, "Failed to get routing group ID", slog.Any("error", err))
		result.Error = fmt.Errorf("database error fetching SP credential routing info: %w", err)
		return result, err
	}
	currentRoutingGroupID = spCredRoutingInfo

	if currentRoutingGroupID == nil && r.defaultRoutingGroupRef != nil {
		slog.DebugContext(logCtx, "Using system default routing group",
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

	config, err := r.dbQueries.GetRoutingConfigForGroup(logCtx, currentRoutingGroupID)
	if err != nil {
		slog.WarnContext(logCtx, "Failed to get routing config, using defaults", slog.Any("error", err))
		config, err = r.dbQueries.GetGlobalRoutingConfig(logCtx)
		if err != nil {
			slog.ErrorContext(logCtx, "Failed to get global routing config", slog.Any("error", err))
			result.Error = fmt.Errorf("routing configuration error: %w", err)
			return result, err
		}
	}

	routingParams := database.GetAllApplicableRoutesForRoutingParams{
		RoutingGroupID:      *currentRoutingGroupID,
		MsisdnPrefixGroupID: matchedPrefixInfo.MsisdnPrefixGroupID,
	}

	slog.DebugContext(logCtx, "Fetching all applicable routes",
		slog.Int("routing_group_id", int(routingParams.RoutingGroupID)),
		slog.Int("msisdn_prefix_group_id", int(routingParams.MsisdnPrefixGroupID)),
	)

	candidates, err := r.dbQueries.GetAllApplicableRoutesForRouting(logCtx, routingParams)
	if err != nil {
		slog.ErrorContext(logCtx, "Database error getting applicable routes", slog.Any("error", err))
		result.Error = fmt.Errorf("database error in route lookup: %w", err)
		return result, err
	}

	if len(candidates) == 0 {
		slog.InfoContext(logCtx, "No active routing assignments found",
			slog.Int("routing_group_id", int(routingParams.RoutingGroupID)),
			slog.Int("msisdn_prefix_group_id", int(routingParams.MsisdnPrefixGroupID)),
		)
		result.Error = fmt.Errorf("no route assignment for prefix group %d in routing group %d",
			routingParams.MsisdnPrefixGroupID, routingParams.RoutingGroupID)
		return result, nil
	}

	slog.InfoContext(logCtx, "Found candidate routes", slog.Int("count", len(candidates)))

	viableCandidates := r.filterViableCandidates(logCtx, candidates, req.SenderID, config)

	if len(viableCandidates) == 0 {
		slog.WarnContext(logCtx, "No healthy routes available after filtering",
			slog.Int("total_candidates", len(candidates)),
		)
		result.Error = fmt.Errorf("no healthy routes available for destination")
		return result, nil
	}

	slog.InfoContext(logCtx, "Viable routes after health filtering", slog.Int("count", len(viableCandidates)))

	strategy := stringPtrToStr(config.SelectionStrategy)
	if strategy == "" {
		strategy = "priority_weighted"
	}
	selectedRoute := r.selectBestRoute(logCtx, viableCandidates, strategy)

	result.ShouldUse = true
	result.MNOID = selectedRoute.MNOID
	result.RoutingAssignmentID = &selectedRoute.RoutingAssignmentID
	result.ConnectionID = selectedRoute.ConnectionID
	result.Protocol = selectedRoute.Protocol

	slog.InfoContext(logCtx, "Route selected successfully",
		slog.Int("mno_id", int(selectedRoute.MNOID)),
		slog.String("mno_name", selectedRoute.MNOName),
		slog.Int("routing_assignment_id", int(selectedRoute.RoutingAssignmentID)),
		slog.Int("priority", int(selectedRoute.Priority)),
		slog.Float64("health_score", selectedRoute.HealthScore),
		slog.String("strategy", strategy),
		slog.Int("alternate_routes", len(viableCandidates)-1),
	)

	return result, nil
}

func (r *IntelligentRouter) RouteWithCandidates(ctx context.Context, req RoutingRequest) (*IntelligentRoutingResult, error) {
	logCtx := logging.ContextWithMSISDN(ctx, req.MSISDN)
	logCtx = logging.ContextWithSPCredentialID(logCtx, req.SPCredentialID)

	slog.DebugContext(logCtx, "Intelligent routing with candidates initiated",
		slog.Int("sp_credential_id", int(req.SPCredentialID)),
		slog.String("msisdn", req.MSISDN),
		slog.String("sender_id", req.SenderID),
		slog.Bool("is_otp", req.IsOTP),
	)

	result := &IntelligentRoutingResult{
		RoutingResult: &mno.RoutingResult{ShouldUse: false},
		Candidates:    []RouteCandidate{},
	}

	normalizedMSISDN := strings.TrimPrefix(req.MSISDN, "+")
	if normalizedMSISDN == "" {
		result.RoutingResult.Error = fmt.Errorf("MSISDN is empty or invalid after normalization: %s", req.MSISDN)
		slog.WarnContext(logCtx, result.RoutingResult.Error.Error())
		return result, nil
	}

	var currentRoutingGroupID *int32

	spCredRoutingInfo, err := r.dbQueries.GetSpCredentialRoutingGroupId(logCtx, req.SPCredentialID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(logCtx, "SP Credential not found",
				slog.Int("sp_credential_id", int(req.SPCredentialID)),
			)
			result.RoutingResult.Error = fmt.Errorf("service provider credential %d not found", req.SPCredentialID)
			return result, nil
		}
		slog.ErrorContext(logCtx, "Failed to get routing group ID", slog.Any("error", err))
		result.RoutingResult.Error = fmt.Errorf("database error fetching SP credential routing info: %w", err)
		return result, err
	}
	currentRoutingGroupID = spCredRoutingInfo

	if currentRoutingGroupID == nil && r.defaultRoutingGroupRef != nil {
		slog.DebugContext(logCtx, "Using system default routing group",
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
		result.RoutingResult.Error = fmt.Errorf("no routing group configured for credential %d and no system default", req.SPCredentialID)
		return result, nil
	}

	result.RoutingGroupID = *currentRoutingGroupID

	slog.DebugContext(logCtx, "Finding MSISDN prefix group", slog.String("normalized_msisdn", normalizedMSISDN))
	matchedPrefixInfo, err := r.dbQueries.FindMatchingPrefixGroupForMSISDN(logCtx, normalizedMSISDN)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(logCtx, "No MSISDN prefix entry matched", slog.String("normalized_msisdn", normalizedMSISDN))
			result.RoutingResult.Error = fmt.Errorf("no matching MSISDN prefix group for %s", normalizedMSISDN)
			return result, nil
		}
		slog.ErrorContext(logCtx, "Database error finding matching MSISDN prefix group", slog.Any("error", err))
		result.RoutingResult.Error = fmt.Errorf("database error finding prefix group: %w", err)
		return result, err
	}

	result.PrefixGroupID = matchedPrefixInfo.MsisdnPrefixGroupID

	slog.DebugContext(logCtx, "Matched MSISDN prefix group",
		slog.Int("prefix_group_id", int(matchedPrefixInfo.MsisdnPrefixGroupID)),
		slog.String("matched_prefix", matchedPrefixInfo.MatchedPrefix),
	)

	config, err := r.dbQueries.GetRoutingConfigForGroup(logCtx, currentRoutingGroupID)
	if err != nil {
		slog.WarnContext(logCtx, "Failed to get routing config, using defaults", slog.Any("error", err))
		config, err = r.dbQueries.GetGlobalRoutingConfig(logCtx)
		if err != nil {
			slog.ErrorContext(logCtx, "Failed to get global routing config", slog.Any("error", err))
			result.RoutingResult.Error = fmt.Errorf("routing configuration error: %w", err)
			return result, err
		}
	}

	result.RoutingConfig = &config

	routingParams := database.GetAllApplicableRoutesForRoutingParams{
		RoutingGroupID:      *currentRoutingGroupID,
		MsisdnPrefixGroupID: matchedPrefixInfo.MsisdnPrefixGroupID,
	}

	slog.DebugContext(logCtx, "Fetching all applicable routes",
		slog.Int("routing_group_id", int(routingParams.RoutingGroupID)),
		slog.Int("msisdn_prefix_group_id", int(routingParams.MsisdnPrefixGroupID)),
	)

	candidates, err := r.dbQueries.GetAllApplicableRoutesForRouting(logCtx, routingParams)
	if err != nil {
		slog.ErrorContext(logCtx, "Database error getting applicable routes", slog.Any("error", err))
		result.RoutingResult.Error = fmt.Errorf("database error in route lookup: %w", err)
		return result, err
	}

	if len(candidates) == 0 {
		slog.InfoContext(logCtx, "No active routing assignments found",
			slog.Int("routing_group_id", int(routingParams.RoutingGroupID)),
			slog.Int("msisdn_prefix_group_id", int(routingParams.MsisdnPrefixGroupID)),
		)
		result.RoutingResult.Error = fmt.Errorf("no route assignment for prefix group %d in routing group %d",
			routingParams.MsisdnPrefixGroupID, routingParams.RoutingGroupID)
		return result, nil
	}

	slog.InfoContext(logCtx, "Found candidate routes", slog.Int("count", len(candidates)))

	viableCandidates := r.filterViableCandidates(logCtx, candidates, req.SenderID, config)

	if len(viableCandidates) == 0 {
		slog.WarnContext(logCtx, "No healthy routes available after filtering",
			slog.Int("total_candidates", len(candidates)),
		)
		result.RoutingResult.Error = fmt.Errorf("no healthy routes available for destination")
		return result, nil
	}

	slog.InfoContext(logCtx, "Viable routes after health filtering", slog.Int("count", len(viableCandidates)))

	strategy := stringPtrToStr(config.SelectionStrategy)
	if strategy == "" {
		strategy = "priority_weighted"
	}
	selectedRoute := r.selectBestRoute(logCtx, viableCandidates, strategy)

	result.RoutingResult.ShouldUse = true
	result.RoutingResult.MNOID = selectedRoute.MNOID
	result.RoutingResult.RoutingAssignmentID = &selectedRoute.RoutingAssignmentID
	result.RoutingResult.ConnectionID = selectedRoute.ConnectionID
	result.RoutingResult.Protocol = selectedRoute.Protocol
	result.Candidates = viableCandidates

	slog.InfoContext(logCtx, "Route selected successfully with failover candidates",
		slog.Int("mno_id", int(selectedRoute.MNOID)),
		slog.String("mno_name", selectedRoute.MNOName),
		slog.Int("routing_assignment_id", int(selectedRoute.RoutingAssignmentID)),
		slog.Int("priority", int(selectedRoute.Priority)),
		slog.Float64("health_score", selectedRoute.HealthScore),
		slog.String("strategy", strategy),
		slog.Int("total_candidates", len(viableCandidates)),
		slog.Int("alternate_routes", len(viableCandidates)-1),
	)

	return result, nil
}

func (r *IntelligentRouter) filterViableCandidates(
	logCtx context.Context,
	candidates []database.GetAllApplicableRoutesForRoutingRow,
	senderID string,
	config database.RoutingConfig,
) []RouteCandidate {
	rand.Seed(time.Now().UnixNano())
	filtered := make([]RouteCandidate, 0, len(candidates))

	maxConsecutiveFailures := int32(5)
	if config.RouteFailureThreshold != nil {
		maxConsecutiveFailures = *config.RouteFailureThreshold
	}

	for _, c := range candidates {
		candidate := RouteCandidate{
			RoutingAssignmentID: c.RoutingAssignmentID,
			MNOID:               c.MnoID,
			MNOName:             c.MnoName,
			ConnectionID:        c.MnoConnectionID,
			Priority:            c.Priority,
			Weight:              1,
			MaxRetryAttempts:    maxConsecutiveFailures,
		}

		if c.MnoProtocol != nil {
			candidate.Protocol = *c.MnoProtocol
		}

		if c.MaxRetryAttempts != nil {
			candidate.MaxRetryAttempts = *c.MaxRetryAttempts
		}

		if c.ConsecutiveFailures >= maxConsecutiveFailures {
			slog.WarnContext(logCtx, "Route exceeds max consecutive failures, skipping",
				slog.Int("mno_id", int(c.MnoID)),
				slog.Int("failures", int(c.ConsecutiveFailures)),
				slog.Int("max_allowed", int(maxConsecutiveFailures)),
			)
			continue
		}

		candidate.ConsecutiveFailures = c.ConsecutiveFailures

		candidate.HealthScore, _ = c.RouteHealthScore.Float64()
		candidate.IsAvailable = c.RouteIsAvailable

		if !candidate.IsAvailable {
			slog.DebugContext(logCtx, "Route health below threshold, skipping",
				slog.Int("mno_id", int(c.MnoID)),
				slog.Float64("health_score", candidate.HealthScore),
			)
			continue
		}

		if c.Weight != nil {
			candidate.Weight = *c.Weight
		}

		filtered = append(filtered, candidate)
	}

	return filtered
}

func (r *IntelligentRouter) selectBestRoute(logCtx context.Context, candidates []RouteCandidate, strategy string) *RouteCandidate {
	if len(candidates) == 0 {
		return nil
	}

	switch strategy {
	case "random":
		selected := candidates[rand.Intn(len(candidates))]
		slog.DebugContext(logCtx, "Selected route using random strategy", slog.Int("mno_id", int(selected.MNOID)))
		return &selected

	case "least_used":
		selected := candidates[0]
		minWeight := selected.Weight
		for i := range candidates {
			if candidates[i].Weight < minWeight {
				minWeight = candidates[i].Weight
				selected = candidates[i]
			}
		}
		slog.DebugContext(logCtx, "Selected route using least_used strategy", slog.Int("mno_id", int(selected.MNOID)))
		return &selected

	case "priority_only":
		selected := candidates[0]
		highestPriority := selected.Priority
		for i := range candidates {
			if candidates[i].Priority > highestPriority {
				highestPriority = candidates[i].Priority
				selected = candidates[i]
			}
		}
		slog.DebugContext(logCtx, "Selected route using priority_only strategy", slog.Int("mno_id", int(selected.MNOID)))
		return &selected

	case "priority_weighted":
		totalWeight := int64(0)
		for _, c := range candidates {
			totalWeight += int64(c.Weight)
		}

		if totalWeight == 0 {
			totalWeight = int64(len(candidates))
			for i := range candidates {
				candidates[i].Weight = 1
			}
		}

		randomValue := rand.Int63n(totalWeight)
		currentWeight := int64(0)
		for _, c := range candidates {
			currentWeight += int64(c.Weight)
			if randomValue < currentWeight {
				slog.DebugContext(logCtx, "Selected route using priority_weighted strategy",
					slog.Int("mno_id", int(c.MNOID)),
					slog.Int("weight", int(c.Weight)),
				)
				return &c
			}
		}

		return &candidates[0]

	default:
		slog.WarnContext(logCtx, "Unknown selection strategy, using priority_weighted", slog.String("strategy", strategy))
		return r.selectBestRoute(logCtx, candidates, "priority_weighted")
	}
}
