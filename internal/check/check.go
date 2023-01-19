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
	now         func() time.Time
	Name        string
}

func New() *Checker {
	log.Debugf("New Checker")
	return &Checker{
		&sync.Mutex{},
		[]time.Time{},
		make(map[time.Time][]string),
		func() time.Time { return time.Now() },
		"New",
	}
}

func (c *Checker) SetNow(now func() time.Time) {
	c.Lock()
	defer c.Unlock()
	c.now = now
}

// For external use only
func (c *Checker) Now() time.Time {
	c.Lock()
	defer c.Unlock()
	return c.now()
}

func (c *Checker) Clean() {
	c.Lock()
	defer c.Unlock()
	c.Times = []time.Time{}
	c.Values = make(map[time.Time][]string)
}

func (c *Checker) WithName(name string) *Checker {
	c.Lock()
	defer c.Unlock()
	c.Name = name
	return c
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

				v := c.GetExpired()
				if len(v) > 0 {
					log.Infof("Expired %d bookings", len(v))
					expired <- v
				}
			}
		}
	}()
}

func (c *Checker) WithNow(now func() time.Time) *Checker {
	c.Lock()
	defer c.Unlock()
	c.now = now
	return c
}

func (c *Checker) Push(t time.Time, v string) error {
	log.Debugf("awaiting lock to add booking %s to cancellation check list", v)
	c.Lock()
	defer c.Unlock()
	log.Debugf("adding booking %s to cancellation check list", v)
	// time must be in the future
	if t.Before(c.now()) { //use internal version to avoid hang over locks
		return errors.New("time is in the past")
	}

	//check if we already have this time?
	if _, ok := c.Values[t]; !ok {
		log.Debugf("checker new time")
		c.Times = insertSorted(c.Times, t)
		c.Values[t] = []string{v}
	} else {
		values := c.Values[t]
		values = append(values, v)
		c.Values[t] = values
	}
	log.Debugf("Checker(%s) has %d times and %d values", c.Name, len(c.Times), len(c.Values))
	return nil
}

func (c *Checker) GetExpired() []string {

	c.Lock()
	defer c.Unlock()

	log.Debugf("Checker(%s) grace checking %d bookings", c.Name, len(c.Times))

	expired := []string{}
	toDelete := []time.Time{}

	expiredIdx := -1

	for idx, t := range c.Times {
		if t.Before(c.now()) { //use internal version to avoid hang over locks
			log.Debugf("checker: index %d is expired at time %s", idx, c.now())
			expiredIdx = idx
			if values, ok := c.Values[t]; ok {
				for _, v := range values {
					expired = append(expired, v)
				}
				toDelete = append(toDelete, t)
			}
		} else {

			log.Debugf("checker: index %d is ok at time %s", idx, c.now())

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
