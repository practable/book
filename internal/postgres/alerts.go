package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/practable/book/internal/store"
)

const alertColumns = `alert_id,resource_name,stream_name,result_code,message,coalesce(job_id,''),manifest_version,status,
	occurrences,first_seen_at,last_seen_at,acknowledged_at,acknowledged_by,resolved_at,resolved_by`

const alertSelect = `SELECT alert_id,resource_name,stream_name,result_code,message,coalesce(job_id,''),manifest_version,status,
	occurrences,first_seen_at,last_seen_at,acknowledged_at,acknowledged_by,resolved_at,resolved_by FROM public.operational_alerts`

func scanAlert(row pgx.Row) (store.OperationalAlert, error) {
	var value store.OperationalAlert
	err := row.Scan(&value.ID, &value.Resource, &value.Stream, &value.Code, &value.Message, &value.JobID, &value.ManifestVersion,
		&value.Status, &value.Occurrences, &value.FirstSeen, &value.LastSeen, &value.AcknowledgedAt, &value.AcknowledgedBy,
		&value.ResolvedAt, &value.ResolvedBy)
	return value, err
}

func (r *Repository) ListOperationalAlerts(ctx context.Context, status string, limit int) ([]store.OperationalAlert, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "open" && status != "acknowledged" && status != "resolved" && status != "all" {
		return nil, errors.New("invalid alert status")
	}
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	rows, err := r.pool.Query(ctx, alertSelect+` WHERE ($1='all' OR ($1='active' AND status IN ('open','acknowledged')) OR status=$1)
		ORDER BY CASE status WHEN 'open' THEN 0 WHEN 'acknowledged' THEN 1 ELSE 2 END,last_seen_at DESC LIMIT $2`, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]store.OperationalAlert, 0)
	for rows.Next() {
		value, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *Repository) ListOperationalStreamHealth(ctx context.Context) ([]store.OperationalStreamHealth, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	rows, err := r.pool.Query(ctx, `SELECT resource_name,stream_name,status,result_code,message,coalesce(job_id,''),manifest_version,checked_at
		FROM public.operational_stream_health ORDER BY resource_name,stream_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]store.OperationalStreamHealth, 0)
	for rows.Next() {
		var value store.OperationalStreamHealth
		if err := rows.Scan(&value.Resource, &value.Stream, &value.Status, &value.Code, &value.Message, &value.JobID, &value.ManifestVersion, &value.CheckedAt); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *Repository) SetOperationalAlertStatus(ctx context.Context, id int64, status, actor string, at time.Time) (store.OperationalAlert, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	if status != "acknowledged" && status != "resolved" {
		return store.OperationalAlert{}, errors.New("invalid alert transition")
	}
	query := `UPDATE public.operational_alerts SET status='acknowledged',acknowledged_at=$2,acknowledged_by=$3,updated_at=$2
		WHERE alert_id=$1 AND status='open' RETURNING ` + alertColumns
	if status == "resolved" {
		query = `UPDATE public.operational_alerts SET status='resolved',resolved_at=$2,resolved_by=$3,updated_at=$2
		WHERE alert_id=$1 AND status IN ('open','acknowledged') RETURNING ` + alertColumns
	}
	value, err := scanAlert(r.pool.QueryRow(ctx, query, id, at.UTC(), actor))
	if errors.Is(err, pgx.ErrNoRows) {
		return store.OperationalAlert{}, store.ErrPersistentNotFound
	}
	return value, err
}

func (r *Repository) ListResourceHolds(ctx context.Context) ([]store.ResourceHold, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	rows, err := r.pool.Query(ctx, `SELECT resource_name,reason,held_since,held_by FROM public.resource_availability
		WHERE NOT available ORDER BY held_since,resource_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]store.ResourceHold, 0)
	for rows.Next() {
		var value store.ResourceHold
		if err := rows.Scan(&value.Resource, &value.Reason, &value.HeldSince, &value.HeldBy); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *Repository) SetResourceAvailabilityBy(ctx context.Context, resource string, available bool, reason, actor string, manifestVersion int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", maintenanceLock); err != nil {
		return err
	}
	if err := assertManifestVersion(ctx, tx, manifestVersion); err != nil {
		return err
	}
	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `INSERT INTO public.resource_availability(resource_name,available,reason,updated_at_ns,held_since,held_by)
		VALUES($1,$2,$3,$4,CASE WHEN $2 THEN NULL::timestamptz ELSE $5::timestamptz END,CASE WHEN $2 THEN '' ELSE $6::text END)
		ON CONFLICT (resource_name) DO UPDATE SET available=EXCLUDED.available,reason=EXCLUDED.reason,
		updated_at=clock_timestamp(),updated_at_ns=EXCLUDED.updated_at_ns,
		held_since=CASE WHEN EXCLUDED.available THEN NULL WHEN resource_availability.available THEN EXCLUDED.held_since ELSE resource_availability.held_since END,
		held_by=CASE WHEN EXCLUDED.available THEN '' WHEN resource_availability.available THEN EXCLUDED.held_by ELSE resource_availability.held_by END`,
		resource, available, reason, now.UnixNano(), now, actor)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
