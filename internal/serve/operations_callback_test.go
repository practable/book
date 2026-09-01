package serve

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/practable/book/internal/config"
	operational "github.com/practable/book/internal/operations"
	"github.com/practable/book/internal/webhook"
)

type callbackRepo struct {
	callback  operational.Callback
	hash      string
	duplicate bool
	err       error
}

func (r *callbackRepo) CreateJob(context.Context, operational.Job, operational.Delivery) (operational.Job, bool, error) {
	panic("unused")
}
func (r *callbackRepo) ClaimDeliveries(context.Context, string, time.Time, time.Duration, int) ([]operational.Delivery, error) {
	panic("unused")
}
func (r *callbackRepo) CompleteDelivery(context.Context, string, string, bool, int, string, time.Time, time.Time) error {
	panic("unused")
}
func (r *callbackRepo) ApplyCallback(_ context.Context, callback operational.Callback, hash string) (operational.Job, bool, error) {
	r.callback, r.hash = callback, hash
	return operational.Job{ID: callback.JobID, State: callback.State}, r.duplicate, r.err
}
func (r *callbackRepo) GetJob(context.Context, string) (operational.Job, error) { panic("unused") }

func TestOperationsCallbackAuthenticatesAndApplies(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	secret := []byte("01234567890123456789012345678901")
	body := []byte(`{"version":1,"job_id":"job-1","state":"running","at":"2026-09-01T12:00:00Z"}`)
	repo := &callbackRepo{}
	h := operationsCallbackMiddleware(config.ServerConfig{OperationsRepository: repo, WebhookSecret: secret, Now: func() time.Time { return now }}, http.NotFoundHandler())
	req := httptest.NewRequest(http.MethodPost, callbackPrefix+"job-1/callbacks", bytes.NewReader(body))
	headers, _ := webhook.Sign(secret, webhook.DirectionCallback, "callback-1", now, body)
	req.Header.Set(webhook.HeaderTimestamp, headers.Timestamp)
	req.Header.Set(webhook.HeaderDeliveryID, headers.DeliveryID)
	req.Header.Set(webhook.HeaderDirection, headers.Direction)
	req.Header.Set(webhook.HeaderSignature, headers.Signature)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	if repo.callback.JobID != "job-1" || repo.callback.DeliveryID != "callback-1" || repo.hash == "" {
		t.Fatalf("callback = %#v, hash=%q", repo.callback, repo.hash)
	}
}

func TestOperationsCallbackRejectsTampering(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	secret := []byte("01234567890123456789012345678901")
	signed := []byte(`{"version":1,"job_id":"job-1","state":"running","at":"2026-09-01T12:00:00Z"}`)
	tampered := bytes.Replace(signed, []byte("running"), []byte("failed"), 1)
	h := operationsCallbackMiddleware(config.ServerConfig{OperationsRepository: &callbackRepo{}, WebhookSecret: secret, Now: func() time.Time { return now }}, http.NotFoundHandler())
	req := httptest.NewRequest(http.MethodPost, callbackPrefix+"job-1/callbacks", bytes.NewReader(tampered))
	headers, _ := webhook.Sign(secret, webhook.DirectionCallback, "callback-1", now, signed)
	req.Header.Set(webhook.HeaderTimestamp, headers.Timestamp)
	req.Header.Set(webhook.HeaderDeliveryID, headers.DeliveryID)
	req.Header.Set(webhook.HeaderDirection, headers.Direction)
	req.Header.Set(webhook.HeaderSignature, headers.Signature)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, req)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestOperationsCallbackRejectsJobMismatch(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	secret := []byte("01234567890123456789012345678901")
	body := []byte(`{"version":1,"job_id":"job-2","state":"running","at":"2026-09-01T12:00:00Z"}`)
	h := operationsCallbackMiddleware(config.ServerConfig{OperationsRepository: &callbackRepo{}, WebhookSecret: secret, Now: func() time.Time { return now }}, http.NotFoundHandler())
	req := httptest.NewRequest(http.MethodPost, callbackPrefix+"job-1/callbacks", bytes.NewReader(body))
	headers, _ := webhook.Sign(secret, webhook.DirectionCallback, "callback-1", now, body)
	req.Header.Set(webhook.HeaderTimestamp, headers.Timestamp)
	req.Header.Set(webhook.HeaderDeliveryID, headers.DeliveryID)
	req.Header.Set(webhook.HeaderDirection, headers.Direction)
	req.Header.Set(webhook.HeaderSignature, headers.Signature)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, req)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}
