// package store holds bookings with arbitrary durations
package store

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/timdrysdale/interval/filter"
	"github.com/timdrysdale/interval/interval"
	"github.com/timdrysdale/interval/resource"
)

var errNotFound = errors.New("resource not found")

// Booking represents a promise to access an equipment that
// provided by the pool referenced in the resource of the slot
type Booking struct {
	Cancelled bool
	ID        string //booking uid
	Slot      string // slot name
	Started   bool
	User      string // user pseudo id
	When      interval.Interval
}

// Group represents a list of slots that
type Group struct {
	*sync.RWMutex `json:"-"`
	Description   `json:"description"`
	Name          string
	Slots         []string
}

// OldBooking represents a promise to access an equipment that
// provided by the pool referenced in the resource of the slot
// include references to the activity and description to avoid
// blank info when looking at old bookings.
// Don't include as copies, to avoid excessive memory leak
// note, these references will point to updated descriptions
// rather than the description at the time - note that the descriptions
// reference external images anyway, so this is to a certain extent
// desirable behaviour, although a resource that no longer exists will not
// be able to be shown in the history very well, if the images etc are gone
// resetting the store ought to delete history, so that memory
// can be cleaned up.
// TODO - log bookings or store in DB,so that history can be restored if needed?!
type OldBooking struct {
	Activity    *Activity
	Booking     *Booking
	Description *Description
}

// Resource represents a single virtual equipment that will (eventually) be
// fulfilled by a real equipment from the pool, and accessed using a particular
// UISet. This lets different users see different UI sets for the same equipment,
// e.g. students in different classes, staff, and developers all have different
// UIs they want to access
type Resource struct {
	Description Description
	Name        string
	Pool        string             // the name of the resource in the poolstore
	UISet       string             // the name of the list of UIs that can be used for this resource
	Resource    *resource.Resource // holds the bookings we've stored
}

// get description from resource
type Slot struct {
	Name          string
	TimePolicies  []string
	UsagePolicies []string
	Resource      string
}

// TODO if the information about an activity is changed or deleted then the old bookings fail...
type Store struct {
	*sync.RWMutex `json:"-"`

	// Bookings represents all the live bookings, indexed by booking id
	Bookings map[string]*Booking

	// Descriptions represents all the descriptions of various entities, indexed by description name
	Descriptions map[string]*shared.Description

	// Equipments represents all the actual equipments in the store, indexed by equipment name
	Equipments map[string]*pool.Equipment

	// Groups is list of all the groups in the store
	Groups map[string]*Group

	// Now is a function for getting the time - useful for mocking in test
	Now func() int64 `json:"-" yaml:"-"`

	//useful for admin dashboard - don't need to also parse logs if keep old bookings
	// Old Bookings represents the
	OldBookings map[string]*Booking

	// Pools represents the types of real equipment available (a pool has zero or more equipments)
	Pools map[string]*pool.Pool

	// Resources represent all the individual virtual equipments that can be booked
	Resources map[string]*Resource

	// Slots represent the combinations of virtual equipments and booking policies that apply to them
	Slots map[string]*Slot

	// TimePolicies represents all the TimePolicy(ies) in use
	TimePolicies map[string]*TimePolicy

	// UIs represents all the user interfaces that are available
	UIs map[string]*UI

	// UISets represents the lists of user interfaces for particular slots
	UISets map[string]*UISet

	// UsagePolicies represents all the UsagePolicy(ies) in use
	UsagePolicies map[string]*UsagePolicy
}

// TimePolicy represents the maximum number of future bookings that can be held
// and how long/short they can be held for. The limits are optionally enforced.
type TimePolicy struct {
	Description        Description
	EnforceMaxBookings bool
	EnforceMaxDuration bool
	EnforceMinDuration bool
	Filter             filter.Filter
	MaxBookings        int64
	MaxDuration        time.Duration
	MinDuration        time.Duration
	Name               string
}

// UsagePolciy represents how long a user can use equipments from a particular slot, i.e.
// cancelled bookings are not counted if they are cancelled before they are started
type UsagePolicy struct {
	Description Description
	Enforce     bool
	MaxSeconds  int64
	Name        string
}

// User represents bookings and usage information associated with a single user
type User struct {
	Name        string
	Usage       map[string]*Usage
	Bookings    map[string]*Booking
	OldBookings map[string]*Booking //useful for user dashboard- don't need to also parse logs if keep old bookings
}

// New returns an empty store
func New() *Store {
	return &Store{
		&sync.RWMutex{},
		make(map[string]*Booking),
		make(map[string]*shared.Description),
		make(map[string]*pool.Equipment),
		make(map[string]*Group),
		func() time.Time { return time.Now() },
		make(map[string]*Booking),
		make(map[string]*pool.Pool),
		make(map[string]*Resource),
		make(map[string]*Slot),
		make(map[string]*TimePolicy),
		make(map[string]*UI),
		make(map[string]UISet),
		make(map[string]*UsagePolicy),
	}
}

func (s *Store) Add() uuid.UUID {

	u := uuid.New()

	r := resource.New()

	s.Resources[u] = r

	return u
}

func (s *Store) ClearBeforeAll(t time.Time) {

	for _, r := range s.Resources {
		r.ClearBefore(t)
	}
}

func (s *Store) Request(rID uuid.UUID, when Interval) (uuid.NullUUID, error) {

	nu := uuid.NullUUID{}

	if r, ok := s.Resources[rID]; ok {

		u, err := r.Request(Interval{
			Start: when.Start,
			End:   when.End,
		})

		if err != nil {
			return nu, err
		}

		nu.UUID = u
		nu.Valid = true
		return nu, nil

	}

	return nu, errNotFound
}

func (s *Store) Cancel(rID uuid.UUID, bID uuid.UUID) error {

	if r, ok := s.Resources[rID]; ok {

		return r.Delete(bID)

	}

	return errNotFound

}

func (s *Store) GetBookings(rID uuid.UUID) ([]Booking, error) {

	bookings := []Booking{}

	if r, ok := s.Resources[rID]; ok {

		bb, err := r.GetBookings()

		if err != nil {
			return bookings, err
		}

		for _, b := range bb {
			bookings = append(bookings,
				Booking{
					When: Interval{
						Start: b.When.Start,
						End:   b.When.End,
					},
					ID: b.ID,
				})
		}

		return bookings, nil

	}

	return bookings, errNotFoundfde3db

}
