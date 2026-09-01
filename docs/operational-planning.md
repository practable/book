# Manifest-driven operational guards

Operational workflow definitions are owner-approved task contracts. They bound
duration and provide a stable name understood by the job runner; they do not
contain commands, scripts, secrets, or webhook URLs.

```yaml
operational_workflows:
  fill-tank:
    description: Fill and verify the ripple tank
    expected_duration: 10m
    maximum_duration: 15m
  drain-tank:
    description: Drain the ripple tank
    expected_duration: 10m
    maximum_duration: 15m
```

A resource can refer to an existing manifest window as its normal operating
window and declare guards around user bookings:

```yaml
resources:
  ripple-tank-01:
    # existing resource fields omitted
    operations:
      operating_window: semester-weekdays
      before_booking:
        - workflow: fill-tank
          duration: 10m
          applies: outside_operating_window
          reclaimable: true
      after_booking:
        - workflow: drain-tank
          duration: 10m
          applies: outside_operating_window
          reclaimable: true
```

`applies` is either `always` or `outside_operating_window`. An outside-window
rule applies when the complete user booking is not contained in the operating
window. A reclaimable before/after guard is omitted when an adjacent booking
already overlaps the proposed guard interval. This supports keeping a ripple
tank ready between nearby out-of-hours bookings. Fridge settling should use
`always` with `reclaimable: false`, so the settling interval remains mandatory.

Manifest validation rejects missing workflows, unknown operating windows,
non-positive durations, durations beyond the workflow maximum, and invalid
conditions. Existing manifests need no new fields and remain compatible.

When a profile contains multiple before or after rules, their manifest order is
their execution order. The planner places them as adjacent half-open intervals:
all before-work finishes at the booking start and all after-work begins at the
booking end.

The implementation validates and persists this vocabulary, computes
deterministic plans, and transactionally creates the user booking, operational
guard reservations, jobs, audit events, and delivery outbox records. Any
failure rolls back the entire plan. Reservations use half-open `[start,end)`
semantics in the AVL diary, filters, and PostgreSQL.

Reclaiming an undispatched guard is also transactional. A following booking may
supersede a `reclaimable` reservation while its job remains `scheduled` or
`reserved`; the associated pending delivery is cancelled and the new boundary
is replanned. Once dispatch has begun, the physical action is treated as
committed and the guard cannot be reclaimed merely to fit another booking.
Cancelling a triggering booking uses the same transaction: operational work
that is still `scheduled` or `reserved` is cancelled, its pending delivery is
cancelled, and its guard reservation is superseded. Work that has reached
`dispatched` or a later state is retained because the physical cost may already
have been incurred.

Editing an unstarted booking also replans atomically. Obsolete undispatched
guards are retired, boundary guards made reclaimable by the new interval are
superseded, and the replacement booking and new guard jobs are committed as one
unit. An edit that conflicts with leased or dispatched physical work is rejected
and leaves the previous booking and plan unchanged.

Each guard job also receives an operational-usage ledger entry in the same
transaction as the user booking, reservation, and webhook delivery. Setup is
reported as preparation and teardown or settling as cleanup. The ledger points
to the exact triggering booking revision and charges its opaque user by
default, but does not add guard time to the user's experiment-usage total or
policy allowance. Runner activation and terminal callbacks record the actual
operational duration. A guard cancelled before dispatch remains auditable with
zero actual duration.
