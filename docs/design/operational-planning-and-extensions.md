# Operational planning, schedules, and booking extensions

Status: discussion record; implementation deferred.

The booking domain should distinguish user bookings, exclusive operational
reservations, and durable task intents. Resources may have configurable state
transitions such as startup, ready, settling, teardown, unavailable, and
faulted. Operational reservations may be mandatory (fridge settling),
conditional (out-of-hours tank filling), or deferrable/supersedable (tank
emptying when another booking follows).

Schedules describe desired resource-state intervals rather than creating fake
bookings. Weekly patterns are effective-dated so semester schedules can apply
between named start and end dates, in an explicit timezone, with dated
exceptions and administrative overrides. Overlapping schedules are combined by
defined priority and compatible ready intervals may be unioned.

Planning a creation, cancellation, edit, shortening, or extension recomputes
the affected resource plan in one database transaction. A shared resource-claim
model should enforce exclusion for user bookings, setup, teardown, settling,
and maintenance. Task intents have stable idempotency keys, plan revisions,
dispatch deadlines, lifecycle/audit states, and expected input/output resource
states. A task accepted by the job runner is the boundary at which its
operational cost becomes committed. Late deliveries for superseded plan
revisions must be rejected.

Ordinary users may extend pending or active bookings. The revised total remains
subject to policy maximum duration and total usage; additional usage is debited
atomically and repeated extensions are allowed within those limits. Policies
may define a request window, minimum/increment/maximum extension sizes,
organisation-allocation enforcement, and whether an extension may cross into
out-of-hours operation. The planner calculates the maximum end from the next
booking, settling and preparation requirements, schedules, suspensions,
maintenance, usage, policy limits, organisation allocation, and already
committed tasks. Preview results include all limiting factors and any operational
cost requiring explicit acceptance.

Bookings may be shortened, including after operational work is committed, but
committed setup/teardown costs remain chargeable. An active booking cannot be
shortened into the past. Undispatched work may be removed or superseded; running
physical work is immutable history and the remaining plan is recomputed from
the latest desired state.

Setup failure normally prevents activation and faults or suspends the resource.
Teardown failure does not invalidate completed use but may block later use and
must alert operators. Maintenance bypasses are capabilities attached to
individual rules, not one universal exemption, because some guards may be
safety-critical.
