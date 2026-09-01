ALTER TABLE public.operational_schedule_occurrences
    DROP COLUMN IF EXISTS workflow_name,
    DROP COLUMN IF EXISTS resource_name,
    DROP COLUMN IF EXISTS slot_name;
