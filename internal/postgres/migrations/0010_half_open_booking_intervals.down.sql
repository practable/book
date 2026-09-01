ALTER TABLE public.bookings DROP CONSTRAINT bookings_no_resource_overlap;

-- This may intentionally fail if adjacent bookings were created after the up
-- migration. Resolve those bookings before attempting a semantic rollback.
ALTER TABLE public.bookings ADD CONSTRAINT bookings_no_resource_overlap
    EXCLUDE USING gist (
        resource_name WITH =,
        int8range(starts_ns, ends_ns, '[]') WITH &&
    ) WHERE (collection = 'live' AND NOT superseded AND resource_constrained);
