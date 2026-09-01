ALTER TABLE resource_availability
    ADD COLUMN held_since timestamptz,
    ADD COLUMN held_by text NOT NULL DEFAULT '';

UPDATE resource_availability SET held_since = updated_at WHERE NOT available;
