UPDATE public.operational_schedule_occurrences SET state='skipped',detail='missed before rollback: ' || detail WHERE state='missed';
ALTER TABLE public.operational_schedule_occurrences DROP CONSTRAINT operational_schedule_occurrences_state_check;
ALTER TABLE public.operational_schedule_occurrences ADD CONSTRAINT operational_schedule_occurrences_state_check
    CHECK (state IN ('planned','skipped','conflict'));
