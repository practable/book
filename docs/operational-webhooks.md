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

## Failure and recovery

Booking commits do not wait for the runner. Network and non-2xx failures are
recorded and retried with bounded exponential backoff. A delivery eventually
becomes `dead` for operator inspection; its job and error history remain
durable. Restarting book or moving dispatch to another instance does not lose
pending work.

Migration 0009 creates `operational_jobs`, `webhook_deliveries`, and
`webhook_callback_receipts`. Apply migrations using the normal process in
[PostgreSQL](postgresql.md). Rollback is safe only after dispatch is disabled
and retained operational history has been exported or deliberately discarded.
