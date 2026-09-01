# ADR 0003: Recurring operations, automation bookings, and health jobs

- Status: accepted; recurrence, durable authenticated job transport, and booking guard creation implemented
- Date: 2026-08-31

## Context

Physical laboratory resources have regular operating patterns that are tedious
and error-prone to express as manually enumerated intervals. A ripple tank, for
example, may be bookable from 09:00 to 17:00 on weekdays, require filling before
opening, and require draining after closing. Other equipment may be prepared on
demand shortly before a booking and shut down after the last adjacent booking.

Resources also need routine quality-control jobs and on-demand checks before a
user is given relay access. A separate job-runner service is intended to accept
authenticated HTTP commands for pre-approved tasks, assign an agent, and let
that agent interact with equipment through the normal booking and relay path.

These requirements are deliberately deferred while the current implementation
focuses on persisting one active manifest in PostgreSQL. This record preserves
the intended boundaries so the smaller manifest change does not foreclose them.

## Proposed decision

### Recurrence is an interval language, not cron

Availability patterns describe recurring intervals with an IANA timezone,
effective dates, exclusions, and one-off overrides. A human-oriented weekly
form is preferred in manifests and may normalize internally to RFC 5545-style
recurrence rules. Existing explicit allowed and denied intervals remain valid.

The effective availability is the union of recurring and explicit allowed
intervals minus recurring and explicit denied intervals. Global resource
suspension takes precedence over every rule. Recurrences are evaluated in local
civil time and occurrences are stored as UTC instants.

Generated occurrences are materialized at least as far ahead as the greatest
booking horizon. Each has a deterministic identity derived from the schedule
version, resource, workflow, and occurrence time. Schedule changes cancel and
rebuild only pending occurrences after the same preview and compatibility checks
used for manifest changes.

### Physical operation is separate from bookable availability

A recurrence expresses when bookings may occur; it does not execute arbitrary
commands. Resource-owner configuration associates approved workflow names with
schedule or booking boundaries. Examples include preparation before opening,
shutdown after closing, preparation before a booking, and delayed shutdown after
the final adjacent booking.

The global resource authority owns these workflows because it can see claims
from every booking service sharing the equipment. A local service may narrow the
resource owner's availability but cannot extend it or nominate an unapproved
workflow.

### Durable jobs and transactional dispatch

Scheduled work is represented by durable database rows rather than process
timers. Jobs record their due time, resource, schedule and manifest versions,
workflow, triggering booking or occurrence, lifecycle state, lease, attempts,
and idempotency key. Authority replicas claim work with database row locking and
leases, retry failures, and reconcile desired state after restart.

Creating an automation booking, its physical resource claim, its audit event,
and the dispatch-outbox record is one PostgreSQL transaction. The command is
sent only after commit. HTTP delivery is at least once because a sender may
crash after delivery but before recording success; the job runner and agent must
therefore deduplicate the stable job identifier.

The job lifecycle distinguishes scheduled, reserved, dispatched, accepted,
running, succeeded, failed, cancelled, and expired. An HTTP success means only
that a command was accepted unless the workflow explicitly completes
synchronously. Completion is authenticated and reported separately.

### Automation bookings and the job runner

Every task that needs exclusive equipment access receives an automation booking
of an approved, bounded duration. It uses an opaque system principal and does
not consume a student's ordinary booking count or usage allowance. It remains
subject to physical overlap, suspension, allocation containment, owner operating
rules, and version fencing.

The authority sends the runner the approved workflow, job and booking IDs,
resource, interval, permitted parameters, and a short-lived capability. The
runner assigns an agent. The agent takes up the existing automation booking—it
does not create a second booking—and obtains the normal activity and relay
details before issuing equipment commands.

Task definitions specify expected and maximum durations. An occurrence may
request a custom duration only within those bounds. Early completion may release
unused capacity; retrying execution within the same interval does not create a
new booking.

### Health observations and activation checks

Daily quality-control jobs and booking preflight jobs report structured health
observations. An observation identifies the resource, test definition and
version, job, result time, expiry, outcome, measurements or diagnostics, and the
manifest/resource-contract version. Results contain no student account data.

Resource health is derived from current observations and explicit operator
state rather than a single unqualified boolean. At minimum it distinguishes
unknown, healthy, degraded, unhealthy, and administratively suspended states.
An observation expires so old success cannot certify equipment indefinitely.

Before take-up, policy may require a sufficiently recent successful observation
or an on-demand preflight. The booking already holds the physical resource while
the preflight agent runs, but user activity credentials are withheld until the
check succeeds. Failure leaves the user booking unstarted, records a clear
outcome, and normally suspends the resource for investigation. Refund, grace,
and retry behavior must be explicit before this path is implemented.

The current manifest's resource `tests` list is treated as a reference to
approved test definitions, not executable code or arbitrary URLs.

### Desired and observed resource state

The authority records desired operational state separately from observed state.
A schedule may require a resource to be ready while the observed state remains
preparing or faulted. Bookings and take-up are permitted only when the required
observed state is confirmed. Manual suspension overrides scheduled transitions.

Preparation or shutdown failure creates a durable failure state and operator
signal. Relay revocation and other external effects use durable outbox commands
and remain visibly pending until acknowledged.

### Attribution rather than immediate billing

The executor, initiator, beneficiary, and sponsor of an operational job are
separate. Normal published-hours preparation is ordinarily sponsored by the
resource operator; preparation for an exceptional out-of-hours booking or an
organisation allocation may be sponsored by that organisation. A preparation
cycle may benefit several bookings.

The authority records immutable attribution entries containing the operational
cycle, sponsor, cause, beneficiaries, quantity, policy version, and any later
adjustment or reversal. Quantities may initially be preparation time or resource
units rather than money. Pricing, invoices, accounts, and payment are out of
scope.

## Safety and security constraints

- Manifests reference owner-approved workflow identifiers, never shell commands
  or unrestricted webhook URLs.
- Runner endpoints and callbacks use authenticated service identities; secrets
  are referenced from secret management and are not stored in manifests.
- Job capabilities are scoped to one booking, resource, action, and validity
  interval.
- Duplicate delivery and completion reports are idempotent.
- Preparation failure fails closed: it cannot make a resource appear ready.
- A cancelled running job triggers agent cancellation and relay revocation.
- User-visible booking commits never wait on an outbound webhook transaction.

## Open questions before implementation

- Interval handover uses half-open `[start,end)` reservations, allowing
  preparation, user activity, and shutdown to share exact boundaries.
- Whether a preflight consumes part of the user's interval or requires a
  separately reserved lead interval.
- Failure, refund, substitution, and notification policy after a failed
  preflight.
- Rules for coalescing preparation across adjacent bookings and dividing its
  attribution.
- Whether any existing workflow requires atomic claims on several resources.
- Health-test result schemas, thresholds, validity periods, and manual override
  authority.

## Deferred implementation sequence

1. Establish durable resource claims and resource operational state.
2. Add recurrence evaluation and materialized occurrences.
3. Add durable scheduled jobs and a transactional outbox.
4. Implement one authenticated, idempotent HTTP command/callback transport.
5. Add automation bookings and job-runner capabilities.
6. Add structured health observations and preflight activation policy.
7. Add attribution records before introducing financial charging.

## Consequences

The future booking system can schedule routine quality checks, ensure equipment
is prepared before use, and preserve exclusive access while execution remains in
a dedicated job runner. The cost is an explicit resource state machine and
at-least-once distributed workflow rather than simple cron callbacks.

Finite weekly `recurring_allowed` and `recurring_denied` rules are supported on
existing manifest windows. Operational job persistence, leased outbox delivery,
direction-bound HMAC authentication, retries, and idempotent callbacks are also
implemented. Manifest-driven booking guards are created atomically with their
triggering booking, undispatched reclaimable guards can be superseded by a
following booking, cancellation retires undispatched guards atomically, and
unstarted booking edits atomically replace undispatched guard plans.
The HMAC-authenticated runner activation action transactionally starts only the
reservation bound to an accepted job and returns its activity configuration.
Terminal callbacks atomically close that reservation, record actual occupied
time, and durably revoke relay access for a started job.
Finite weekly operational schedules with explicit conflict recording are
implemented using the same reservation, job, and outbox lifecycle.
Operational-state schedules, flexible conflict windows, health policy,
charging, organisation accounts, and block booking remain deferred.

## Related decisions

- [ADR 0001: PostgreSQL as the authoritative booking store](0001-postgresql-booking-persistence.md)
- [ADR 0002: Durable manifests and shared-resource authority boundaries](0002-manifest-and-resource-authority.md)
