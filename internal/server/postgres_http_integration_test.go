package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/phayes/freeport"
	"github.com/practable/book/internal/config"
	"github.com/practable/book/internal/login"
	"github.com/practable/book/internal/postgres"
	"github.com/practable/book/internal/store"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestPostgresBackedReleaseHTTPContract(t *testing.T) {
	databaseURL := os.Getenv("BOOK_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("BOOK_TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithCancel(context.Background())
	pool, err := pgxpool.Connect(ctx, databaseURL)
	require.NoError(t, err)
	lock, err := pool.Acquire(ctx)
	require.NoError(t, err)
	_, err = lock.Exec(ctx, "SELECT pg_advisory_lock(1651470187)")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = lock.Exec(context.Background(), "SELECT pg_advisory_unlock(1651470187)")
		lock.Release()
		pool.Close()
	})
	_, err = pool.Exec(ctx, "TRUNCATE public.resource_release_events, public.resource_release_state, public.operational_alerts, public.operational_stream_health, public.operational_usage_ledger, public.booking_activation_stages, public.booking_activation_runs, public.operational_schedule_occurrences, public.webhook_callback_receipts, public.webhook_deliveries, public.operational_jobs, public.relay_revocations, public.booking_events, public.booking_replacements, public.bookings, public.user_groups, public.resource_availability, public.slot_availability, public.service_state, public.active_manifest, public.manifest_versions RESTART IDENTITY")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "INSERT INTO public.service_state(singleton,updated_at_ns) VALUES(true,0)")
	require.NoError(t, err)

	repository, err := postgres.Open(ctx, databaseURL, 4, 10*time.Second)
	require.NoError(t, err)
	t.Cleanup(repository.Close)
	port, err := freeport.GetFreePort()
	require.NoError(t, err)
	host := fmt.Sprintf("http://127.0.0.1:%d", port)
	now := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	secret := []byte("postgres-http-test-secret")
	server, err := NewWithError(config.ServerConfig{Host: host, Port: port, StoreSecret: secret, RelaySecret: []byte("relay-test-secret"),
		Repository: repository, OperationsRepository: repository, Now: func() time.Time { return now }, PruneEvery: time.Hour, CheckEvery: time.Hour})
	require.NoError(t, err)

	var manifest store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestYAML, &manifest))
	manifest.OperationalWorkflows = map[string]store.OperationalWorkflow{"video-check": {Description: "Check video", Kind: "health_check", ExpectedDuration: time.Second, MaximumDuration: 5 * time.Second}}
	manifest.OperationalJobTemplates = map[string]store.OperationalJobTemplate{"video-check": {Workflow: "video-check", Timeout: 5 * time.Second}}
	manifest.OperationalPipelineTemplates = map[string]store.OperationalPipelineTemplate{"video": {Stages: []store.OperationalPipelineStage{{Name: "check", JobTemplate: "video-check"}}}}
	resource := manifest.Resources["r-a"]
	resource.StreamOperations = map[string]store.OperationalStreamBinding{"st-b": {ActivationPipeline: "video"}}
	manifest.Resources["r-a"] = resource
	require.NoError(t, server.Store.ReplaceManifest(manifest))
	require.NoError(t, server.Store.SetResourceIsAvailableBy("r-a", false, "camera fault", "technician"))

	go server.Run(ctx)
	t.Cleanup(cancel)
	client := &http.Client{Timeout: 2 * time.Second}
	for deadline := time.Now().Add(2 * time.Second); ; {
		response, requestErr := client.Get(host + "/api/v1/status")
		if requestErr == nil {
			response.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not start: %v", requestErr)
		}
		time.Sleep(10 * time.Millisecond)
	}
	token := login.New(host, "admin", []string{"booking:admin"}, now.Unix()-1, now.Unix()-1, now.Add(time.Hour).Unix())
	signed, err := login.Sign(token, string(secret))
	require.NoError(t, err)

	do := func(method, path string) (*http.Response, []byte) {
		t.Helper()
		request, err := http.NewRequest(method, host+path, nil)
		require.NoError(t, err)
		request.Header.Set("Authorization", signed)
		response, err := client.Do(request)
		require.NoError(t, err)
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		require.NoError(t, err)
		return response, body
	}

	response, body := do(http.MethodPost, "/api/v1/admin/resource-holds/r-a/release")
	require.Equal(t, http.StatusAccepted, response.StatusCode, string(body))
	var release map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &release))
	require.Equal(t, "pending_checks", release["state"])
	require.Equal(t, []interface{}{"st-b"}, release["required_streams"])

	response, body = do(http.MethodPost, "/api/v1/admin/resource-holds/r-a/release?override_reason=video+unavailable%3B+data+usable")
	require.Equal(t, http.StatusAccepted, response.StatusCode, string(body))
	require.NoError(t, json.Unmarshal(body, &release))
	require.Equal(t, "degraded_override", release["state"])
	require.Equal(t, []interface{}{"st-b"}, release["failing_streams"])

	response, body = do(http.MethodGet, "/api/v1/calendar/catalog/g-a")
	require.Equal(t, http.StatusOK, response.StatusCode, string(body))
	var catalogue []struct {
		Resources []struct {
			Name               string   `json:"name"`
			ActivationStreams  []string `json:"activation_streams"`
			Degraded           bool     `json:"degraded"`
			DegradedReason     string   `json:"degraded_reason"`
			UnavailableStreams []string `json:"unavailable_streams"`
		} `json:"resources"`
	}
	require.NoError(t, json.Unmarshal(body, &catalogue))
	require.True(t, catalogue[0].Resources[0].Degraded)
	require.Equal(t, []string{"st-b"}, catalogue[0].Resources[0].ActivationStreams)
	require.Equal(t, "video unavailable; data usable", catalogue[0].Resources[0].DegradedReason)
	require.Equal(t, []string{"st-b"}, catalogue[0].Resources[0].UnavailableStreams)
}
