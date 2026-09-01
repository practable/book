# PostgreSQL operations

PostgreSQL is mandatory for the production `book serve` command. Library users
and unit tests that construct `store.New()` directly retain the prior in-memory
mode, which is not safe for multiple instances and loses state on restart.

## Configuration

Set these environment variables in the service secret/configuration system:

```sh
export BOOK_DATABASE_URL='postgres://book:REDACTED@db.example/book?sslmode=require'
export BOOK_DATABASE_MAX_CONNECTIONS=10
export BOOK_DATABASE_OPERATION_TIMEOUT=10s
```

`BOOK_DATABASE_URL` is never logged. Do not place it in source control. The
database role needs normal DML/DDL access to its database and permission to
install `btree_gist` for the first migration. A database administrator can
install that extension in advance and remove extension-creation privilege from
the runtime role afterward.

## Local setup and tests

A disposable local instance can be started with Docker (choose your own local
password):

```sh
docker run --rm --name book-postgres \
  -e POSTGRES_PASSWORD=local-only-password \
  -p 5432:5432 -d postgres:16
export BOOK_TEST_DATABASE_URL='postgres://postgres:local-only-password@127.0.0.1:5432/postgres?sslmode=disable'
export BOOK_TEST_ADMIN_DATABASE_URL="$BOOK_TEST_DATABASE_URL"
GOCACHE=/tmp/book-go-cache go test -count=1 -v ./internal/postgres
```

The integration suite truncates booking tables. Point it only at a disposable
database. `BOOK_TEST_ADMIN_DATABASE_URL` must be a disposable administrative
connection: the empty-database migration test uses it to create and drop one
uniquely named temporary database. Run the remaining checks with:

```sh
GOCACHE=/tmp/book-go-cache go test ./internal/store ./internal/check ./internal/diary ./internal/filter ./internal/interval
GOCACHE=/tmp/book-go-cache go test -race ./internal/store ./internal/postgres
GOCACHE=/tmp/book-go-cache go vet ./...
```

Some existing API tests open loopback listeners. In a restricted sandbox they
must be run outside the network sandbox; on a normal development host use
`GOCACHE=/tmp/book-go-cache go test ./internal/server ./pkg/...`.

## Manifest activation and replicas

The existing administrative workflow remains the entry point:

```sh
book manifest check manifest.yaml
book manifest replace manifest.yaml
book manifest export
```

Replacement validates and prepares the candidate before opening a database
transaction. PostgreSQL then serializes activation against manifest-dependent
writes and replays authoritative bookings and historical usage through the
candidate policies. If any booking is incompatible, the transaction rolls back
and the previous active manifest remains unchanged. Uploading the exact same
manifest again is idempotent and does not create another version.

Each committed version is immutable and identified by a SHA-256 checksum;
`active_manifest` selects the singleton active version. Booking, cancellation,
take-up, import, and group mutations carry the version used for validation and
are rejected if activation won the race. Running instances poll the active
version every five seconds and reload committed state when another replica has
changed it. The database fence, rather than polling, provides correctness.

An empty database intentionally starts with no bookable configuration. Load the
first manifest using the same `book manifest replace` command. To exercise a
large local manifest in the destructive disposable-database integration suite:

```sh
export BOOK_TEST_MANIFEST_PATH=/absolute/path/to/manifest.yaml
GOCACHE=/tmp/book-go-cache go test -count=1 \
  -run TestExternalManifestPersistenceRoundTrip ./internal/postgres
```

## Migrations

Numbered migrations are embedded from `internal/postgres/migrations`. Startup
opens one dedicated connection, takes a migration advisory lock, verifies the
SHA-256 checksum of every applied version, and applies each pending version in a
transaction. Starting the service against an empty database creates the entire
schema. A changed checksum is a startup error: add a new migration instead of
editing one already applied.

## Resource and slot suspension

The existing admin endpoints use names from the active manifest.  Suspend a
physical resource when it is unsafe or a health check fails:

```text
PUT /admin/resources/{resource_name}?available=false&reason=failed+health+check
```

That state is durable and applies to every slot using the resource.  The legacy
slot endpoint remains available but is now deliberately narrower:

```text
PUT /admin/slots/{slot_name}?available=false&reason=UI+maintenance
```

It suspends only that named slot, which is appropriate for a broken UI,
configuration, or teaching offering while the physical resource remains usable.
Both changes are serialized against booking creation and take-up in PostgreSQL.
They do not cancel pending bookings or revoke an activity that has already
started.  Other book replicas reload the authoritative state before booking
operations and through their five-second reconciliation loop.

## Backup, recovery, and rollback

Use PostgreSQL backups and point-in-time recovery appropriate to the deployment.
To recover a service, restore the database first, verify `schema_migrations`, then
start one application instance and confirm booking counts before restoring normal
capacity. The application reconstructs the active manifest first, followed by
current/historical bookings, resource diaries, usage, group grants, and grace
checks from PostgreSQL. After the first manifest has been activated, a restart
does not require the manifest Git checkout or another upload.

Never roll back to a pre-database or pre-durable-manifest binary while writers
are enabled: it cannot see authoritative bookings or the active manifest and may
admit invalid work. Roll back only to a tested database-aware binary compatible
with the applied schema. Supplied `.down.sql` files are documentation for an
exceptional, offline destructive rollback; take and verify a backup before using
them. Rolling back migration 0003 removes stored manifest versions, so export
and retain the active manifest first.

## Idempotency and audit

The public API has no idempotency-key field and successful booking creation has
no response body. Therefore an exact active request (same pseudonymous user,
policy, slot, start, and end) is treated as a retry and returns success with the
original durable booking. A different overlap fails. Reusing an administrative
booking name with different live data fails except as part of the transactional
whole-set replacement operation.

`booking_events` records creation, take-up, cancellation, expiry, import, and
supersession in the same transaction as the state change. It stores only the
existing opaque user/booking values and adds no student account data.

## Correcting one booking

Use the individual administrative workflow when a timetable entry is wrong but
the remaining bookings must stay untouched:

```sh
book bookings export-one booking-id > booking-edit.yaml
# edit booking.name, booking.user, booking.slot, booking.policy, or booking.when
book bookings replace-one booking-edit.yaml
```

Keep `original_name` and `revision` unchanged. The command uses a stable YAML
shape with snake_case fields, avoiding the generated-client YAML naming problem
in the older whole-collection export. The server accepts only unstarted,
uncancelled replacements. It validates policy, windows, historical usage and
resource conflicts, then supersedes the old row and inserts the replacement in
one transaction. A retry of the same file is safe; an edit based on a stale
revision returns a conflict rather than overwriting another administrator's
work.
