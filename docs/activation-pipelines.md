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

operational_pipeline_templates:
  standard-video-activation:
    stages:
      - {name: switch-on, job_template: standard-video-on, wait_after: 500ms}
      - {name: check-video, job_template: standard-video-check}

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
