package sms

import (
	"context"
	"errors"
	"log/slog"

	"github.com/thrillee/aegisbox/internal/database"
	"github.com/thrillee/aegisbox/internal/logging"
	"github.com/thrillee/aegisbox/pkg/codes"
)

type RoutingService struct {
	dbQueries database.Querier
	router    Router
	validator Validator
}

type RoutingServiceDependencies struct {
	DBQueries database.Querier
	Router    Router
	Validator Validator
}

func NewRoutingService(deps RoutingServiceDependencies) *RoutingService {
	return &RoutingService{
		dbQueries: deps.DBQueries,
		router:    deps.Router,
		validator: deps.Validator,
	}
}

func (s *RoutingService) RouteMessage(
	ctx context.Context,
	msg database.GetMessagesByStatusRow,
) *RoutingOutcome {
	logCtx := logging.ContextWithMessageID(ctx, msg.ID)
	logCtx = logging.ContextWithSPID(logCtx, msg.ServiceProviderID)

	slog.DebugContext(logCtx, "Starting validation and routing")

	outcome := &RoutingOutcome{
		Status:  codes.MsgStatusFailedValidation,
		IsValid: false,
	}

	validationRes, valErr := s.validator.Validate(
		logCtx,
		msg.ServiceProviderID,
		msg.OriginalSourceAddr,
	)
	if valErr != nil {
		slog.ErrorContext(logCtx, "Validation internal error", slog.Any("error", valErr))
		outcome.Error = valErr
		return outcome
	}
	if validationRes == nil {
		outcome.Error = errors.New("validation returned nil result")
		outcome.IsLogicalFailure = true
		slog.ErrorContext(logCtx, "Validation returned nil result")
		return outcome
	}

	if !validationRes.IsValid {
		outcome.Status = codes.MsgStatusFailedValidation
		outcome.Error = validationRes.Error
		outcome.IsLogicalFailure = true
		slog.WarnContext(logCtx, "Validation failed", slog.Any("error", validationRes.Error))

		errDesc := validationRes.Error.Error()
		errCode := "VAL_FAILED"
		params := database.UpdateMessageValidatedRoutedParams{
			ProcessingStatus:    codes.MsgStatusFailedValidation,
			RoutedMnoID:         nil,
			RoutingAssignmentID: nil,
			ApprovedSenderID:    nil,
			TemplateID:          nil,
			ErrorDescription:    &errDesc,
			ErrorCode:           &errCode,
			ID:                  msg.ID,
		}
		if dbErr := s.dbQueries.UpdateMessageValidatedRouted(logCtx, params); dbErr != nil {
			slog.ErrorContext(logCtx, "Failed to update message to failed_validation status", slog.Any("db_error", dbErr))
		}
		return outcome
	}

	routingRes, routeErr := s.router.Route(ctx, RoutingRequest{
		SPCredentialID: msg.SpCredentialID,
		MSISDN:         msg.OriginalDestinationAddr,
		SenderID:       msg.OriginalSourceAddr,
	})
	if routeErr != nil {
		outcome.Error = routeErr
		slog.ErrorContext(logCtx, "Routing internal error", slog.Any("error", routeErr))
		return outcome
	}
	if routingRes == nil || !routingRes.ShouldUse {
		outcome.Status = codes.MsgStatusFailedRouting
		if routingRes != nil {
			outcome.Error = routingRes.Error
		} else {
			outcome.Error = errors.New("routing returned nil result")
		}
		outcome.IsLogicalFailure = true
		slog.WarnContext(logCtx, "Routing failed", slog.Any("error", outcome.Error))

		errDesc := "NO_ROUTE"
		errCode := "ROUTE_FAILED"
		params := database.UpdateMessageValidatedRoutedParams{
			ProcessingStatus:    codes.MsgStatusFailedRouting,
			RoutedMnoID:         nil,
			RoutingAssignmentID: nil,
			ApprovedSenderID:    nil,
			TemplateID:          nil,
			ErrorDescription:    &errDesc,
			ErrorCode:           &errCode,
			ID:                  msg.ID,
		}
		if dbErr := s.dbQueries.UpdateMessageValidatedRouted(ctx, params); dbErr != nil {
			slog.ErrorContext(ctx, "Failed to update message to failed_routing status", slog.Any("db_error", dbErr))
		}
		return outcome
	}

	dbErrMsg := ""
	params := database.UpdateMessageValidatedRoutedParams{
		ProcessingStatus:    codes.MsgStatusRouted,
		RoutedMnoID:         &routingRes.MNOID,
		RoutingAssignmentID: routingRes.RoutingAssignmentID,
		ApprovedSenderID:    &validationRes.ApprovedSenderID,
		TemplateID:          nil,
		ErrorDescription:    &dbErrMsg,
		ErrorCode:           nil,
		ID:                  msg.ID,
	}

	updateErr := s.dbQueries.UpdateMessageValidatedRouted(ctx, params)
	if updateErr != nil {
		slog.ErrorContext(ctx, "Failed to update message status", slog.Any("db_error", updateErr))
		outcome.Error = updateErr
		return outcome
	}

	outcome.Status = codes.MsgStatusRouted
	outcome.MNOID = routingRes.MNOID
	outcome.ApprovedSenderID = validationRes.ApprovedSenderID
	outcome.IsValid = true
	outcome.IsLogicalFailure = false

	slog.InfoContext(logCtx, "Routing completed successfully",
		slog.Int("mno_id", int(routingRes.MNOID)),
		slog.Any("routing_assignment_id", routingRes.RoutingAssignmentID),
	)

	return outcome
}
