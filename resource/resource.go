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
	bookings      *avl.Tree
}

// Booking represents a booking. This is not used internally, it's just for
// returning booking info from GetBookings()
type Booking struct {
	When interval.Interval
	ID   uuid.UUID
}

// New creates a new resource with no bookings
func New() *Resource {
	return &Resource{
		&sync.RWMutex{},
		avl.NewWith(interval.Comparator),
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

// Request returns a booking, if it can be made
func (r *Resource) Request(when interval.Interval) (uuid.UUID, error) {
	r.Lock()
	defer r.Unlock()

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
