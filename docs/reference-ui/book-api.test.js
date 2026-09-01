import test from "node:test";
import assert from "node:assert/strict";
import { BookApi, BookApiError } from "./book-api.js";

function response(status, payload) {
  return { ok: status >= 200 && status < 300, status, statusText: "failure", text: async () => payload === undefined ? "" : JSON.stringify(payload) };
}

test("calendar creation sends bearer and idempotency headers", async () => {
  let request;
  const api = new BookApi({ baseUrl: "https://example.test/book/", token: "Bearer token", fetchImpl: async (url, options) => {
    request = { url, options }; return response(200, { name: "booking" });
  }});
  const payload = { user_name: "user", selector: { policy: "spinner" }, when: {} };
  assert.deepEqual(await api.createBooking(payload, "retry-key"), { name: "booking" });
  assert.equal(request.url, "https://example.test/book/api/v1/calendar/bookings");
  assert.equal(request.options.headers.Authorization, "Bearer token");
  assert.equal(request.options.headers["Idempotency-Key"], "retry-key");
});

test("admin filters are encoded and API errors retain status", async () => {
  const calls = [];
  const api = new BookApi({ fetchImpl: async (url) => { calls.push(url); return response(409, { message: "conflict" }); }});
  await assert.rejects(() => api.bookingRecords({ resource: "spin 66", state: "current", user: "" }), error => error instanceof BookApiError && error.status === 409);
  assert.equal(calls[0], "/api/v1/admin/booking-records?resource=spin+66&state=current");
});

test("maintenance lock uses the legacy-compatible query names", async () => {
  let called;
  const api = new BookApi({ fetchImpl: async (url) => { called = url; return response(200, {}); }});
  await api.setMaintenance(true, "Manifest update");
  assert.equal(called, "/api/v1/admin/status?lock=true&msg=Manifest+update");
});
