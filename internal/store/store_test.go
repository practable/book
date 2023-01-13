package store

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/timdrysdale/interval/internal/diary"
	"github.com/timdrysdale/interval/internal/interval"
	"gopkg.in/yaml.v2"
)

var manifestYAML = []byte(`descriptions:
  d-p-a:
    name: policy-a
    type: policy
    short: a
  d-p-b:
    name: policy-b
    type: policy
    short: b  
  d-p-instant:
    name: policy-instant
    type: policy
    short: instant
  d-p-simulation:
    name: policy-simulation
    type: policy
    short: simulation
  d-p-start-in-past:
    name: policy-start-in-past
    type: policy
    short: start-in-past
  d-r-a:
    name: resource-a
    type: resource
    short: a
  d-r-b:
    name: resource-b
    type: resource
    short: b
  d-r-simulation:
    name: resource-simulation
    type: resource
    short: simulation
  d-sl-a:
    name: slot-a
    type: slot
    short: a
  d-sl-b:
    name: slot-b
    type: slot
    short: b  
  d-sl-instant:
    name: slot-instant
    type: slot
    short: instant
  d-sl-simulation:
    name: slot-simulation
    type: slot
    short: simulation
  d-sl-start-in-past:
    name: slot-start-in-past
    type: slot
    short: start-in-past
  d-ui-a:
    name: ui-a
    type: ui
    short: a
  d-ui-b:
    name: ui-b
    type: ui
    short: b 
  d-ui-simulation:
    name: ui-simulation
    type: ui
    short: simulation   
display_guides:
  6m:
    book_ahead: 1h
    duration: 6m
    max_slots: 12
  8m:
    book_ahead: 2h
    duration: 8m
    max_slots: 8
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
    display_guides:
      - 6m
      - 8m
    enforce_book_ahead: true
    enforce_max_bookings: true
    enforce_max_duration: true
    enforce_min_duration: true
    enforce_max_usage: true
    max_bookings: 2
    max_duration: 10m0s
    min_duration: 5m0s
    max_usage: 30m0s
    slots:
    - sl-b
  p-instant:
    book_ahead: 2h0m0s
    description: d-p-b
    display_guides:
      - 6m
      - 8m
    enforce_book_ahead: true
    enforce_max_bookings: true
    enforce_max_duration: true
    enforce_min_duration: true
    enforce_max_usage: true
    enforce_starts_within: true
    max_bookings: 2
    max_duration: 10m0s
    min_duration: 5m0s
    max_usage: 30m0s
    slots:
    - sl-instant
    starts_within: 1m
  p-simulation:
    book_ahead: 2h0m0s
    description: d-p-simulation
    display_guides:
      - 6m
      - 8m
    enforce_book_ahead: true
    enforce_max_bookings: true
    enforce_max_duration: true
    enforce_min_duration: true
    enforce_max_usage: true
    enforce_unlimited_users: true
    max_bookings: 2
    max_duration: 10m0s
    min_duration: 5m0s
    max_usage: 30m0s
    slots:
    - sl-simulation
  p-start-in-past:
    allow_start_in_past_within: 1m
    book_ahead: 2h0m0s
    description: d-p-start-in-past
    display_guides:
      - 6m
      - 8m
    enforce_allow_start_in_past: true
    enforce_book_ahead: true
    enforce_max_bookings: true
    enforce_max_duration: true
    enforce_min_duration: true
    enforce_max_usage: true
    max_bookings: 2
    max_duration: 10m0s
    min_duration: 5m0s
    max_usage: 30m0s
    slots:
    - sl-start-in-past
resources:
  r-a:
    description: d-r-a
    streams:
    - st-a
    - st-b
    topic_stub: aaaa00
  r-b:
    description: d-r-b
    streams:
    - st-a
    - st-b
    topic_stub: bbbb00
  r-simulation:
    description: d-r-simulation
    streams:
    - st-log
    topic_stub: simu00
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
  sl-instant:
    description: d-sl-instant
    policy: p-instant
    resource: r-b
    ui_set: us-b
    window: w-b
  sl-simulation:
    description: d-sl-simulation
    policy: p-simulation
    resource: r-simulation
    ui_set: us-simulation
    window: w-b
  sl-start-in-past:
    description: d-sl-start-in-past
    policy: p-start-in-past
    resource: r-a
    ui_set: us-a
    window: w-a
streams:
  st-a:
    audience: a
    connection_type: a
    for: a
    scopes:
    - r
    - w
    topic: a
    url: a
  st-b:
    audience: b
    connection_type: b
    for: b
    scopes:
    - r
    - w
    topic: b
    url: b
  st-log:
    audience: some_audience
    connection_type: session
    for: log
    scopes:
    - r
    - w
    topic: some_topic
    url: some_url
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
  ui-simulation:
    description: d-ui-simulation
    url: https://some_url.org
    streams_required:
    - st-log
ui_sets:
  us-a:
    uis:
    - ui-a
  us-b:
    uis:
    - ui-a
    - ui-b
  us-simulation:
    uis:
    - ui-simulation
windows:
  w-a:
    allowed:
    - start: 2022-11-04T00:00:00Z
      end: 2022-11-06T00:00:00Z
    denied: []
  w-b:
    allowed:
    - start: 2022-11-04T00:00:00Z
      end: 2022-11-06T00:00:00Z
    denied: []`)

var debug bool

func init() {
	debug = false
	if debug {
		log.SetReportCaller(true)
		log.SetLevel(log.TraceLevel)
		log.SetFormatter(&log.TextFormatter{FullTimestamp: false, DisableColors: true})
		defer log.SetOutput(os.Stdout)

	} else {
		log.SetLevel(log.WarnLevel)
		var ignore bytes.Buffer
		logignore := bufio.NewWriter(&ignore)
		log.SetOutput(logignore)
	}

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

func TestReplaceManifest(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	err, msg := checkManifest(testManifest.Manifest)

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

func TestReplaceManifestFromYAML(t *testing.T) {
	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)
	s := New()
	err = s.ReplaceManifest(m)
	assert.NoError(t, err)
	if err != nil { //print errors (useful during manifest evolution to add new tests)
		_, list := checkManifest(m) //err same as before
		for _, item := range list {
			t.Log(item)
		}
	}
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

	a := availability(bk, start, end)

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

	a := availability(bk, start, end)

	d := diary.New("test")

	err := d.Request(bk[0].When, "test00")
	assert.NoError(t, err)
	err = d.Request(bk[1].When, "test01")
	assert.NoError(t, err)

	// request the whole middle interval that is available
	err = d.Request(a[2], "test02")
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
			End:   s.Now().Add(interval.Century), //don't use infinity because it does not parse well in the API
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

func TestMakeBooking(t *testing.T) {

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }

	policy := "p-b"
	slot := "sl-b"
	user := "test" //does not yet exist in store
	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}

	b, err := s.MakeBookingWithName(policy, slot, user, when, "test00")

	assert.NoError(t, err)

	assert.Equal(t, policy, b.Policy)
	assert.Equal(t, slot, b.Slot)
	assert.Equal(t, user, b.User)
	assert.Equal(t, when, b.When)
	assert.Equal(t, "test00", b.Name)
	assert.False(t, b.Cancelled)
	assert.False(t, b.Started)
	assert.False(t, b.Unfulfilled)
}

func TestDenyBookingOfUnavailable(t *testing.T) {

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	s.SetSlotIsAvailable("sl-b", false, "foo")

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }

	policy := "p-b"
	slot := "sl-b"
	user := "test" //does not yet exist in store
	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}

	_, err = s.MakeBooking(policy, slot, user, when)

	assert.Error(t, err)
	assert.Equal(t, "unavailable because foo", err.Error())

}

func TestPolicyChecks(t *testing.T) {

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }

	// Check denied outside slot's window
	policy := "p-a"
	slot := "sl-a"
	user := "test" //does not yet exist in store

	when := interval.Interval{
		Start: time.Date(2022, 11, 20, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 20, 2, 10, 0, 0, time.UTC),
	}

	_, err = s.MakeBooking(policy, slot, user, when)

	assert.Error(t, err)
	assert.Equal(t, "bookings cannot be made outside the window for the slot", err.Error())

	// Check denied outside bookahed window
	policy = "p-b"
	slot = "sl-b"
	user = "test" //does not yet exist in store

	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 12, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 12, 10, 0, 0, time.UTC),
	}

	_, err = s.MakeBooking(policy, slot, user, when)

	assert.Error(t, err)
	assert.Equal(t, "bookings cannot be made more than 2h0m0s ahead of the current time", err.Error())

	// Too many bookings (ignoring attempted bookings)

	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}
	_, err = s.MakeBooking(policy, slot, user, when)
	assert.NoError(t, err)

	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 10, 0, 1, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 20, 0, 0, time.UTC),
	}
	_, err = s.MakeBooking(policy, slot, user, when)
	assert.NoError(t, err)

	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 20, 0, 1, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 30, 0, 0, time.UTC),
	}

	_, err = s.MakeBooking(policy, slot, user, when)
	assert.Error(t, err)
	assert.Equal(t, "you currently have 2 current/future bookings which is at or exceeds the limit of 2 for policy p-b", err.Error())

	// advance time to after both previous bookings
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 3, 0, 0, 0, time.UTC) }

	// a further booking must now succeed
	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 3, 10, 0, 1, time.UTC),
		End:   time.Date(2022, 11, 5, 3, 20, 0, 0, time.UTC),
	}
	_, err = s.MakeBooking(policy, slot, user, when)
	assert.NoError(t, err)

	// we now exceed the available usage, so should be denied
	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 3, 30, 0, 1, time.UTC),
		End:   time.Date(2022, 11, 5, 3, 40, 0, 0, time.UTC),
	}
	_, err = s.MakeBooking(policy, slot, user, when)
	assert.Error(t, err)
	assert.Equal(t, "requested duration of 10m0s exceeds remaining usage limit of 0s", err.Error())

	// another user can book (check usage is applied per user)
	user = "bar"
	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 3, 30, 0, 1, time.UTC),
		End:   time.Date(2022, 11, 5, 3, 36, 0, 0, time.UTC),
	}
	_, err = s.MakeBooking(policy, slot, user, when)
	assert.NoError(t, err)

	// user books too short a duration
	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 3, 37, 0, 1, time.UTC),
		End:   time.Date(2022, 11, 5, 3, 38, 0, 0, time.UTC),
	}
	_, err = s.MakeBooking(policy, slot, user, when)
	assert.Error(t, err)
	assert.Equal(t, "requested duration of 1m0s shorter than minimum permitted duration of 5m0s", err.Error())

	// user books too long a duration
	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 3, 40, 0, 1, time.UTC),
		End:   time.Date(2022, 11, 5, 3, 55, 0, 0, time.UTC),
	}
	_, err = s.MakeBooking(policy, slot, user, when)
	assert.Error(t, err)
	assert.Equal(t, "requested duration of 15m0s longer than maximum permitted duration of 10m0s", err.Error())

	// user books ok, using up usage allowance
	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 3, 40, 0, 1, time.UTC),
		End:   time.Date(2022, 11, 5, 3, 50, 0, 0, time.UTC),
	}
	_, err = s.MakeBooking(policy, slot, user, when)
	assert.NoError(t, err)

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 4, 0, 0, 0, time.UTC) }

	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 4, 10, 0, 1, time.UTC),
		End:   time.Date(2022, 11, 5, 4, 20, 0, 0, time.UTC),
	}
	bc, err := s.MakeBooking(policy, slot, user, when)
	assert.NoError(t, err)

	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 4, 30, 0, 1, time.UTC),
		End:   time.Date(2022, 11, 5, 4, 40, 0, 0, time.UTC),
	}
	_, err = s.MakeBooking(policy, slot, user, when)
	assert.Error(t, err)
	assert.Equal(t, "requested duration of 10m0s exceeds remaining usage limit of 4m0s", err.Error())

	// free up some allocation and try again, must succeed
	err = s.CancelBooking(bc, "test")
	assert.NoError(t, err)
	_, err = s.MakeBooking(policy, slot, user, when)
	assert.NoError(t, err)

	// indirect check on remaining usage, to ensure cancellation refund was accurate amount
	// move forward in time to avoid limit on current/future bookings
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 6, 0, 0, 0, time.UTC) }

	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 6, 45, 0, 1, time.UTC),
		End:   time.Date(2022, 11, 5, 6, 55, 0, 0, time.UTC),
	}

	_, err = s.MakeBooking(policy, slot, user, when)
	assert.Error(t, err)
	assert.Equal(t, "requested duration of 10m0s exceeds remaining usage limit of 4m0s", err.Error())

	// make a booking then try to cancel it with incomplete information, must fail
	user = "test1"
	b, err := s.MakeBooking(policy, slot, user, when)
	assert.NoError(t, err)

	fake := Booking{
		Name: b.Name,
	}
	err = s.CancelBooking(fake, "test")
	assert.Error(t, err)
	assert.Equal(t, "could not verify booking details", err.Error())

}

func TestGetActivity(t *testing.T) {

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }

	policy := "p-b"
	slot := "sl-b"
	user := "test" //does not yet exist in store
	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}

	b, err := s.MakeBooking(policy, slot, user, when)

	assert.NoError(t, err)

	assert.Equal(t, policy, b.Policy)
	assert.Equal(t, slot, b.Slot)
	assert.Equal(t, user, b.User)
	assert.Equal(t, when, b.When)
	assert.NotEqual(t, "", b.Name) //non null name
	assert.False(t, b.Cancelled)
	assert.False(t, b.Started)
	assert.False(t, b.Unfulfilled)

	// advance time, but still before booking is live
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 59, 0, 0, time.UTC) }

	_, err = s.GetActivity(b)
	assert.Error(t, err)
	assert.Equal(t, "too early", err.Error())

	// advance time, but after booking is finished (edge case where booking not pruned yet)
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 2, 11, 0, 0, time.UTC) }

	_, err = s.GetActivity(b)
	assert.Error(t, err)
	assert.Equal(t, "too late", err.Error())

	// incomplete booking
	badb := Booking{
		Name: b.Name,
	}
	_, err = s.GetActivity(badb)
	assert.Error(t, err)
	assert.Equal(t, "could not verify booking details", err.Error())

	// shift to time within booking, but make resource unavailable.
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 2, 02, 0, 0, time.UTC) }
	s.SetSlotIsAvailable("sl-b", false, "test")

	_, err = s.GetActivity(b)
	assert.Error(t, err)
	assert.Equal(t, "unavailable because test", err.Error())

	// now make resource available, must get activity now
	s.SetSlotIsAvailable("sl-b", true, "test")

	a, err := s.GetActivity(b)

	assert.NoError(t, err)
	exp := Activity{
		Description: Description{
			Name:    "slot-b",
			Type:    "slot",
			Short:   "b",
			Long:    "",
			Further: "",
			Thumb:   "",
			Image:   ""},
		ConfigURL: "",
		Streams: map[string]Stream{
			"st-a": Stream{
				Audience:       "a",
				ConnectionType: "a",
				For:            "a",
				Scopes:         []string{"r", "w"},
				Topic:          "bbbb00-st-a",
				URL:            "a"},
			"st-b": Stream{
				Audience:       "b",
				ConnectionType: "b",
				For:            "b",
				Scopes:         []string{"r", "w"},
				Topic:          "bbbb00-st-b",
				URL:            "b"}},
		UIs: []UIDescribed{
			UIDescribed{
				Description: Description{
					Name:    "ui-a",
					Type:    "ui",
					Short:   "a",
					Long:    "",
					Further: "",
					Thumb:   "",
					Image:   ""},
				URL:             "a",
				StreamsRequired: []string{"st-a", "st-b"},
			},
			UIDescribed{
				Description: Description{
					Name:    "ui-b",
					Type:    "ui",
					Short:   "b",
					Long:    "",
					Further: "",
					Thumb:   "",
					Image:   ""},
				URL:             "b",
				StreamsRequired: []string{"st-a", "st-b"}}},
		NotBefore: time.Date(2022, time.November, 5, 2, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2022, time.November, 5, 2, 10, 0, 0, time.UTC),
	}

	assert.Equal(t, exp, a)

	// must not cancel started activity
	err = s.CancelBooking(b, "test")
	assert.Error(t, err)
	assert.Equal(t, "cannot cancel booking that has already been used", err.Error())

	// TODO - set up a user with two short bookings, then try to make third booking that is within total usage allowance, but outside maxBookings, so it must fail. then cancel a booking, and try again. Third booking should suceed now that MaxBookings limit does not prevent it.

}

func TestCheckBooking(t *testing.T) {

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	policy := "p-b"
	slot := "sl-b"
	user := "test" //does not yet exist in store
	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 1, 10, 0, 0, time.UTC),
	}

	b, err := s.MakeBooking(policy, slot, user, when)

	assert.NoError(t, err)

	err, msg := s.checkBooking(b)

	assert.NoError(t, err)
	assert.Equal(t, []string{}, msg)

	b.Policy = ""
	err, msg = s.checkBooking(b)
	assert.Error(t, err)
	assert.Equal(t, []string{b.Name + " missing policy"}, msg)
	b.Policy = "foo"
	err, msg = s.checkBooking(b)
	assert.Error(t, err)
	assert.Equal(t, []string{b.Name + " policy foo not found"}, msg)
	b.Policy = policy

	b.Slot = ""
	err, msg = s.checkBooking(b)
	assert.Error(t, err)
	assert.Equal(t, []string{b.Name + " missing slot"}, msg)
	b.Slot = "foo"
	err, msg = s.checkBooking(b)
	assert.Error(t, err)
	assert.Equal(t, []string{b.Name + " slot foo not found"}, msg)
	b.Slot = slot

	b.User = ""
	err, msg = s.checkBooking(b)
	assert.Error(t, err)
	assert.Equal(t, []string{b.Name + " missing user"}, msg)
	// no need to check for user not found - this is ok, as
	// they are created as needed when bookings are made

	b.User = user

	name := b.Name
	b.Name = ""
	err, msg = s.checkBooking(b)
	assert.Error(t, err)
	assert.Equal(t, []string{"missing name"}, msg)
	b.Name = name

	b.When = interval.Interval{}
	err, msg = s.checkBooking(b)
	assert.Error(t, err)
	assert.Equal(t, []string{b.Name + " missing when"}, msg)
	b.When = when

}

func TestExportBookings(t *testing.T) {

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	policy0 := "p-a"
	slot0 := "sl-a"
	user0 := "u-a" //does not yet exist in store
	when0 := interval.Interval{
		Start: time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 1, 10, 0, 0, time.UTC),
	}

	b0, err := s.MakeBooking(policy0, slot0, user0, when0)

	assert.NoError(t, err)

	policy1 := "p-b"
	slot1 := "sl-b"
	user1 := "u-b" //does not yet exist in store
	when1 := interval.Interval{
		Start: time.Date(2022, 11, 5, 1, 5, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 1, 15, 0, 0, time.UTC),
	}

	b1, err := s.MakeBooking(policy1, slot1, user1, when1)

	assert.NoError(t, err)

	bm := s.ExportBookings()

	exp := make(map[string]Booking)

	exp[b0.Name] = b0
	exp[b1.Name] = b1

	assert.Equal(t, exp, bm)

}

func TestReplaceBookings(t *testing.T) {

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	policy0 := "p-a"
	slot0 := "sl-a"
	user0 := "u-a" //does not yet exist in store
	when0 := interval.Interval{
		Start: time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 1, 10, 0, 0, time.UTC),
	}

	b0, err := s.MakeBooking(policy0, slot0, user0, when0)

	assert.NoError(t, err)

	policy1 := "p-b"
	slot1 := "sl-b"
	user1 := "u-b" //does not yet exist in store
	when1 := interval.Interval{
		Start: time.Date(2022, 11, 5, 1, 5, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 1, 15, 0, 0, time.UTC),
	}

	b1, err := s.MakeBooking(policy1, slot1, user1, when1)

	assert.NoError(t, err)

	bm := s.ExportBookings()

	exp := make(map[string]Booking)

	exp[b0.Name] = b0
	exp[b1.Name] = b1

	assert.Equal(t, exp, bm)

	// Now prepare replacement bookings

	policy2 := "p-a"
	slot2 := "sl-a"
	user2 := "u-c" //does not yet exist in store
	when2 := interval.Interval{
		Start: time.Date(2022, 11, 5, 1, 2, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 1, 12, 0, 0, time.UTC),
	}

	policy3 := "p-b"
	slot3 := "sl-b"
	user3 := "u-d" //does not yet exist in store
	when3 := interval.Interval{
		Start: time.Date(2022, 11, 5, 1, 6, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 1, 16, 0, 0, time.UTC),
	}

	b2 := Booking{
		Name:   "b2",
		Policy: policy2,
		Slot:   slot2,
		User:   user2,
		When:   when2,
	}

	b3 := Booking{
		Name:   "b3",
		Policy: policy3,
		Slot:   slot3,
		User:   user3,
		When:   when3,
	}

	nb := make(map[string]Booking)
	nb["b2"] = b2
	nb["b3"] = b3

	err, msg := s.ReplaceBookings(nb)

	assert.NoError(t, err)
	assert.Equal(t, []string{}, msg)

	exp = make(map[string]Booking)
	exp[b2.Name] = b2
	exp[b3.Name] = b3

	bm = s.ExportBookings()

	assert.Equal(t, exp, bm)

}

func TestOldBookings(t *testing.T) {

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	policy0 := "p-a"
	slot0 := "sl-a"
	user0 := "u-a" //does not yet exist in store
	when0 := interval.Interval{
		Start: time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 1, 10, 0, 0, time.UTC),
	}

	b0, err := s.MakeBooking(policy0, slot0, user0, when0)

	assert.NoError(t, err)

	policy1 := "p-b"
	slot1 := "sl-b"
	user1 := "u-b" //does not yet exist in store
	when1 := interval.Interval{
		Start: time.Date(2022, 11, 5, 1, 5, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 1, 15, 0, 0, time.UTC),
	}

	b1, err := s.MakeBooking(policy1, slot1, user1, when1)

	assert.NoError(t, err)

	bm := s.ExportBookings()

	exp := make(map[string]Booking)

	exp[b0.Name] = b0
	exp[b1.Name] = b1

	assert.Equal(t, exp, bm)

	// Now move time forward to make these old bookings
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC) }

	s.pruneBookings()

	// check our bookings are now old bookings
	bm = s.ExportOldBookings()
	assert.Equal(t, exp, bm)

	// check they are not present in the current bookings anymore
	bm = s.ExportBookings()
	exp = make(map[string]Booking)
	assert.Equal(t, exp, bm)

	// Prepare replacement old bookings

	policy2 := "p-a"
	slot2 := "sl-a"
	user2 := "u-c" //does not yet exist in store
	when2 := interval.Interval{
		Start: time.Date(2022, 11, 5, 1, 2, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 1, 12, 0, 0, time.UTC),
	}

	policy3 := "p-b"
	slot3 := "sl-b"
	user3 := "u-d" //does not yet exist in store
	when3 := interval.Interval{
		Start: time.Date(2022, 11, 5, 1, 6, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 1, 16, 0, 0, time.UTC),
	}

	b2 := Booking{
		Name:   "b2",
		Policy: policy2,
		Slot:   slot2,
		User:   user2,
		When:   when2,
	}

	b3 := Booking{
		Name:   "b3",
		Policy: policy3,
		Slot:   slot3,
		User:   user3,
		When:   when3,
	}

	nb := make(map[string]Booking)
	nb["b2"] = b2
	nb["b3"] = b3

	err, msg := s.ReplaceOldBookings(nb)

	assert.NoError(t, err)
	assert.Equal(t, []string{}, msg)

	exp = make(map[string]Booking)
	exp[b2.Name] = b2
	exp[b3.Name] = b3

	bm = s.ExportOldBookings()

	assert.Equal(t, exp, bm)

}

func TestExportUsers(t *testing.T) {

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }

	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}

	_, err = s.MakeBookingWithName("p-a", "sl-a", "user-a", when, "test00")
	_, err = s.MakeBookingWithName("p-b", "sl-b", "user-b", when, "test01")

	um := s.ExportUsers()

	exp := make(map[string]UserExternal)

	exp["user-a"] = UserExternal{
		Bookings:    []string{"test00"},
		OldBookings: []string{},
		Policies:    []string{"p-a"},
		Usage: map[string]string{
			"p-a": "10m0s",
		},
	}
	exp["user-b"] = UserExternal{
		Bookings:    []string{"test01"},
		OldBookings: []string{},
		Policies:    []string{"p-b"},
		Usage: map[string]string{
			"p-b": "10m0s",
		},
	}

	assert.Equal(t, exp, um)

}

func TestReplaceBookingsUsageRefunded(t *testing.T) {

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)
	s := New()
	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	// make a booking

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }

	policy := "p-b"
	slot := "sl-b"
	user := "u-b" //does not yet exist in store
	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}

	b, err := s.MakeBooking(policy, slot, user, when)

	assert.NoError(t, err)

	// Check booking is as expected
	assert.Equal(t, policy, b.Policy)
	assert.Equal(t, slot, b.Slot)
	assert.Equal(t, user, b.User)
	assert.Equal(t, when, b.When)
	assert.NotEqual(t, "", b.Name) //non null name
	assert.False(t, b.Cancelled)
	assert.False(t, b.Started)
	assert.False(t, b.Unfulfilled)

	// Check user usages
	um := s.ExportUsers()

	// check test user a does not exist
	_, ok := um["u-a"]
	assert.False(t, ok)

	// check test user b exists
	utb, ok := um["u-b"]
	assert.True(t, ok)

	// check usage of user b is correct
	assert.Equal(t, "10m0s", utb.Usage["p-b"])

	// modify the booking to belong to user-a
	bm := s.ExportBookings()

	newb := bm[b.Name]

	newb.User = "u-a"

	bm[newb.Name] = newb

	err, msgs := s.ReplaceBookings(bm)

	if err != nil {
		t.Log(msgs)
	}

	assert.NoError(t, err)

	bm = s.ExportBookings()

	// Check user usages
	um = s.ExportUsers()

	// check test user a exists
	uta, ok := um["u-a"]
	assert.True(t, ok)

	// check usage of user a is correct
	assert.Equal(t, "10m0s", uta.Usage["p-b"])

	// check test user b exists still
	utb, ok = um["u-b"]
	assert.True(t, ok)

	// check usage of user b has been refunded the cancelled booking
	assert.Equal(t, "0s", utb.Usage["p-b"])
}

func TestReplaceOldBookings(t *testing.T) {
	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)
	s := New()
	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	// make a booking

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }

	policy := "p-b"
	slot := "sl-b"
	user := "u-b" //does not yet exist in store
	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}

	b, err := s.MakeBooking(policy, slot, user, when)

	assert.NoError(t, err)

	// Check booking is as expected
	assert.Equal(t, policy, b.Policy)
	assert.Equal(t, slot, b.Slot)
	assert.Equal(t, user, b.User)
	assert.Equal(t, when, b.When)
	assert.NotEqual(t, "", b.Name) //non null name
	assert.False(t, b.Cancelled)
	assert.False(t, b.Started)
	assert.False(t, b.Unfulfilled)

	// Check user usages
	um := s.ExportUsers()

	// check test user a does not exist
	_, ok := um["u-a"]
	assert.False(t, ok)

	// check test user b exists
	utb, ok := um["u-b"]
	assert.True(t, ok)

	// check usage of user b is correct
	assert.Equal(t, "10m0s", utb.Usage["p-b"])

	// Move one day to the future, to make the booking old
	s.Now = func() time.Time { return time.Date(2022, 12, 5, 1, 0, 0, 0, time.UTC) }

	s.pruneBookings()

	// modify the booking to belong to user-a
	bm := s.ExportOldBookings()

	newb := bm[b.Name]

	newb.User = "u-a"

	bm[newb.Name] = newb

	err, msgs := s.ReplaceOldBookings(bm)

	if err != nil {
		t.Log(msgs)
	}

	assert.NoError(t, err)

	bm = s.ExportBookings()

	// Check user usages
	um = s.ExportUsers()

	// check test user a exists
	uta, ok := um["u-a"]
	assert.True(t, ok)

	// check usage of user a is correct
	assert.Equal(t, "10m0s", uta.Usage["p-b"])

	// check test user b now does not exist (unlike replacebookings, users without an oldbooking are deleted during the old bookings replacement process)
	_, ok = um["u-b"]
	assert.False(t, ok)

	// check that bookings are indeed old bookings
	ps, err := s.GetPolicyStatusFor("u-a", policy)

	assert.NoError(t, err)

	assert.Equal(t, int64(0), ps.CurrentBookings)
	assert.Equal(t, int64(1), ps.OldBookings)
	d, err := time.ParseDuration("10m0s")
	assert.NoError(t, err)
	assert.Equal(t, d, ps.Usage)
}

func TestGetBookingsForGetOldBookingsFor(t *testing.T) {
	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)
	s := New()
	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	// make a booking

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }

	policy := "p-a"
	slot := "sl-a"
	user := "u-a"
	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}

	_, err = s.MakeBooking(policy, slot, user, when)
	assert.NoError(t, err)

	policy = "p-b"
	slot = "sl-b"
	user = "u-b" //does not yet exist in store
	when = interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}
	_, err = s.MakeBooking(policy, slot, user, when)
	assert.NoError(t, err)

	bm, err := s.GetBookingsFor("u-a")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(bm))
	assert.Equal(t, "sl-a", bm[0].Slot)

	bm, err = s.GetBookingsFor("u-b")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(bm))
	assert.Equal(t, "sl-b", bm[0].Slot)

	bm, err = s.GetOldBookingsFor("u-a")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(bm))

	bm, err = s.GetOldBookingsFor("u-b")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(bm))

	// move forward a day to make bookings old
	s.Now = func() time.Time { return time.Date(2022, 12, 5, 1, 0, 0, 0, time.UTC) }
	s.pruneBookings()

	bm, err = s.GetBookingsFor("u-a")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(bm))

	bm, err = s.GetBookingsFor("u-b")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(bm))

	bm, err = s.GetOldBookingsFor("u-a")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(bm))
	assert.Equal(t, "sl-a", bm[0].Slot)

	bm, err = s.GetOldBookingsFor("u-b")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(bm))
	assert.Equal(t, "sl-b", bm[0].Slot)

}

func TestGetPolicyStatusFor(t *testing.T) {

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)
	s := New()
	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	// make a booking

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }

	policy := "p-b"
	slot := "sl-b"
	user := "u-b" //does not yet exist in store
	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}

	b, err := s.MakeBooking(policy, slot, user, when)

	assert.NoError(t, err)

	// Check booking is as expected
	assert.Equal(t, policy, b.Policy)
	assert.Equal(t, slot, b.Slot)
	assert.Equal(t, user, b.User)
	assert.Equal(t, when, b.When)
	assert.NotEqual(t, "", b.Name) //non null name
	assert.False(t, b.Cancelled)
	assert.False(t, b.Started)
	assert.False(t, b.Unfulfilled)

	ps, err := s.GetPolicyStatusFor(user, policy)

	assert.NoError(t, err)

	assert.Equal(t, int64(1), ps.CurrentBookings)
	assert.Equal(t, int64(0), ps.OldBookings)
	d, err := time.ParseDuration("10m0s")
	assert.NoError(t, err)
	assert.Equal(t, d, ps.Usage)

}

func TestGetPoliciesFor(t *testing.T) {

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)
	s := New()
	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	// booking details
	policy := "p-b"
	slot := "sl-b"
	user := "u-b" //does not yet exist in store
	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}

	// before we book, user does not exist
	_, err = s.GetPoliciesFor(user)
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())

	// make a booking
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }
	_, err = s.MakeBooking(policy, slot, user, when)
	assert.NoError(t, err)

	// check policy now listed for user
	p, err := s.GetPoliciesFor(user)
	assert.NoError(t, err)
	assert.Equal(t, []string{"p-b"}, p)

}

func TestStoreStatusAdminUser(t *testing.T) {

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)
	s := New()
	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }

	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}

	_, err = s.MakeBookingWithName("p-a", "sl-a", "user-a", when, "test00")
	_, err = s.MakeBookingWithName("p-b", "sl-b", "user-b", when, "test01")

	sa := s.GetStoreStatusAdmin()
	esa := StoreStatusAdmin{
		Locked:       false,
		Message:      "Welcome to the interval booking store",
		Now:          time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC),
		Bookings:     2,
		Descriptions: 16,
		Filters:      2,
		OldBookings:  0,
		Policies:     5,
		Resources:    3,
		Slots:        5,
		Streams:      3,
		UIs:          3,
		UISets:       3,
		Users:        2,
		Windows:      2}
	assert.Equal(t, esa, sa)

	su := s.GetStoreStatusUser()
	esu := StoreStatusUser{
		Locked:  false,
		Message: "Welcome to the interval booking store",
		Now:     time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC),
	}
	assert.Equal(t, esu, su)

}

func TestExportManifest(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	s := New()
	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	// make diary pointers nil as expected for exported version
	rm := make(map[string]Resource)
	for k, v := range m.Resources {
		rm[k] = Resource{
			ConfigURL:   v.ConfigURL,
			Description: v.Description,
			Streams:     v.Streams,
			TopicStub:   v.TopicStub,
		}
	}

	m.Resources = rm

	exportedm := s.ExportManifest()
	assert.Equal(t, m, exportedm)

}

// Note that complex types and slices are shallow copied so changes are visible
// to other tests. Since tests may eventually run in parallel, add a mutex
// All tests must restore any changes they make to the manifest
// Note :- the mutex might have been an over-reaction to a confusing
// test result .... but it's in there now.
type MutexManifest struct {
	*sync.Mutex
	Manifest Manifest
}

var testManifest = MutexManifest{
	&sync.Mutex{},
	Manifest{
		Descriptions: map[string]Description{
			"d-p-a": Description{
				Name:  "policy-a",
				Type:  "policy",
				Short: "a",
			},
			"d-p-b": Description{
				Name:  "policy-b",
				Type:  "policy",
				Short: "b",
			},
			"d-r-a": Description{
				Name:  "resource-a",
				Type:  "resource",
				Short: "a",
			},
			"d-r-b": Description{
				Name:  "resource-b",
				Type:  "resource",
				Short: "b",
			},
			"d-sl-a": Description{
				Name:  "slot-a",
				Type:  "slot",
				Short: "a",
			},
			"d-sl-b": Description{
				Name:  "slot-b",
				Type:  "slot",
				Short: "b",
			},
			"d-ui-a": Description{
				Name:  "ui-a",
				Type:  "ui",
				Short: "a",
			},
			"d-ui-b": Description{
				Name:  "ui-b",
				Type:  "ui",
				Short: "b",
			},
		},
		Policies: map[string]Policy{
			"p-a": Policy{
				Description: "d-p-a",
				Slots:       []string{"sl-a"},
			},
			"p-b": Policy{
				BookAhead:          time.Duration(2 * time.Hour),
				Description:        "d-p-b",
				EnforceBookAhead:   true,
				EnforceMaxBookings: true,
				EnforceMinDuration: true,
				EnforceMaxDuration: true,
				EnforceMaxUsage:    true,
				MaxUsage:           time.Duration(30 * time.Minute),
				MaxBookings:        2,
				MaxDuration:        time.Duration(10 * time.Minute),
				MinDuration:        time.Duration(5 * time.Minute),
				Slots:              []string{"sl-b"},
			},
		},
		Resources: map[string]Resource{
			"r-a": Resource{
				Description: "d-r-a",
				Streams:     []string{"st-a", "st-b"},
				TopicStub:   "aaaa00",
			},
			"r-b": Resource{
				Description: "d-r-b",
				Streams:     []string{"st-a", "st-b"},
				TopicStub:   "bbbb00",
			},
		},
		Slots: map[string]Slot{
			"sl-a": Slot{
				Description: "d-sl-a",
				Policy:      "p-a",
				Resource:    "r-a",
				UISet:       "us-a",
				Window:      "w-a",
			},
			"sl-b": Slot{
				Description: "d-sl-b",
				Policy:      "p-b",
				Resource:    "r-b",
				UISet:       "us-b",
				Window:      "w-b",
			},
		},
		Streams: map[string]Stream{
			"st-a": Stream{
				Audience:       "a",
				ConnectionType: "a",
				For:            "a",
				Scopes:         []string{"r", "w"},
				Topic:          "a",
				URL:            "a",
			},
			"st-b": Stream{
				Audience:       "b",
				ConnectionType: "b",
				For:            "b",
				Scopes:         []string{"r", "w"},
				Topic:          "b",
				URL:            "b",
			},
		},
		UIs: map[string]UI{
			"ui-a": UI{
				Description:     "d-ui-a",
				StreamsRequired: []string{"st-a", "st-b"},
				URL:             "a",
			},
			"ui-b": UI{
				Description:     "d-ui-b",
				StreamsRequired: []string{"st-a", "st-b"},
				URL:             "b",
			},
		},
		UISets: map[string]UISet{
			"us-a": UISet{
				UIs: []string{"ui-a"},
			},
			"us-b": UISet{
				UIs: []string{"ui-a", "ui-b"},
			},
		},
		Windows: map[string]Window{
			"w-a": Window{
				Allowed: []interval.Interval{
					interval.Interval{
						Start: time.Date(2022, 11, 4, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			"w-b": Window{
				Allowed: []interval.Interval{
					interval.Interval{
						Start: time.Date(2022, 11, 4, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	},
}

func TestCheckOKManifest(t *testing.T) {

	err, msg := checkManifest(testManifest.Manifest)

	assert.NoError(t, err)
	assert.Equal(t, []string{}, msg)
}

func TestCheckManifestCatchMissingUI(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()
	m := testManifest.Manifest

	m.UISets["us-b"].UIs[1] = "ui-c" //ui-c does not exist

	err, msg := checkManifest(m)

	assert.Error(t, err)
	assert.Equal(t, []string{"ui_set us-b references non-existent ui: ui-c"}, msg)

	//fix manifest for other tests
	m.UISets["us-b"].UIs[1] = "ui-b"

	err, _ = checkManifest(m)
	assert.NoError(t, err)
}

func TestCheckManifestCatchMissingResource(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	testManifest.Manifest.Resources["r-c"] = testManifest.Manifest.Resources["r-b"]
	delete(testManifest.Manifest.Resources, "r-b")

	err, msg := checkManifest(testManifest.Manifest)

	assert.Error(t, err)
	assert.Equal(t, []string{"slot sl-b references non-existent resource: r-b"}, msg)

	// fix manifest
	testManifest.Manifest.Resources["r-b"] = testManifest.Manifest.Resources["r-c"]
	delete(testManifest.Manifest.Resources, "r-c")

	err, _ = checkManifest(testManifest.Manifest)
	assert.NoError(t, err)
}

func TestCheckManifestCatchMissingDescriptions(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	dsla := testManifest.Manifest.Descriptions["d-sl-a"]
	delete(testManifest.Manifest.Descriptions, "d-sl-a")

	err, msg := checkManifest(testManifest.Manifest)

	assert.Error(t, err)
	assert.Equal(t, []string{"slot sl-a references non-existent description: d-sl-a"}, msg)

	//fix manifest for other tests
	testManifest.Manifest.Descriptions["d-sl-a"] = dsla
	err, _ = checkManifest(testManifest.Manifest)
	assert.NoError(t, err)

}

func TestCheckManifestCatchMissingStream(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	u := testManifest.Manifest.UIs["ui-b"]
	s := u.StreamsRequired
	u.StreamsRequired = []string{"st-c"}
	testManifest.Manifest.UIs["ui-b"] = u

	err, msg := checkManifest(testManifest.Manifest)

	assert.Error(t, err)
	assert.Equal(t, []string{"ui ui-b references non-existent stream: st-c"}, msg)

	//fix manifest for other tests
	u.StreamsRequired = s
	testManifest.Manifest.UIs["ui-b"] = u
	err, _ = checkManifest(testManifest.Manifest)
	assert.NoError(t, err)

}

func TestDeletePolicyAddPolicy(t *testing.T) {

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }

	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}

	_, err = s.MakeBookingWithName("p-a", "sl-a", "user-a", when, "test00")
	assert.NoError(t, err)

	_, err = s.MakeBookingWithName("p-b", "sl-b", "user-b", when, "test01")
	assert.NoError(t, err)

	bm := s.ExportBookings()
	assert.Equal(t, 2, len(bm))

	// check that deleting an unused policy and does not affect bookings
	// note that policy is known to store, so no error because delete from
	// map operation does not care whether item to be deleted existed
	err = s.DeletePolicyFor("user-a", "p-b")
	assert.NoError(t, err)

	bm = s.ExportBookings()
	assert.Equal(t, 2, len(bm))

	// check that deleting a used policy deletes associated booking test00 but keeps test01
	err = s.DeletePolicyFor("user-a", "p-a")
	assert.NoError(t, err)

	bm = s.ExportBookings()
	assert.Equal(t, 1, len(bm))
	_, ok := bm["test01"]
	assert.True(t, ok)

	um := s.ExportUsers()
	assert.Equal(t, []string{"p-b"}, um["user-b"].Policies)

	err = s.AddPolicyFor("user-b", "p-a")
	assert.NoError(t, err)
	um = s.ExportUsers()
	assert.NoError(t, err)

	//make a map of the responses to avoid ordering issues in checking the test
	epm := make(map[string]bool)
	epm["p-a"] = true
	epm["p-b"] = true

	apm := make(map[string]bool)
	for _, v := range um["user-b"].Policies {
		apm[v] = true
	}

	assert.Equal(t, epm, apm)

	// check the usage tracker has been initialised
	ps, err := s.GetPolicyStatusFor("user-b", "p-a")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), ps.Usage)

}

func TestPruneDiaries(t *testing.T) {
	s := New()
	s.pruneDiaries()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)
	s.pruneDiaries()

}

func TestGetPolicy(t *testing.T) {

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	p, err := s.GetPolicy("p-b")

	exp := Policy{
		BookAhead:     time.Duration(2 * time.Hour),
		Description:   "d-p-b",
		DisplayGuides: []string{"6m", "8m"},
		DisplayGuidesMap: map[string]DisplayGuide{
			"6m": DisplayGuide{
				BookAhead: time.Duration(1 * time.Hour),
				Duration:  time.Duration(6 * time.Minute),
				MaxSlots:  12,
			},
			"8m": DisplayGuide{
				BookAhead: time.Duration(2 * time.Hour),
				Duration:  time.Duration(8 * time.Minute),
				MaxSlots:  8,
			},
		},
		EnforceBookAhead:   true,
		EnforceMaxBookings: true,
		EnforceMinDuration: true,
		EnforceMaxDuration: true,
		EnforceMaxUsage:    true,
		MaxBookings:        2,
		MaxDuration:        time.Duration(10 * time.Minute),
		MinDuration:        time.Duration(5 * time.Minute),
		MaxUsage:           time.Duration(30 * time.Minute),
		Slots:              []string{"sl-b"},
	}

	assert.Equal(t, exp, p)

	if debug {
		y, err := yaml.Marshal(exp)
		assert.NoError(t, err)

		t.Log(string(y))
	}
}

func TestGetBooking(t *testing.T) {

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }

	policy := "p-b"
	slot := "sl-b"
	user := "test" //does not yet exist in store
	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}

	b, err := s.MakeBookingWithName(policy, slot, user, when, "test00")
	assert.NoError(t, err)

	b2, err := s.GetBooking("test00")
	assert.NoError(t, err)

	assert.Equal(t, b, b2)

	_, err = s.GetBooking("nosuchbooking")
	assert.Error(t, err)

}

func TestCalculateUsage(t *testing.T) {

	tEarly := time.Date(2022, 11, 4, 0, 0, 0, 0, time.UTC)
	tStart := time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC)
	tDuringGrace := time.Date(2022, 11, 5, 1, 3, 0, 0, time.UTC)
	tAutoGrace := time.Date(2022, 11, 5, 1, 5, 1, 0, time.UTC) //juuuust after the grace period to avoid equal_or_greater than comparison
	tAfterGrace := time.Date(2022, 11, 5, 1, 22, 0, 0, time.UTC)
	tAfterBooking := time.Date(2022, 11, 7, 1, 22, 0, 0, time.UTC)
	tEnd := time.Date(2022, 11, 5, 1, 30, 0, 0, time.UTC)
	w := interval.Interval{
		Start: tStart,
		End:   tEnd,
	}

	nograce := Policy{}
	grace := Policy{
		EnforceGracePeriod: true,
		GracePeriod:        time.Duration(5 * time.Minute),
		// make different to GracePeriod for checking correct member of struct is used
		GracePenalty: time.Duration(6 * time.Minute),
	}

	cancelledEarly := Booking{
		Cancelled:   true,
		CancelledAt: tEarly,
		When:        w,
	}

	// shouldn't be allowed to set this, but check we handle it correctly and charge for full usage
	cancelledLate := Booking{
		Cancelled:   true,
		CancelledAt: tAfterBooking,
		When:        w,
	}

	completed := Booking{
		Started: true,
		When:    w,
	}

	unfulfilled := Booking{
		// check that we do not charge for unfulfilled bookings that are incorrectly set as started
		Started:     true,
		Unfulfilled: true,
		When:        w,
	}

	noShow := Booking{
		Cancelled:   true,
		CancelledAt: tAutoGrace,
		When:        w,
	}

	cancelledDuringGraceUnstarted := Booking{
		Cancelled:   true,
		CancelledAt: tDuringGrace,
		When:        w,
	}
	cancelledDuringGraceStarted := Booking{
		Cancelled:   true,
		CancelledAt: tDuringGrace,
		Started:     true,
		When:        w,
	}
	cancelledAfterGraceUnstarted := Booking{
		Cancelled:   true,
		CancelledAt: tAfterGrace,
		When:        w,
	}
	cancelledAfterGraceStarted := Booking{
		Cancelled:   true,
		CancelledAt: tAfterGrace,
		Started:     true,
		When:        w,
	}

	tests := map[string]struct {
		booking Booking
		policy  Policy
		err     error
		minutes int
	}{
		"grace:completed":                       {completed, grace, nil, 30},
		"grace:unfulfilled":                     {unfulfilled, grace, nil, 0},
		"grace:cancelledEarly":                  {cancelledEarly, grace, nil, 0},
		"grace:cancelledLate":                   {cancelledLate, grace, nil, 30},
		"grace:noShow":                          {noShow, grace, nil, 11}, //penalty applied
		"grace:cancelledDuringGraceUnstarted":   {cancelledDuringGraceUnstarted, grace, nil, 5},
		"grace:cancelledDuringGraceStarted":     {cancelledDuringGraceStarted, grace, nil, 5},
		"grace:cancelledAfterGraceUnstarted":    {cancelledAfterGraceUnstarted, grace, nil, 11}, //auto-cancel will have happened
		"grace:cancelledAfterGraceStarted":      {cancelledAfterGraceStarted, grace, nil, 22},   //session ran for a while then user cancelled
		"nograce:completed":                     {completed, nograce, nil, 30},
		"nograce:unfulfilled":                   {unfulfilled, nograce, nil, 0},
		"nograce:cancelledEarly":                {cancelledEarly, nograce, nil, 0},
		"nograce:cancelledLate":                 {cancelledLate, nograce, nil, 30},
		"nograce:cancelledDuringGraceUnstarted": {cancelledDuringGraceUnstarted, nograce, nil, 3},
		"nograce:cancelledDuringGraceStarted":   {cancelledDuringGraceStarted, nograce, nil, 3},
		"nograce:cancelledAfterGraceUnstarted":  {cancelledAfterGraceUnstarted, nograce, nil, 22}, //no grace, no auto-cancel
		"nograce:cancelledAfterGraceStarted":    {cancelledAfterGraceStarted, nograce, nil, 22},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			usage, err := calculateUsage(tc.booking, tc.policy)

			want := time.Duration(time.Duration(tc.minutes) * time.Minute)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, want, usage)
		})
	}

}

func TestEnforceUnlimitedUsers(t *testing.T) {

	// derived from TestGetActivity, modified to have second user book at same time as first user
	// both should be able to get Activity successfully at the same time.

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC) }

	policy := "p-simulation"
	slot := "sl-simulation"
	user := "sim-user-0"
	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC),
		End:   time.Date(2022, 11, 5, 2, 10, 0, 0, time.UTC),
	}

	b0, err := s.MakeBooking(policy, slot, user, when)

	assert.NoError(t, err)

	assert.Equal(t, policy, b0.Policy)
	assert.Equal(t, slot, b0.Slot)
	assert.Equal(t, user, b0.User)
	assert.Equal(t, when, b0.When)
	assert.NotEqual(t, "", b0.Name) //non null name
	assert.False(t, b0.Cancelled)
	assert.False(t, b0.Started)
	assert.False(t, b0.Unfulfilled)

	// make second booking at same time for different user
	user = "sim-user-1"
	b1, err := s.MakeBooking(policy, slot, user, when)

	assert.NoError(t, err)

	assert.Equal(t, policy, b1.Policy)
	assert.Equal(t, slot, b1.Slot)
	assert.Equal(t, user, b1.User)
	assert.Equal(t, when, b1.When)
	assert.NotEqual(t, "", b1.Name) //non null name
	assert.False(t, b1.Cancelled)
	assert.False(t, b1.Started)
	assert.False(t, b1.Unfulfilled)

	// shift to time within booking
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 2, 02, 0, 0, time.UTC) }

	a0, err := s.GetActivity(b0)
	assert.NoError(t, err)
	a1, err := s.GetActivity(b1)
	assert.NoError(t, err)

	exp := Activity{
		Description: Description{
			Name:    "slot-simulation",
			Type:    "slot",
			Short:   "simulation",
			Long:    "",
			Further: "",
			Thumb:   "",
			Image:   ""},
		ConfigURL: "",
		Streams: map[string]Stream{
			"st-log": Stream{
				Audience:       "some_audience",
				ConnectionType: "session",
				For:            "log",
				Scopes:         []string{"r", "w"},
				Topic:          "simu00-st-log",
				URL:            "some_url"}},
		UIs: []UIDescribed{
			UIDescribed{
				Description: Description{
					Name:    "ui-simulation",
					Type:    "ui",
					Short:   "simulation",
					Long:    "",
					Further: "",
					Thumb:   "",
					Image:   ""},
				URL:             "https://some_url.org",
				StreamsRequired: []string{"st-log"},
			}},
		NotBefore: time.Date(2022, time.November, 5, 2, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2022, time.November, 5, 2, 10, 0, 0, time.UTC),
	}

	assert.Equal(t, exp, a0)
	assert.Equal(t, exp, a1)
}

func TestAllowStartInPast(t *testing.T) {

	// derived from TestGetActivity, modified to have second user book at same time as first user
	// both should be able to get Activity successfully at the same time.

	s := New()

	// fix time for ease of checking results
	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }

	m := Manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	assert.NoError(t, err)

	err = s.ReplaceManifest(m)
	assert.NoError(t, err)

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 0, 30, 0, time.UTC) } // move forward 30sec in time

	policy := "p-instant"
	slot := "sl-instant"
	user := "user-0"
	when := interval.Interval{
		Start: time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC), //now 30 sec in the past
		End:   time.Date(2022, 11, 5, 0, 10, 0, 0, time.UTC),
	}

	_, err = s.MakeBooking(policy, slot, user, when)

	assert.Error(t, err)

	assert.Equal(t, "booking cannot start in the past", err.Error())

	policy = "p-start-in-past"
	slot = "sl-start-in-past"
	user = "user-0"

	b, err := s.MakeBooking(policy, slot, user, when)

	if err != nil {
		t.Log(err.Error())
	}

	assert.Equal(t, policy, b.Policy)
	assert.Equal(t, slot, b.Slot)
	assert.Equal(t, user, b.User)
	assert.Equal(t, when, b.When)
	assert.NotEqual(t, "", b.Name) //non null name
	assert.False(t, b.Cancelled)
	assert.False(t, b.Started)
	assert.False(t, b.Unfulfilled)

	s.Now = func() time.Time { return time.Date(2022, 11, 5, 0, 2, 0, 0, time.UTC) } // move forward 2min  in time, outside allowed window for starting booking in the past

	_, err = s.MakeBooking(policy, slot, user, when)

	assert.Error(t, err)

	assert.Equal(t, "booking cannot start more than 1m0s in the past", err.Error())

}
