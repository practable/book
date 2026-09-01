package serve

import (
	"encoding/json"
	"errors"

	"github.com/go-openapi/runtime/middleware"
	"github.com/practable/book/internal/config"
	"github.com/practable/book/internal/operations"
	"github.com/practable/book/internal/serve/models"
	"github.com/practable/book/internal/serve/restapi/operations/users"
	"github.com/practable/book/internal/store"
)

func authorizeBookingActivation(principal interface{}, userName string) error {
	admin, claims, err := isActivityCaller(principal)
	if err != nil {
		return err
	}
	if !admin && claims.Subject != userName {
		return errors.New("user_name in path does not match subject in token")
	}
	return nil
}

func activationToModel(run operations.ActivationRun) *models.BookingActivation {
	id, booking, stream, state, cleanupState, progress := run.ID, run.BookingName, run.Stream, run.State, run.CleanupState, run.ProgressMessage
	current := int64(run.CurrentStage)
	recoveryAttempt, maximumRecoveryAttempts := int64(run.RecoveryAttempt), int64(run.MaximumRecoveryAttempts)
	result := &models.BookingActivation{ID: &id, BookingName: &booking, Stream: &stream, State: &state, CleanupState: &cleanupState, CurrentStage: &current,
		RecoveryAttempt: &recoveryAttempt, MaximumRecoveryAttempts: &maximumRecoveryAttempts,
		ProgressMessage: &progress, FailureCode: run.FailureCode, FailureMessage: run.FailureMessage}
	convertStages := func(values []operations.ActivationStage) []*models.BookingActivationStage {
		converted := make([]*models.BookingActivationStage, 0, len(values))
		for _, item := range values {
			index, attempt, maximum := int64(item.Index), int64(item.Attempt), int64(item.MaximumAttempts)
			name, stageState := item.Name, item.State
			converted = append(converted, &models.BookingActivationStage{Index: &index, Name: &name, State: &stageState, Attempt: &attempt,
				MaximumAttempts: &maximum, ProgressMessage: item.ProgressMessage, LastErrorCode: item.LastErrorCode, LastError: item.LastError})
		}
		return converted
	}
	result.Stages = convertStages(run.Stages)
	result.RecoveryStages = convertStages(run.RecoveryStages)
	result.CleanupStages = convertStages(run.CleanupStages)
	if len(run.FailureGuidance) > 0 && string(run.FailureGuidance) != "null" {
		var guidance store.OperationalFailureGuidance
		if json.Unmarshal(run.FailureGuidance, &guidance) == nil {
			converted := &models.OperationalFailureGuidance{Title: guidance.Title, Message: guidance.Message}
			for _, item := range guidance.Actions {
				actionType, label := item.Type, item.Label
				converted.Actions = append(converted.Actions, &models.OperationalFailureAction{Type: &actionType, Label: &label, URL: item.URL})
			}
			result.FailureGuidance = converted
		}
	}
	return result
}

func beginBookingActivationHandler(config config.ServerConfig) func(users.BeginBookingActivationParams, interface{}) middleware.Responder {
	return func(params users.BeginBookingActivationParams, principal interface{}) middleware.Responder {
		if err := authorizeBookingActivation(principal, params.UserName); err != nil {
			code, message := "401", err.Error()
			return users.NewBeginBookingActivationUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		run, _, err := config.Store.BeginBookingActivation(params.HTTPRequest.Context(), params.BookingName, *params.Request.Stream, params.IdempotencyKey)
		if err == nil {
			return users.NewBeginBookingActivationAccepted().WithPayload(activationToModel(run))
		}
		code, message := "500", err.Error()
		if errors.Is(err, operations.ErrActivationConflict) || errors.Is(err, store.ErrBookingConflict) {
			code = "409"
			return users.NewBeginBookingActivationConflict().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		if errors.Is(err, store.ErrPersistentNotFound) {
			code = "404"
			return users.NewBeginBookingActivationNotFound().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		return users.NewBeginBookingActivationInternalServerError().WithPayload(&models.Error{Code: &code, Message: &message})
	}
}

func getBookingActivationHandler(config config.ServerConfig) func(users.GetBookingActivationParams, interface{}) middleware.Responder {
	return func(params users.GetBookingActivationParams, principal interface{}) middleware.Responder {
		if err := authorizeBookingActivation(principal, params.UserName); err != nil {
			code, message := "401", err.Error()
			return users.NewGetBookingActivationUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		run, err := config.Store.GetBookingActivation(params.HTTPRequest.Context(), params.ActivationID)
		if errors.Is(err, operations.ErrNotFound) || (err == nil && (run.BookingName != params.BookingName || run.User != params.UserName)) {
			code, message := "404", "booking activation not found"
			return users.NewGetBookingActivationNotFound().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		if err != nil {
			code, message := "500", err.Error()
			return users.NewGetBookingActivationInternalServerError().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		return users.NewGetBookingActivationOK().WithPayload(activationToModel(run))
	}
}
