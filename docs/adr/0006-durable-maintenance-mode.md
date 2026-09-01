# ADR 0006: Durable maintenance mode and welcome message

- Status: accepted
- Date: 2026-09-01

## Decision

The singleton service state is stored in PostgreSQL: `booking_creation_paused`
and the welcome message. It is recovered on restart and refreshed by replicas.
Maintenance mode rejects only new user booking creation. Reads, cancellation,
and take-up remain available; resource or slot suspension is the separate,
durable control that prevents unsafe take-up. Updates serialize with the
maintenance advisory lock and are exposed through the existing admin status
endpoint for compatibility.
