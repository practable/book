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
      end: 2022-11-06T00:00:00Z
    denied: []
  w-b:
    allowed:
    - start: 2022-11-04T00:00:00Z
      end: 2022-11-06T00:00:00Z
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

	b, err := s.MakeBooking(policy, slot, user, when)

	assert.NoError(t, err)

	assert.Equal(t, policy, b.Policy)
	assert.Equal(t, slot, b.Slot)
	assert.Equal(t, user, b.User)
	assert.Equal(t, when, b.When)
	assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", b.ID.String()) //non null ID
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
	s.CancelBooking(bc)
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

}
