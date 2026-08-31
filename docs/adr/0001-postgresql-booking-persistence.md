# ADR 0001: PostgreSQL as the authoritative booking store

- Status: accepted
- Date: 2026-08-31

## Context and current state

`internal/store.Store` currently owns all mutable state in Go maps guarded by one
process-local `sync.RWMutex`. Resource occupancy is duplicated in an in-memory
AVL-backed `diary.Diary`. Making, cancelling, taking up, grace-expiring, and
pruning bookings mutate only those structures. The mutex and diary prevent races
inside one process, but state is lost on restart and two server processes can
make overlapping bookings. `BOOK_PERSIST_DIR` is read by the CLI but persistence
is explicitly not implemented.

The public API does not accept a client booking identifier or idempotency key.
Successful booking creation returns HTTP 204. Existing intervals conflict when
they touch at an endpoint, and timestamps are exposed as RFC3339 values.

## Decision

PostgreSQL is authoritative when `BOOK_DATABASE_URL` is configured. The domain
store retains manifest data and in-memory projections needed by the established
policy and availability code, but it reads the durable booking projection before
booking-sensitive operations and writes every state transition through a narrow
`BookingRepository` interface. The default in-memory mode remains available to
library users and unit tests; the production `book serve` command requires a
database URL.

The database model contains:

- `bookings`, with the public booking identifier, user pseudonym, policy, slot,
  resolved resource, UTC start/end timestamps, state, started/cancelled metadata,
  usage charge, and whether the booking consumes a resource;
- `booking_events`, an append-only audit record for created, started, cancelled,
  expired, imported, and otherwise transitioned bookings, containing the booking
  pseudonyms but no student account or profile data;
- `user_groups`, for durable pseudonymous group grants needed to authorize a
  recovered user's future bookings; and
- `schema_migrations`, recording each successfully applied versioned migration.

Database timestamps use `timestamptz` for operations and audit readability, while
parallel epoch-nanosecond integers are authoritative for identity and interval
comparison. This avoids PostgreSQL's microsecond timestamp precision changing the
existing nanosecond and closed-endpoint behaviour. Callers and API conversion
continue to use Go `time.Time`. Durations are stored as integer nanoseconds.

## Transactions and concurrency

Creating a booking is one transaction. It takes transaction-scoped advisory locks
for the `(user, policy)` and resource keys, checks an exact active request first,
enforces maximum booking and usage limits against committed database state, then
inserts both the booking and audit event. An exclusion constraint on
`(resource_name WITH =, int8range(starts_ns, ends_ns, '[]') WITH &&)` for active,
resource-constrained rows is the final invariant preventing overlap across server
instances. Closed ranges intentionally retain the service's current rule that
bookings sharing an endpoint conflict. Constraint or serialization failures are
reported as the same booking failure class the API already exposes.

Because the API has no request key, an exact active request—same user, slot,
policy, start, and end—is defined as a retry and returns the original booking as
success. A supplied administrative booking identifier is immutable during normal
creation: repeating the same identifier and payload succeeds, while reusing it
for different data is an error. Internally versioned rows let transactional
administrative replacement reuse/edit a public name without losing audit history.
Cancellation and take-up are idempotent state transitions in the
repository, although the HTTP cancellation endpoint retains its historical 404
response contract. Expiry and bulk replacement are transactional and write audit
events in the same transaction. Failed event or state writes roll back together.

The in-process mutex still protects each projection, but correctness does not
depend on it. A process refreshes its projection from PostgreSQL around durable
operations. PostgreSQL constraints and locks are authoritative under contention.

## Migrations, recovery, and rollback

Numbered SQL migrations are embedded in the binary and applied in order under a
PostgreSQL advisory lock. Each version is applied in its own transaction and is
recorded only after success. Starting from an empty database therefore produces
the complete schema deterministically.

Application rollback means first stopping writers, then deploying an earlier
database-aware binary known to understand the applied schema. Rolling back to the
pre-database service is unsafe because it would start empty and could double-book.
Additive migrations remain in place and are tolerated. Destructive schema
rollback is deliberately not automatic: restore a tested database backup or run
a separately reviewed down migration after confirming no newer data would be
lost. Operational backups and point-in-time recovery remain PostgreSQL concerns.

On process start, and after a manifest is installed, current and historical
bookings and pseudonymous group grants are rebuilt from PostgreSQL. Resource
diaries and usage totals are derived from those rows, so no snapshot file is
required.

## Compatibility risks

- Exact duplicate creation changes from an overlap error to idempotent success;
  this is intentionally unobservable to clients because successful creation has
  no response body. Without a client idempotency key this is best-effort: a
  delayed retry after cancellation can create a new booking, while two intentional
  exact submissions collapse into one.
- The schema stores the resource resolved at booking time. A manifest replacement
  that remaps a slot with live bookings must be rejected or explicitly reconciled;
  silently changing the resource would weaken exclusivity.
- External relay denial happens before durable cancellation because it cannot be
  part of a database transaction. A relay failure leaves the booking active. A
  database failure after a successful denial also leaves it active and retryable,
  which is safer than freeing occupied time without a durable cancellation.
- A take-up request on another instance can race the external relay-denial phase
  of cancellation. Eliminating that cross-system race needs a future relay-aware
  cancelling protocol.
- Manifests remain process-local. Operators must coordinate manifest changes
  across instances; resolved resources protect existing bookings, but divergent
  policies can still produce inconsistent admission decisions.
- PostgreSQL is required for multi-instance safety. In-memory mode is documented
  as test/development compatibility only and provides the prior single-process
  guarantees.

## Consequences

The HTTP schema and generated clients remain unchanged. Production gains durable
recovery, auditable transitions, transactional policy accounting, and a database
constraint that prevents double booking. The cost is a PostgreSQL operational
dependency and explicit projection-refresh/error handling in the existing store.
