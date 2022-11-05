package interval

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var manifest = []byte{`descriptions:
- name: d-p-a
  type: policy
  short: a
  long: ""
  further: ""
  thumb: ""
  image: ""
- name: d-p-b
  type: policy
  short: b
  long: ""
  further: ""
  thumb: ""
  image: ""
- name: d-r-a
  type: resource
  short: a
  long: ""
  further: ""
  thumb: ""
  image: ""
- name: d-r-b
  type: resource
  short: b
  long: ""
  further: ""
  thumb: ""
  image: ""
- name: d-sl-a
  type: slot
  short: a
  long: ""
  further: ""
  thumb: ""
  image: ""
- name: d-sl-b
  type: slot
  short: b
  long: ""
  further: ""
  thumb: ""
  image: ""
- name: d-ui-a
  type: ui
  short: a
  long: ""
  further: ""
  thumb: ""
  image: ""
- name: d-ui-b
  type: ui
  short: b
  long: ""
  further: ""
  thumb: ""
  image: ""
policies:
- description: d-p-a
  enforce_max_bookings: false
  enforce_max_duration: false
  enforce_min_duration: false
  enforce_max_usage: false
  max_bookings: 0
  max_duration: 0s
  min_duration: 0s
  name: p-a
  max_usage: 0s
- description: d-p-b
  enforce_max_bookings: false
  enforce_max_duration: false
  enforce_min_duration: false
  enforce_max_usage: false
  max_bookings: 0
  max_duration: 0s
  min_duration: 0s
  name: p-b
  max_usage: 0s
resources:
- description: d-r-a
  name: r-a
  streams:
  - st-a
  - st-b
- description: d-r-b
  name: r-b
  streams:
  - st-a
  - st-b
slots:
- description: d-sl-a
  name: sl-a
  policy: p-a
  resource: r-a
  ui_set: us-a
  window: w-a
- description: d-sl-b
  name: sl-b
  policy: p-b
  resource: r-b
  ui_set: us-b
  window: w-b
streams:
- name: st-a
  audience: a
  ct: a
  for: a
  scopes:
  - r
  - w
  topic: a
  url: a
- name: st-b
  audience: b
  ct: b
  for: b
  scopes:
  - r
  - w
  topic: b
  url: b
uis:
- name: ui-a
  description: d-ui-a
  url: a
  streams_required:
  - st-a
  - st-b
- name: ui-b
  description: d-ui-b
  url: b
  streams_required:
  - st-a
  - st-b
ui_sets:
- name: us-a
  uis:
  - ui-a
- name: us-b
  uis:
  - ui-a
  - ui-b
windows:
- name: w-a
  allowed:
  - start: 2022-11-05T00:18:02.215634079Z
    end: 2022-11-05T01:18:02.215634346Z
  denied: []
- name: w-b
  allowed:
  - start: 2022-11-05T00:18:02.215634977Z
    end: 2022-11-05T01:18:02.215635182Z
  denied: []`}

func TestReplaceManifest(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	err, msg := CheckManifest(testManifest.Manifest)

	assert.NoError(t, err)
	assert.Equal(t, []string{}, msg)

	s := New()

	err = s.ReplaceManifest(testManifest.Manifest)

	assert.NoError(t, err)

	assert.Equal(t, 8, len(s.Descriptions))
	assert.Equal(t, 2, len(s.Filters))
	assert.Equal(t, 2, len(s.Policies))
	assert.Equal(t, 2, len(s.Resources))
	assert.Equal(t, 2, len(s.Slots))
	assert.Equal(t, 2, len(s.Streams))
	assert.Equal(t, 2, len(s.UIs))
	assert.Equal(t, 2, len(s.UISets))
	assert.Equal(t, 2, len(s.Windows))

}

// rename as Test... if required to update the yaml file for testing manifest ingest
func TestCreateManifestYAML(t *testing.T) {

	d, err := yaml.Marshal(&testManifest.Manifest)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	fmt.Printf("\n%s\n", string(d))
}
