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

The current implementation validates and persists this vocabulary and computes
deterministic plans. Connecting plans to automation reservations is the next
stage. That stage must resolve the legacy closed-interval rule: bookings whose
end and start timestamps are equal currently conflict in both the AVL diary and
PostgreSQL. Operational guards naturally share an exact boundary with the user
booking, so changing to conventional half-open `[start,end)` reservations is
preferable to exposing artificial one-nanosecond gaps.
