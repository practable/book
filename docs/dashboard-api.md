# Dashboard API

The OpenAPI definition in `api/booking.yml` is authoritative. The additions are
optional, additive endpoints; existing slot-oriented clients continue to work.

## Student calendar

- `GET /calendar/catalog/{group_name}` describes policy-backed experiment tiles,
  recommended durations, images, candidate resources, and structured resource
  properties.
- `POST /calendar/availability` samples the count of matching resources for an
  exact requested duration. Queries are bounded to 62 days and one-minute or
  coarser resolution.
- `POST /calendar/preview` explains whether an exact interval is currently
  bookable, including safe policy reason codes and projected usage.
- `POST /calendar/bookings` assigns one matching resource transactionally. The
  required `Idempotency-Key` is scoped to the user; an exact retry returns the
  original booking and reuse for a different request is rejected.
- `PATCH /calendar/bookings/{booking_name}` moves or resizes an unstarted
  booking using its opaque revision. Ownership and all current policy, usage,
  window, and overlap rules are rechecked atomically.

Class and `properties` are optional fields on manifest resources. Policies
remain the fungible experiment-class boundary, so no duplicate catalogue must
be maintained. Times are accepted and returned as RFC 3339 instants and are
stored in UTC.

## Administrator operations

- `GET /admin/operations` returns the current manifest version, persistent
  booking lock/message, and effective resource and slot availability.
- `GET /admin/operational-occurrences` returns authoritative scheduled work.
  Optional `from`, `until`, `state`, and `limit` parameters support upcoming,
  conflict, and missed-work queues. The default range is the preceding day
  through the next seven days and callers may request at most 500 rows.
- `GET /admin/booking-records` searches live and historical bookings by
  resource, slot, policy, opaque user, state, and overlapping time range. The
  result is bounded by `limit` and includes actual usage.
- `GET /admin/bookings/{booking_name}/events` returns the durable PostgreSQL
  lifecycle audit trail.
- `GET /admin/usage` accepts the same identifying filters plus a reporting
  interval; actual usage is clipped to that interval.
- `POST /admin/resources/{resource_name}/maintenance-bookings` creates a
  resource-level maintenance reservation without requiring the dashboard to
  understand legacy slots. Suspension is bypassed, but overlap protection is
  not.

Resource suspension, slot suspension, booking creation pause/message, and
override cancellation remain the previously documented endpoints. Relay health
checks and restart commands belong to the relay/job-runner integration, not to
this booking API.
