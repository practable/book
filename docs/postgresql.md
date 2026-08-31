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
GOCACHE=/tmp/book-go-cache go test -count=1 -v ./internal/postgres
```

The integration suite truncates booking tables. Point it only at a disposable
database. Run the remaining checks with:

```sh
GOCACHE=/tmp/book-go-cache go test ./internal/store ./internal/check ./internal/diary ./internal/filter ./internal/interval
GOCACHE=/tmp/book-go-cache go test -race ./internal/store ./internal/postgres
GOCACHE=/tmp/book-go-cache go vet ./...
```

Some existing API tests open loopback listeners. In a restricted sandbox they
must be run outside the network sandbox; on a normal development host use
`GOCACHE=/tmp/book-go-cache go test ./internal/server ./pkg/...`.

## Migrations

Numbered migrations are embedded from `internal/postgres/migrations`. Startup
opens one dedicated connection, takes a migration advisory lock, verifies the
SHA-256 checksum of every applied version, and applies each pending version in a
transaction. Starting the service against an empty database creates the entire
schema. A changed checksum is a startup error: add a new migration instead of
editing one already applied.

## Backup, recovery, and rollback

Use PostgreSQL backups and point-in-time recovery appropriate to the deployment.
To recover a service, restore the database first, verify `schema_migrations`, then
start one application instance and confirm booking counts before restoring normal
capacity. The application reconstructs current/historical bookings, resource
diaries, usage, group grants, and grace checks from PostgreSQL.

Never roll back to a pre-database binary while writers are enabled: it cannot see
durable bookings and can double-book resources. Roll back only to a tested
database-aware binary compatible with the applied schema. The supplied `.down.sql`
file is documentation for an exceptional, offline destructive rollback; take and
verify a backup before using it.

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
