# Operational activation pipelines

Activation pipelines describe owner-approved preparation and health-check work
without placing executable commands or webhook URLs in the booking manifest.
The job runner maps each `workflow` identifier to a pre-approved task.

A reusable job template fixes the workflow, timeout, retry behaviour, user
progress text, and safe failure guidance. A pipeline orders templates. A
resource binds one of its streams to a pipeline and supplies only parameters
explicitly allowed by those templates.

```yaml
operational_workflows:
  video-on:
    description: Switch on an experiment camera
    kind: action
    expected_duration: 1s
    maximum_duration: 5s
  video-health:
    description: Verify that a camera stream is usable
    kind: health_check
    expected_duration: 1s
    maximum_duration: 5s
  video-reset:
    description: Reset an experiment camera
    kind: action
    expected_duration: 1s
    maximum_duration: 5s

operational_job_templates:
  standard-video-on:
    workflow: video-on
    timeout: 4s
    allowed_overrides: {stream: string}
  standard-video-check:
    workflow: video-health
    timeout: 4s
    allowed_overrides: {stream: string}
    retry:
      attempts: 3
      initial_delay: 250ms
      backoff: 2
      maximum_delay: 1s
      total_timeout: 4s
      retryable_codes: [not_ready]
    progress_messages:
      initial: Starting the video check
      retry: The video is not ready yet; trying again
    failure_guidance:
      title: Video unavailable
      message: Please choose another experiment or contact the laboratory.
      actions:
        - type: choose_another
          label: Choose another experiment
  standard-video-reset:
    workflow: video-reset
    timeout: 4s
    allowed_overrides: {stream: string}

operational_pipeline_templates:
  standard-video-activation:
    stages:
      - {name: switch-on, job_template: standard-video-on, wait_after: 500ms}
      - {name: check-video, job_template: standard-video-check}
    recovery:
      - {name: reset-video, job_template: standard-video-reset, wait_after: 500ms}
    recovery_attempts: 1

resources:
  spinner-66:
    # Existing required resource fields are omitted here.
    streams: [spinner-66-video]
    properties: {video_stream: spinner-66-video}
    stream_operations:
      spinner-66-video:
        activation_pipeline: standard-video-activation
        parameters:
          stream: {from: resource.properties.video_stream}
```

Bindings use either a literal `value` or a `from` value of the form
`resource.properties.NAME`, never both. Import validation rejects unknown
templates, unapproved parameters, missing properties, invalid or conflicting
parameter types, unsafe failure-action URLs, and retry or timeout values outside
the workflow contract. Failure-action URLs, when present, must use HTTPS.

These fields are optional, so existing manifests remain valid. They define the
plan but do not execute it; booking activation state, attempts, deadlines, and
progress are persisted separately by the activation engine.

Pipelines may also contain a cleanup plan. Its resolved stages are persisted
with the activation run before work starts, so a later manifest change cannot
alter cleanup already committed for a booking. Cleanup begins after
cancellation or expiry, and also after a partially completed preparation fails.
Its state is reported independently as `not_required`, `pending`, `running`,
`succeeded`, or `failed`; cleanup failure never disguises the original
activation result.

Cleanup after a completed or cancelled user session receives a short-lived,
resource-constrained maintenance reservation. That prevents a new booking from
overlapping work that is physically returning the equipment to its safe state.
If preparation fails while the user's booking still owns the resource, cleanup
uses that existing reservation. Retry attempts reuse the same reservation.

When an activation check exhausts its retries, Book records a durable
retry policy, a pipeline with `recovery` stages first runs those owner-approved
action workflows, waits as configured, and repeats the failed health-check
stage. Recovery stages and the target check are persisted before execution;
callbacks, retries, process restarts, and multiple Book instances therefore
cannot accidentally run an unapproved command or skip the recheck. The bounded
`recovery_attempts` value is between one and five.

Only after recovery is absent, fails, or exhausts its attempts does Book record
an `unhealthy` decision for that resource and stream and open (or increment) a
deduplicated technician alert. The activation response exposes
`recovery_stages` so the booking page can keep the user informed, and retains
its stable failure code, user-safe guidance, and actions such as choosing
another experiment. A successful recheck restores automated health and resolves
the active alert.

Automated health is independent of technician availability controls. A resource
held offline by a technician remains unavailable even when checks pass. The
final activation transaction rechecks that manual state, so a suspension made
while preparation is running cannot accidentally admit the user.

Administrators can poll:

- `GET /api/v1/admin/operational-alerts` for open, acknowledged, resolved, or all alerts;
- `POST /api/v1/admin/operational-alerts/{id}/acknowledge`;
- `POST /api/v1/admin/operational-alerts/{id}/resolve`;
- `GET /api/v1/admin/operational-health` for the latest automated stream decisions;
- `GET /api/v1/admin/resource-holds` for technician-held resources, including reason, actor, and original hold time.

Releasing a technician hold continues to use the resource availability endpoint
and is always an explicit technician action. Detailed metrics and diagnostics
remain the responsibility of the job runner and monitoring system.
