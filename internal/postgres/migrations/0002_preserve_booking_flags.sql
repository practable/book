ALTER TABLE public.bookings
    ADD COLUMN started boolean NOT NULL DEFAULT false,
    ADD COLUMN started_at_text text NOT NULL DEFAULT '',
    ADD COLUMN cancelled boolean NOT NULL DEFAULT false,
    ADD COLUMN cancelled_by_text text NOT NULL DEFAULT '';

UPDATE public.bookings
SET started = started_at IS NOT NULL,
    cancelled = cancelled_at IS NOT NULL,
    cancelled_by_text = cancelled_by;
