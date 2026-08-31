-- Destructive rollback is intentionally manual. Stop every writer and take a
-- verified backup before executing this file.
DROP TABLE IF EXISTS user_groups;
DROP TABLE IF EXISTS booking_events;
ALTER TABLE IF EXISTS bookings DROP CONSTRAINT IF EXISTS bookings_no_resource_overlap;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS schema_migrations;
