// package resource holds non-overlapping bookings with arbitrary durations
package diary

import (
	"errors"
	"sync"
	"time"

	avl "github.com/timdrysdale/interval/internal/trees/avltree"

	"github.com/google/uuid"
	"github.com/timdrysdale/interval/interval"
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
	ID   uuid.UUID
}

// New creates a new resource with no bookings
func New(name string) *Diary {
	return &Diary{
		&sync.RWMutex{},
		name,
		avl.NewWith(interval.Comparator),
		true,
		"",
	}
}

// Delete removes a booking, if it exists
func (d *Diary) Delete(delete uuid.UUID) error {

	d.Lock()
	defer d.Unlock()

	slots := d.bookings.Keys() //these are given in order
	IDs := d.bookings.Values()

	for idx, ID := range IDs {
		if delete == ID {
			d.bookings.Remove(slots[idx])
			return nil
		}
	}

	return errors.New("ID not found")

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
func (d *Diary) Request(when interval.Interval) (uuid.UUID, error) {

	u := uuid.New()

	err := d.RequestWithID(when, u)

	if err != nil {
		return [16]byte{}, err
	}

	return u, nil
}

// Request returns a booking, if it can be made
// using a given uuid
func (d *Diary) RequestWithID(when interval.Interval, u uuid.UUID) error {
	d.Lock()
	defer d.Unlock()

	if ok, msg := d.IsAvailable(); !ok {
		return errors.New(msg)
	}

	_, err := d.bookings.Put(when, u)

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
	IDs := d.bookings.Values()

	if len(slots) != len(IDs) {
		return b, errors.New("number of slots and IDs are not the same")
	}

	for idx, when := range slots {
		b = append(b, Booking{
			When: when.(interval.Interval),
			ID:   (IDs[idx]).(uuid.UUID),
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

	id, found := d.bookings.Get(b.When)

	if !found {
		return false, errors.New("not found")
	}

	if ok, msg := d.IsAvailable(); !ok {
		return false, errors.New(msg)
	}

	if id != b.ID {
		return false, errors.New("ID mismatch")
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
