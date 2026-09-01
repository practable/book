export class BookApiError extends Error {
  constructor(status, message, payload) {
    super(message || `Booking API returned ${status}`);
    this.name = "BookApiError";
    this.status = status;
    this.payload = payload;
  }
}

export class BookApi {
  constructor({ baseUrl = "", token = "", fetchImpl = globalThis.fetch } = {}) {
    this.baseUrl = baseUrl.replace(/\/$/, "");
    this.token = token;
    this.fetchImpl = fetchImpl;
  }

  async request(path, { method = "GET", body, headers = {} } = {}) {
    const response = await this.fetchImpl(`${this.baseUrl}/api/v1${path}`, {
      method,
      headers: {
        Accept: "application/json",
        ...(body === undefined ? {} : { "Content-Type": "application/json" }),
        ...(this.token ? { Authorization: this.token } : {}),
        ...headers,
      },
      ...(body === undefined ? {} : { body: JSON.stringify(body) }),
    });
    const text = await response.text();
    let payload;
    if (text) {
      try { payload = JSON.parse(text); } catch { payload = text; }
    }
    if (!response.ok) {
      throw new BookApiError(response.status, payload?.message || response.statusText, payload);
    }
    return payload;
  }

  calendarCatalogue(group) {
    return this.request(`/calendar/catalog/${encodeURIComponent(group)}`);
  }

  calendarAvailability(request) {
    return this.request("/calendar/availability", { method: "POST", body: request });
  }

  previewBooking(request) {
    return this.request("/calendar/preview", { method: "POST", body: request });
  }

  createBooking(request, idempotencyKey = crypto.randomUUID()) {
    return this.request("/calendar/bookings", {
      method: "POST", body: request, headers: { "Idempotency-Key": idempotencyKey },
    });
  }

  beginBookingActivation(user, booking, stream, idempotencyKey = crypto.randomUUID()) {
    return this.request(`/users/${encodeURIComponent(user)}/bookings/${encodeURIComponent(booking)}/activations`, {
      method: "POST", body: { stream }, headers: { "Idempotency-Key": idempotencyKey },
    });
  }

  beginExperimentActivation(user, booking, idempotencyKey = crypto.randomUUID()) {
    return this.request(`/users/${encodeURIComponent(user)}/bookings/${encodeURIComponent(booking)}/activations`, {
      method: "POST", body: {}, headers: { "Idempotency-Key": idempotencyKey },
    });
  }

  bookingActivation(user, booking, activation) {
    return this.request(`/users/${encodeURIComponent(user)}/bookings/${encodeURIComponent(booking)}/activations/${encodeURIComponent(activation)}`);
  }

  rescheduleBooking(name, request) {
    return this.request(`/calendar/bookings/${encodeURIComponent(name)}`, { method: "PATCH", body: request });
  }

  operationalStatus() { return this.request("/admin/operations"); }
  resources() { return this.request("/admin/resources"); }
  operationalHealth() { return this.request("/admin/operational-health"); }
  beginOperationalHealthCheck(resource, stream, idempotencyKey = crypto.randomUUID()) {
    return this.request(`/admin/resources/${encodeURIComponent(resource)}/streams/${encodeURIComponent(stream)}/health-checks`, {
      method: "POST", headers: { "Idempotency-Key": idempotencyKey },
    });
  }
  resourceHolds() { return this.request("/admin/resource-holds"); }
  resourceReleases() { return this.request("/admin/resource-releases"); }
  requestResourceRelease(resource, overrideReason = "") {
    const query = overrideReason ? `?${new URLSearchParams({ override_reason: overrideReason })}` : "";
    return this.request(`/admin/resource-holds/${encodeURIComponent(resource)}/release${query}`, { method: "POST" });
  }
  operationalAlerts(status = "active") {
    return this.request(`/admin/operational-alerts?${new URLSearchParams({ status })}`);
  }
  acknowledgeOperationalAlert(id) {
    return this.request(`/admin/operational-alerts/${encodeURIComponent(id)}/acknowledge`, { method: "POST" });
  }
  resolveOperationalAlert(id) {
    return this.request(`/admin/operational-alerts/${encodeURIComponent(id)}/resolve`, { method: "POST" });
  }

  bookingRecords(filters = {}) {
    const query = new URLSearchParams(Object.entries(filters).filter(([, value]) => value !== "" && value != null));
    return this.request(`/admin/booking-records?${query}`);
  }

  bookingEvents(name) { return this.request(`/admin/bookings/${encodeURIComponent(name)}/events`); }
  usage(filters = {}) {
    const query = new URLSearchParams(Object.entries(filters).filter(([, value]) => value !== "" && value != null));
    return this.request(`/admin/usage?${query}`);
  }

  setMaintenance(paused, message) {
    const query = new URLSearchParams({ lock: String(paused), msg: message });
    return this.request(`/admin/status?${query}`, { method: "PUT" });
  }

  setResourceAvailable(resource, available, reason) {
    const query = new URLSearchParams({ available: String(available), reason });
    return this.request(`/admin/resources/${encodeURIComponent(resource)}?${query}`, { method: "PUT" });
  }

  cancelWithOverride(name, reason) {
    const query = new URLSearchParams({ reason });
    return this.request(`/admin/booking-overrides/${encodeURIComponent(name)}/cancel?${query}`, { method: "POST" });
  }

  createResourceMaintenance(resource, from, to) {
    const query = new URLSearchParams({ from, to });
    return this.request(`/admin/resources/${encodeURIComponent(resource)}/maintenance-bookings?${query}`, { method: "POST" });
  }
}

export function loadConnection(prefix) {
  return {
    baseUrl: localStorage.getItem(`${prefix}.baseUrl`) || location.origin,
    token: localStorage.getItem(`${prefix}.token`) || "",
  };
}

export function saveConnection(prefix, value) {
  localStorage.setItem(`${prefix}.baseUrl`, value.baseUrl);
  localStorage.setItem(`${prefix}.token`, value.token);
}
