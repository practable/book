# ADR 0002: Durable manifests and shared-resource authority boundaries

- Status: accepted
- Date: 2026-08-31

## Context

ADR 0001 made PostgreSQL authoritative for bookings but deliberately left the
manifest and resource availability state inside each `book` process. That is
safe only while one process is running with one operator-supplied manifest.
With several replicas, one instance can admit a booking using policy data that
another instance has already replaced. A restart also still depends on an
operator or external process loading the correct manifest before bookings can be
reconstructed.

The service is expected eventually to support equipment shared between
organisations. An organisation may receive a block of resource time and later
subdivide it between its users or delegate part of it. Independently
administered booking services may refer to the same physical resource. Giving
all such services direct write access to one database would make entitlement
enforcement a matter of trust, even if PostgreSQL continued to prevent interval
overlap.

These future capabilities are not part of the current implementation scope, but
the persistence boundary and identifiers chosen now must not make them needlessly
difficult to add.

## Decision

### Distinct identities and scopes

The design distinguishes the following identities:

- `service_id` identifies one logical booking service, its active manifest, and
  its administrative boundary. High-availability replicas of that service use
  the same `service_id`; it is not a process or host identifier.
- `resource_id` identifies one physical piece of equipment. It is the global
  collision and suspension boundary, even when several services or
  organisations refer to the resource using different local names.
- `organisation_id` will identify the party whose entitlement is exercised by a
  future booking or allocation. It must eventually come from authenticated
  service or user context and must never be trusted merely because a client sent
  it.

The current API continues to use manifest resource names. A service-local name
maps to a stable `resource_id`. Legacy, unshared resources may initially derive a
private stable ID from the service and resource name; sharing a resource between
services requires an explicit common ID.

### Durable, versioned manifests

PostgreSQL will hold immutable manifest versions scoped by `service_id`. A
version retains the original document, format, SHA-256 checksum, normalized
representation, optional Git revision or source label, timestamps, state, and
validation report. Exactly one version is active for a configured service.

Git remains the source of the proposed configuration. Administrators load a
file with the existing command-line workflow rather than an automatic webhook,
so validation and booking conflicts remain visible. `book manifest check`,
`replace`, and `export` remain the primary interface; replacement gains a remote
preview and safe activation workflow.

On an empty installation, the server exposes administrative configuration
operations but no booking operations until a manifest is activated. On every
subsequent start it loads the active database manifest before reconstructing
bookings. A manifest is parsed and prepared completely before it replaces the
in-memory immutable snapshot.

### Activation and multi-instance consistency

Manifest activation is serialized in PostgreSQL. Activation takes an exclusive
transaction-scoped lock; booking creation takes the corresponding shared lock.
After obtaining the exclusive lock, activation rereads authoritative booking
state and replays every active booking against the candidate manifest and
historical usage. By default, any incompatible active booking rejects the whole
activation and leaves the prior manifest and bookings unchanged.

Every manifest-dependent write carries the manifest version used for its domain
checks. PostgreSQL rejects the write if that version is no longer active. Thus a
replica using stale configuration cannot commit a booking even if it missed a
change notification.

Activation publishes a PostgreSQL notification containing only the service and
new version identifiers. Each replica uses a dedicated listener connection,
then reloads the committed version from PostgreSQL. Notification is an
optimization rather than the correctness mechanism: startup, listener
reconnection, and periodic polling also compare versions. A replica that cannot
load the active version becomes unready and refuses manifest-dependent writes
instead of continuing with stale policy.

### Administrative creation pause

A durable, service-scoped creation pause may be enabled while an administrator
reviews and retries a manifest change. It blocks new bookings across all
replicas but continues to allow reads, cancellation, and take-up of otherwise
valid existing bookings. It is separate from both the short transaction lock and
the existing broad process-local maintenance lock. The pause records a reason,
time, and optional expiry so an abandoned administrative session need not leave
creation disabled indefinitely.

### Incompatible bookings and replacement batches

A preview identifies incompatible bookings and binds the reviewed plan to the
candidate checksum, active manifest version, booking-state revision, and exact
booking identifiers. Applying a stale plan fails and requires another preview;
newly appeared bookings are never silently added to a destructive action.

An explicit `activate-and-supersede-invalid` operation may transition the
reviewed, unstarted bookings and activate the candidate in one transaction.
Superseded bookings are not physically deleted: they cease reserving resources,
do not contribute usage, remain available to administrative audit, and record
the replacement batch and manifest version. Started bookings require explicit
resolution because relay revocation is external to the database transaction.

Large timetable imports are represented as identifiable batches with source
checksums. Replacing a batch validates the complete new set, supersedes the old
unstarted set, installs the replacement, and records audit events atomically.
Repeating the same batch checksum is idempotent.

### Resource status and the take-up race

Temporary equipment failure is operational resource state, not a reason to
remove a resource from a manifest. Resource availability is stored durably by
`resource_id` and checked transactionally during both booking creation and
take-up. The existing slot-oriented administrative operation remains compatible
by resolving the slot to its resource; a resource-oriented operation will make
the actual scope explicit and report every affected slot.

Starting a booking and suspending its resource lock the same database resource
state. If suspension wins, take-up fails. If take-up wins, suspension observes
the newly started booking and records relay-revocation work. Relay effects cannot
be committed atomically with PostgreSQL, so durable outbox jobs, retries, and a
visible `revocation pending` state are required. A future resource-level relay
suspension is preferable to one denial per booking.

### Hierarchical resource claims

Physical occupancy is represented separately from booking metadata. A resource
claim contains a stable claim ID, resource ID, interval, lifecycle state,
contention scope, and optional parent claim. PostgreSQL exclusion constraints
prevent overlapping live claims with the same resource and contention scope.

Current direct bookings create top-level claims that contend on the physical
resource. A future organisation block also creates a top-level claim. User
bookings inside that block create child claims that contend with their siblings,
not with their parent. Transactions additionally require children to use the
same resource, remain within the parent interval, and reference an active parent.
The same representation can support further delegation without weakening global
physical exclusivity.

### Trust boundary and future Resource Authority

Direct PostgreSQL access is permitted only for replicas of one trusted logical
service. It is not the federation interface for independently administered
services.

Before shared equipment is offered across independent trust domains, global
resource operations will move behind an authenticated Resource Authority. It is
one logical service but may have several replicas. Only authority replicas write
the claim database. Organisation-facing `book` services authenticate as service
identities and request idempotent claims through a narrow API.

The authority enforces resource identity, ownership, service-to-organisation
authority, organisation-to-resource grants, allocation containment, suspension,
global conflicts, and audit. Local services retain descriptions, UI, slots, and
local user policies, which may restrict but cannot enlarge the authority's
grants. A local manifest's reference to a resource never grants access by
itself.

The initial PostgreSQL claim implementation may remain embedded behind a narrow
`ClaimAuthority` interface for the current single-trust-domain deployment. That
interface, rather than direct table assumptions in domain code, is the migration
seam to a network authority. Services sharing physical equipment across separate
databases would require such a common authority because local exclusion
constraints cannot provide global protection.

## Alternatives considered

- **Reload manifests independently from Git on every replica.** Rejected because
  replicas can observe different revisions and restart correctness depends on an
  external system.
- **Allow every organisation's service to write the shared database.** Rejected
  because overlap constraints do not enforce ownership, grants, or delegation
  and a faulty or hostile service could alter another organisation's state.
- **Maintain one global user-facing manifest.** Rejected because unrelated
  administrators would have to coordinate every local policy or presentation
  change. The global authority needs a resource catalogue and grants, not one
  combined UI manifest.
- **Treat organisation as the exclusion boundary.** Rejected because bookings
  from different organisations still conflict when they use the same physical
  resource.
- **Hard-delete erroneous timetable bookings.** Rejected for committed data
  because deletion loses audit, retry, and rollback evidence. Supersession gives
  the same active-state result without losing provenance.
- **Implement organisation accounts and block-booking APIs immediately.**
  Rejected because their authentication and delegation model is not yet defined.
  Only stable identifiers and the resource-claim seam are committed now.

## Migration and implementation sequence

1. Introduce stable service and resource identities and move physical overlap
   enforcement behind resource claims while preserving the current HTTP API.
2. Add versioned manifests, the active pointer, creation-pause state, activation
   validation, and manifest-version fencing.
3. Add listener/reconnect/poll synchronization and extend the CLI with preview,
   explicit supersession, source metadata, and actionable reports.
4. Persist resource status and enforce it during creation and take-up, retaining
   the existing slot command as a compatibility alias.
5. Add organisation grants, allocations, and the authenticated Resource
   Authority only when federation enters scope.

Each schema change is a forward versioned migration. Existing records are placed
under a configured default `service_id`; existing resource names receive stable
private IDs; and existing live bookings receive equivalent top-level claims in
the same migration transaction. Rollback must not return to code that ignores
claims or manifest versions while writers remain enabled.

## Consequences and compatibility

The current public booking API need not change. Existing manifest commands and
the slot availability command remain valid, although new administrative commands
and explicit destructive flags will be added. A production restart no longer
requires Git or an operator to reload the last active manifest.

The design adds identifiers, version checks, cache synchronization, and an
eventual network dependency for federation. It also makes the security boundary
explicit: PostgreSQL constraints protect trusted replicas today, while future
independent services receive only authenticated authority operations. No
organisation accounts, institutional allocation behavior, calendar UI, or
federation deployment is introduced by this decision alone.

## Related decisions

- [ADR 0001: PostgreSQL as the authoritative booking store](0001-postgresql-booking-persistence.md)

This decision supersedes ADR 0001 only where it describes manifests and resource
availability as permanently process-local. ADR 0001 remains authoritative for
the implemented booking persistence and compatibility behavior.
