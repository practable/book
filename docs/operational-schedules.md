# Operational schedules

Finite recurring operational schedules create the same durable maintenance
reservations, jobs, and signed webhook deliveries as per-booking guards. Times
are civil times in an explicit IANA time zone, so a 09:00 schedule remains at
09:00 across daylight-saving changes.

```yaml
operational_workflows:
  fill:
    description: Fill the ripple tank
    expected_duration: 10m
    maximum_duration: 15m

resources:
  ripple-tank-resource:
    # Existing resource fields omitted.
    operations:
      cost_owner: physics-teaching-lab

operational_schedules:
  weekday-fill:
    slot: ripple-tank-01
    workflow: fill
    duration: 10m
    conflict: require
    recurrence:
      timezone: Europe/London
      start_date: "2026-09-21"
      end_date: "2026-12-18"
      weekdays: [monday, tuesday, wednesday, thursday, friday]
      time: "08:50"
      exceptions: ["2026-10-26"]
```

`end_date` is inclusive. `exceptions` omit individual civil dates. The slot is
explicit because it determines the activity configuration supplied to the job
runner as well as the underlying resource.

`cost_owner` is an opaque identifier for the team or service responsible for
the experiment. A manifest that schedules work for a resource without a cost
owner is rejected. Completed scheduled runner time is reported separately as
`scheduled_usage`, attributed to `experiment_owner`; it is never charged to a
student or included in student usage-policy limits. The owner identifier is
accounting metadata and does not grant access to the resource.

Conflict modes are:

- `require`: record a durable `conflict` occurrence when the resource is busy.
- `skip`: record a durable `skipped` occurrence when the resource is busy.

The scheduler never silently moves work to a different time. A future flexible
window mode can do that without changing the meaning of existing manifests.

Each Book instance scans a rolling horizon, but PostgreSQL advisory locks and a
unique occurrence key ensure only one instance creates a particular booking,
job, and outbox delivery. Restarting or rescanning is idempotent. Occurrence
identity includes the active manifest version; an unchanged occurrence in a
new manifest cannot double-book because the resource constraint remains
authoritative and is recorded as a conflict instead.

Each scan also covers the preceding horizon. Occurrences whose complete window
passed during an outage are recorded as `missed`; Book never dispatches stale
physical work merely to catch up.

Configuration defaults are:

```sh
BOOK_OPERATIONAL_SCHEDULE_EVERY=1m
BOOK_OPERATIONAL_SCHEDULE_HORIZON=168h
```

Scheduling runs when the job-runner webhook integration is configured. Keep
the horizon comfortably longer than the longest expected Book outage. Migration
0012 stores occurrence decisions and their links to bookings and jobs; 0013
adds explicit missed-occurrence recovery records.

Administrators can inspect these decisions through
`GET /api/v1/admin/operational-occurrences`. Responses include schedule, UTC
occurrence time, manifest version, state, slot, resource, workflow, and—where
planned—the booking and job identifiers. Migration 0014 persists this display
context so skipped and conflicted occurrences remain explainable after a
manifest change.

Migration 0021 permits explicit experiment-owner attribution in the operational
usage ledger. The scheduled reservation, job, delivery, and owner-funded ledger
entry are committed atomically.
