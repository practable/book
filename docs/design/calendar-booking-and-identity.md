# Calendar booking and recoverable ownership

Status: calendar catalogue, availability, preview, idempotent creation, and optimistic move/resize APIs implemented; anonymous recovery remains deferred.

## Calendar interaction

The proposed booking UI is time-centric rather than slot-centric. An experiment
class supplies a recommended duration, but users may select any policy-permitted
start time on a continuous calendar, initially snapping to a manifest-defined
increment. The proposed interval can be moved or resized, including across day
boundaries. Day, week, and month views are projections of the same interval
model rather than different booking semantics.

Availability queries may select any resource in a class, one named resource, or
resources matching structured manifest properties. Results must expose an
aggregate suitable-resource count without revealing another user's identity.
The booking preview resolves the precise interval, policy allowance, usage
charge, candidate resources, and operational effects before creation. Final
creation remains transactional and idempotent; a preview is advisory and cannot
reserve capacity.

Working hours are presentation and operational context, not an implicit ban.
Experiments such as spinners may be booked at any time. Other experiments may
require setup, teardown, settling, or additional usage outside configured
working hours. Separately, policies and resource state may declare hard
unavailable windows. Availability responses therefore distinguish normal,
out-of-hours-but-bookable, and prohibited intervals and include safe reason
codes and previewable operational effects.

## Anonymous recovery

Anonymous booking remains supported. A browser-held identity or booking ID is
not sufficient cancellation authority. The service may issue a high-entropy
management capability and a human-enterable recovery code, storing only a
verifier. A user may optionally request a short-lived emailed receipt containing
the management link and recovery code, and may download an iCalendar event. The
calendar event contains the non-secret booking reference but not management
authority because calendars are commonly shared.

An email address is a transient delivery address rather than an account. If
email delivery is implemented, the address is retained encrypted only while a
bounded, idempotent delivery attempt is pending. Durable booking audit records
retain delivery outcome, not the address.

## Account migration boundary

The current PostgreSQL and calendar work must not assume that browser storage is
a permanent owner identity. A later account feature will use external OpenID
Connect rather than local passwords. The durable identity key is an internal
random `account_id` associated with the provider's immutable `(issuer, subject)`
pair; email is not an identity key. Bookings may then receive nullable account
ownership, while existing anonymous management capabilities continue to work.

The intended migration is:

1. add account and provider-subject records in a versioned migration;
2. add nullable account ownership to bookings without changing anonymous API
   behaviour;
3. attach new bookings to the authenticated account when present;
4. allow a valid anonymous recovery capability to claim its bookings into an
   account; and
5. keep ordinary identity separate from administrator and maintenance
   authorization.

No unused account column, login endpoint, password handling, OIDC provider, or
student profile data is added during the current database-backing goal. Adding
the foreign key together with the account table later is safer than committing
an unconstrained placeholder column now, and does not require rewriting booking
identity or interval semantics.

## Expected API capabilities

The eventual calendar client requires narrowly scoped operations to:

- describe experiment classes, images, recommended durations, structured
  properties, working hours, and permitted booking increments;
- query aggregate availability for a time range and selection predicate;
- explain why a precise interval is limited or prohibited;
- preview usage, policy, assignment, setup, teardown, and settling effects;
- create a booking with an idempotency key;
- move, shorten, or extend a booking with optimistic concurrency; and
- recover or claim anonymous bookings using a management capability.

Calendar discovery, creation, and owner-authorized move/resize operations are
now exposed under `/calendar`. The existing API and `bookjs` compatibility
behaviour remain unchanged. Anonymous recovery remains deliberately deferred
until its capability lifecycle and delivery mechanism are designed.
