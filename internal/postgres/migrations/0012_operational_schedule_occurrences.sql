CREATE TABLE public.operational_schedule_occurrences (
    schedule_name text NOT NULL,
    occurrence_at timestamptz NOT NULL,
    occurrence_at_ns bigint NOT NULL,
    manifest_version bigint NOT NULL,
    state text NOT NULL CHECK (state IN ('planned','skipped','conflict')),
    booking_row_id bigint REFERENCES public.bookings(row_id),
    job_id text REFERENCES public.operational_jobs(job_id),
    detail text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (schedule_name, occurrence_at_ns, manifest_version)
);
CREATE INDEX operational_schedule_occurrences_state_time
    ON public.operational_schedule_occurrences(state, occurrence_at_ns);
