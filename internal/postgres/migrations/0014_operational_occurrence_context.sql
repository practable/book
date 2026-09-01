ALTER TABLE public.operational_schedule_occurrences
    ADD COLUMN slot_name text NOT NULL DEFAULT '',
    ADD COLUMN resource_name text NOT NULL DEFAULT '',
    ADD COLUMN workflow_name text NOT NULL DEFAULT '';
