# PostgreSQL backing: readiness report

Status: implementation complete on `database-backing`; not deployed.

## Outcome

PostgreSQL is authoritative for bookings, manifest versions, user-group grants,
maintenance state and message, resource/slot holds, operational work,
activation/cleanup, health and alerts, usage accounting, relay revocations, and
verified/degraded technician release. Bookings and manifests recover after a
new process opens the database. PostgreSQL transactions, manifest fencing, and
half-open exclusion constraints protect concurrent instances from overlapping
physical-resource reservations.

Legacy booking APIs remain available. Additive calendar, administrator,
maintenance, health, job-runner, and activation APIs support the reference
student and technician interfaces. Domain policy remains in `internal/store`;
durable mutations cross narrow repository interfaces in
`internal/store/persistence.go` and `internal/store/admin.go`.
Student activation prepares all manifest-bound streams in one durable run.
Technician release similarly queues every required stream check atomically;
the dashboard is an observer and a browser restart cannot strand the workflow.

## Changed files

The branch contains generated server/client changes for the expanded
`api/booking.yml`; those files are intentionally generated and should be
reviewed as contract output. The hand-maintained production changes are grouped
as follows:

- PostgreSQL: `internal/postgres/postgres.go`, `operations.go`, `activation.go`,
  `alerts.go`, `releases.go`, `usage.go`, and migrations `0001` through `0027`
  (each with a down migration).
- Storage/domain: `internal/store/store.go`, `persistence.go`, `admin.go`,
  `calendar.go`, `operational.go`, and `recurrence.go`.
- HTTP/service: `internal/server/server.go`; `internal/serve/serve.go`,
  `adminhandlers.go`, `usershandlers.go`, `booking_activation.go`,
  `operational_alerts.go`, `operational_manifest.go`, `operations_callback.go`,
  and `operations_runner.go`.
- Authentication and delivery: `internal/webhook/*`, `internal/operations/*`,
  `internal/config/config.go`, and `cmd/book/cmd/serve.go`.
- Supporting behavior: `internal/interval/*`, `internal/diary/diary.go`,
  `internal/deny/*`, and individual-booking commands under `cmd/book/cmd`.
- Contract/UI: `api/booking.yml`, generated `internal/client/**` and
  `internal/serve/{models,restapi}/**`, and `docs/reference-ui/**`.
- Decisions and operations: `docs/adr/0001` through `0008`,
  `docs/postgresql.md`, `docs/activation-pipelines.md`, and the other
  operational/calendar/dashboard documents under `docs/`.
- Tests: colocated unit tests plus
  `internal/postgres/postgres_integration_test.go` and
  `internal/server/postgres_http_integration_test.go`.

Use `git diff --name-only main...database-backing` for the exhaustive generated
file list.

## Verification evidence

On 2026-09-01, against disposable PostgreSQL on `127.0.0.1:5432`:

- `go test ./...` passed with PostgreSQL integration enabled.
- `go vet ./...` passed.
- `node --test docs/reference-ui/book-api.test.js` passed (8 tests).
- `go test -race ./internal/postgres ./internal/server ./internal/store ./internal/serve`
  passed.
- `TestMigrationsFromEmptyDatabase` passed through migration `0027` in a newly
  created temporary database.
- An external production-sized manifest passed CLI validation, database
  activation, and recovery through a newly opened repository with resource,
  slot, policy, group, and stream inventories preserved. Its path and contents
  are not stored in this repository.
- Tests cover restart recovery, simultaneous overlap rejection, cancellation
  and rebooking, exact retry/idempotency behavior, historical usage and policy
  limits, transaction rollback, manifest fencing, half-open adjacency,
  maintenance access, arbitrary-stream experiment activation, atomically
  queued multi-stream verified release, degraded override, and PostgreSQL-backed
  HTTP disclosure.

Commands and disposable-database warnings are in `docs/postgresql.md`.

## Unresolved risks and deferred work

- This is a large branch (including generated API code) and should be reviewed
  and rolled out in stages, not merged and deployed as an unobserved one-step
  change.
- Existing manifests without `stream_operations` remain compatible. Health
  checks, activation preparation, and verified release remain dormant until
  experiment owners add reviewed workflow bindings. Book deliberately refuses
  to invent executable health checks.
- The external job runner and its approved task implementations are separate
  deployment work. Book persists, signs, retries, and audits jobs but does not
  execute hardware commands itself.
- PostgreSQL is now on the booking write path. Existing relay sessions have
  their separately documented availability behavior, but new Book mutations
  cannot safely proceed while the authoritative database is unavailable.
- Multi-organisation allocation and a central cross-service Resource Authority
  remain future boundaries. Shared trusted Book instances are protected by the
  same database constraints today, but this is not an untrusted multi-tenant
  authorization system.
- The reference UI demonstrates the APIs; it is not yet the final branded,
  accessible production student/technician application.
- Video safety-audit access is documented as deferred and needs a privacy,
  proportionality, retention, and regulatory decision before implementation.
- Down migration `0010` cannot restore closed intervals while adjacent
  reservations exist; this is a safe refusal and requires moving/cancelling
  those reservations before rollback.

## Recommended deployment sequence

1. Review the ADRs, API diff, migrations, and production configuration. Confirm
   backups and rehearse restore on a copy of production data.
2. Provision PostgreSQL with TLS, backups, monitoring, connection limits, and
   `btree_gist`. Use a migration role to install schema, then a least-privilege
   runtime role. Do not reuse the Docker test password.
3. Run all migrations once before starting new Book instances. Verify schema
   version `0027` and run the documented smoke/restart checks.
4. Validate the current production manifest with the new binary. Activate it
   while booking creation is paused, inspect replay findings, then resume. No
   operational bindings are required for the initial database migration.
5. Deploy one Book instance first, verify booking creation/cancellation,
   restart recovery, resource holds, welcome message, audit events, and relay
   revocation delivery; then add further instances and run a live concurrency
   smoke test.
6. Configure separate direction-bound 32-byte HMAC secrets for the job-runner
   command and callback directions, scoped dashboard tokens, rotation
   procedure, alerting, and outbox backlog monitoring.
7. Add activation/health bindings to a small pilot equipment class only after
   its job-runner tasks are approved. Exercise healthy failure, automatic
   recovery, technician hold, verified release, and explicit degraded release
   before wider adoption.
8. Keep the previous application artifact and a tested database restore plan.
   Prefer forward fixes after writes begin; use down migrations only after
   checking their documented data preconditions.
