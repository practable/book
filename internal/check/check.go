package check

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Checker struct {
	*sync.Mutex `json:"-" yaml:"-"`
	Times       []time.Time
	Values      map[time.Time][]string
	Period      time.Duration
	Now         func() time.Time
}

func New() *Checker {

	return &Checker{
		&sync.Mutex{},
		[]time.Time{},
		make(map[time.Time][]string),
		time.Duration(time.Minute),
		func() time.Time { return time.Now() },
	}
}

func (c *Checker) Run(ctx context.Context, checkEvery time.Duration, expired chan []string) {

	go func() {
		log.Debug("checker will check expiry every " + checkEvery.String())
		for {

			select {
			case <-ctx.Done():
				log.Trace("checker stopped permanently")
				return
			case <-time.After(checkEvery):
				log.Trace("checker checking expiry at time " + c.Now().String())
				expired <- c.GetExpired()
			}
		}
	}()
}

func (c *Checker) WithNow(now func() time.Time) *Checker {
	c.Now = now
	return c
}

func (c *Checker) WithPeriod(period time.Duration) *Checker {
	c.Period = period
	return c
}

func (c *Checker) Push(t time.Time, v string) error {
	c.Lock()
	defer c.Unlock()

	// time must be in the future
	if t.Before(c.Now()) {
		return errors.New("time is in the past")
	}

	c.Times = insertSorted(c.Times, t)

	if _, ok := c.Values[t]; !ok {
		c.Values[t] = []string{v}
	} else {
		vv := c.Values[t]
		vv = append(vv, v)
		c.Values[t] = vv
	}

	return nil
}

func (c *Checker) GetExpired() []string {

	c.Lock()
	defer c.Unlock()

	expired := []string{}
	toDelete := []time.Time{}

	expiredIdx := -1

	for idx, t := range c.Times {
		if t.Before(c.Now()) {
			expiredIdx = idx
			if values, ok := c.Values[t]; ok {
				for _, v := range values {
					expired = append(expired, v)
				}
				toDelete = append(toDelete, t)
			}
		}
	}

	if expiredIdx > -1 {
		c.Times = c.Times[(expiredIdx + 1):]
	}

	for _, t := range toDelete {
		delete(c.Values, t)
	}

	return expired
}

//modified from int version here: https://stackoverflow.com/questions/42746972/golang-insert-to-a-sorted-slice

// insertAt inserts v into s at index i and returns the new slice.
func insertAt(data []time.Time, i int, v time.Time) []time.Time {
	if i == len(data) {
		// Insert at end is the easy case.
		return append(data, v)
	}

	// Make space for the inserted element by shifting
	// values at the insertion index up one index. The call
	// to append does not allocate memory when cap(data) is
	// greater ​than len(data).
	data = append(data[:i+1], data[i:]...)

	// Insert the new element.
	data[i] = v

	// Return the updated slice.
	return data
}

func insertSorted(data []time.Time, t time.Time) []time.Time {
	i := sort.Search(len(data), func(i int) bool { return data[i].Equal(t) || data[i].After(t) })
	return insertAt(data, i, t)
}
