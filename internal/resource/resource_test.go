package resource

import (
	"interval/internal/interval"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestRequest(t *testing.T) {

	r := New()

	// request first interval - must succeed
	ua, err := r.Request(a)
	assert.NoError(t, err)
	assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", ua.String())

	// repeat request - must fail
	u, err := r.Request(a)
	assert.Error(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", u.String())

	// request a different non-overlapping interval - must succeed
	ub, err := r.Request(b)
	assert.NoError(t, err)
	assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", ub.String())

	// request a partly overlapping interval with a - must fail
	u, err = r.Request(c)
	assert.Error(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", u.String())

	bookings, err := r.GetBookings()
	assert.NoError(t, err)
	assert.Equal(t, bookings[0].When.Start, a.Start)
	assert.Equal(t, bookings[0].ID, ua)
	assert.Equal(t, bookings[1].When.Start, b.Start)
	assert.Equal(t, bookings[1].ID, ub)

}
