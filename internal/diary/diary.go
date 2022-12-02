// package resource holds non-overlapping bookings with arbitrary durations
package diary

import (
	"errors"
	"sync"
	"time"

	avl "github.com/timdrysdale/interval/internal/trees/avltree"

	"github.com/timdrysdale/interval/internal/interval"
)

// Diary represents the bookings of a resources
type Diary struct {
	*sync.RWMutex `json:"-"`
	Name          string // unique, persistant name
	bookings      *avl.Tree
	available     bool   // must be true to be booked - we don't know when it might be available again
	status        string // optional status message, to explain lack of availability
}

// Booking represents a booking. This is not used internally, it's just for
// returning booking info from GetBookings() and
// validating bookings with ValidateBooking()
type Booking struct {
	When interval.Interval
	Name string
}

// New creates a new resource with no bookings
func New(name string) *Diary {
	return &Diary{
		&sync.RWMutex{},
		name,
		avl.NewWith(interval.Comparator),
		true,
		"new",
	}
}

// Delete removes a booking, if it exists
func (d *Diary) Delete(delete string) error {

	d.Lock()
	defer d.Unlock()

	slots := d.bookings.Keys() //these are given in order
	Names := d.bookings.Values()

	for idx, Name := range Names {
		if delete == Name {
			d.bookings.Remove(slots[idx])
			return nil
		}
	}

	return errors.New("not found")

}

func (d *Diary) SetUnavailable(reason string) {
	d.available = false
	d.status = reason
}

func (d *Diary) SetAvailable(reason string) {
	d.available = true
	d.status = reason
}

// Request returns a booking, if it can be made
// name must be specified
func (d *Diary) Request(when interval.Interval, name string) error {

	if name == "" {
		return errors.New("must not have empty name")
	}

	bs, err := d.GetBookings()

	if err != nil {
		return err
	}

	d.Lock()
	defer d.Unlock()

	for _, b := range bs {
		if b.Name == name {
			return errors.New("name already in use")
		}
	}

	if ok, msg := d.IsAvailable(); !ok {
		return errors.New(msg)
	}

	_, err = d.bookings.Put(when, name)

	return err
}

// GetCount returns the number of live bookings
func (d *Diary) GetCount() int {
	d.RLock()
	defer d.RUnlock()
	return d.bookings.Size()
}

// GetBookings returns all bookings
func (d *Diary) GetBookings() ([]Booking, error) {
	d.RLock()
	defer d.RUnlock()
	b := []Booking{}

	slots := d.bookings.Keys() //these are given in order
	names := d.bookings.Values()

	if len(slots) != len(names) { //prevent segfault in next step
		return b, errors.New("diary consistency error")
	}

	for idx, when := range slots {
		b = append(b, Booking{
			When: when.(interval.Interval),
			Name: (names[idx]).(string),
		})
	}

	return b, nil

}

func (d *Diary) IsAvailable() (bool, string) {

	if d.available {
		return true, d.status
	}

	msg := "unavailable"

	if d.status != "" {
		msg += " because " + d.status
	}

	return false, msg
}

// ValidateBooking checks if a given booking matches an existing booking
// Returns false if the resource is not available so that it can be
// used as a check on whether to supply connection info to user
// (don't if resource if not available)
func (d *Diary) ValidateBooking(b Booking) (bool, error) {

	name, found := d.bookings.Get(b.When)

	if !found {
		return false, errors.New("not found")
	}

	if ok, msg := d.IsAvailable(); !ok {
		return false, errors.New(msg)
	}

	if name != b.Name {
		return false, errors.New("name mismatch")
	}

	return true, nil

}

// ClearBefore removes all old bookings
func (d *Diary) ClearBefore(t time.Time) {
	d.Lock()
	defer d.Unlock()
	slots := d.bookings.Keys() //these are given in order

	for _, when := range slots {
		if when.(interval.Interval).End.Before(t) {
			d.bookings.Remove(when)
		}
	}
}
