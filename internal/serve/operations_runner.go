package serve

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/icza/gog"
	serverconfig "github.com/practable/book/internal/config"
	operational "github.com/practable/book/internal/operations"
	"github.com/practable/book/internal/serve/models"
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
		if config.Store == nil || len(config.WebhookSecret) != 32 || len(config.RelaySecret) == 0 {
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
			response, err := operationalActivityModel(activity, config.RelaySecret)
			if err != nil {
				http.Error(w, "could not prepare operational activity", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}
	})
}

// operationalActivityModel adds relay credentials to the activity returned to
// the runner. All time-dependent claims are derived from the persisted booking
// interval so an exact activation retry produces the same signed credentials.
func operationalActivityModel(activity store.Activity, relaySecret []byte) (*models.Activity, error) {
	issuedAt := jwt.NewNumericDate(activity.NotBefore.Add(-time.Second))
	expiresAt := jwt.NewNumericDate(activity.ExpiresAt)
	keys := make([]string, 0, len(activity.Streams))
	for key := range activity.Streams {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	streams := make([]*models.ActivityStream, 0, len(keys))
	for _, key := range keys {
		stream := activity.Streams[key]
		permission := Permission{
			BookingID: activity.BookingID,
			Topic:     stream.Topic,
			Prefix:    stream.ConnectionType,
			Scopes:    stream.Scopes,
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  issuedAt,
				NotBefore: issuedAt,
				ExpiresAt: expiresAt,
				Subject:   "__operations__",
				Audience:  jwt.ClaimStrings{stream.URL},
			},
		}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, permission).SignedString(relaySecret)
		if err != nil {
			return nil, err
		}
		accessURL := strings.TrimRight(stream.URL, "/") + "/" + stream.ConnectionType + "/" + stream.Topic
		streams = append(streams, &models.ActivityStream{
			Audience:       gog.Ptr(stream.URL),
			ConnectionType: gog.Ptr(stream.ConnectionType),
			For:            gog.Ptr(stream.For),
			Prefix:         stream.ConnectionType,
			Scopes:         append([]string(nil), stream.Scopes...),
			Token:          token,
			Topic:          gog.Ptr(stream.Topic),
			URL:            gog.Ptr(accessURL),
		})
	}
	uis := make([]*models.UIDescribed, 0, len(activity.UIs))
	for _, ui := range activity.UIs {
		uis = append(uis, &models.UIDescribed{
			Description: &models.Description{
				Name: &ui.Description.Name, Type: &ui.Description.Type,
				Short: ui.Description.Short, Long: ui.Description.Long,
				Further: ui.Description.Further, Thumb: ui.Description.Thumb,
				Image: ui.Description.Image,
			},
			URL: &ui.URL, StreamsRequired: append([]string(nil), ui.StreamsRequired...),
		})
	}
	nbf, exp := float64(activity.NotBefore.Unix()), float64(activity.ExpiresAt.Unix())
	return &models.Activity{
		Config: activity.ConfigURL,
		Description: &models.Description{
			Name: &activity.Description.Name, Type: &activity.Description.Type,
			Short: activity.Description.Short, Long: activity.Description.Long,
			Further: activity.Description.Further, Thumb: activity.Description.Thumb,
			Image: activity.Description.Image,
		},
		Nbf: &nbf, Exp: &exp, Streams: streams, Uis: uis,
	}, nil
}

func operationalActionJobID(path, action string) (string, bool) {
	suffix := "/" + action
	if !strings.HasPrefix(path, callbackPrefix) || !strings.HasSuffix(path, suffix) {
		return "", false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(path, callbackPrefix), suffix)
	return id, id != "" && !strings.Contains(id, "/")
}
