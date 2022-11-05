// package store holds bookings with arbitrary durations
package interval

import (
	"errors"
	"sync"
	"time"

	"github.com/timdrysdale/interval/diary"
	"github.com/timdrysdale/interval/filter"
	"github.com/timdrysdale/interval/interval"
)

// Booking represents a promise to access an equipment that
// provided by the pool referenced in the resource of the slot
type Booking struct {
	Cancelled   bool
	ID          string // booking uid
	Policy      string // reference to policy it was booked under
	Slot        string // slot name
	Started     bool
	Unfulfilled bool   //when the resource was unavailable
	User        string // user pseudo id
	When        interval.Interval
}

// Description represents information to display to a user about an entity
type Description struct {
	Type    string `json:"type"`
	Short   string `json:"short"`
	Long    string `json:"long,omitempty"`
	Further string `json:"further,omitempty"`
	Thumb   string `json:"thumb,omitempty"`
	Image   string `json:"image,omitempty"`
}

// Policy represents what a user can book, and any limits on bookings/usage
// Unmarshaling of time.Duration works in yaml.v3, https://play.golang.org/p/-6y0zq96gVz"
type Policy struct {
	BookAhead          time.Duration `json:"book_ahead"  yaml:"book_ahead"`
	Description        string        `json:"description"  yaml:"description"`
	EnforceBookAhead   bool          `json:"enforce_book_ahead"  yaml:"enforce_book_ahead"`
	EnforceMaxBookings bool          `json:"enforce_max_bookings"  yaml:"enforce_max_bookings"`
	EnforceMaxDuration bool          `json:"enforce_max_duration"  yaml:"enforce_max_duration"`
	EnforceMinDuration bool          `json:"enforce_min_duration"  yaml:"enforce_min_duration"`
	EnforceMaxUsage    bool          `json:"enforce_max_usage"  yaml:"enforce_max_usage"`
	MaxBookings        int64         `json:"max_bookings"  yaml:"max_bookings"`
	MaxDuration        time.Duration `json:"max_duration"  yaml:"max_duration"`
	MinDuration        time.Duration `json:"min_duration"  yaml:"min_duration"`
	MaxUsage           time.Duration `json:"max_usage"  yaml:"max_usage"`
	Slots              []string      `json:"slots" yaml:"slots"`
}

// Resource represents a physical entity that can be booked
type Resource struct {

	// ConfigURL represents a hardware configuration file URL
	// that may be useful to a UI
	ConfigURL string `json:"config_url,omitempty"  yaml:"config_url,omitempty"`

	// Description is a reference to a named description of the resource
	// that will probably only be shown on admin dashboards (not to students)
	Description string `json:"description"  yaml:"description"`

	// Diary is held in memory, not in the manifest, so don't unmarshall it.
	Diary *diary.Diary `json:"-"  yaml:"-"`

	// Streams is a list of stream types used by this resource, e.g. data, video, logging
	// We autogenerate the full stream details needed by the UI  when making a live activity,
	// using a rule to generate the topic and filling in the other details from the stream prototype
	// Streams are required because sims would still use logging, and if not
	// just add a dummy stream called null so that we have a check on streams
	// being included for the main use case.
	Streams []string `json:"streams"  yaml:"streams"`
}

// use separate description from resource, because UISet
// All of the strings, except Name, are references to other entries
// but we can do our own consistency checking rather
// than having to replace the yaml unmarshal process
// if we used pointers and big structs as before
type Slot struct {
	Description string `json:"description"  yaml:"description"`
	Policy      string `json:"policy"  yaml:"policy"`
	Resource    string `json:"resource"  yaml:"resource"`
	UISet       string `json:"ui_set"  yaml:"ui_set"`
	Window      string `json:"window"  yaml:"window"`
}

// Stream represents a prototype for a type of stream from a relay
// Streams will typically be either data, video, or logging.
// If multiple relay access servers r1, r2 etc are used,just define separate prototypes for
// each type of stream, on each relay, e.g. data-r0, data-r1 etc. Note that in future, a single
// access point will reverse proxy for multiple actual relays, so it's only if there
// are multiple access points that this would be needed.
// Streams are typically accessed via POST with bearer token to an access API
type Stream struct {

	// Audience is the URL of the relay server e.g. https://relay-access.practable.io
	Audience string `json:"audience"  yaml:"audience"`

	// ConnectionType is whether for session or shell e.g. session
	ConnectionType string `json:"ct"  yaml:"ct"`

	// For is the key in the UI's URL in which the client puts
	// the relay (wss) address and code after getting them
	// from the relay, e.g. data
	For string `json:"for"  yaml:"for"`

	// Scopes represent what the client can do e.g. read, write
	Scopes []string `json:"scopes"  yaml:"scopes"`

	// Topic is the relay topic, usually <resource name>-<for>. e.g. pend03-data
	Topic string `json:"topic"  yaml:"topic"`

	// URL of the relay access point for this stream
	URL string `json:"url"  yaml:"url"`
}

// Store represents entities required to make bookings, including resources, slots, descriptions, users, policies, and bookings
// any maps to values are data that are not mutated except when the manifest is replaced so do not need to be maps to pointers
type Store struct {
	*sync.RWMutex `json:"-"`

	// Bookings represents all the live bookings, indexed by booking id
	Bookings map[string]*Booking

	// Descriptions represents all the descriptions of various entities, indexed by description name
	Descriptions map[string]Description

	// Filters are how the windows are checked, mapped by window name (populated after loading window info from manifest)
	Filters map[string]*filter.Filter

	// Now is a function for getting the time - useful for mocking in test
	Now func() time.Time `json:"-" yaml:"-"`

	//useful for admin dashboard - don't need to also parse logs if keep old bookings
	// Old Bookings represents the
	OldBookings map[string]*Booking

	// TimePolicies represents all the TimePolicy(ies) in use
	Policies map[string]Policy

	// Resources represent all the actual physical experiments, indexed by name
	Resources map[string]Resource

	// Slots represent the combinations of virtual equipments and booking policies that apply to them
	Slots map[string]Slot

	Streams map[string]Stream

	// UIs represents all the user interfaces that are available
	UIs map[string]UI

	// UISets represents the lists of user interfaces for particular slots
	UISets map[string]UISet

	// UsagePolicies represents all the UsagePolicy(ies) in use
	Users map[string]*User

	// Window represents allowed and denied time periods for slots
	Windows map[string]Window
}

// User represents bookings and usage information associated with a single user
// remembering policies allows us to direct a user to link to a policy for a course just once, and then have that remembered
// at least until a system restart -> should be logged as a transaction
type User struct {
	Bookings    map[string]*Booking      //map by id for retrieval
	OldBookings map[string]*Booking      //map by id, for admin dashboards
	Policies    map[string]bool          //map of policies that apply to the user
	Usage       map[string]time.Duration //map by policy for checking usage
}

// UI represents a UI that can be used with a resource, for a given slot
type UI struct {
	Description string `json:"description"  yaml:"description"`
	// URL with moustache {{key}} templating for stream connections
	URL             string   `json:"url"  yaml:"url"`
	StreamsRequired []string `json:"streams_required"  yaml:"streams_required"`
}

// UISet represents UIs that can be used with a slot
type UISet struct {
	UIs []string
}

// Window represents allowed and denied periods for slots
type Window struct {
	Allowed []interval.Interval `json:"allowed"  yaml:"allowed"`
	Denied  []interval.Interval `json:"denied"  yaml:"denied"`
}

// New returns an empty store
func New() *Store {
	return &Store{
		&sync.RWMutex{},
		make(map[string]*Booking),
		make(map[string]Description),
		make(map[string]*filter.Filter),
		func() time.Time { return time.Now() },
		make(map[string]*Booking),
		make(map[string]Policy),
		make(map[string]Resource),
		make(map[string]Slot),
		make(map[string]Stream),
		make(map[string]UI),
		make(map[string]UISet),
		make(map[string]*User),
		make(map[string]Window),
	}
}

// PruneDiaries is a maintenance operation to prune old bookings from diaries
// to make booking decisions faster. There is an overhead to pruning trees
// because they are rebalanced, so don't do too frequently.
func (s *Store) PruneDiaries() {
	for _, r := range s.Resources {
		r.Diary.ClearBefore(s.Now())
	}
}

// PruneDiaries is maintenance operation that moves expired bookings from
// the map of current bookings to the map of old bookings
func (s *Store) PruneBookings() {

	stale := make(map[string]*Booking)

	for k, v := range s.Bookings {
		if v.When.End.After(s.Now()) {
			stale[k] = v
		}
	}

	for k, v := range stale {
		s.OldBookings[k] = v
		delete(s.Bookings, k)
	}

}

// Operations required by users
// Get information on a policy
// Get information on the availability of a resource in a slot within an interval
// Book a particular slot for a particular time

// Optional extensions
// Find all slots that are free for a particular period?
// Find a random slot that can fulfil a particular request
// Present an aggregate availability for a set of slots

// Let consumer of this package, e.g. the API, define some types that contain both the description and the contents
// of the entities, if required - not much point doing it here because the openAPI generator will create its own
// types anyway.

// GetDescription returns a description if found
func (s *Store) GetDescription(name string) (Description, error) {

	d, ok := s.Descriptions[name]

	if !ok {
		return Description{}, errors.New("not found")
	}

	return d, nil
}

// GetPolicy returns a policy if found
func (s *Store) GetPolicy(name string) (Policy, error) {

	p, ok := s.Policies[name]

	if !ok {
		return Policy{}, errors.New("not found")
	}

	return p, nil
}

// Availability returns a slice of available intervals between start and end, given a set of bookings
func Availability(bk []diary.Booking, start, end time.Time) []interval.Interval {

	// strip the intervals from the bookings
	bi := []interval.Interval{}

	for _, b := range bk {
		bi = append(bi, b.When)
	}

	// interval.Invert merges and sort intervals
	// so we don't need to check for overlaps and ordering
	a := interval.Invert(bi)

	// The inverted list will start at zero time and end at infinity
	// so make a filtered list with no values before start or after end
	fa := []interval.Interval{}

	for _, i := range a {

		//ignore availability intervals that end before our start
		if i.End.Before(start) {
			continue
		}
		//ignore availability intervals that start after our end
		if i.Start.After(end) {
			continue
		}

		//trim an interval if it overlaps start
		if i.Start.Before(start) {
			fa = append(fa, interval.Interval{
				Start: start,
				End:   i.End,
			})
		} else if i.End.After(end) { //trim an interval if it overlaps end
			fa = append(fa, interval.Interval{
				Start: i.Start,
				End:   end,
			})
		} else { // ok interval, append it
			fa = append(fa, i)
		}
	}

	return fa

}

func (s *Store) GetAvailability(policy, slot string) ([]interval.Interval, error) {

	p, ok := s.Policies[policy]

	if !ok {
		return []interval.Interval{}, errors.New("policy " + policy + " not found")
	}

	_, ok = s.Slots[slot]

	if !ok {
		return []interval.Interval{}, errors.New("slot " + slot + " not found")
	}

	bk, err := s.GetSlotBookings(slot)

	if err != nil {
		return []interval.Interval{}, err
	}

	start := s.Now()
	end := interval.Infinity

	if p.EnforceBookAhead {
		end = start.Add(p.BookAhead)
	}

	fa := Availability(bk, start, end)

	return fa, nil

}

// GetSlotIsAvailable checks the underlying resource's availability
func (s *Store) GetSlotIsAvailable(slot string) (bool, string, error) {

	sl, ok := s.Slots[slot]

	if !ok {
		return false, "", errors.New("slot " + slot + " not found")
	}

	r, ok := s.Resources[sl.Resource]

	if !ok {
		return false, "", errors.New("resource " + sl.Resource + " not found")
	}

	ok, reason := r.Diary.IsAvailable()

	return ok, reason, nil

}

// GetSlotBookings gets bookings as far as ahead as the policy will let you book ahead
// It's up to the consumer to handle any pagination
func (s *Store) GetSlotBookings(slot string) ([]diary.Booking, error) {

	sl, ok := s.Slots[slot]

	if !ok {
		return []diary.Booking{}, errors.New("slot " + slot + " not found")
	}

	r, ok := s.Resources[sl.Resource]

	if !ok {
		return []diary.Booking{}, errors.New("resource " + sl.Resource + " not found")
	}

	b, err := r.Diary.GetBookings()

	if err != nil {
		return []diary.Booking{}, err
	}

	// if unavailable, return bookings with error to indicate requests will be unsuccessful
	ok, reason := r.Diary.IsAvailable()

	if !ok {
		return b, errors.New("not available because " + reason)
	}

	// available, return bookings
	return b, nil

}

/*
// RequestSlotBooking
// Note that we must ensure we book on the actual resource, not a copy of it
// consider refactor to maps of pointers if there are issues here
func (s *Store) Request(r string, when interval.Interval) (uuid.NullUUID, error) {

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
*/
