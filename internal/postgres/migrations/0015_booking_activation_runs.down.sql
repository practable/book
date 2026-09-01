ALTER TABLE operational_jobs
    DROP CONSTRAINT operational_jobs_activation_pair,
    DROP CONSTRAINT operational_jobs_activation_stage,
    DROP COLUMN activation_stage_index,
    DROP COLUMN activation_run_id;

DROP TABLE booking_activation_stages;
DROP TABLE booking_activation_runs;
