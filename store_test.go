package interval

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/timdrysdale/interval/diary"
	"github.com/timdrysdale/interval/interval"
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
  d-r-a:
    name: resource-a
    type: resource
    short: a
  d-r-b:
    name: resource-b
    type: resource
    short: b
  d-sl-a:
    name: slot-a
    type: slot
    short: a
  d-sl-b:
    name: slot-b
    type: slot
    short: b
  d-ui-a:
    name: ui-a
    type: ui
    short: a
  d-ui-b:
    name: ui-b
    type: ui
    short: b
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
    topic_stub: aaaa00
  r-b:
    description: d-r-b
    streams:
    - st-a
    - st-b
    topic_stub: bbbb00
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
	err = s.CancelBooking(bc)
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
	err = s.CancelBooking(fake)
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
	err = s.CancelBooking(b)
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

	err, msg := s.CheckBooking(b)

	assert.NoError(t, err)
	assert.Equal(t, []string{}, msg)

	b.Policy = ""
	err, msg = s.CheckBooking(b)
	assert.Error(t, err)
	assert.Equal(t, []string{b.Name + " missing policy"}, msg)
	b.Policy = "foo"
	err, msg = s.CheckBooking(b)
	assert.Error(t, err)
	assert.Equal(t, []string{b.Name + " policy foo not found"}, msg)
	b.Policy = policy

	b.Slot = ""
	err, msg = s.CheckBooking(b)
	assert.Error(t, err)
	assert.Equal(t, []string{b.Name + " missing slot"}, msg)
	b.Slot = "foo"
	err, msg = s.CheckBooking(b)
	assert.Error(t, err)
	assert.Equal(t, []string{b.Name + " slot foo not found"}, msg)
	b.Slot = slot

	b.User = ""
	err, msg = s.CheckBooking(b)
	assert.Error(t, err)
	assert.Equal(t, []string{b.Name + " missing user"}, msg)
	// no need to check for user not found - this is ok, as
	// they are created as needed when bookings are made

	b.User = user

	name := b.Name
	b.Name = ""
	err, msg = s.CheckBooking(b)
	assert.Error(t, err)
	assert.Equal(t, []string{"missing name"}, msg)
	b.Name = name

	b.When = interval.Interval{}
	err, msg = s.CheckBooking(b)
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

	s.PruneBookings()

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

	s.PruneBookings()

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
	s.PruneBookings()

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
		Descriptions: 8,
		Filters:      2,
		OldBookings:  0,
		Policies:     2,
		Resources:    2,
		Slots:        2,
		Streams:      2,
		UIs:          2,
		UISets:       2,
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

	err, msg := CheckManifest(testManifest.Manifest)

	assert.NoError(t, err)
	assert.Equal(t, []string{}, msg)
}

func TestCheckManifestCatchMissingUI(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()
	m := testManifest.Manifest

	m.UISets["us-b"].UIs[1] = "ui-c" //ui-c does not exist

	err, msg := CheckManifest(m)

	assert.Error(t, err)
	assert.Equal(t, []string{"ui_set us-b references non-existent ui: ui-c"}, msg)

	//fix manifest for other tests
	m.UISets["us-b"].UIs[1] = "ui-b"

	err, _ = CheckManifest(m)
	assert.NoError(t, err)
}

func TestCheckManifestCatchMissingResource(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	testManifest.Manifest.Resources["r-c"] = testManifest.Manifest.Resources["r-b"]
	delete(testManifest.Manifest.Resources, "r-b")

	err, msg := CheckManifest(testManifest.Manifest)

	assert.Error(t, err)
	assert.Equal(t, []string{"slot sl-b references non-existent resource: r-b"}, msg)

	// fix manifest
	testManifest.Manifest.Resources["r-b"] = testManifest.Manifest.Resources["r-c"]
	delete(testManifest.Manifest.Resources, "r-c")

	err, _ = CheckManifest(testManifest.Manifest)
	assert.NoError(t, err)
}

func TestCheckManifestCatchMissingDescriptions(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	dsla := testManifest.Manifest.Descriptions["d-sl-a"]
	delete(testManifest.Manifest.Descriptions, "d-sl-a")

	err, msg := CheckManifest(testManifest.Manifest)

	assert.Error(t, err)
	assert.Equal(t, []string{"slot sl-a references non-existent description: d-sl-a"}, msg)

	//fix manifest for other tests
	testManifest.Manifest.Descriptions["d-sl-a"] = dsla
	err, _ = CheckManifest(testManifest.Manifest)
	assert.NoError(t, err)

}

func TestCheckManifestCatchMissingStream(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	u := testManifest.Manifest.UIs["ui-b"]
	s := u.StreamsRequired
	u.StreamsRequired = []string{"st-c"}
	testManifest.Manifest.UIs["ui-b"] = u

	err, msg := CheckManifest(testManifest.Manifest)

	assert.Error(t, err)
	assert.Equal(t, []string{"ui ui-b references non-existent stream: st-c"}, msg)

	//fix manifest for other tests
	u.StreamsRequired = s
	testManifest.Manifest.UIs["ui-b"] = u
	err, _ = CheckManifest(testManifest.Manifest)
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
