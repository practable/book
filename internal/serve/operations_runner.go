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
	"github.com/practable/book/internal/store"
	"github.com/practable/book/internal/webhook"
)

type operationalActivationPayload struct {
	Version int       `json:"version"`
	JobID   string    `json:"job_id"`
	At      time.Time `json:"at"`
}

func operationsRunnerMiddleware(config serverconfig.ServerConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobID, matched := operationalActionJobID(r.URL.Path, "activate")
		if !matched {
			next.ServeHTTP(w, r)
			return
		}
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if config.Store == nil || len(config.WebhookSecret) != 32 {
			http.Error(w, "operational runner actions are not configured", http.StatusServiceUnavailable)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, callbackBodyLimit+1))
		if err != nil || len(body) > callbackBodyLimit {
			http.Error(w, "invalid activation body", http.StatusBadRequest)
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
		if err := webhook.Verify(config.WebhookSecret, webhook.DirectionCallback, webhook.Headers{
			Timestamp: r.Header.Get(webhook.HeaderTimestamp), DeliveryID: r.Header.Get(webhook.HeaderDeliveryID),
			Direction: r.Header.Get(webhook.HeaderDirection), Signature: r.Header.Get(webhook.HeaderSignature),
		}, now, tolerance, body); err != nil {
			http.Error(w, "invalid webhook authentication", http.StatusUnauthorized)
			return
		}
		var payload operationalActivationPayload
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil || ensureJSONEnd(decoder) != nil || payload.Version != 1 || payload.JobID != jobID || payload.At.IsZero() {
			http.Error(w, "invalid activation", http.StatusBadRequest)
			return
		}
		hash := sha256.Sum256(body)
		activity, err := config.Store.ActivateOperationalJob(r.Context(), jobID, r.Header.Get(webhook.HeaderDeliveryID), hex.EncodeToString(hash[:]))
		switch {
		case errors.Is(err, operational.ErrNotFound):
			http.Error(w, "operational job not found", http.StatusNotFound)
		case errors.Is(err, operational.ErrInvalidTransition), errors.Is(err, operational.ErrCallbackConflict), errors.Is(err, store.ErrBookingConflict):
			http.Error(w, err.Error(), http.StatusConflict)
		case err != nil:
			http.Error(w, "could not activate operational job", http.StatusInternalServerError)
		default:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(activity)
		}
	})
}

func operationalActionJobID(path, action string) (string, bool) {
	suffix := "/" + action
	if !strings.HasPrefix(path, callbackPrefix) || !strings.HasSuffix(path, suffix) {
		return "", false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(path, callbackPrefix), suffix)
	return id, id != "" && !strings.Contains(id, "/")
}
