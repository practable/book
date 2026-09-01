# ADR 0004: Individual administrative booking corrections

- Status: accepted
- Date: 2026-08-31

## Context

The original administrative correction workflow exports every current booking,
requires an operator to edit the whole collection, and replaces the collection
in one operation. This is unnecessarily broad for correcting one unstarted
booking and makes two unrelated administrative changes contend with each other.
PostgreSQL now provides transactions, immutable superseded rows, and stable row
identifiers that can support a narrower operation.

## Decision

Add an administrative editable-booking representation containing the original
booking name, an opaque revision, and the proposed booking. A GET operation
exports this envelope; a PUT operation accepts the same format so the file can
be edited and uploaded directly. Existing bulk endpoints remain compatible.

The revision is the durable booking row identifier and is used only as an
optimistic concurrency token. The replacement transaction:

1. takes the manifest maintenance lock and checks the active manifest version;
2. locks the affected user-policy and resource collision domains;
3. verifies that the named live booking still has the expected revision and has
   not started;
4. validates the replacement against the active manifest, policy limits,
   historical usage, and every other active booking;
5. marks the old row superseded, inserts the replacement, links the two rows,
   and writes audit events; and
6. commits all changes together or none of them.

The replacement may retain or change the booking name, user, slot, policy, and
interval. It must describe a new, unstarted, uncancelled booking. Started
bookings are never edited through this operation because equipment access may
already have escaped the database transaction.

An exact retry using the original revision and identical replacement is
idempotent and returns the already-created booking. Reusing the revision with a
different replacement fails as a conflict. A stale manifest, conflicting
booking, policy failure, missing booking, or started booking leaves the original
booking active.

## Consequences

Operators no longer need to replace unrelated bookings to make one correction,
and erroneous timetable entries remain auditable without hard deletion. The
revision is intentionally opaque: clients must preserve it but must not infer
ordering or identity semantics from it. Whole-batch replacement remains useful
for atomic timetable imports and is not reimplemented as a loop of individual
edits.
