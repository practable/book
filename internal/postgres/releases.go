package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/practable/book/internal/store"
)

func (r *Repository) RequestVerifiedResourceRelease(ctx context.Context, resource string, streams []string, actor string, manifestVersion int64, at time.Time) (store.ResourceReleaseState, error) {
	return r.setResourceRelease(ctx, resource, streams, nil, actor, "", manifestVersion, at, false)
}

func (r *Repository) OverrideResourceRelease(ctx context.Context, resource string, required, failing []string, actor, reason string, manifestVersion int64, at time.Time) (store.ResourceReleaseState, error) {
	if reason == "" {
		return store.ResourceReleaseState{}, errors.New("degraded release reason is required")
	}
	return r.setResourceRelease(ctx, resource, required, failing, actor, reason, manifestVersion, at, true)
}

func (r *Repository) setResourceRelease(ctx context.Context, resource string, required, failing []string, actor, reason string, manifestVersion int64, at time.Time, override bool) (store.ResourceReleaseState, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return store.ResourceReleaseState{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", maintenanceLock); err != nil {
		return store.ResourceReleaseState{}, err
	}
	if err = assertManifestVersion(ctx, tx, manifestVersion); err != nil {
		return store.ResourceReleaseState{}, err
	}
	var available bool
	if err = tx.QueryRow(ctx, `SELECT available FROM public.resource_availability WHERE resource_name=$1 FOR UPDATE`, resource).Scan(&available); errors.Is(err, pgx.ErrNoRows) || available {
		return store.ResourceReleaseState{}, errors.New("resource is not technician-held")
	} else if err != nil {
		return store.ResourceReleaseState{}, err
	}
	requiredJSON, _ := json.Marshal(required)
	failingJSON, _ := json.Marshal(failing)
	state, event := "pending_checks", "verification_requested"
	var released interface{}
	if override {
		state, event, released = "degraded_override", "degraded_override", at.UTC()
	}
	_, err = tx.Exec(ctx, `INSERT INTO public.resource_release_state(resource_name,state,required_streams,requested_at,requested_by,manifest_version,override_reason,failing_streams,released_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(resource_name) DO UPDATE SET state=EXCLUDED.state,required_streams=EXCLUDED.required_streams,
		requested_at=EXCLUDED.requested_at,requested_by=EXCLUDED.requested_by,manifest_version=EXCLUDED.manifest_version,override_reason=EXCLUDED.override_reason,
		failing_streams=EXCLUDED.failing_streams,released_at=EXCLUDED.released_at,updated_at=clock_timestamp()`, resource, state, string(requiredJSON), at.UTC(), actor, manifestVersion, reason, string(failingJSON), released)
	if err != nil {
		return store.ResourceReleaseState{}, err
	}
	if override {
		_, err = tx.Exec(ctx, `UPDATE public.resource_availability SET available=true,reason=$2,updated_at=clock_timestamp(),updated_at_ns=$3 WHERE resource_name=$1`, resource, "Degraded override: "+reason, at.UnixNano())
		if err != nil {
			return store.ResourceReleaseState{}, err
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO public.resource_release_events(resource_name,event_type,occurred_at,actor,reason,required_streams,failing_streams,manifest_version) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, resource, event, at.UTC(), actor, reason, string(requiredJSON), string(failingJSON), manifestVersion)
	if err != nil {
		return store.ResourceReleaseState{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return store.ResourceReleaseState{}, err
	}
	return store.ResourceReleaseState{Resource: resource, State: state, RequiredStreams: required, FailingStreams: failing, RequestedAt: at.UTC(), RequestedBy: actor, ManifestVersion: manifestVersion, OverrideReason: reason}, nil
}

func (r *Repository) ListResourceReleaseStates(ctx context.Context) ([]store.ResourceReleaseState, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	rows, err := r.pool.Query(ctx, `SELECT resource_name,state,required_streams,failing_streams,requested_at,requested_by,manifest_version,override_reason,released_at FROM public.resource_release_state ORDER BY requested_at,resource_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []store.ResourceReleaseState
	for rows.Next() {
		var value store.ResourceReleaseState
		var required, failing []byte
		if err := rows.Scan(&value.Resource, &value.State, &required, &failing, &value.RequestedAt, &value.RequestedBy, &value.ManifestVersion, &value.OverrideReason, &value.ReleasedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(required, &value.RequiredStreams); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(failing, &value.FailingStreams); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}
