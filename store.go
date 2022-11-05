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

var errNotFound = errors.New("resource not found")

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
	Name    string `json:"name"`
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
	Description        string        `json:"description"  yaml:"description"`
	EnforceMaxBookings bool          `json:"enforce_max_bookings"  yaml:"enforce_max_bookings"`
	EnforceMaxDuration bool          `json:"enforce_max_duration"  yaml:"enforce_max_duration"`
	EnforceMinDuration bool          `json:"enforce_min_duration"  yaml:"enforce_min_duration"`
	EnforceMaxUsage    bool          `json:"enforce_max_usage"  yaml:"enforce_max_usage"`
	MaxBookings        int64         `json:"max_bookings"  yaml:"max_bookings"`
	MaxDuration        time.Duration `json:"max_duration"  yaml:"max_duration"`
	MinDuration        time.Duration `json:"min_duration"  yaml:"min_duration"`
	Name               string        `json:"name"  yaml:"name"`
	MaxUsage           time.Duration `json:"max_usage"  yaml:"max_usage"`
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

	// Name is the resource's unique name
	Name string `json:"name"  yaml:"name"`

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
	Name        string `json:"name"  yaml:"name"`
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

	// Name is unique reference to the stream prototype
	Name string `json:"name"  yaml:"name"`

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

// There is no need for a description in the resource, because the slot holds the description, so
// we can just use the resource.Resource directly in the store.

// Store represents entities required to make bookings, including resources, slots, descriptions, users, policies, and bookings
type Store struct {
	*sync.RWMutex `json:"-"`

	// Bookings represents all the live bookings, indexed by booking id
	Bookings map[string]*Booking

	// Descriptions represents all the descriptions of various entities, indexed by description name
	Descriptions map[string]*Description

	// Filters are how the windows are checked, mapped by window name (populated after loading window info from manifest)
	Filters map[string]*filter.Filter

	// Now is a function for getting the time - useful for mocking in test
	Now func() time.Time `json:"-" yaml:"-"`

	//useful for admin dashboard - don't need to also parse logs if keep old bookings
	// Old Bookings represents the
	OldBookings map[string]*Booking

	// TimePolicies represents all the TimePolicy(ies) in use
	Policies map[string]*Policy

	// Resources represent all the actual physical experiments, indexed by name
	Resources map[string]*Resource

	// Slots represent the combinations of virtual equipments and booking policies that apply to them
	Slots map[string]*Slot

	Streams map[string]*Stream

	// UIs represents all the user interfaces that are available
	UIs map[string]*UI

	// UISets represents the lists of user interfaces for particular slots
	UISets map[string]*UISet

	// UsagePolicies represents all the UsagePolicy(ies) in use
	Users map[string]*User

	// Window represents allowed and denied time periods for slots
	Windows map[string]*Window
}

// User represents bookings and usage information associated with a single user
// remembering policies allows us to direct a user to link to a policy for a course just once, and then have that remembered
// at least until a system restart -> should be logged as a transaction
type User struct {
	Name        string
	Bookings    map[string]*Booking      //map by id for retrieval
	OldBookings map[string]*Booking      //map by id, for admin dashboards
	Policies    map[string]bool          //map of policies that apply to the user
	Usage       map[string]time.Duration //map by policy for checking usage
}

// UI represents a UI that can be used with a resource, for a given slot
type UI struct {
	Name        string `json:"name"  yaml:"name"`
	Description string `json:"description"  yaml:"description"`
	// URL with moustache {{key}} templating for stream connections
	URL             string   `json:"url"  yaml:"url"`
	StreamsRequired []string `json:"streams_required"  yaml:"streams_required"`
}

// UISet represents UIs that can be used with a slot
type UISet struct {
	Name string
	UIs  []string
}

// Window represents allowed and denied periods for slots
type Window struct {
	Name    string              `json:"name"  yaml:"name"`
	Allowed []interval.Interval `json:"allowed"  yaml:"allowed"`
	Denied  []interval.Interval `json:"denied"  yaml:"denied"`
}

// New returns an empty store
func New() *Store {
	return &Store{
		&sync.RWMutex{},
		make(map[string]*Booking),
		make(map[string]*Description),
		make(map[string]*filter.Filter),
		func() time.Time { return time.Now() },
		make(map[string]*Booking),
		make(map[string]*Policy),
		make(map[string]*Resource),
		make(map[string]*Slot),
		make(map[string]*Stream),
		make(map[string]*UI),
		make(map[string]*UISet),
		make(map[string]*User),
		make(map[string]*Window),
	}
}

/* TODO change to name indexed format
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
*/
