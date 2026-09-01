# Recurring booking windows

Existing explicit `allowed` and `denied` intervals remain supported. A window
may additionally contain finite weekly rules. This is intended for semester and
other effective-dated policy availability, not for triggering equipment tasks.

```yaml
windows:
  semester-weekdays:
    recurring_allowed:
      - timezone: Europe/London
        start_date: "2026-09-14"
        end_date: "2026-12-18"
        weekdays: [monday, tuesday, wednesday, thursday, friday]
        start_time: "09:00"
        end_time: "17:00"
        exceptions: ["2026-10-26", "2026-10-27"]
    recurring_denied:
      - timezone: Europe/London
        start_date: "2026-09-14"
        end_date: "2026-12-18"
        weekdays: [wednesday]
        start_time: "12:00"
        end_time: "13:00"
```

Dates are interpreted in the named IANA timezone and `end_date` is inclusive.
Civil times therefore remain at 09:00 after a daylight-saving transition even
though their UTC representation changes. When `end_time` is equal to or earlier
than `start_time`, the interval ends on the next civil day.

Rules must be finite and no longer than ten years. Invalid timezones, dates,
clock values or weekday names reject the whole manifest update. Occurrences are
materialized as UTC intervals while the candidate manifest is validated, before
it becomes authoritative.

Recurring windows are hard booking policy. A resource that is bookable outside
normal working hours should retain an allowed booking window covering those
hours. Its working-hours context and any required setup/teardown workflow belong
to the operational planner rather than `recurring_denied`.
