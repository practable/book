ALTER TABLE public.booking_events DROP CONSTRAINT booking_events_event_type_check;
ALTER TABLE public.booking_events ADD CONSTRAINT booking_events_event_type_check CHECK (event_type IN
    ('created', 'started', 'cancelled', 'expired', 'imported', 'superseded', 'completed'));
