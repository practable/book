# 0008: Durable relay revocations with availability-preserving fallback

When a started booking is cancelled, the booking transaction records an
unexpired `relay_revocations` row. A relay configured with
`RELAY_REVOCATION_DATABASE_URL` uses a PostgreSQL role with **SELECT only** to
consult this table before issuing a new session code. This makes a cancelled
booking remain denied after a relay restart.

The existing authenticated relay deny request remains the immediate mechanism:
it closes currently connected sockets through the relay's in-memory deny list.
If the revocation database cannot be reached, the relay logs a warning and
continues using that in-memory list. It neither disconnects live users nor
rejects otherwise valid new sessions. Consequently, the explicitly accepted
residual risk is that a relay restart during a database outage loses prior
revocations until PostgreSQL is restored. This protects timetabled operation
from a database outage while retaining durable restart protection whenever the
database is available.

Migration 0008 is additive. A rollback drops only revocation records and must
not be performed while relay instances depend on durable restart protection.
