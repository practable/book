package diary

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/timdrysdale/interval/internal/interval"
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
	assert.Equal(t, msg, "new")

	// request first interval - must succeed
	err := d.Request(a, "test00")
	assert.NoError(t, err)

	d.SetUnavailable("Offline")

	// request a different non-overlapping interval
	// would succeed if available but must fail because unavailable
	err = d.Request(b, "test01")
	assert.Error(t, err)

	ok, msg = d.IsAvailable()

	assert.False(t, ok)
	assert.Equal(t, msg, "unavailable because Offline")

}

func TestBooking(t *testing.T) {

	d := New("test")

	// request first interval - must succeed
	err := d.Request(a, "test00")
	assert.NoError(t, err)

	// repeat request - must fail
	err = d.Request(a, "test01")
	assert.Error(t, err)

	// request a different non-overlapping interval - must succeed
	err = d.Request(b, "test02")
	assert.NoError(t, err)

	// request a partly overlapping interval with a - must fail
	err = d.Request(c, "test03")
	assert.Error(t, err)

	// Get current bookings
	bookings, err := d.GetBookings()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(bookings))
	assert.Equal(t, bookings[0].When.Start, a.Start)
	assert.Equal(t, bookings[0].Name, "test00")
	assert.Equal(t, bookings[1].When.Start, b.Start)
	assert.Equal(t, bookings[1].Name, "test02")

	// Delete a booking
	assert.Equal(t, 2, d.GetCount())
	err = d.Delete("test00")
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.Equal(t, 1, d.GetCount())
	bookings, err = d.GetBookings()
	assert.NoError(t, err)
	assert.Equal(t, bookings[0].When.Start, b.Start)
	assert.Equal(t, bookings[0].Name, "test02")

	// add another booking back for testing clear before
	err = d.Request(a, "test04")
	assert.NoError(t, err)
	assert.Equal(t, 2, d.GetCount())
	// clear from before a time in the middle of a booking - must keep that booking
	d.ClearBefore(w.Add(3 * time.Second))
	assert.NoError(t, err)
	assert.Equal(t, 2, d.GetCount())
	// clear the first booking only (i.e. clear test00)
	d.ClearBefore(w.Add(6 * time.Second))
	assert.NoError(t, err)
	assert.Equal(t, 1, d.GetCount())
	bookings, err = d.GetBookings()
	assert.NoError(t, err)
	assert.Equal(t, bookings[0].When.Start, b.Start)
	assert.Equal(t, "test02", bookings[0].Name)

}

func TestValidateBooking(t *testing.T) {

	d := New("test")

	ok, msg := d.IsAvailable()

	assert.True(t, ok)
	assert.Equal(t, msg, "new")

	// request first interval - must succeed
	err := d.Request(a, "test00")
	assert.NoError(t, err)

	// Booking is valid
	ok, err = d.ValidateBooking(Booking{
		When: a,
		Name: "test00",
	})

	assert.True(t, ok)
	assert.Equal(t, nil, err)

	// Check invalid if interval is not present
	ok, err = d.ValidateBooking(Booking{
		When: b, //this interval not present
		Name: "test00",
	})
	assert.False(t, ok)
	assert.Equal(t, "not found", err.Error())

	// Check invalid if ID and interval from different bookings
	// add a second booking to do this check
	err = d.Request(b, "test01")
	assert.NoError(t, err)

	ok, err = d.ValidateBooking(Booking{
		When: a, //this interval from first booking
		Name: "test01",
	})

	// Make booking invalid by setting machine unavailable
	d.SetUnavailable("offline")

	ok, err = d.ValidateBooking(Booking{
		When: a,
		Name: "test00",
	})

	assert.False(t, ok)
	assert.Equal(t, err.Error(), "unavailable because offline")

}

func TestName(t *testing.T) {
	d := New("test")
	assert.Equal(t, d.Name, "test")
}

func TestDenySameNameBooking(t *testing.T) {

	d := New("test")

	// request first interval - must succeed
	err := d.Request(a, "test00")
	assert.NoError(t, err)

	// request different interval with same name - must fail
	err = d.Request(b, "test00")
	assert.Error(t, err)
	assert.Equal(t, "name already in use", err.Error())

}
