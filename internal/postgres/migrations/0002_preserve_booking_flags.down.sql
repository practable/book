ALTER TABLE public.bookings
    DROP COLUMN IF EXISTS cancelled_by_text,
    DROP COLUMN IF EXISTS cancelled,
    DROP COLUMN IF EXISTS started_at_text,
    DROP COLUMN IF EXISTS started;
