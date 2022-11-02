// package resource holds non-overlapping bookings with arbitrary durations
package resource

import (
	"errors"
	"sync"
	"time"

	avl "github.com/timdrysdale/interval/internal/trees/avltree"

	"github.com/google/uuid"
	"github.com/timdrysdale/interval/interval"
)

// Resource represents the bookings of a resources
type Resource struct {
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
func New(name string) *Resource {
	return &Resource{
		&sync.RWMutex{},
		name,
		avl.NewWith(interval.Comparator),
		true,
		"",
	}
}

// Delete removes a booking, if it exists
func (r *Resource) Delete(delete uuid.UUID) error {

	r.Lock()
	defer r.Unlock()

	slots := r.bookings.Keys() //these are given in order
	IDs := r.bookings.Values()

	for idx, ID := range IDs {
		if delete == ID {
			r.bookings.Remove(slots[idx])
			return nil
		}
	}

	return errors.New("ID not found")

}

func (r *Resource) SetUnavailable(reason string) {
	r.available = false
	r.status = reason
}

func (r *Resource) SetAvailable(reason string) {
	r.available = true
	r.status = reason
}

// Request returns a booking, if it can be made
func (r *Resource) Request(when interval.Interval) (uuid.UUID, error) {
	r.Lock()
	defer r.Unlock()

	if ok, msg := r.IsAvailable(); !ok {
		return [16]byte{}, errors.New(msg)
	}

	u := uuid.New()

	_, err := r.bookings.Put(when, u)

	if err != nil {
		//return a zero-value UUID if there is an error
		return [16]byte{}, err
	}

	return u, err

}

// GetCount returns the number of live bookings
func (r *Resource) GetCount() int {
	r.RLock()
	defer r.RUnlock()
	return r.bookings.Size()
}

// GetBookings returns all bookings
func (r *Resource) GetBookings() ([]Booking, error) {
	r.RLock()
	defer r.RUnlock()
	b := []Booking{}

	slots := r.bookings.Keys() //these are given in order
	IDs := r.bookings.Values()

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

func (r *Resource) IsAvailable() (bool, string) {

	if r.available {
		return true, r.status
	}

	msg := "Unavailable"

	if r.status != "" {
		msg += " (" + r.status + ")"
	}

	return false, msg
}

// ValidateBooking checks if a given booking matches an existing booking
// Returns false if the resource is not available so that it can be
// used as a check on whether to supply connection info to user
// (don't if resource if not available)
func (r *Resource) ValidateBooking(b Booking) (bool, error) {

	id, found := r.bookings.Get(b.When)

	if !found {
		return false, errors.New("Not Found")
	}

	if ok, msg := r.IsAvailable(); !ok {
		return false, errors.New(msg)
	}

	if id != b.ID {
		return false, errors.New("ID mismatch")
	}

	return true, nil

}

// ClearBefore removes all old bookings
func (r *Resource) ClearBefore(t time.Time) {
	r.Lock()
	defer r.Unlock()
	slots := r.bookings.Keys() //these are given in order

	for _, when := range slots {
		if when.(interval.Interval).End.Before(t) {
			r.bookings.Remove(when)
		}
	}
}
