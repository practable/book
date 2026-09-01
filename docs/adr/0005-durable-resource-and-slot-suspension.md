# ADR 0005: Durable resource and slot suspension

- Status: accepted
- Date: 2026-09-01

## Context

The original availability endpoints update only an in-memory resource diary.
Consequently a suspension is lost on restart and is invisible to another book
server connected to the same PostgreSQL database.  The legacy slot endpoint
also resolves a slot to its resource and unexpectedly suspends every slot that
uses that resource.

## Decision

PostgreSQL is the authority for explicit operator suspensions.  Resource and
slot suspensions are stored independently, keyed by the manifest names, with
their availability, reason and UTC update time.  Absence of a row means
available.  Resource suspension takes precedence over a slot suspension.

The existing resource endpoint remains resource-wide.  The existing slot
endpoint remains for API compatibility but now controls only its named slot.
Slot suspension is useful where two booking experiences share hardware but one
is unsafe, under maintenance, or temporarily withdrawn without taking the
other out of service.

Creation and take-up recheck both status rows inside their booking transaction,
after taking the normal maintenance lock.  This prevents a stale process from
creating or starting a booking after another server suspends the target.  Store
projections reload durable state before public operations and during the normal
five-second reconciliation loop, so read views converge across instances.

Suspension is administrative state, not a health observation.  It deliberately
does not cancel pending bookings or terminate an activity already started.
Health observations, preflight policy, and automatic suspension remain deferred
under ADR 0003.

## Migration and rollback

Migration 0005 adds `resource_availability` and `slot_availability`; both are
small, independent tables.  Rolling it back drops those tables, which loses
only explicit suspension state.  Operators should therefore clear or record
suspensions before a downgrade.  Migrations remain transactional and versioned.

## Compatibility risks

The slot endpoint changes from an accidental resource-wide effect to its
documented slot-scoped meaning.  Clients relying on the old effect must use the
resource endpoint explicitly.  Manifest names remain the key until the future
global resource authority introduces stable resource identifiers.
