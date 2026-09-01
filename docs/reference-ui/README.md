# Calendar and operations reference UI

This dependency-free browser client exercises the new calendar and administrator
APIs without changing `bookjs`. It is an integration reference, not a deployed
production frontend.

Serve this directory through the same origin as `book`, or configure a local
development reverse proxy for `/api/v1`. A different-origin service must
explicitly allow that origin; the booking server does not currently enable
general browser CORS.

The connection form accepts the existing bearer token format. Tokens and the
opaque student identifier are retained only in browser local storage. Do not use
a production administrator token in an untrusted browser or shared profile.

The student view exercises catalogue, manifest images, aggregate availability,
preview and idempotent creation. The administrator view exercises operational
status, resource metadata and suspension, booking pause/message, bounded booking
search, booking audit, override cancellation, technician holds, stream health,
alert acknowledgement, and manifest-approved manual health checks. Relay
restart intentionally remains outside this client.

Run the adapter tests with:

```sh
npm test
```

For a local preview:

```sh
python3 -m http.server 4177
```
