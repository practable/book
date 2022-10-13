// package store holds bookings with arbitrary durations
package interval

/*
import (
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
)

var errNotFound = errors.New("resource not found")

//UserCommands represents commands that users will send to the system
var UserCommands = []string{
	"cancelBooking",  //cancel an existing booking. Fails if booking has been collected already.
	"collectBooking", //get access to the experiment (may trigger some checks on video/data)
	"requestBooking", // make a new booking, only completes if within allowable number of booked slots, and slot is free
	"swapBooking",    // atomic action to cancel an existing booking and replace with a new one, only completes if new booking succesful
	"getBookings",    // return my current bookings
	"listBookings",   // return bookings for a given slot, in a particular interval.

}

// AdminCommands represents commands used by the administrator
// Notes ... do we separate configuration from user commands in the transaction history (probably no, for testing reasons)
//
var AdminCommands = []string{
	"importTransactions",
	"addSlot",
	"deleteSlot",
}

// Action represents a booking action, including the time it was taken
// so as to allow history-replay to rebuild the booking status based on
// a record of past actions. This may also help with testing?
type Action struct {
	IssuedAt  time.Time
	Do        string
	When      Interval
	SlotID    uuid.UUID
	BookingID uuid.UUID
	UserID    uuid.UUID
}

type Slot struct {
	*sync.Mutex `json:"-"`
	ID          uuid.UUID
	Resource    *Resource
	TimePolicy  *TimePolicy // usually points to to the DefaultTimePolicy
	Pool        string      // the pool name to get the activity from
	Available   *Resource   // the window set by the time policy

	Blocked []*Interval // any times when the slot cannot be offered
}

// Booking represents additional information about a booking
// The resource only holds a UUID, so we use a map to find
// information for a given booking
type Booking struct {
	ID      uuid.UUID
	SlotID  uuid.UUID
	Started bool
	UserID  uuid.UUID
	When    Interval
}

// Policy represents limits on when a booking can be made
// EnforceInAdvance: set to True to limit how far in advance bookings can be made
// Expiry: the latest possible datetime that a booking can end
// InAdvance: how far in advance
// MaxDuration: longest individual booking
// NotBefore: the earliest possible datetime that a booking can start
type TimePolicy struct {
	EnforceInAdvance bool          `json:"enforce_in_advance"`
	Expiry           time.Time     `json:"exp"`
	InAdvance        time.Duration `json:"in_advance"`
	MaxDuration      time.Duration `json:"max_duration"`
	NotBefore        time.Time     `json:"nbf"`
}

// meh, time policy needs intervals ...

// BookingPolicy represents a limit of the max number of live bookings
type BookingPolicy struct {
	Enforce     bool `json:"enforce"`
	MaxBookings `json:"max_bookings"`
}

// what's this do??
type Diary struct {
	*sync.Mutex `json:"-"`
	Bookings    []Booking
}

type Store struct {
	*sync.RWMutex `json:"-" yaml:"-"`

	Slots map[uuid.UUID]*Slot `json:"slots"`

	DefaultTimePolicy TimePolicy `json:"default_time_policy"`

	UserBookingPolicy BookingPolicy `json:"user_booking_policy"`

	AdminBookingPolicy BookingPolicy `json:"admin_booking_policy"`

	// Now is a function for getting the time - useful for mocking in test
	Now func() time.Time `json:"-" yaml:"-"`
}

// Note - we probably want to rename resources as slots.

// Fulfil represents where to get an experiment to fulfil a booking
// this needs to work with the current API only (for now...)
// but for backwards compatibility as routings change, it may be necessary
// to have some sort of templating ....
// Would be helpful to have a booking client library that did this stuff
// rather than hardcoding it all here ...
type Fulfil struct {
	Host  url.URL //where is the experiment access obtained from
	API   string  //what API &version is it running? (for backwards compatibility
	Token String  //jwt token for requesting access to experiment
	// how do you get information on the experiment?
	// how do you request the experiment?
}

// Slot represents a bookable
type Slot struct {
	Resource *Resource
	Policies []*TimePolicy
	// we might later put some info here as to where to get the item .... but for MVP, it is going to be local, using go.
}

func New() *Store {
	return &Store{
		Resources: make(map[uuid.UUID]*resource.Resource),
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

	return bookings, errNotFound

}
*/
