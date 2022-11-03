package diary

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/timdrysdale/interval/interval"
)

var w = time.Now()

var a = interval.Interval{
	Start: w,
	End:   w.Add(5 * time.Second),
}

// does not overlap a
var b = interval.Interval{
	Start: w.Add(10 * time.Second),
	End:   w.Add(20 * time.Second),
}

// overlaps a
var c = interval.Interval{
	Start: w.Add(3 * time.Second),
	End:   w.Add(12 * time.Second),
}

func TestIsAvailable(t *testing.T) {

	d := New("test")

	ok, msg := d.IsAvailable()

	assert.True(t, ok)
	assert.Equal(t, msg, "")

	// request first interval - must succeed
	ua, err := d.Request(a)
	assert.NoError(t, err)
	assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", ua.String())

	d.SetUnavailable("Offline")

	// request a different non-overlapping interval
	// would succeed if available but must fail because unavailable
	ub, err := d.Request(b)
	assert.Error(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", ub.String())

	ok, msg = d.IsAvailable()

	assert.False(t, ok)
	assert.Equal(t, msg, "Unavailable (Offline)")

}

func TestBooking(t *testing.T) {

	d := New("test")

	// request first interval - must succeed
	ua, err := d.Request(a)
	assert.NoError(t, err)
	assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", ua.String())

	// repeat request - must fail
	u, err := d.Request(a)
	assert.Error(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", u.String())

	// request a different non-overlapping interval - must succeed
	ub, err := d.Request(b)
	assert.NoError(t, err)
	assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", ub.String())

	// request a partly overlapping interval with a - must fail
	u, err = d.Request(c)
	assert.Error(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", u.String())

	// Get current bookings
	bookings, err := d.GetBookings()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(bookings))
	assert.Equal(t, bookings[0].When.Start, a.Start)
	assert.Equal(t, bookings[0].ID, ua)
	assert.Equal(t, bookings[1].When.Start, b.Start)
	assert.Equal(t, bookings[1].ID, ub)

	// Delete a booking
	assert.Equal(t, 2, d.GetCount())
	err = d.Delete(ua)
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.Equal(t, 1, d.GetCount())
	bookings, err = d.GetBookings()
	assert.NoError(t, err)
	assert.Equal(t, bookings[0].When.Start, b.Start)
	assert.Equal(t, bookings[0].ID, ub)

	// add another booking back for testing clear before
	_, err = d.Request(a)
	assert.NoError(t, err)
	assert.Equal(t, 2, d.GetCount())
	// clear from before a time in the middle of a booking - must keep that booking
	d.ClearBefore(w.Add(3 * time.Second))
	assert.NoError(t, err)
	assert.Equal(t, 2, d.GetCount())
	// clear the first booking only
	d.ClearBefore(w.Add(6 * time.Second))
	assert.NoError(t, err)
	assert.Equal(t, 1, d.GetCount())
	bookings, err = d.GetBookings()
	assert.NoError(t, err)
	assert.Equal(t, bookings[0].When.Start, b.Start)
	assert.Equal(t, bookings[0].ID, ub)

}

func TestValidateBooking(t *testing.T) {

	d := New("test")

	ok, msg := d.IsAvailable()

	assert.True(t, ok)
	assert.Equal(t, msg, "")

	// request first interval - must succeed
	ua, err := d.Request(a)
	assert.NoError(t, err)
	assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", ua.String())

	// Booking is valid
	ok, err = d.ValidateBooking(Booking{
		When: a,
		ID:   ua,
	})

	assert.True(t, ok)
	assert.Equal(t, nil, err)

	// Check invalid if interval is not present
	ok, err = d.ValidateBooking(Booking{
		When: b, //this interval not present
		ID:   ua,
	})
	assert.False(t, ok)
	assert.Equal(t, "Not Found", err.Error())

	// Check invalid if ID and interval from different bookings
	// add a second booking to do this check
	ub, err := d.Request(b)
	assert.NoError(t, err)
	assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", ub.String())

	ok, err = d.ValidateBooking(Booking{
		When: a,  //this interval from first booking
		ID:   ub, //this id from second booking
	})

	// Make booking invalid by setting machine unavailable
	d.SetUnavailable("Offline")

	ok, err = d.ValidateBooking(Booking{
		When: a,
		ID:   ua,
	})

	assert.False(t, ok)
	assert.Equal(t, err.Error(), "Unavailable (Offline)")

}

func TestName(t *testing.T) {
	d := New("test")
	assert.Equal(t, d.Name, "test")
}
