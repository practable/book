package postgres

import (
	"context"
	"time"

	"github.com/practable/book/internal/store"
)

func (r *Repository) GetOperationalUsageSummary(ctx context.Context, query store.UsageQuery) (time.Duration, time.Duration, time.Duration, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	var preparation, cleanup, scheduled int64
	var jobs int64
	err := r.pool.QueryRow(ctx, `SELECT
		COALESCE(sum(CASE WHEN l.phase='preparation' THEN
			LEAST(l.actual_duration_ns,GREATEST(0,EXTRACT(EPOCH FROM (LEAST(COALESCE(l.completed_at,l.started_at),COALESCE($6,l.completed_at,l.started_at))-GREATEST(l.started_at,COALESCE($5,l.started_at))))*1000000000)::bigint) ELSE 0 END),0),
		COALESCE(sum(CASE WHEN l.phase='cleanup' THEN
			LEAST(l.actual_duration_ns,GREATEST(0,EXTRACT(EPOCH FROM (LEAST(COALESCE(l.completed_at,l.started_at),COALESCE($6,l.completed_at,l.started_at))-GREATEST(l.started_at,COALESCE($5,l.started_at))))*1000000000)::bigint) ELSE 0 END),0),
		COALESCE(sum(CASE WHEN l.phase='scheduled' THEN
			LEAST(l.actual_duration_ns,GREATEST(0,EXTRACT(EPOCH FROM (LEAST(COALESCE(l.completed_at,l.started_at),COALESCE($6,l.completed_at,l.started_at))-GREATEST(l.started_at,COALESCE($5,l.started_at))))*1000000000)::bigint) ELSE 0 END),0),
		count(*)
		FROM public.operational_usage_ledger l
		JOIN public.operational_jobs j ON j.job_id=l.job_id
		JOIN public.bookings b ON b.row_id=l.triggering_booking_row_id
		WHERE l.actual_duration_ns IS NOT NULL AND ($1='' OR j.resource_name=$1) AND ($2='' OR b.slot_name=$2)
		AND ($3='' OR b.policy_name=$3) AND ($4='' OR l.user_name=$4)
		AND ($5::timestamptz IS NULL OR COALESCE(l.completed_at,l.started_at)>$5)
		AND ($6::timestamptz IS NULL OR l.started_at<$6)`, query.Resource, query.Slot, query.Policy, query.User, query.From, query.To).
		Scan(&preparation, &cleanup, &scheduled, &jobs)
	return time.Duration(preparation), time.Duration(cleanup), time.Duration(scheduled), jobs, err
}
