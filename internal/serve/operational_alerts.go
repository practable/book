package serve

import (
	"errors"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/practable/book/internal/config"
	"github.com/practable/book/internal/serve/models"
	"github.com/practable/book/internal/serve/restapi/operations/admin"
	"github.com/practable/book/internal/store"
)

func beginOperationalHealthCheckHandler(cfg config.ServerConfig) func(admin.BeginOperationalHealthCheckParams, interface{}) middleware.Responder {
	return func(params admin.BeginOperationalHealthCheckParams, principal interface{}) middleware.Responder {
		claims, err := isMaintenanceOrAdmin(principal)
		if err != nil {
			code, message := "401", "no scope booking:maintenance"
			return admin.NewBeginOperationalHealthCheckUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		operator := claims.Subject
		if operator == "" {
			operator = "maintenance"
		}
		run, _, err := cfg.Store.BeginOperationalHealthCheck(params.HTTPRequest.Context(), params.ResourceName, params.StreamName, operator, params.IdempotencyKey)
		if err == nil {
			return admin.NewBeginOperationalHealthCheckAccepted().WithPayload(activationToModel(run))
		}
		code, message := "500", err.Error()
		if errors.Is(err, store.ErrPersistentNotFound) {
			code = "404"
			return admin.NewBeginOperationalHealthCheckNotFound().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		if errors.Is(err, store.ErrBookingConflict) || errors.Is(err, store.ErrBookingIDConflict) {
			code = "409"
			return admin.NewBeginOperationalHealthCheckConflict().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		return admin.NewBeginOperationalHealthCheckInternalServerError().WithPayload(&models.Error{Code: &code, Message: &message})
	}
}

func operationalAlertToModel(value store.OperationalAlert) *models.OperationalAlert {
	first, last := strfmt.DateTime(value.FirstSeen), strfmt.DateTime(value.LastSeen)
	result := &models.OperationalAlert{ID: &value.ID, Resource: &value.Resource, Stream: &value.Stream, Code: &value.Code,
		Message: &value.Message, JobID: &value.JobID, ManifestVersion: &value.ManifestVersion, Status: &value.Status,
		Occurrences: &value.Occurrences, FirstSeen: &first, LastSeen: &last, AcknowledgedBy: &value.AcknowledgedBy, ResolvedBy: &value.ResolvedBy}
	if value.AcknowledgedAt != nil {
		result.AcknowledgedAt = strfmt.DateTime(*value.AcknowledgedAt)
	}
	if value.ResolvedAt != nil {
		result.ResolvedAt = strfmt.DateTime(*value.ResolvedAt)
	}
	return result
}

func listOperationalAlertsHandler(cfg config.ServerConfig) func(admin.ListOperationalAlertsParams, interface{}) middleware.Responder {
	return func(params admin.ListOperationalAlertsParams, principal interface{}) middleware.Responder {
		if _, err := isAdmin(principal); err != nil {
			code, message := "401", "no scope booking:admin"
			return admin.NewListOperationalAlertsUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		status, limit := "active", 200
		if params.Status != nil {
			status = *params.Status
		}
		if params.Limit != nil {
			limit = int(*params.Limit)
		}
		values, err := cfg.Store.ListOperationalAlerts(params.HTTPRequest.Context(), status, limit)
		if err != nil {
			code, message := "500", "could not read operational alerts"
			return admin.NewListOperationalAlertsInternalServerError().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		payload := make([]*models.OperationalAlert, 0, len(values))
		for _, value := range values {
			payload = append(payload, operationalAlertToModel(value))
		}
		return admin.NewListOperationalAlertsOK().WithPayload(payload)
	}
}

func acknowledgeOperationalAlertHandler(cfg config.ServerConfig) func(admin.AcknowledgeOperationalAlertParams, interface{}) middleware.Responder {
	return func(params admin.AcknowledgeOperationalAlertParams, principal interface{}) middleware.Responder {
		claims, err := isAdmin(principal)
		if err != nil {
			code, message := "401", "no scope booking:admin"
			return admin.NewAcknowledgeOperationalAlertUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		value, err := cfg.Store.SetOperationalAlertStatus(params.HTTPRequest.Context(), params.AlertID, "acknowledged", claims.Subject)
		if errors.Is(err, store.ErrPersistentNotFound) {
			code, message := "404", "alert not found or no longer open"
			return admin.NewAcknowledgeOperationalAlertNotFound().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		if err != nil {
			code, message := "500", "could not acknowledge alert"
			return admin.NewAcknowledgeOperationalAlertInternalServerError().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		return admin.NewAcknowledgeOperationalAlertOK().WithPayload(operationalAlertToModel(value))
	}
}

func resolveOperationalAlertHandler(cfg config.ServerConfig) func(admin.ResolveOperationalAlertParams, interface{}) middleware.Responder {
	return func(params admin.ResolveOperationalAlertParams, principal interface{}) middleware.Responder {
		claims, err := isAdmin(principal)
		if err != nil {
			code, message := "401", "no scope booking:admin"
			return admin.NewResolveOperationalAlertUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		value, err := cfg.Store.SetOperationalAlertStatus(params.HTTPRequest.Context(), params.AlertID, "resolved", claims.Subject)
		if errors.Is(err, store.ErrPersistentNotFound) {
			code, message := "404", "alert not found or already resolved"
			return admin.NewResolveOperationalAlertNotFound().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		if err != nil {
			code, message := "500", "could not resolve alert"
			return admin.NewResolveOperationalAlertInternalServerError().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		return admin.NewResolveOperationalAlertOK().WithPayload(operationalAlertToModel(value))
	}
}

func listOperationalHealthHandler(cfg config.ServerConfig) func(admin.ListOperationalHealthParams, interface{}) middleware.Responder {
	return func(params admin.ListOperationalHealthParams, principal interface{}) middleware.Responder {
		if _, err := isAdmin(principal); err != nil {
			code, message := "401", "no scope booking:admin"
			return admin.NewListOperationalHealthUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		values, err := cfg.Store.ListOperationalStreamHealth(params.HTTPRequest.Context())
		if err != nil {
			code, message := "500", "could not read operational health"
			return admin.NewListOperationalHealthInternalServerError().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		payload := make([]*models.OperationalStreamHealth, 0, len(values))
		for _, value := range values {
			checked := strfmt.DateTime(value.CheckedAt)
			payload = append(payload, &models.OperationalStreamHealth{Resource: &value.Resource, Stream: &value.Stream,
				Status: &value.Status, Code: &value.Code, Message: &value.Message, JobID: &value.JobID,
				ManifestVersion: &value.ManifestVersion, CheckedAt: &checked})
		}
		return admin.NewListOperationalHealthOK().WithPayload(payload)
	}
}

func listResourceHoldsHandler(cfg config.ServerConfig) func(admin.ListResourceHoldsParams, interface{}) middleware.Responder {
	return func(params admin.ListResourceHoldsParams, principal interface{}) middleware.Responder {
		if _, err := isAdmin(principal); err != nil {
			code, message := "401", "no scope booking:admin"
			return admin.NewListResourceHoldsUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		values, err := cfg.Store.ListResourceHolds(params.HTTPRequest.Context())
		if err != nil {
			code, message := "500", "could not read resource holds"
			return admin.NewListResourceHoldsInternalServerError().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		payload := make([]*models.ResourceHold, 0, len(values))
		for _, value := range values {
			heldSince := strfmt.DateTime(value.HeldSince)
			payload = append(payload, &models.ResourceHold{Resource: &value.Resource, Reason: &value.Reason,
				HeldSince: &heldSince, HeldBy: &value.HeldBy})
		}
		return admin.NewListResourceHoldsOK().WithPayload(payload)
	}
}

func resourceReleaseToModel(value store.ResourceReleaseState) *models.ResourceReleaseState {
	requested := strfmt.DateTime(value.RequestedAt)
	result := &models.ResourceReleaseState{Resource: &value.Resource, State: &value.State, RequiredStreams: value.RequiredStreams,
		FailingStreams: value.FailingStreams, RequestedAt: &requested, RequestedBy: &value.RequestedBy,
		ManifestVersion: &value.ManifestVersion, OverrideReason: &value.OverrideReason}
	if value.ReleasedAt != nil {
		result.ReleasedAt = strfmt.DateTime(*value.ReleasedAt)
	}
	return result
}

func requestResourceReleaseHandler(cfg config.ServerConfig) func(admin.RequestResourceReleaseParams, interface{}) middleware.Responder {
	return func(params admin.RequestResourceReleaseParams, principal interface{}) middleware.Responder {
		claims, err := isMaintenanceOrAdmin(principal)
		if err != nil {
			code, message := "401", "no scope booking:maintenance"
			return admin.NewRequestResourceReleaseUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		reason := ""
		if params.OverrideReason != nil {
			reason = *params.OverrideReason
		}
		value, err := cfg.Store.RequestResourceRelease(params.HTTPRequest.Context(), params.ResourceName, claims.Subject, reason)
		if err == nil {
			return admin.NewRequestResourceReleaseAccepted().WithPayload(resourceReleaseToModel(value))
		}
		code, message := "409", err.Error()
		if errors.Is(err, store.ErrPersistentNotFound) {
			code = "404"
			return admin.NewRequestResourceReleaseNotFound().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		return admin.NewRequestResourceReleaseConflict().WithPayload(&models.Error{Code: &code, Message: &message})
	}
}

func listResourceReleasesHandler(cfg config.ServerConfig) func(admin.ListResourceReleasesParams, interface{}) middleware.Responder {
	return func(params admin.ListResourceReleasesParams, principal interface{}) middleware.Responder {
		if _, err := isMaintenanceOrAdmin(principal); err != nil {
			code, message := "401", "no scope booking:maintenance"
			return admin.NewListResourceReleasesUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		values, err := cfg.Store.ListResourceReleaseStates(params.HTTPRequest.Context())
		if err != nil {
			code, message := "500", err.Error()
			return admin.NewListResourceReleasesInternalServerError().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		payload := make([]*models.ResourceReleaseState, 0, len(values))
		for _, value := range values {
			payload = append(payload, resourceReleaseToModel(value))
		}
		return admin.NewListResourceReleasesOK().WithPayload(payload)
	}
}
