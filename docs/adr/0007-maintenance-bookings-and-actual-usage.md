# ADR 0007: maintenance bookings and actual usage

## Status

Accepted.

## Context

An unavailable resource must remain unavailable to ordinary bookings and
take-up, while a technician sometimes needs a time-bounded connection to
repair it. Policy usage (`usage_charged`) is an entitlement/accounting value,
not a report of actual equipment use.

## Decision

Add an additive `maintenance` booking kind. It bypasses only resource and slot
availability checks. It remains resource-constrained, and the existing
PostgreSQL exclusion constraint prevents overlap with all ordinary and
maintenance bookings. It does not use student policy limits or alter policy
usage. A maintenance booking can take up activity while its resource is
suspended.

Use permanent, narrowly scoped technician authority: `booking:maintenance`
creates and takes up the technician's maintenance bookings; the additionally
granted `booking:booking-override` permits cancellation of an eligible booking
when a non-empty reason is supplied. Neither scope permits manifest, policy,
lock, or availability administration. `booking:admin` remains the broad role.
The dashboard may present these controls together according to token scope.

Actual usage is derived from durable lifecycle timestamps: first activation to
cancellation or scheduled end, whichever is earlier. Live bookings accrue only
to the report time. This is restart-safe and cannot double count retries.

## Migration and rollback

Migration 0007 adds a default-false `maintenance` column. Its down migration
removes the column. Rollback must be performed only after taking a backup and
stopping binaries that rely on maintenance bookings.

## Compatibility risks

Existing endpoints retain their behaviour. Booking JSON gains only an optional
additive field. The existing external token issuer must issue the new scopes;
this service does not introduce technician accounts.
