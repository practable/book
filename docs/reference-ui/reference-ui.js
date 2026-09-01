import { BookApi, loadConnection, saveConnection } from "./book-api.js";

const $ = id => document.getElementById(id);
const escapeHTML = value => String(value ?? "").replace(/[&<>"']/g, character => ({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"})[character]);
const durationMs = value => ({"10m":600000,"20m":1200000,"30m":1800000,"1h":3600000})[value] || 600000;
const setStatus = (element, message, error = false) => { element.textContent = message; element.classList.toggle("error", error); };
const localDate = date => `${date.getFullYear()}-${String(date.getMonth()+1).padStart(2,"0")}-${String(date.getDate()).padStart(2,"0")}`;
const formatTime = value => new Intl.DateTimeFormat(undefined,{weekday:"short",hour:"2-digit",minute:"2-digit"}).format(new Date(value));

document.querySelectorAll("[data-view]").forEach(button => button.addEventListener("click", () => {
  document.querySelectorAll("[data-view]").forEach(candidate => candidate.classList.toggle("active", candidate === button));
  document.querySelectorAll(".panel").forEach(panel => panel.classList.toggle("active", panel.id === button.dataset.view));
}));

const studentConnection = loadConnection("book.student");
$("s-base").value = studentConnection.baseUrl;
$("s-token").value = studentConnection.token;
$("s-user").value = localStorage.getItem("book.student.user") || "";
$("s-group").value = localStorage.getItem("book.student.group") || "";
$("date").value = localDate(new Date());

let studentApi;
let catalogue = [];
let selectedItem;
let proposedRequest;
let currentBooking;
let activationPoll;

function studentSelector() {
  const resource = $("resource").value;
  return { policy: $("policy").value, ...(resource ? { resource } : {}) };
}

function configurePlanner(item) {
  selectedItem = item;
  $("policy").innerHTML = catalogue.map(entry => `<option value="${escapeHTML(entry.policy)}">${escapeHTML(entry.description?.short || entry.policy)}</option>`).join("");
  $("policy").value = item.policy;
  renderResourceChoices(item);
  if (item.recommended_duration) {
    const matching = [...$("duration").options].find(option => option.value === item.recommended_duration);
    if (matching) $("duration").value = matching.value;
  }
  $("planner").classList.remove("hidden");
  $("preview").classList.add("hidden");
  refreshAvailability();
}

function renderResourceChoices(item) {
  $("resource").innerHTML = `<option value="">Any ${escapeHTML(item.description?.short || item.policy)}</option>` +
    (item.resources || []).map(resource => `<option value="${escapeHTML(resource.name)}">Specific: ${escapeHTML(resource.name)}</option>`).join("");
}

$("policy").addEventListener("change", () => {
  selectedItem = catalogue.find(item => item.policy === $("policy").value);
  renderResourceChoices(selectedItem);
  refreshAvailability();
});

$("s-connect").addEventListener("click", async () => {
  const connection = { baseUrl: $("s-base").value.trim(), token: $("s-token").value.trim() };
  saveConnection("book.student", connection);
  localStorage.setItem("book.student.user", $("s-user").value.trim());
  localStorage.setItem("book.student.group", $("s-group").value.trim());
  studentApi = new BookApi(connection);
  setStatus($("s-status"), "Loading catalogue…");
  try {
    catalogue = await studentApi.calendarCatalogue($("s-group").value.trim());
    $("catalogue").innerHTML = catalogue.map((item,index) => { const degraded = (item.resources || []).filter(resource => resource.degraded).length; return `<button class="card tile" data-item="${index}">${item.description?.image || item.description?.thumb ? `<img src="${escapeHTML(item.description.image || item.description.thumb)}" alt="">` : ""}<div class="tile-body"><h2>${escapeHTML(item.description?.short || item.policy)}</h2><p class="muted">${escapeHTML(item.description?.long || item.description?.further || "Choose an available resource and time.")}</p><p>${item.resources?.length || 0} matching resources · suggested ${escapeHTML(item.recommended_duration || "policy duration")}</p>${degraded ? `<p class="warning">${degraded} currently offered with reduced capability</p>` : ""}</div></button>`; }).join("");
    document.querySelectorAll("[data-item]").forEach(tile => tile.addEventListener("click", () => configurePlanner(catalogue[Number(tile.dataset.item)])));
    setStatus($("s-status"), `${catalogue.length} experiment classes loaded.`);
  } catch (error) { setStatus($("s-status"), error.message, true); }
});

async function refreshAvailability() {
  if (!studentApi || !selectedItem) return;
  const day = new Date(`${$("date").value}T00:00:00`);
  const from = new Date(day); from.setHours(8);
  const to = new Date(day); to.setHours(22);
  const request = { user_name: $("s-user").value.trim(), selector: studentSelector(), from: from.toISOString(), to: to.toISOString(), duration: $("duration").value, resolution: "10m" };
  setStatus($("s-status"), "Checking availability…");
  try {
    const bands = await studentApi.calendarAvailability(request);
    $("bands").innerHTML = bands.map((band,index) => `<button class="band ${band.degraded_resources > 0 ? "degraded" : band.matching_resources > 2 ? "good" : band.matching_resources > 0 ? "limited" : "none"}" data-band="${index}" ${band.bookable ? "" : "disabled"}>${formatTime(band.start)}<br>${band.matching_resources} available${band.degraded_resources ? `<br>${band.degraded_resources} degraded` : ""}</button>`).join("");
    document.querySelectorAll("[data-band]").forEach(button => button.addEventListener("click", () => previewBand(bands[Number(button.dataset.band)])));
    setStatus($("s-status"), "Availability is current; final creation is rechecked transactionally.");
  } catch (error) { setStatus($("s-status"), error.message, true); }
}

async function previewBand(band) {
  const start = new Date(band.start);
  const end = new Date(start.getTime() + durationMs($("duration").value));
  proposedRequest = { user_name: $("s-user").value.trim(), selector: studentSelector(), when: { start: start.toISOString(), end: end.toISOString() } };
  try {
    const preview = await studentApi.previewBooking(proposedRequest);
    $("preview-title").textContent = `${formatTime(start)}–${new Intl.DateTimeFormat(undefined,{hour:"2-digit",minute:"2-digit"}).format(end)}`;
    const degraded = preview.degraded_resources || [];
    const warning = degraded.length ? ` · Reduced capability: ${degraded.map(resource => `${resource.name} (${(resource.unavailable_streams || []).join(", ") || resource.degraded_reason})`).join("; ")}` : "";
    $("preview-detail").textContent = preview.bookable ? `${preview.matching_resources.length} possible resources · usage after booking ${preview.usage_after}${warning}` : `Cannot book: ${(preview.reasons || []).join(", ")}`;
	$("preview-detail").classList.toggle("warning", degraded.length > 0);
    $("book").disabled = !preview.bookable;
    $("preview").classList.remove("hidden");
  } catch (error) { setStatus($("s-status"), error.message, true); }
}

$("availability").addEventListener("click", refreshAvailability);
$("resource").addEventListener("change", refreshAvailability);
$("date").addEventListener("change", refreshAvailability);
$("duration").addEventListener("change", refreshAvailability);
$("book").addEventListener("click", async () => {
  try {
    const booking = await studentApi.createBooking(proposedRequest);
    currentBooking = booking;
    setStatus($("s-status"), `Booked ${booking.name} on ${booking.slot}. Save this reference.`);
    $("preview").classList.add("hidden");
    const selectedResource = $("resource").value;
    const candidates = (selectedItem.resources || []).filter(resource => !selectedResource || resource.name === selectedResource);
    const streams = [...new Set(candidates.flatMap(resource => resource.activation_streams || []))].sort();
    $("activation-booking").textContent = `Booking ${booking.name}. Preparation is idempotent and may be retried safely.`;
    $("activation-start").disabled = streams.length === 0;
    setStatus($("activation-status"), streams.length ? `Start preparation for all ${streams.length} configured connection${streams.length === 1 ? "" : "s"}.` : "This experiment has no managed preparation pipeline; use the normal booking link when its time starts.");
    $("activation").classList.remove("hidden");
    await refreshAvailability();
  } catch (error) { setStatus($("s-status"), error.message, true); }
});

function showActivation(run) {
  const stages = [...(run.stages || []), ...(run.recovery_stages || []), ...(run.cleanup_stages || [])];
  const activeStage = stages.find(stage => ["pending", "dispatched", "accepted", "running"].includes(stage.state));
  const degraded = run.degraded ? ` Reduced capability: ${run.degraded_reason || "some connections are unavailable"}${run.unavailable_streams?.length ? ` (${run.unavailable_streams.join(", ")})` : ""}.` : "";
  const failure = run.failure_message ? ` ${run.failure_message}` : "";
  setStatus($("activation-status"), `${run.progress_message || activeStage?.progress_message || run.state}.${degraded}${failure}`, run.state === "failed" || run.state === "cleanup_failed");
  $("activation-status").classList.toggle("warning", Boolean(run.degraded) && run.state !== "failed");
}

async function pollActivation(run) {
  clearTimeout(activationPoll);
  showActivation(run);
  if (["active", "failed", "cancelled", "expired", "closed", "cleanup_failed"].includes(run.state)) return;
  activationPoll = setTimeout(async () => {
    try {
      pollActivation(await studentApi.bookingActivation($("s-user").value.trim(), currentBooking.name, run.id));
    } catch (error) { setStatus($("activation-status"), error.message, true); }
  }, 750);
}

$("activation-start").addEventListener("click", async () => {
  if (!studentApi || !currentBooking) return;
  $("activation-start").disabled = true;
  setStatus($("activation-status"), "Starting preparation…");
  try {
    const run = await studentApi.beginExperimentActivation($("s-user").value.trim(), currentBooking.name);
    pollActivation(run);
  } catch (error) {
    const message = error.status === 409 ? `This experiment cannot be started: ${error.message}` : error.message;
    setStatus($("activation-status"), message, true);
    $("activation-start").disabled = false;
  }
});

const adminConnection = loadConnection("book.admin");
$("a-base").value = adminConnection.baseUrl;
$("a-token").value = adminConnection.token;
let adminApi;
let operations;
let resourceMetadata = {};

async function loadOperations() {
  const [status, metadata, holds, health, alerts, releases] = await Promise.all([
    adminApi.operationalStatus(), adminApi.resources(), adminApi.resourceHolds(), adminApi.operationalHealth(), adminApi.operationalAlerts(), adminApi.resourceReleases(),
  ]);
  operations = status;
  resourceMetadata = metadata;
  const statuses = Object.entries(operations.resources || {});
  $("a-lock").textContent = operations.status.locked ? "Paused" : "Open";
  $("a-toggle").textContent = operations.status.locked ? "Resume" : "Pause";
  $("a-resources").textContent = `${statuses.filter(([,status]) => status.available).length}/${statuses.length}`;
  $("a-bookings").textContent = operations.status.bookings;
  renderResources(statuses);
  renderHolds(holds, releases);
  renderHealth(health);
  renderAlerts(alerts);
  setStatus($("a-status"), `Manifest v${operations.manifest_version}; refreshed ${new Date().toLocaleTimeString()}.`);
}

function renderHolds(holds, releases) {
  const releaseByResource = new Map(releases.map(item => [item.resource, item]));
  $("a-hold-rows").innerHTML = holds.length ? holds.map(hold => { const release = releaseByResource.get(hold.resource); return `<tr><td>${escapeHTML(hold.resource)}</td><td>${escapeHTML(new Date(hold.held_since).toLocaleString())}</td><td>${escapeHTML(hold.held_by)}</td><td>${escapeHTML(release?.state === "pending_checks" ? `Waiting for: ${release.required_streams.join(", ")}` : hold.reason)}</td><td><button class="button" data-release-hold="${escapeHTML(hold.resource)}">Check & release</button> <button class="button danger" data-override-release="${escapeHTML(hold.resource)}">Release degraded</button></td></tr>`; }).join("") : `<tr><td colspan="5" class="muted">No technician-held resources.</td></tr>`;
  document.querySelectorAll("[data-release-hold]").forEach(button => button.addEventListener("click", async () => {
    const resource = button.dataset.releaseHold;
    if (!confirm(`Run every required health check and release ${resource} only if all pass?`)) return;
    try { const release = await adminApi.requestResourceRelease(resource); setStatus($("a-status"), `Queued ${release.required_streams.length} required checks for ${resource}. It remains held until all pass.`); await loadOperations(); } catch (error) { setStatus($("a-status"), error.message, true); }
  }));
  document.querySelectorAll("[data-override-release]").forEach(button => button.addEventListener("click", async () => {
    const reason = prompt(`Reason for releasing ${button.dataset.overrideRelease} with failing checks:`, "Required service can continue without the failed capability");
    if (!reason?.trim()) return;
    try { await adminApi.requestResourceRelease(button.dataset.overrideRelease, reason.trim()); await loadOperations(); } catch (error) { setStatus($("a-status"), error.message, true); }
  }));
}

function renderHealth(health) {
  const known = new Map(health.map(item => [`${item.resource}\0${item.stream}`, item]));
  Object.entries(resourceMetadata).forEach(([resource, metadata]) => (metadata.streams || []).forEach(stream => {
    const key = `${resource}\0${stream}`;
    if (!known.has(key)) known.set(key, { resource, stream, status: "unknown" });
  }));
  const rows = [...known.values()];
  $("a-health-rows").innerHTML = rows.length ? rows.map(item => `<tr><td>${escapeHTML(item.resource)}</td><td>${escapeHTML(item.stream)}</td><td><span class="badge ${item.status === "healthy" ? "good" : "bad"}">${escapeHTML(item.status)}</span></td><td>${item.checked_at ? escapeHTML(new Date(item.checked_at).toLocaleString()) : "Never"}</td><td>${escapeHTML(item.code || item.message || "—")}</td><td><button class="button" data-run-check-resource="${escapeHTML(item.resource)}" data-run-check-stream="${escapeHTML(item.stream)}">Run check</button></td></tr>`).join("") : `<tr><td colspan="6" class="muted">No configured streams.</td></tr>`;
  document.querySelectorAll("[data-run-check-resource]").forEach(button => button.addEventListener("click", async () => {
    button.disabled = true;
    setStatus($("a-status"), `Reserving ${button.dataset.runCheckResource} and starting ${button.dataset.runCheckStream} check…`);
    try {
      const run = await adminApi.beginOperationalHealthCheck(button.dataset.runCheckResource, button.dataset.runCheckStream);
      setStatus($("a-status"), `Health check ${run.id} queued. Refreshing status…`);
      setTimeout(() => loadOperations().catch(error => setStatus($("a-status"), error.message, true)), 1500);
    } catch (error) { button.disabled = false; setStatus($("a-status"), error.message, true); }
  }));
}

function renderAlerts(alerts) {
  $("a-alert-rows").innerHTML = alerts.length ? alerts.map(alert => `<tr><td>${escapeHTML(alert.resource)} / ${escapeHTML(alert.stream)}</td><td>${escapeHTML(alert.status)}</td><td>${escapeHTML(new Date(alert.last_seen).toLocaleString())}</td><td>${escapeHTML(alert.occurrences)}</td><td>${escapeHTML(alert.message || alert.code)}</td><td>${alert.status === "open" ? `<button class="button" data-ack-alert="${escapeHTML(alert.id)}">Acknowledge</button>` : "Being investigated"}</td></tr>`).join("") : `<tr><td colspan="6" class="muted">No active operational alerts.</td></tr>`;
  document.querySelectorAll("[data-ack-alert]").forEach(button => button.addEventListener("click", async () => {
    try { await adminApi.acknowledgeOperationalAlert(button.dataset.ackAlert); await loadOperations(); } catch (error) { setStatus($("a-status"), error.message, true); }
  }));
}

function renderResources(entries = Object.entries(operations?.resources || {})) {
  const search = $("a-search").value.toLowerCase();
  const filter = $("a-filter").value;
  $("a-resource-rows").innerHTML = entries.filter(([name,status]) => (!search || name.toLowerCase().includes(search)) && (!filter || (filter === "available") === status.available)).map(([name,status]) => {
    const metadata = resourceMetadata[name] || {};
    const details = [metadata.class, ...Object.entries(metadata.properties || {}).map(([key,value]) => `${key}=${value}`)].filter(Boolean).join(" · ");
    return `<tr><td>${escapeHTML(name)}</td><td class="muted">${escapeHTML(details)}</td><td><span class="badge ${status.available ? "good" : "bad"}">${status.available ? "Available" : "Unavailable"}</span></td><td>${escapeHTML(status.reason)}</td><td><button class="button" data-resource-toggle="${escapeHTML(name)}">${status.available ? "Suspend" : "Restore"}</button></td></tr>`;
  }).join("");
  document.querySelectorAll("[data-resource-toggle]").forEach(button => button.addEventListener("click", async () => {
    const name = button.dataset.resourceToggle;
    const current = operations.resources[name];
    const reason = current.available ? prompt(`Reason for suspending ${name}:`, "Maintenance") : "Restored after maintenance";
    if (reason === null) return;
    try { await adminApi.setResourceAvailable(name, !current.available, reason); await loadOperations(); } catch (error) { setStatus($("a-status"), error.message, true); }
  }));
}

$("a-connect").addEventListener("click", async () => {
  const connection = { baseUrl: $("a-base").value.trim(), token: $("a-token").value.trim() };
  saveConnection("book.admin", connection); adminApi = new BookApi(connection);
  try { await loadOperations(); await findBookings(); } catch (error) { setStatus($("a-status"), error.message, true); }
});
$("a-search").addEventListener("input", () => renderResources());
$("a-filter").addEventListener("change", () => renderResources());
$("a-toggle").addEventListener("click", async () => {
  const paused = !operations.status.locked;
  const message = prompt(paused ? "Message shown while booking creation is paused:" : "Welcome message after resuming:", operations.status.message || "");
  if (message === null) return;
  try { await adminApi.setMaintenance(paused, message); await loadOperations(); } catch (error) { setStatus($("a-status"), error.message, true); }
});

async function findBookings() {
  const records = await adminApi.bookingRecords({ resource: $("a-booking-resource").value.trim(), state: $("a-booking-state").value, limit: 200 });
  $("a-booking-rows").innerHTML = records.map(record => `<tr><td>${escapeHTML(record.booking.name)}</td><td>${escapeHTML(record.resource)}</td><td>${formatTime(record.booking.when.start)}–${new Intl.DateTimeFormat(undefined,{hour:"2-digit",minute:"2-digit"}).format(new Date(record.booking.when.end))}</td><td>${escapeHTML(record.actual_usage)}</td><td><button class="button" data-audit="${escapeHTML(record.booking.name)}">Audit</button> <button class="button danger" data-cancel="${escapeHTML(record.booking.name)}">Cancel</button></td></tr>`).join("");
  document.querySelectorAll("[data-audit]").forEach(button => button.addEventListener("click", async () => {
    try { const events = await adminApi.bookingEvents(button.dataset.audit); alert(events.map(event => `${event.occurred_at}  ${event.type}  ${event.actor}`).join("\n")); } catch (error) { setStatus($("a-status"), error.message, true); }
  }));
  document.querySelectorAll("[data-cancel]").forEach(button => button.addEventListener("click", async () => {
    const reason = prompt(`Operational reason for cancelling ${button.dataset.cancel}:`, "Urgent equipment maintenance");
    if (!reason) return;
    try { await adminApi.cancelWithOverride(button.dataset.cancel, reason); await findBookings(); await loadOperations(); } catch (error) { setStatus($("a-status"), error.message, true); }
  }));
}
$("a-find").addEventListener("click", () => findBookings().catch(error => setStatus($("a-status"), error.message, true)));
