package interval

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/timdrysdale/interval/diary"
	"github.com/timdrysdale/interval/interval"
	"gopkg.in/yaml.v2"
)

var manifestYAML = []byte(`descriptions:
  d-p-a:
    type: policy
    short: a
    long: ""
    further: ""
    thumb: ""
    image: ""
  d-p-b:
    type: policy
    short: b
    long: ""
    further: ""
    thumb: ""
    image: ""
  d-r-a:
    type: resource
    short: a
    long: ""
    further: ""
    thumb: ""
    image: ""
  d-r-b:
    type: resource
    short: b
    long: ""
    further: ""
    thumb: ""
    image: ""
  d-sl-a:
    type: slot
    short: a
    long: ""
    further: ""
    thumb: ""
    image: ""
  d-sl-b:
    type: slot
    short: b
    long: ""
    further: ""
    thumb: ""
    image: ""
  d-ui-a:
    type: ui
    short: a
    long: ""
    further: ""
    thumb: ""
    image: ""
  d-ui-b:
    type: ui
    short: b
    long: ""
    further: ""
    thumb: ""
    image: ""
policies:
  p-a:
    book_ahead: 0s
    description: d-p-a
    enforce_book_ahead: false
    enforce_max_bookings: false
    enforce_max_duration: false
    enforce_min_duration: false
    enforce_max_usage: false
    max_bookings: 0
    max_duration: 0s
    min_duration: 0s
    max_usage: 0s
    slots:
    - sl-a
  p-b:
    book_ahead: 2h0m0s
    description: d-p-b
    enforce_book_ahead: true
    enforce_max_bookings: false
    enforce_max_duration: false
    enforce_min_duration: false
    enforce_max_usage: false
    max_bookings: 0
    max_duration: 0s
    min_duration: 0s
    max_usage: 0s
    slots:
    - sl-b
resources:
  r-a:
    description: d-r-a
    streams:
    - st-a
    - st-b
  r-b:
    description: d-r-b
    streams:
    - st-a
    - st-b
slots:
  sl-a:
    description: d-sl-a
    policy: p-a
    resource: r-a
    ui_set: us-a
    window: w-a
  sl-b:
    description: d-sl-b
    policy: p-b
    resource: r-b
    ui_set: us-b
    window: w-b
streams:
  st-a:
    audience: a
    ct: a
    for: a
    scopes:
    - r
    - w
    topic: a
    url: a
  st-b:
    audience: b
    ct: b
    for: b
    scopes:
    - r
    - w
    topic: b
    url: b
uis:
  ui-a:
    description: d-ui-a
    url: a
    streams_required:
    - st-a
    - st-b
  ui-b:
    description: d-ui-b
    url: b
    streams_required:
    - st-a
    - st-b
ui_sets:
  us-a:
    uis:
    - ui-a
  us-b:
    uis:
    - ui-a
    - ui-b
windows:
  w-a:
    allowed:
    - start: 2022-11-04T00:00:00Z 
      end:   2022-11-06T00:00:00Z
    denied: []
  w-b:
    allowed:
    - start: 2022-11-04T00:00:00Z 
      end:   2022-11-06T00:00:00Z
    denied: []`)

func TestReplaceManifest(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	err, msg := CheckManifest(testManifest.Manifest)

	assert.NoError(t, err)
	assert.Equal(t, []string{}, msg)

	s := New()
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

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

	// check Diaries
	for _, v := range s.Resources {
		ok, reason := v.Diary.IsAvailable()
		assert.True(t, ok)
		assert.Equal(t, "Loaded at 2022-11-05T00:00:00Z", reason)
	}

	// check Filters
	for _, v := range s.Filters {
		assert.NotEqual(t, nil, v)
	}

	// check SlotMaps
	sml := make(map[string]int)
	for k, v := range s.Policies {
		sml[k] = len(v.SlotMap)
	}
	exp := map[string]int{
		"p-a": 1,
		"p-b": 1,
	}
	assert.Equal(t, exp, sml)

}

// rename as Test... if required to update the yaml file for testing manifest ingest
func testCreateManifestYAML(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	d, err := yaml.Marshal(&testManifest.Manifest)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	fmt.Printf("\n%s\n", string(d))
}

func TestReplaceManifestFromYAML(t *testing.T) {
	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)
	s := New()
	err = s.ReplaceManifest(m)
	assert.NoError(t, err)
}

func TestAvailability(t *testing.T) {

	start := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	s1 := start.Add(5 * time.Minute)
	e1 := start.Add(10 * time.Minute)
	s2 := start.Add(20 * time.Minute)
	e2 := start.Add(30 * time.Minute)

	bk := []diary.Booking{
		diary.Booking{
			When: interval.Interval{
				Start: s1,
				End:   e1,
			},
		},
		diary.Booking{
			When: interval.Interval{
				Start: s2,
				End:   e2,
			},
		},
	}

	exp := []interval.Interval{
		interval.Interval{
			Start: start,
			End:   s1.Add(-time.Nanosecond),
		},
		interval.Interval{
			Start: e1.Add(time.Nanosecond),
			End:   s2.Add(-time.Nanosecond),
		},
		interval.Interval{
			Start: e2.Add(time.Nanosecond),
			End:   end,
		},
	}

	a := Availability(bk, start, end)

	assert.Equal(t, exp, a)

}

// TestBooking checks whether the availability calculations result in bookable
// sessions that do not overlap the existing booked sessions.
func TestAvailabilityTimeBoundaries(t *testing.T) {

	start := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	s1 := start.Add(5 * time.Minute)
	e1 := start.Add(10 * time.Minute)
	s2 := start.Add(20 * time.Minute)
	e2 := start.Add(30 * time.Minute)

	bk := []diary.Booking{
		diary.Booking{
			When: interval.Interval{
				Start: s1,
				End:   e1,
			},
		},
		diary.Booking{
			When: interval.Interval{
				Start: s2,
				End:   e2,
			},
		},
	}

	a := Availability(bk, start, end)

	d := diary.New("test")

	_, err := d.Request(bk[0].When)
	assert.NoError(t, err)
	_, err = d.Request(bk[1].When)
	assert.NoError(t, err)

	// request the whole middle interval that is available
	_, err = d.Request(a[2])
	assert.NoError(t, err)
}

func TestGetSlotIsAvailable(t *testing.T) {

	s := New()

	// fix time for ease of testing reason string
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	ok, reason, err := s.GetSlotIsAvailable("sl-a")

	assert.NoError(t, err)
	assert.Equal(t, true, ok)
	assert.Equal(t, "Loaded at 2022-11-05T00:00:00Z", reason)

}

func TestSetSlotIsAvailable(t *testing.T) {

	s := New()

	// fix time for ease of testing reason string
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	ok, reason, err := s.GetSlotIsAvailable("sl-a")

	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "Loaded at 2022-11-05T00:00:00Z", reason)

	s.SetSlotIsAvailable("sl-a", false, "foo")

	ok, reason, err = s.GetSlotIsAvailable("sl-a")

	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, "unavailable because foo", reason)

	s.SetSlotIsAvailable("sl-a", true, "bar")

	ok, reason, err = s.GetSlotIsAvailable("sl-a")

	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "bar", reason)

}

func TestGetSlotAvailabilityWithNoBookings(t *testing.T) {

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	ok, reason, err := s.GetSlotIsAvailable("sl-a")

	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "Loaded at 2022-11-05T00:00:00Z", reason)

	// no lookahead limit in policy
	a, err := s.GetAvailability("p-a", "sl-a")
	assert.NoError(t, err)
	exp := []interval.Interval{
		interval.Interval{
			Start: s.Now(),
			End:   interval.Infinity,
		},
	}
	assert.Equal(t, exp, a)

	// 2-hour lookahead limit in policy
	a, err = s.GetAvailability("p-b", "sl-b")
	assert.NoError(t, err)
	exp = []interval.Interval{
		interval.Interval{
			Start: s.Now(),
			End:   s.Now().Add(2 * time.Hour),
		},
	}
	assert.Equal(t, exp, a)

	// slot not part of policy
	a, err = s.GetAvailability("p-b", "sl-a")
	assert.Error(t, err)
	assert.Equal(t, "slot sl-a not in policy p-b", err.Error())

}
