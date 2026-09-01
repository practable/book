# Operational job webhooks

Book persists operational jobs and outbound messages in PostgreSQL before any
HTTP request is attempted. Each running book instance may poll the same outbox;
row leases ensure only one instance owns a delivery attempt. A runner must still
deduplicate stable delivery and job IDs because delivery is at least once.

The feature is disabled unless both variables are set:

```sh
export BOOK_JOB_RUNNER_URL=https://job-runner.example/commands
export BOOK_WEBHOOK_SECRET='<secret from deployment secret management>'
```

Generate the shared secret as 32 random bytes, represented as either 64 hex
characters or unpadded base64url. For example, `openssl rand -hex 32` prints a
suitable value. A UUID is not accepted. Configure the same value in book and
the job runner; do not store it in a manifest, source repository, image, or log.

Optional settings are:

```sh
export BOOK_WEBHOOK_TOLERANCE=5m
export BOOK_WEBHOOK_POLL_EVERY=1s
```

Production runner URLs must use HTTPS. Plain HTTP is accepted only for loopback
development addresses.

## Authentication contract

Requests carry `X-Book-Timestamp`, `X-Book-Delivery-ID`, `X-Book-Direction`,
and `X-Book-Signature`. The signature is `v1=` followed by a hex HMAC-SHA256
over `direction.timestamp.delivery-id.` and then the exact raw HTTP body.
Directions are `book-to-runner` and `runner-to-book`. Both use the same secret,
but direction binding prevents a callback from being replayed as a command.
Receivers compare signatures in constant time and reject stale timestamps.

The runner reports lifecycle changes to:

```text
POST /api/v1/operations/jobs/{job_id}/callbacks
```

with direction `runner-to-book` and a JSON body such as:

```json
{"version":1,"job_id":"job-id-from-command","state":"running","at":"2026-09-01T12:00:00Z"}
```

Allowed reported states are `accepted`, `running`, `succeeded`, `failed`,
`cancelled`, and `expired`. Repeating the same delivery ID with the exact same
body is idempotent. Reusing it with different content returns 409, as does an
invalid lifecycle transition.

Failed health checks should provide a stable `code` separately from their
human-readable `error`, for example:

```json
{"version":1,"job_id":"job-id-from-command","state":"failed","at":"2026-09-01T12:00:03Z","code":"video_not_ready","error":"The camera has not produced a frame"}
```

Activation retry policies match `code`; changing explanatory prose therefore
does not change retry behaviour. For compatibility, Book falls back to the
`error` value when an older runner omits `code`.

After reporting `accepted`, the runner activates the job's reserved maintenance
booking with:

```text
POST /api/v1/operations/jobs/{job_id}/activate
```

using the same `runner-to-book` HMAC headers and a body such as:

```json
{"version":1,"job_id":"job-id-from-command","at":"2026-09-01T12:00:00Z"}
```

Book verifies the job-to-booking relationship and permitted lifecycle state in
the same transaction that starts the reservation and moves the job to
`running`. Activation is allowed only during the reservation's half-open time
interval. An exact retry returns the same activity without changing its
original start time. A booking identifier by itself grants no runner access.

Preflight jobs are different: the runner calls the activation endpoint to gain
scoped equipment access, but that call does not mark the user's booking started
or begin user usage accounting. Book advances the persisted preparation
pipeline and starts the user's booking only after its final stage succeeds.
Overdue preflight stages are claimed with database row locking by any available
Book instance. They are retried when `activation_timeout` is configured as
retryable, or are failed with persisted user guidance. This recovery also runs
immediately after a Book restart.

Teardown jobs implement the persisted cleanup plan. They use the same signed
runner protocol and close their maintenance reservation on a terminal callback.
Cleanup is triggered durably by booking cancellation or expiry, and by a
terminal preparation failure when cleanup work was resolved for that run.

Terminal callbacks (`succeeded`, `failed`, `cancelled`, or `expired`) also close
the bound reservation transactionally. Book records actual occupied time from
activation until the callback (capped at the planned end). If the reservation
was active, a durable relay revocation is inserted before the callback is
acknowledged, so access is withdrawn even when a different Book instance or
relay observes the result. Successful work is recorded as completed history;
other terminal outcomes retain their failure/cancellation state and audit
actor.

## Usage and cost accounting

User equipment usage remains the interval from successful booking activation
to cancellation or booking end. Preparation and cleanup never inflate that
figure or the policy usage limits derived from it.

Each activation, cleanup, or per-booking guard attempt also writes an
operational usage ledger entry. It records the exact triggering booking
revision, opaque user identifier, preparation/cleanup phase, planned duration,
actual runner-active duration, and outcome. Operational work is chargeable to
that user by default, while
remaining separately reportable as `preparation_usage`, `cleanup_usage`, and
`operational_jobs` in the administrator usage summary. The payer fields are
explicit so a future policy can charge an organisation or service without
rewriting historical rows.

## Failure and recovery

Booking commits do not wait for the runner. Network and non-2xx failures are
recorded and retried with bounded exponential backoff. A delivery eventually
becomes `dead` for operator inspection; its job and error history remain
durable. Restarting book or moving dispatch to another instance does not lose
pending work.

Migration 0009 creates `operational_jobs`, `webhook_deliveries`, and
`webhook_callback_receipts`; migration 0011 adds the completed booking audit
event; migrations 0015 and 0016 add durable booking activation runs, stages,
deadlines, retry policy, and failure guidance. Migrations 0017 and 0018 add the
independent cleanup plan and lifecycle; migration 0019 adds operational usage
accounting; migration 0020 extends that accounting to per-booking setup and
teardown guards; migration 0021 adds experiment-owner-funded scheduled work.
Apply migrations using the normal process in
[PostgreSQL](postgresql.md). Rollback is safe only after dispatch is disabled
and retained operational history has been exported or deliberately discarded.
