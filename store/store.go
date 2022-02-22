// package store holds bookings with arbitrary durations
package store

import (
	"errors"
	"net/url"
	"sync"
	"time"

	"interval/internal/resource"

	"interval/internal/interval"

	"github.com/google/uuid"

	bc "github.com/timdrysdale/relay/pkg/bc/client"
)

var errNotFound = errors.New("resource not found")

// it would be helpful if people were warned about non-redundant equipment
// e.g. messaging them when equipment is known to be offline and unavailable for their booking
// or alerting them to a status page.

// Granularity of booking, and display of booking slots.
// we can let the user interface invert bookings information to show availability.
// If we assume "bookable unless booked", then graphically, put the background colour
// to the "free" colour, and overlay bookings in "booked" colour. Save processing it.

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

type Action struct {
	At        Time
	Do        string
	When      Interval
	SlotID    uuid.UUID
	BookingID uuid.UUID
	UserID    uuid.UUID
}

type Interval struct {
	Start time.Time
	End   time.Time
}

type Slot struct {
	*sync.Mutex `json:"-"`
	ID          uuid.UUID
	Bookings    *resource.Resource
	Experiment  *bc.Bc
}

// Booking represents additional information about a booking
// The resource only holds a UUID, so we use a map to find
// information for a given booking
type Booking struct {
	User    uuid.UUID
	When    Interval
	Started bool
}

// Duration represents a duration
// defined so we can (un)marshal
// https://stackoverflow.com/questions/48050945/how-to-unmarshal-json-into-durations
type Duration struct {
	time.Duration
}

// Duration represents a datetime
// defined so we can (un)marshal
// https://ukiahsmith.com/blog/go-marshal-and-unmarshal-json-with-time-and-url-data/
type Time struct {
	time.Time
}

// Policy represents limits on what a user can book
// Window: how far in advance a booking can end
// Expiry: the latest possible datetime that a booking can end
// MaxBookings: the maximum number of bookings that can be made
type Policy struct {
	Window      Duration `json:"duration"`
	Expiry      Time
	MaxBookings int
}

type Diary struct {
	*sync.Mutex `json:"-"`
	Bookings    []Booking
}

type UserID struct {
	uuid.UUID
}

type SlotID struct {
	uuid.UUID
}

type Store struct {
	Slots         map[SlotID]*Slot
	Sessions      map[UserID]*Diary
	Policies      map[UserID]*Policy
	DefaultPolicy Policy
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

// Slot represents a bookable slot
type Slot struct {
	Resource *resource.Resource
	BC       *booking.Client
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

		u, err := r.Request(interval.Interval{
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
