package check

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var currentTime *time.Time
var ct time.Time
var now func() time.Time

func init() {

	ct = time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	currentTime = &ct

	now = func() time.Time {
		return *currentTime

	}
}

func TestNew(t *testing.T) {

	c := New().WithNow(now)

	assert.Equal(t, ct, c.Now())

}

func TestPush(t *testing.T) {

	c := New().WithNow(now)

	t1 := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)

	c.Push(t1, "test")

	et := []time.Time{t1}

	assert.Equal(t, et, c.Times)

	ev := make(map[time.Time][]string)
	ev[t1] = []string{"test"}

	assert.Equal(t, ev, c.Values)

	c.Push(t1, "foo")

	// check we don't have doublers
	et = []time.Time{t1}
	assert.Equal(t, et, c.Times)

	ev[t1] = []string{"test", "foo"}
	assert.Equal(t, ev, c.Values)

}

func TestExpired(t *testing.T) {

	c := New().WithNow(now)

	got := c.GetExpired()
	want := []string{}
	assert.Equal(t, got, want)

	t1 := time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC)

	c.Push(t1, "test")
	c.Push(t1, "foo")

	t2 := time.Date(2022, 11, 5, 2, 0, 0, 0, time.UTC)

	t3 := time.Date(2022, 11, 5, 3, 0, 0, 0, time.UTC)

	c.Push(t3, "thing")
	c.Push(t3, "bar")

	t4 := time.Date(2022, 11, 5, 4, 0, 0, 0, time.UTC)

	// move time forward to t2

	currentTime = &t2

	got = c.GetExpired()
	want = []string{"test", "foo"}
	assert.Equal(t, got, want)

	currentTime = &t4
	got = c.GetExpired()
	want = []string{"thing", "bar"}
	assert.Equal(t, got, want)

	t5 := time.Date(2022, 11, 5, 5, 0, 0, 0, time.UTC)

	currentTime = &t5

	got = c.GetExpired()
	want = []string{}
	assert.Equal(t, got, want)

	// check we are removing old values from memory
	assert.Equal(t, len(c.Times), 0)
	assert.Equal(t, len(c.Values), 0)
}

func TestRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	c := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expired := make(chan []string)
	d := time.Duration(10 * time.Millisecond)

	c.Run(ctx, d, expired)

	t0 := time.Now()
	t1 := t0.Add(time.Duration(100 * time.Millisecond))
	c.Push(t1, "test")
	got := <-expired
	t2 := time.Now()

	assert.Equal(t, []string{"test"}, got)
	assert.Less(t, time.Duration(90*time.Millisecond), t2.Sub(t0))
	assert.Greater(t, time.Duration(110*time.Millisecond), t2.Sub(t0))

}
