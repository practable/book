package convert

import (
	"os"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamljson "sigs.k8s.io/yaml"
)

// the p-blank policy tests for errors on default intervals, which should become 0s
var manifestYAML = []byte(`
display_guides:
  1mFor20m:
    book_ahead: 20m
    duration: 1m
    label: 1m
  dg-blank:
policies:
  p-modes:
    allow_start_in_past_within: 1m0s
    book_ahead: 2h0m0s
    max_duration: 10m0s
    min_duration: 5m0s
    max_usage: 30m0s
    next_available: 1m0s
    starts_within: 1m0s
  p-blank:
windows:
  w-a:
    allowed:
    - start: 2022-11-04T00:00:00Z
      end: 2022-11-06T00:00:00Z
    denied: []
`)

// Other testing has noted that the durations get missed without UnmarshalJSON functions being provided.
// We have that for the store manifest but the models manifest uses pointers and would need patching
// so we just convert. Here we test whether the durations are picked up correctly.
// This also tests JSONToManifests - test separately if implementation changes
func TestYAMLToManifests(t *testing.T) {

	m, s, err := YAMLToManifests(manifestYAML)

	assert.NoError(t, err)

	assert.Equal(t, time.Duration(20*time.Minute), s.DisplayGuides["1mFor20m"].BookAhead)
	assert.Equal(t, "20m", *(m.DisplayGuides["1mFor20m"].BookAhead))

	assert.Equal(t, time.Duration(1*time.Minute), s.DisplayGuides["1mFor20m"].Duration)
	assert.Equal(t, "1m", *(m.DisplayGuides["1mFor20m"].Duration))

	assert.Equal(t, time.Duration(2*time.Hour), s.Policies["p-modes"].BookAhead)
	assert.Equal(t, "2h0m0s", m.Policies["p-modes"].BookAhead)

	assert.Equal(t, time.Duration(10*time.Minute), s.Policies["p-modes"].MaxDuration)
	assert.Equal(t, "10m0s", m.Policies["p-modes"].MaxDuration)

	assert.Equal(t, time.Duration(5*time.Minute), s.Policies["p-modes"].MinDuration)
	assert.Equal(t, "5m0s", m.Policies["p-modes"].MinDuration)

	assert.Equal(t, time.Duration(30*time.Minute), s.Policies["p-modes"].MaxUsage)
	assert.Equal(t, "30m0s", m.Policies["p-modes"].MaxUsage)

	assert.Equal(t, time.Duration(1*time.Minute), s.Policies["p-modes"].NextAvailable)
	assert.Equal(t, "1m0s", m.Policies["p-modes"].NextAvailable)

	assert.Equal(t, time.Duration(1*time.Minute), s.Policies["p-modes"].StartsWithin)
	assert.Equal(t, "1m0s", m.Policies["p-modes"].StartsWithin)

}

func TestDemoManifestSatisfiesAPIValidation(t *testing.T) {
	manifest, err := os.ReadFile("../../demo/manifest.yaml")
	require.NoError(t, err)

	apiManifest, _, err := YAMLToManifests(manifest)
	require.NoError(t, err)
	require.NoError(t, apiManifest.Validate(strfmt.Default))
}

func TestOperationalDurationsConvertFromYAMLAndJSON(t *testing.T) {
	yamlManifest := []byte(`
operational_workflows:
  test-check:
    description: check a stream
    kind: health_check
    expected_duration: 20ms
    maximum_duration: 5s
operational_job_templates:
  stream-check:
    workflow: test-check
    timeout: 3s
    retry:
      attempts: 2
      initial_delay: 100ms
      backoff: 2
      maximum_delay: 1s
      total_timeout: 4s
operational_pipeline_templates:
  stream-activation:
    stages:
      - name: check
        job_template: stream-check
        wait_after: 250ms
`)

	_, yamlStore, err := YAMLToManifests(yamlManifest)
	require.NoError(t, err)
	assert.Equal(t, 20*time.Millisecond, yamlStore.OperationalWorkflows["test-check"].ExpectedDuration)
	assert.Equal(t, 5*time.Second, yamlStore.OperationalWorkflows["test-check"].MaximumDuration)
	assert.Equal(t, 3*time.Second, yamlStore.OperationalJobTemplates["stream-check"].Timeout)
	assert.Equal(t, 100*time.Millisecond, yamlStore.OperationalJobTemplates["stream-check"].Retry.InitialDelay)
	assert.Equal(t, time.Second, yamlStore.OperationalJobTemplates["stream-check"].Retry.MaximumDelay)
	assert.Equal(t, 4*time.Second, yamlStore.OperationalJobTemplates["stream-check"].Retry.TotalTimeout)
	assert.Equal(t, 250*time.Millisecond, yamlStore.OperationalPipelineTemplates["stream-activation"].Stages[0].WaitAfter)

	jsonManifest, err := yamljson.YAMLToJSON(yamlManifest)
	require.NoError(t, err)
	_, jsonStore, err := JSONToManifests(jsonManifest)
	require.NoError(t, err)
	assert.Equal(t, yamlStore.OperationalWorkflows, jsonStore.OperationalWorkflows)
	assert.Equal(t, yamlStore.OperationalJobTemplates, jsonStore.OperationalJobTemplates)
	assert.Equal(t, yamlStore.OperationalPipelineTemplates, jsonStore.OperationalPipelineTemplates)
}
