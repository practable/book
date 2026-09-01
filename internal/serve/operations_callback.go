package serve

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	serverconfig "github.com/practable/book/internal/config"
	operational "github.com/practable/book/internal/operations"
	"github.com/practable/book/internal/webhook"
)

const callbackBodyLimit = 64 << 10
const callbackPrefix = "/api/v1/operations/jobs/"

func operationsCallbackMiddleware(config serverconfig.ServerConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobID, matched := callbackJobID(r.URL.Path)
		if !matched {
			next.ServeHTTP(w, r)
			return
		}
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if config.OperationsRepository == nil || len(config.WebhookSecret) != 32 {
			http.Error(w, "operational callbacks are not configured", http.StatusServiceUnavailable)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, callbackBodyLimit+1))
		if err != nil || len(body) > callbackBodyLimit {
			http.Error(w, "invalid callback body", http.StatusBadRequest)
			return
		}
		now := time.Now().UTC()
		if config.Now != nil {
			now = config.Now().UTC()
		}
		tolerance := config.WebhookTolerance
		if tolerance <= 0 {
			tolerance = 5 * time.Minute
		}
		err = webhook.Verify(config.WebhookSecret, webhook.DirectionCallback, webhook.Headers{
			Timestamp:  r.Header.Get(webhook.HeaderTimestamp),
			DeliveryID: r.Header.Get(webhook.HeaderDeliveryID),
			Direction:  r.Header.Get(webhook.HeaderDirection),
			Signature:  r.Header.Get(webhook.HeaderSignature),
		}, now, tolerance, body)
		if err != nil {
			http.Error(w, "invalid webhook authentication", http.StatusUnauthorized)
			return
		}
		var payload operational.CallbackPayload
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil || payload.Version != 1 || payload.JobID != jobID || payload.State == "" || payload.At.IsZero() {
			http.Error(w, "invalid callback", http.StatusBadRequest)
			return
		}
		if err := ensureJSONEnd(decoder); err != nil {
			http.Error(w, "invalid callback", http.StatusBadRequest)
			return
		}
		hash := sha256.Sum256(body)
		job, duplicate, err := config.OperationsRepository.ApplyCallback(r.Context(), operational.Callback{
			DeliveryID: r.Header.Get(webhook.HeaderDeliveryID), JobID: jobID,
			State: payload.State, At: payload.At.UTC(), Code: payload.Code, Error: payload.Error,
		}, hex.EncodeToString(hash[:]))
		switch {
		case errors.Is(err, operational.ErrNotFound):
			http.Error(w, "operational job not found", http.StatusNotFound)
		case errors.Is(err, operational.ErrCallbackConflict), errors.Is(err, operational.ErrInvalidTransition):
			http.Error(w, err.Error(), http.StatusConflict)
		case err != nil:
			http.Error(w, "could not record callback", http.StatusInternalServerError)
		default:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"job_id": job.ID, "state": job.State, "duplicate": duplicate})
		}
	})
}

func callbackJobID(path string) (string, bool) {
	if !strings.HasPrefix(path, callbackPrefix) || !strings.HasSuffix(path, "/callbacks") {
		return "", false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(path, callbackPrefix), "/callbacks")
	if id == "" || strings.Contains(id, "/") {
		return "", false
	}
	return id, true
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var extra interface{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON content")
	}
	return nil
}
