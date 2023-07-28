// package store holds bookings with arbitrary durations

//
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
package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/practable/book/internal/check"
	"github.com/practable/book/internal/deny"
	"github.com/practable/book/internal/diary"
	"github.com/practable/book/internal/filter"
	"github.com/practable/book/internal/interval"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

// Activity represents connection details for a live booking
type Activity struct {
	BookingID   string            `json:"booking_id" yaml:"booking_id"`
	Description Description       `json:"description" yaml:"description"`
	ConfigURL   string            `json:"config_url,omitempty"  yaml:"config_url,omitempty"`
	Streams     map[string]Stream `json:"streams" yaml:"streams"`
	UIs         []UIDescribed     `json:"ui" yaml:"ui"`
	NotBefore   time.Time         `json:"nbf" yaml:"nbf"`
	ExpiresAt   time.Time         `json:"exp" yaml:"exp"`
}

// Booking represents a promise to access an equipment that
// provided by the pool referenced in the resource of the slot
type Booking struct {
	// Cancelled indicates if booking cancelled
	Cancelled bool `json:"cancelled" yaml:"cancelled"`
	// CancelledAt represents when the booking was cancelled
	CancelledAt time.Time `json:"cancelled_at" yaml:"cancelled_at"`
	// CancelledBy indicates who cancelled e.g. auto-grace-period, admin or user
	CancelledBy string `json:"cancelled_by" yaml:"cancelled_by"`
	// Group
	Group string `json:"group" yaml:"group"`
	// booking unique reference
	Name string `json:"name" yaml:"name"`
	// reference to policy it was booked under
	Policy string `json:"policy" yaml:"policy"`
	// slot name
	Slot    string `json:"slot" yaml:"slot"`
	Started bool   `json:"started" yaml:"started"`
	//StartedAt is for reporting purposes, do not use to calculate usage
	StartedAt string `json:"started_at" yaml:"started_at"`
	//when the resource was unavailable
	Unfulfilled bool `json:"unfulfilled" yaml:"unfulfilled"`
	// User represents user's name
	User string `json:"user" yaml:"user"`

	// UsageCharged represents how much usage was charged
	// This is updated on cancellation and is for convenience of admin looking at exports/reports
	// Replace(Old)Bookings should calculate the usage to be charged based on the policy
	// That avoids those editing bookings to upload from performing this calculation manually
	UsageCharged time.Duration `json:"usage_charged" yaml:"usage_charged"`

	When interval.Interval `json:"when" yaml:"when"`
}

// Description represents information to display to a user about an entity
type Description struct {
	Name    string `json:"name" yaml:"name"`
	Type    string `json:"type" yaml:"type"`
	Short   string `json:"short" yaml:"short"`
	Long    string `json:"long,omitempty" yaml:"long,omitempty"`
	Further string `json:"further,omitempty" yaml:"further,omitempty"`
	Thumb   string `json:"thumb,omitempty" yaml:"thumb,omitempty"`
	Image   string `json:"image,omitempty" yaml:"image,omitempty"`
}

// DisplayGuide represents guidance to the booking app on what length slots
// to offer, how many, and how far in the future. This is to allow course staff
// to influence the offerings to students in a way that might better suit their
// teaching views.
// remember to update UnmarshalJSON if adding fields
type DisplayGuide struct {
	BookAhead time.Duration `json:"book_ahead" yaml:"book_ahead"`
	Duration  time.Duration `json:"duration" yaml:"duration"`
	Label     string        `json:"label" yaml:"label"`
	MaxSlots  int           `json:"max_slots" yaml:"max_slots"`
}

// Group represents a list of policies for ease of sharing multiple policies with users
// and being able to change the policies that are supplied to a user without having to
// update all the links the user has (important if user is a course organiser on a large course!)
type Group struct {
	Description string   `json:"description" yaml:"description"`
	Policies    []string `json:"policies" yaml:"policies"`
}

// GroupDescribed includes the description to save some overhead, since it will always be
// requested by the user with the description included.
type GroupDescribed struct {
	Description Description `json:"description"  yaml:"description"`
	// keep track of the description reference, needed for manifest export
	DescriptionReference string   `json:"-" yaml:"-"`
	Policies             []string `json:"policies" yaml:"policies"`
}

// Manifest represents all the available equipment and how to access it
// Slots are the primary entities, so reference checking starts with them
type Manifest struct {
	Descriptions  map[string]Description  `json:"descriptions" yaml:"descriptions"`
	DisplayGuides map[string]DisplayGuide `json:"display_guides" yaml:"display_guides"`
	Groups        map[string]Group        `json:"groups" yaml:"groups"`
	Policies      map[string]Policy       `json:"policies" yaml:"policies"`
	Resources     map[string]Resource     `json:"resources" yaml:"resources"`
	Slots         map[string]Slot         `json:"slots" yaml:"slots"`
	Streams       map[string]Stream       `json:"streams" yaml:"streams"`
	UIs           map[string]UI           `json:"uis" yaml:"uis"`
	UISets        map[string]UISet        `json:"ui_sets" yaml:"ui_sets"`
	Windows       map[string]Window       `json:"windows" yaml:"windows"`
}

// Policy represents what a user can book, and any limits on bookings/usage
// Unmarshaling of time.Duration works in yaml.v3, https://play.golang.org/p/-6y0zq96gVz"
// remember to update UnmarshalJSON if adding fields
type Policy struct {
	// AllowStartInPastWithin gives some latitude to accept a booking starting now that gets delayed on the way to the server. A bookng at minimum acceptable duration will be reduced to as much as this duration, so that there is no need to include logic about how to handle a shift in the end time. Typically values might be 10s or 1m.
	AllowStartInPastWithin time.Duration `json:"allow_start_in_past_within"  yaml:"allow_start_in_past_within"`
	//booking must finish within the book_ahead duration, if enforced
	BookAhead     time.Duration `json:"book_ahead"  yaml:"book_ahead"`
	Description   string        `json:"description"  yaml:"description"`
	DisplayGuides []string      `json:"display_guides"  yaml:"display_guides"`
	// In the manifest, we will refer to display guides by reference
	// For users, we want to send policy descriptions that are complete
	// so store a local copy of the displayguides to ease the process of fulfilling GET policy_name requests
	// but don't include this local copy of information in any manifests
	DisplayGuidesMap map[string]DisplayGuide `json:"-"  yaml:"-"` //local copy so that exported policies are complete but exclude from json/yaml so not duplicated in manifests
	// EnforceAllowStartInPast lets a request starting before now (e.g. due to delayed communication of request) be accepted if other policies are still met.
	EnforceAllowStartInPast bool `json:"enforce_allow_start_in_past"  yaml:"enforce_allow_start_in_past"`
	EnforceBookAhead        bool `json:"enforce_book_ahead"  yaml:"enforce_book_ahead"`
	EnforceGracePeriod      bool `json:"enforce_grace_period"  yaml:"enforce_grace_period"`
	EnforceMaxBookings      bool `json:"enforce_max_bookings"  yaml:"enforce_max_bookings"`
	EnforceMaxDuration      bool `json:"enforce_max_duration"  yaml:"enforce_max_duration"`
	EnforceMinDuration      bool `json:"enforce_min_duration"  yaml:"enforce_min_duration"`
	EnforceMaxUsage         bool `json:"enforce_max_usage"  yaml:"enforce_max_usage"`
	EnforceNextAvailable    bool `json:"enforce_next_available"  yaml:"enforce_next_available"`
	EnforceStartsWithin     bool `json:"enforce_starts_within"  yaml:"enforce_starts_within"`
	//EnforceUnlimitedUsers if true, bookings are not checked, and the token is granted if otherwise within policy. This supports hardware-less simulations to be
	// included without needing to specify multiple slots. We don't set a finite limit here to avoid having to track multiple overlapping bookings when usually simulations
	// run entirely in client-side code - if a simulation has a resource limit e.g. due to using some central heavyweight server to crunch data, then slots should be specified
	// same as for hardware, and this option left as false.
	EnforceUnlimitedUsers bool `json:"enforce_unlimited_users"  yaml:"enforce_unlimited_users"`
	// GracePeriod is how long after When.Start that the booking will be kept
	GracePeriod time.Duration `json:"grace_period" yaml:"grace_period"`
	// GracePenalty represents the time lost to finding a new user after auto-cancellation
	GracePenalty time.Duration `json:"grace_penalty" yaml:"grace_penalty"`
	MaxBookings  int64         `json:"max_bookings"  yaml:"max_bookings"`
	MaxDuration  time.Duration `json:"max_duration"  yaml:"max_duration"`
	MinDuration  time.Duration `json:"min_duration"  yaml:"min_duration"`
	MaxUsage     time.Duration `json:"max_usage"  yaml:"max_usage"`
	// NextAvailable allows for a small gap in bookings to give some flex in case the availability windows are presented with reduced resolution at some point in the system
	// i.e. set to 2min to allow a request that is rounded up to start at the next minute after the last booking ends, instead of expecting ms precision from everyone
	// Leaving this to default to zero requires the booking UI to return the exact figure given in the availability list, which probably works for now but might not later when other developers
	// working on other features maybe don't realise how strict the calculation is without this allowance, or we change the precision somewhere in the system for human-readability and lose the
	// exact value that the system would expect due to loss of precision - resulting in a rejected booking that is otherwise within the spirit of the policy.
	// Also, some use cases might actually let this be say 15min or 30min - we can't predict the use cases, but can expect them to vary within the same booking system,
	// so don't make this a system-wide parameter.
	NextAvailable time.Duration   `json:"next_available"  yaml:"next_available"`
	Slots         []string        `json:"slots" yaml:"slots"`
	SlotMap       map[string]bool `json:"-" yaml:"-"` // internal usage, do not populate from file
	// booking must start within this duration from now, if enforced
	StartsWithin time.Duration `json:"starts_within"  yaml:"starts_within"`
}

type PolicyStatus struct {
	CurrentBookings int64         `json:"current_bookings"  yaml:"current_bookings"`
	OldBookings     int64         `json:"old_bookings"  yaml:"old_bookings"`
	Usage           time.Duration `json:"usage"  yaml:"usage"`
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

	Tests []string `json:"tests"  yaml:"tests"`

	//TopicStub is the name that should be used to make the topic <TopicStub>-<for>
	TopicStub string `json:"topic_stub" yaml:"topic_stub"`
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

// Store represents entities required to make bookings, including resources, slots, descriptions, users, policies, and bookings
// any maps to values are data that are not mutated except when the manifest is replaced so do not need to be maps to pointers
type Store struct {
	*sync.RWMutex `json:"-"`

	// Checker does grace checking on bookings
	Checker *check.Checker

	// Bookings represents all the live bookings, indexed by booking id
	Bookings map[string]*Booking

	denyClient *deny.Client

	denyRequests chan deny.Request

	// Descriptions represents all the descriptions of various entities, indexed by description name
	Descriptions map[string]Description

	DisableCancelAfterUse bool

	DisplayGuides map[string]DisplayGuide

	// Filters are how the windows are checked, mapped by window name (populated after loading window info from manifest)
	Filters map[string]*filter.Filter

	// Groups represent groups of policies - we bake in the description to reduce overhead on this common operation
	Groups map[string]GroupDescribed

	// Locked is true when we want to stop making bookings or getting info while we do uploads/maintenance
	// The API handler has to check this, e.g. if locked, do not make bookings or check availability on
	// behalf of users. We can't do this automatically in the methods because then we'd need some sort
	// of admin override, to permit maintenance when locked (which is the whole point of locking the system)

	// GraceRebound represents how long to wait before checking any bookings that were
	// supposed to be checked but the store was locked (see GraceCheck)
	GraceRebound time.Duration

	Locked bool

	// Message represents our message of the day, to send to users (e.g. to explain system is locked)
	Message string

	// now is a function for getting the time - useful for mocking in test
	// to avoid races, we must use a setter and a getter with a mutex
	now func() time.Time `json:"-" yaml:"-"`

	//useful for admin dashboard - don't need to also parse logs if keep old bookings
	// Old Bookings represents the
	OldBookings map[string]*Booking

	// TimePolicies represents all the TimePolicy(ies) in use
	Policies map[string]Policy

	// relaySecret holds the secret for the relays (all relays served by a book instance must share the same secret)
	// Don't expose secret unnecessarily, so don't include when serialising (not that we currently serialise the store anyway)
	relaySecret string `json:"-" yaml:"-"`

	// how long to wait when making requests to external API (e.g. for deny)
	requestTimeout time.Duration

	// Resources represent all the actual physical experiments, indexed by name
	Resources map[string]Resource

	// Slots represent the combinations of virtual equipments and booking policies that apply to them
	Slots map[string]Slot

	Streams map[string]Stream

	// UIs represents all the user interfaces that are available
	UIs map[string]UIDescribed

	// UISets represents the lists of user interfaces for particular slots
	UISets map[string]UISet

	// Users maps all users.
	Users map[string]*User

	// Window represents allowed and denied time periods for slots
	Windows map[string]Window
}

type StoreStatusAdmin struct {
	Bookings     int64     `json:"bookings"  yaml:"bookings"`
	Descriptions int64     `json:"descriptions"  yaml:"descriptions"`
	Filters      int64     `json:"filters" yaml:"filters"`
	Groups       int64     `json:"groups" yaml:"groups"`
	Locked       bool      `json:"locked" yaml:"locked"`
	Message      string    `json:"message" yaml:"message"`
	Now          time.Time `json:"now" yaml:"now"`
	OldBookings  int64     `json:"old_bookings"  yaml:"old_bookings"`
	Policies     int64     `json:"policies" yaml:"policies"`
	Resources    int64     `json:"resources" yaml:"resources"`
	Slots        int64     `json:"slots" yaml:"slots"`
	Streams      int64     `json:"streams" yaml:"streams"`
	UIs          int64     `json:"uis" yaml:"uis"`
	UISets       int64     `json:"ui_sets" yaml:"ui_sets"`
	Users        int64     `json:"users" yaml:"users"`
	Windows      int64     `json:"windows" yaml:"windows"`
}

type StoreStatusUser struct {
	Locked  bool      `json:"locked" yaml:"locked"`
	Message string    `json:"message" yaml:"message"`
	Now     time.Time `json:"now" yaml:"now"`
}

// Stream represents a prototype for a type of stream from a relay
// Streams will typically be either data, video, or logging.
// If multiple relay access servers r1, r2 etc are used,just define separate prototypes for
// each type of stream, on each relay, e.g. data-r0, data-r1 etc. Note that in future, a single
// access point will reverse proxy for multiple actual relays, so it's only if there
// are multiple access points that this would be needed.
// Streams are typically accessed via POST with bearer token to an access API
type Stream struct {
	Audience string `json:"audience" yaml:"audience"`
	// ConnectionType is whether for session or shell e.g. session
	ConnectionType string `json:"connection_type"  yaml:"connection_type"`

	// For is the key in the UI's URL in which the client puts
	// the relay (wss) address and code after getting them
	// from the relay, e.g. data
	For string `json:"for"  yaml:"for"`

	// Scopes represent what the client can do e.g. read, write
	Scopes []string `json:"scopes"  yaml:"scopes"`

	// Topic is the relay topic, usually <resource name>-<for>. e.g. pend03-data
	Topic string `json:"topic"  yaml:"topic"`

	// URL of the relay access point for this stream e.g. https://relay-access.practable.io
	URL string `json:"url"  yaml:"url"`
}

// UI represents a UI that can be used with a resource, for a given slot
type UI struct {
	Description string `json:"description"  yaml:"description"`
	// URL with moustache {{key}} templating for stream connections
	URL             string   `json:"url"  yaml:"url"`
	StreamsRequired []string `json:"streams_required"  yaml:"streams_required"`
}

// UIDescribed represents a UI that can be used with a resource, for a given slot
// with a description - for sending to users
type UIDescribed struct {
	Description Description `json:"description"  yaml:"description"`
	// Keep track of the description's name, needed for ExportManifest
	DescriptionReference string `json:"-" yaml:"-"`
	// URL with moustache {{key}} templating for stream connections
	URL             string   `json:"url"  yaml:"url"`
	StreamsRequired []string `json:"streams_required"  yaml:"streams_required"`
}

// UISet represents UIs that can be used with a slot
type UISet struct {
	UIs []string `json:"uis" yaml:"uis"`
}

// User represents bookings and usage information associated with a single user
// remembering policies allows us to direct a user to link to a policy for a course just once, and then have that remembered
// at least until a system restart -> should be logged as a transaction
type User struct {
	Bookings    map[string]*Booking       //map by id for retrieval
	OldBookings map[string]*Booking       //map by id, for admin dashboards
	Groups      map[string]bool           //map of groups of policies that the user can access
	Usage       map[string]*time.Duration //map by policy for checking usage
}

type UserExternal struct {
	Bookings    []string `json:"bookings" yaml:"bookings"`
	OldBookings []string `json:"old_bookings" yaml:"old_bookings"`
	Groups      []string `json:"groups" yaml:"groups"`
	//map humanised durations by policy name
	Usage map[string]string `json:"usage" yaml:"usage"`
}

// Window represents allowed and denied periods for slots
type Window struct {
	Allowed []interval.Interval `json:"allowed"  yaml:"allowed"`
	Denied  []interval.Interval `json:"denied"  yaml:"denied"`
}

// New returns an empty store
func New() *Store {
	denyClient := deny.New()
	return &Store{
		&sync.RWMutex{},
		check.New().WithNow(func() time.Time { return time.Now() }).WithName("forStore"),
		make(map[string]*Booking),
		denyClient,
		denyClient.Request, //can be overwritten for testing using WithDenyRequests()
		make(map[string]Description),
		false,
		make(map[string]DisplayGuide),
		make(map[string]*filter.Filter),
		make(map[string]GroupDescribed),
		time.Duration(time.Minute),
		false,
		"Welcome to the interval booking store",
		func() time.Time { return time.Now() },
		make(map[string]*Booking),
		make(map[string]Policy),
		"replaceme",
		time.Second,
		make(map[string]Resource),
		make(map[string]Slot),
		make(map[string]Stream),
		make(map[string]UIDescribed),
		make(map[string]UISet),
		make(map[string]*User),
		make(map[string]Window),
	}
}

// for testing purposes, otherwise deny channel set to that of the deny.Client
func (s *Store) WithDenyRequests(d chan deny.Request) *Store {
	s.Lock()
	defer s.Unlock()
	log.Warn("Overriding denyRequests, preventing normal operation - do not use in production")
	s.denyRequests = d
	return s
}

// WithDisableCancelAfterUse stops users from cancelling bookings they already started using
// this is provided in case external API calls to relay cannot be supported (e.g. due to relay version)
// note all relays need to have the same secret!
func (s *Store) WithDisableCancelAfterUse(d bool) *Store {
	s.Lock()
	defer s.Unlock()
	s.DisableCancelAfterUse = d
	return s
}

// WithNow sets the time function
func (s *Store) WithNow(now func() time.Time) *Store {
	s.Lock()
	defer s.Unlock()
	s.now = now
	s.Checker.SetNow(now)
	s.denyClient.SetNow(now)
	return s
}

// SetRelaySecret sets the relay secret
func (s *Store) SetRelaySecret(secret string) *Store {
	s.Lock()
	defer s.Unlock()
	s.relaySecret = secret
	s.denyClient.SetSecret(secret)
	return s
}

// WithRelaySecret sets the relay secret
func (s *Store) WithRelaySecret(secret string) *Store {
	s.Lock()
	defer s.Unlock()
	s.relaySecret = secret
	s.denyClient.SetSecret(secret)
	return s
}

// SetRequestTimeout sets how long to wait for external API requests, e.g. deny requests to relay
func (s *Store) SetRequestTimeout(timeout time.Duration) *Store {
	s.Lock()
	defer s.Unlock()
	s.requestTimeout = timeout
	s.denyClient.SetTimeout(timeout)
	return s
}

// WithRequestTimeout sets how long to wait for external API requests, e.g. deny requests to relay
func (s *Store) WithRequestTimeout(timeout time.Duration) *Store {
	s.Lock()
	defer s.Unlock()
	s.requestTimeout = timeout
	s.denyClient.SetTimeout(timeout)
	return s
}

// RelaySecret returns the relay secret
// don't use in internal functions because it will hang waiting for lock
// just use s.relaySecret directly in internal functions
func (s *Store) RelaySecret() string {
	s.Lock()
	defer s.Unlock()
	return s.relaySecret
}

// SetNow sets the time function (useful for mocking in tests)
// Alternative named version for readability when updating the time
// function multiple times in a test
func (s *Store) SetNow(now func() time.Time) *Store {
	s.Lock()
	defer s.Unlock()
	s.now = now
	s.Checker.SetNow(now)
	s.denyClient.SetNow(now)
	return s
}

func (s *Store) Now() time.Time {
	s.Lock()
	defer s.Unlock()
	return s.now()
}

func (s *Store) WithGraceRebound(d time.Duration) *Store {
	s.Lock()
	defer s.Unlock()
	s.GraceRebound = d
	return s
}

// NewUser returns a pointer to a new User
func NewUser() *User {
	return &User{
		make(map[string]*Booking),
		make(map[string]*Booking),
		make(map[string]bool),
		make(map[string]*time.Duration),
	}
}

// AddGroupForUser adds a group for a user so they can book with it in future
// without having to have the access code to hand
// TODO needs a corresponding DeleteGroupFor
func (s *Store) AddGroupForUser(user, group string) error {

	where := "store.AddGroupFor"
	log.Trace(where + " awaiting lock")
	s.Lock()
	log.Trace(where + " has lock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released lock")
	}()

	_, ok := s.Groups[group]

	if !ok {
		return errors.New("group " + group + " not found")
	}

	u, ok := s.Users[user]

	if !ok { //create user if does not exist
		u = NewUser()
		s.Users[user] = u
	}

	u.Groups[group] = true

	// no need to initialise any usage trackers, because (a) policies could change between now and the
	// first booking, and (b) the makeBooking function initialises any trackers required at time of booking

	s.Users[user] = u

	return nil

}

func (s *Store) GenerateUniqueUser() string {

	return xid.New().String() //Unicity guaranteed for 16,777,216 (24 bits) unique ids per second and per host/process
	// but could be predicted i.e. not cryptographically secure

}

// CancelBooking cancels a booking or returns an error if not found
// Takes a lock - for external usage
func (s *Store) CancelBooking(booking Booking, cancelledBy string) error {
	where := "store.CancelBooking"
	log.Trace(where + " awaiting lock")
	s.Lock()
	log.Trace(where + " has lock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released lock")
	}()

	return s.cancelBooking(booking, cancelledBy)

}

// cancelBooking cancels a booking or returns an error if not found
// does not take a lock, for internal use by functions that handle taking the lock
func (s *Store) cancelBooking(booking Booking, cancelledBy string) error {
	// check if booking exists and details are valid (i.e. must confirm booking contents, not just ID)
	b, ok := s.Bookings[booking.Name]

	if !ok {
		return errors.New("not found")
	}

	// compare the externally relevant fields of the booking (ignore internal boolean fields
	// to prevent status changes in the booking preventing cancellation

	t1 := Booking{
		Name:   b.Name,
		Policy: b.Policy,
		Slot:   b.Slot,
		User:   b.User,
		When:   b.When,
	}
	t2 := Booking{
		Name:   booking.Name,
		Policy: booking.Policy,
		Slot:   booking.Slot,
		User:   booking.User,
		When:   booking.When,
	}

	if t1 != t2 { //spam submission with non-matching details
		return errors.New("could not verify booking details")
	}

	if b.When.End.Before(s.now()) {
		return errors.New("cannot cancel booking that has already ended")
	}

	msg := "cancelling a started booking failed because "

	sl, ok := s.Slots[b.Slot]

	if !ok { //won't happen unless manifest and bookings out of sync
		return errors.New(msg + "slot " + b.Slot + " not found")
	}

	r, ok := s.Resources[sl.Resource]

	if !ok { //won't happen unless manifest and bookings out of sync
		return errors.New(msg + "resource " + sl.Resource + " not found")
	}

	if b.Started {

		if s.DisableCancelAfterUse {
			return errors.New("cannot cancel booking that has already been used")
		}

		// Booking has started so we will need to POST a deny request to the relay(s)
		// assume a manifest may have more than one relay
		// and that therefore even an experiment may have more than one relay
		// although that is more of an edge case.
		// task: map all the relay urls being used
		// slot -> resource -> streams -> url

		um := make(map[string]bool) //map of URLs from streams (this de-duplicates urls)

		// streams
		for _, k := range r.Streams {
			st, ok := s.Streams[k]
			if !ok { //won't happen unless manifest and bookings out of sync
				return errors.New(msg + "stream " + k + " not found")
			}

			um[st.URL] = true
		}

		for URL := range um {

			if s.denyRequests == nil {
				msg = msg + "deny requests channel is nil"
				log.WithFields(log.Fields{"user": b.User, "booking": b.Name}).Error(msg)
				return errors.New(msg)
			}
			c := make(chan string)
			s.denyRequests <- deny.Request{
				Result:    c,
				URL:       strings.TrimPrefix(URL, "http://"), //deny.Client scheme must be http
				BookingID: b.Name,
				ExpiresAt: b.When.End.Unix(),
			}

		DONE:
			for {
				select {
				case result, ok := <-c:
					if ok && result == "ok" {
						// deny request was successful
						log.WithFields(log.Fields{"user": b.User, "booking": b.Name}).Info("access cancelled at relay")
						break DONE
					} else {
						msg = msg + " error cancelling access at relay " + result
						log.WithFields(log.Fields{"user": b.User, "booking": b.Name}).Error(msg)
						return errors.New(msg)
					}
				case <-time.After(s.requestTimeout):
					msg = msg + " timed out cancelling access at relay " + URL
					log.WithFields(log.Fields{"user": b.User, "booking": b.Name}).Error(msg)
					return errors.New(msg)
				}
			}

		}

		// ok to cancel if get to here

	}

	// delete in the resource
	p, ok := s.Policies[booking.Policy]
	if !ok {
		return errors.New(msg + "could not find policy " + booking.Policy)
	}
	if !p.EnforceUnlimitedUsers { //if we aren't allowing unlimited users, then we made a resource booking
		err := r.Diary.Delete(booking.Name) //so delete that booking to allow others to use the cancelled time
		if err != nil {
			return errors.New(msg + "could not delete resource booking " + err.Error())
		}
	}

	delete(s.Bookings, b.Name)

	b.Cancelled = true
	b.CancelledAt = s.now()
	b.CancelledBy = cancelledBy

	s.OldBookings[b.Name] = b

	// adjust usage for user - original usage charge was booking length

	originalCharge := b.When.End.Sub(b.When.Start)
	p, err := s.getPolicy(b.Policy)
	if err != nil {
		msg := "cannot cancel booking because cannot get policy: " + err.Error()
		log.WithFields(log.Fields{"user": b.User, "booking": b.Name}).Error(msg)
		return errors.New(msg)
	}

	usage, err := calculateUsage(*b, p)

	if err != nil {
		msg := "cannot cancel booking because cannot calculate usage to refund: " + err.Error()
		log.WithFields(log.Fields{"user": b.User, "booking": b.Name}).Error(msg)
		return errors.New(msg)
	}

	refund := originalCharge - usage

	// get Usage tracker so we can modify it
	u, ok := s.Users[b.User]

	if !ok { //might happen if server is restarted, old booking restored but user has not made any new bookings yet
		// could be a prompt to create users for restored bookings ....
		msg := "cancelled but could not refund usage to unknown user " + b.User
		log.WithFields(log.Fields{"user": b.User, "booking": b.Name}).Error(msg)
		return errors.New(msg)
	}

	*u.Usage[b.Policy] = *u.Usage[b.Policy] - refund //refund reduces usage

	s.Users[b.User] = u

	log.WithFields(log.Fields{"user": b.User, "booking": b.Name}).Info("booking cancelled")
	return nil

}

// CheckBooking returns nil error if booking is ok, or an error and a slice of messages describing issues
// doesn't need a mutex, as is a support function
func (s *Store) checkBooking(b Booking) (error, []string) {
	msg := []string{}

	if b.Name == "" {
		msg = append(msg, "missing name")
	}

	if b.Policy == "" {
		msg = append(msg, b.Name+" missing policy")
	}
	if b.Slot == "" {
		msg = append(msg, b.Name+" missing slot")
	}
	if b.User == "" {
		msg = append(msg, b.Name+" missing user")
	}

	if (b.When == interval.Interval{}) {
		msg = append(msg, b.Name+" missing when")
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	if _, ok := s.Policies[b.Policy]; !ok {
		msg = append(msg, b.Name+" policy "+b.Policy+" not found")
	}
	if _, ok := s.Slots[b.Slot]; !ok {
		msg = append(msg, b.Name+" slot "+b.Slot+" not found")
	}

	// we don't check whether user exists, because we create them as needed

	if len(msg) > 0 {
		return errors.New("missing references"), msg
	}

	return nil, []string{}
}

// DeleteGroupFor removes the group from the user's list of allowed groups, and deletes any
// current bookings they have policies that are only accessible to the user via that group
func (s *Store) DeleteGroupFor(user, group string) error {

	where := "store.DeleteGroupFor"
	log.Trace(where + " awaiting lock")
	s.Lock()
	log.Trace(where + " has lock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released lock")
	}()

	u, ok := s.Users[user]

	if !ok {
		return errors.New("user " + user + " not found")
	}

	g, ok := s.Groups[group]

	if !ok {
		return errors.New("group " + group + " not found")
	}

	delete(u.Groups, group)

	s.Users[user] = u

	// delete any bookings this user has under this group
	// that are not covered by the same policy appearing
	// in another group the user has

	current := make(map[string]bool)
	remove := make(map[string]bool)

	for k := range u.Groups {
		current[k] = true
	}

	for _, k := range g.Policies { //remove policies not found in the policies of the remaining groups
		if _, ok := current[k]; !ok {
			remove[k] = true
		}
	}

	// get bookings so we can check policies in use
	bm, err := s.getBookingsFor(user)

	if err != nil {
		return err
	}

	for _, v := range bm {

		if _, ok := remove[v.Policy]; ok { //remove booking since its policy is in the remove list

			err = s.cancelBooking(v, "deletePolicy")

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ExportBookings returns a map of all current/future bookings
func (s *Store) ExportBookings() map[string]Booking {
	where := "store.ExportBookings"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	bm := make(map[string]Booking)

	for k, v := range s.Bookings {
		bm[k] = *v
	}

	return bm
}

// ExportManifest returns the manifest from the store
func (s *Store) ExportManifest() Manifest {

	where := "store.ExportManifest"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	// We store the full description in the store for convenience
	// but the manifest only has the name of the description in the Group
	// as a reference to the description elsewhere in the manifest
	// so we restore that format on export by removing all description
	// except for the description reference
	gm := make(map[string]Group)
	for k, v := range s.Groups {
		gm[k] = Group{
			Description: v.DescriptionReference,
			Policies:    v.Policies,
		}
	}

	// We store the full description in the store for convenience
	// but the manifest only has the name of the description in the UI
	// as a reference to the description elsewhere in the manifest
	// so we restore that format on export by removing all description
	// except for the description reference
	uis := make(map[string]UI)
	for k, v := range s.UIs {
		uis[k] = UI{
			Description:     v.DescriptionReference,
			URL:             v.URL,
			StreamsRequired: v.StreamsRequired,
		}
	}

	// Resources have diary pointers which we should nullify by omission for security and readability
	rm := make(map[string]Resource)
	for k, v := range s.Resources {
		rm[k] = Resource{
			ConfigURL:   v.ConfigURL,
			Description: v.Description,
			Streams:     v.Streams,
			Tests:       v.Tests,
			TopicStub:   v.TopicStub,
		}
	}

	return Manifest{
		Descriptions:  s.Descriptions,
		DisplayGuides: s.DisplayGuides,
		Groups:        gm,
		Policies:      s.Policies,
		Resources:     rm,
		Slots:         s.Slots,
		Streams:       s.Streams,
		UIs:           uis,
		UISets:        s.UISets,
		Windows:       s.Windows,
	}
}

// ExportOldBookings returns a map by name of old bookings
func (s *Store) ExportOldBookings() map[string]Booking {

	where := "store.ExportOldBookings"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	bm := make(map[string]Booking)

	for k, v := range s.OldBookings {
		bm[k] = *v
	}

	return bm
}

// ExportUsers returns a map of users, listing the names of bookings, old bookings, policies and
// their usage to date by policy name
func (s *Store) ExportUsers() map[string]UserExternal {

	where := "store.ExportUsers"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	um := make(map[string]UserExternal)

	for k, v := range s.Users {

		bs := []string{}
		obs := []string{}
		gs := []string{}
		ds := make(map[string]string)

		for k := range v.Bookings {
			b := k
			bs = append(bs, b)
		}

		for k := range v.OldBookings {
			ob := k
			obs = append(obs, ob)
		}
		for k := range v.Groups {
			g := k
			gs = append(gs, g)
		}
		for k, v := range v.Usage {
			ds[k] = HumaniseDuration(*v)
		}

		um[k] = UserExternal{
			Bookings:    bs,
			Groups:      gs,
			OldBookings: obs,
			Usage:       ds,
		}
	}

	return um
}

// GetActivity returns an activity associated with a booking, or an error
// if the booking is invalid in some way
func (s *Store) GetActivity(booking Booking) (Activity, error) {

	where := "store.GetActivity"
	log.Trace(where + " awaiting lock")
	s.Lock()
	log.Trace(where + " has lock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released lock")
	}()

	err := s.validateBooking(booking)

	if err != nil {
		return Activity{}, err
	}

	b, ok := s.Bookings[booking.Name]
	if !ok {
		return Activity{}, errors.New("not found")
	}

	b.Started = true

	s.Bookings[booking.Name] = b

	sl, ok := s.Slots[b.Slot]

	if !ok {
		return Activity{}, errors.New("slot " + b.Slot + " not found")
	}

	d, ok := s.Descriptions[sl.Description]

	if !ok {
		return Activity{}, errors.New("description " + sl.Description + " not found")
	}

	r, ok := s.Resources[sl.Resource]

	if !ok {
		return Activity{}, errors.New("resource " + sl.Resource + " not found")
	}

	a := Activity{
		BookingID:   b.Name,
		Description: d,
		ConfigURL:   r.ConfigURL,
		NotBefore:   b.When.Start,
		ExpiresAt:   b.When.End,
		Streams:     make(map[string]Stream),
		UIs:         []UIDescribed{},
	}

	// streams
	for _, k := range r.Streams {
		st, ok := s.Streams[k]
		if !ok {
			return Activity{}, errors.New("stream " + k + " not found")
		}
		//Streams are prototypes, so make the specific topic
		st.Topic = r.TopicStub + "-" + k
		a.Streams[k] = st
	}

	//UIs
	uis, ok := s.UISets[sl.UISet]
	if !ok {
		return Activity{}, errors.New("ui_set" + sl.UISet + " not found")
	}

	for _, k := range uis.UIs {
		ui, ok := s.UIs[k]
		if !ok {
			return Activity{}, errors.New("ui" + k + " not found")
		}

		// omit the DescriptionReference field for readability
		a.UIs = append(a.UIs, UIDescribed{
			Description:     ui.Description,
			URL:             ui.URL,
			StreamsRequired: ui.StreamsRequired,
		})
	}

	return a, nil
}

// GetAvailability returns a list of intervals for which a given slot is available under a given policy, or an error if the slot or policy is not found. The policy contains aspects such as look-ahead which may limit the window of availability.
func (s *Store) GetAvailability(slot string) ([]interval.Interval, error) {

	where := "store.GetAvailability"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	return s.getAvailability(slot)

}

// getAvailability is for internal use only (e.g. in MakeBooking)
func (s *Store) getAvailability(slot string) ([]interval.Interval, error) {

	sl, ok := s.Slots[slot]

	if !ok {
		return []interval.Interval{}, errors.New("slot " + slot + " not found")
	}

	p, ok := s.Policies[sl.Policy]
	if !ok {
		return []interval.Interval{}, errors.New("policy " + sl.Policy + " not found")
	}

	bk, err := s.getSlotBookings(slot)

	if err != nil {
		return []interval.Interval{}, err
	}

	// strip the intervals from the bookings
	bi := []interval.Interval{}

	for _, b := range bk {
		bi = append(bi, b.When)
	}

	// get pointer to filter for policy
	fp, ok := s.Filters[sl.Window]

	if !ok {
		return []interval.Interval{}, errors.New("cannot find filter for window " + sl.Window + " in slot " + slot)
	}

	dl := fp.Export()

	// merge bookings with the times that are blocked by the policy
	unavailable := interval.Merge(append(bi, dl...))

	start := s.now()

	end := interval.DistantFuture //interval.Infinity causes parsing problems in API, so choose something saner (from a parsing point of view)

	if p.EnforceBookAhead {
		fmt.Println()
		end = start.Add(p.BookAhead)
	}

	if len(unavailable) == 0 { // no bookings, no blocked periods
		a := []interval.Interval{
			interval.Interval{
				Start: start,
				End:   end,
			},
		}

		return a, nil
	}

	fa := availability(unavailable, start, end)

	return fa, nil

}

//GetBooking returns a booking given a bookingname
func (s *Store) GetBooking(booking string) (Booking, error) {
	where := "store.GetBooking"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	v, ok := s.Bookings[booking]

	if !ok {
		return Booking{}, errors.New("booking not found")
	}

	return *v, nil
}

// GetBookingsFor returns a slice of all the current bookings for the given user
// don't use mutex because called from functions that do
func (s *Store) GetBookingsFor(user string) ([]Booking, error) {
	where := "store.GetBookingsFor"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()
	return s.getBookingsFor(user)
}

// getBookingsFor returns a slice of all the current bookings for the given user
// don't use mutex because called from functions that do
// for internal use - calling function must take the lock
func (s *Store) getBookingsFor(user string) ([]Booking, error) {

	if _, ok := s.Users[user]; !ok {
		return []Booking{}, errors.New("user not found")
	}

	s.pruneUserBookings(user)

	b := []Booking{}

	for _, v := range s.Bookings {
		if user == v.User {
			b = append(b, *v)
		}
	}

	return b, nil

}

// GetDescription returns a description if found
func (s *Store) GetDescription(name string) (Description, error) {

	where := "store.GetDescription"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	return s.getDescription(name)

}

// getDescription returns a description if found
// no lock - internal use only
func (s *Store) getDescription(name string) (Description, error) {

	d, ok := s.Descriptions[name]

	if !ok {
		return Description{}, errors.New("not found")
	}

	return d, nil
}

// GetDisplayGuide returns a diplay guide if found
func (s *Store) GetDisplayGuide(name string) (DisplayGuide, error) {

	where := "store.GetDisplayGuide"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	return s.getDisplayGuide(name)

}

// getDescription returns a description if found
// no lock - internal use only
func (s *Store) getDisplayGuide(name string) (DisplayGuide, error) {

	d, ok := s.DisplayGuides[name]

	if !ok {
		return DisplayGuide{}, errors.New("not found")
	}

	return d, nil
}

func (s *Store) getGroup(name string) (GroupDescribed, error) {
	g, ok := s.Groups[name]

	if !ok {
		return GroupDescribed{}, errors.New("not found")
	}

	return g, nil
}

// GetGroup returns a group if found
// do not use internally - it takes the lock
func (s *Store) GetGroup(name string) (GroupDescribed, error) {

	where := "store.GetGroup"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	return s.getGroup(name)
}

// GetOldBookingsFor returns a slice of all the old bookings for the given user
// don't use mutex because called from functions that do
func (s *Store) GetOldBookingsFor(user string) ([]Booking, error) {
	where := "store.GetOldBookingsFor"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()
	return s.getOldBookingsFor(user)
}

// getOldBookingsFor returns a slice of all the old bookings for the given user
// don't use mutex because called from functions that do
// internal use only - calling function must handle taking the lock
func (s *Store) getOldBookingsFor(user string) ([]Booking, error) {

	if _, ok := s.Users[user]; !ok {
		return []Booking{}, errors.New("user not found")
	}

	s.pruneUserBookings(user)

	b := []Booking{}

	for _, v := range s.OldBookings {
		if user == v.User {
			b = append(b, *v)
		}
	}

	return b, nil
}

func (s *Store) getPolicy(name string) (Policy, error) {
	p, ok := s.Policies[name]

	// remove the slotmap, not for external use
	// this uninitilialised form is easier to test
	// because you can just omit the SlotMap field
	// from the expected object you are checking against
	// in the test and it will be the same as this now
	var sm map[string]bool
	p.SlotMap = sm

	if !ok {
		return Policy{}, errors.New("not found")
	}

	return p, nil
}

// GetPolicy returns a policy if found
// this is not used internally
func (s *Store) GetPolicy(name string) (Policy, error) {

	where := "store.GetPolicy"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	return s.getPolicy(name)
}

// GetGroupsFor returns a list of groups that a user has access to
func (s *Store) GetGroupsFor(user string) ([]string, error) {

	where := "store.GetGroupsFor"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	if _, ok := s.Users[user]; !ok {
		return []string{}, errors.New("user not found")
	}

	g := []string{}

	for k := range s.Users[user].Groups {
		g = append(g, k) //append group name
	}
	return g, nil
}

// GetPolicyStatusFor returns usage, and counts of current and old bookings
// Needs a write lock because it prunes
func (s *Store) GetPolicyStatusFor(user, policy string) (PolicyStatus, error) {

	where := "store.GetPolicyStatusFor"
	log.Trace(where + " awaiting lock")
	s.Lock()
	log.Trace(where + " has lock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released lock")
	}()

	if _, ok := s.Users[user]; !ok {
		return PolicyStatus{}, errors.New("user not found")
	}

	if _, ok := s.Policies[policy]; !ok {
		return PolicyStatus{}, errors.New("policy not found")
	}

	if _, ok := s.Users[user].Usage[policy]; !ok { // no usage according to that policy
		ut, err := time.ParseDuration("0s")
		if err != nil { // shouldn't get this error because parsing "0s" is known good
			return PolicyStatus{}, errors.New("no usage found")
		}
		// return a successful result with zero usage because every new
		// user will at some point have zero usage on a new policy and this
		// makes the display logic easier to handle on the client side
		// Do NOT add a new tracker to the store because a GET query should not mutate state
		// For example, it could imply that the user once had permission to book this policy
		// but that concern is handled in the authorisation middlelayer outside this package,
		// and it is possible that the middleware implementation might let users check policy
		// status without having permission to book, because they rightly assume a GET must not
		// mutate state, so we don't want to create a privilege escalation by creating a usage tracker
		// that imples permission to book was once held, when perhaps it wasn't.
		return PolicyStatus{Usage: ut}, nil

	}

	bp := []Booking{}
	obp := []Booking{}

	b, err := s.getBookingsFor(user)
	if err != nil {
		return PolicyStatus{}, err
	}

	for _, v := range b {
		if policy == v.Policy {
			bp = append(bp, v)
		}
	}

	ob, err := s.getOldBookingsFor(user)
	if err != nil {
		return PolicyStatus{}, err
	}
	for _, v := range ob {
		if policy == v.Policy {
			obp = append(obp, v)
		}
	}

	ps := PolicyStatus{
		CurrentBookings: int64(len(bp)),
		OldBookings:     int64(len(obp)),
		Usage:           *(s.Users[user].Usage[policy]),
	}
	return ps, nil
}

// GetSlot returns a slot if found
func (s *Store) GetSlot(name string) (Slot, error) {

	where := "store.GetSlot"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	return s.getSlot(name)

}

// ExportManifest returns the manifest from the store
func (s *Store) GetResources() map[string]Resource {

	where := "store.GetResources"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	// Resources have diary pointers which we should nullify by omission for security and readability
	rm := make(map[string]Resource)
	for k, v := range s.Resources {
		rm[k] = Resource{
			ConfigURL:   v.ConfigURL,
			Description: v.Description,
			Streams:     v.Streams,
			Tests:       v.Tests,
			TopicStub:   v.TopicStub,
		}
	}

	return rm

}

// getSlot returns a slot if found
// no lock - internal use only
func (s *Store) getSlot(name string) (Slot, error) {

	d, ok := s.Slots[name]

	if !ok {
		return Slot{}, errors.New("not found")
	}

	return d, nil
}

// GetSlotBookings gets bookings as far as ahead as the diary holds them
// It's up to the caller to handle any pagination that might be required
// Does not take a lock because the calling function(s) handles that
// for interal use only - calling function must take the lock
func (s *Store) getSlotBookings(slot string) ([]diary.Booking, error) {

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

// getResourceIsAvailable checks the underlying resource's availability
// this one does not take a lock, so it can be used within other functions
// that already take a lock
func (s *Store) getResourceIsAvailable(resource string) (bool, string, error) {

	r, ok := s.Resources[resource]

	if !ok {
		return false, "", errors.New("resource " + resource + " not found")
	}

	ok, reason := r.Diary.IsAvailable()

	return ok, reason, nil

}

// GetResourceIsAvailable checks the underlying resource's availability
// Use this version when calling externally
func (s *Store) GetResourceIsAvailable(resource string) (bool, string, error) {
	where := "store.GetResourceIsAvailable"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	return s.getResourceIsAvailable(resource)
}

// getSlotIsAvailable checks the underlying resource's availability
// this one does not take a lock, so it can be used within other functions
// that already take a lock
func (s *Store) getSlotIsAvailable(slot string) (bool, string, error) {
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

// GetSlotIsAvailable checks the underlying resource's availability
// Use this version when calling externally
func (s *Store) GetSlotIsAvailable(slot string) (bool, string, error) {
	where := "store.GetSlotIsAvailable"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	return s.getSlotIsAvailable(slot)
}

// GetStoreStatusAdmin returns the status of the store with entity counts
func (s *Store) GetStoreStatusAdmin() StoreStatusAdmin {

	where := "store.GetStoreStatusAdmin"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	return StoreStatusAdmin{
		Locked:       s.Locked,
		Message:      s.Message,
		Now:          s.now(),
		Bookings:     int64(len(s.Bookings)),
		Descriptions: int64(len(s.Descriptions)),
		Filters:      int64(len(s.Filters)),
		OldBookings:  int64(len(s.OldBookings)),
		Policies:     int64(len(s.Policies)),
		Resources:    int64(len(s.Resources)),
		Slots:        int64(len(s.Slots)),
		Streams:      int64(len(s.Streams)),
		UIs:          int64(len(s.UIs)),
		UISets:       int64(len(s.UISets)),
		Users:        int64(len(s.Users)),
		Windows:      int64(len(s.Windows)),
	}
}

// GetStoreStatusUser returns the store status without entity counts
func (s *Store) GetStoreStatusUser() StoreStatusUser {

	where := "store.GetStoreStatusUser"
	log.Trace(where + " awaiting Rlock")
	s.Lock()
	log.Trace(where + " has Rlock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released Rlock")
	}()

	return StoreStatusUser{
		Locked:  s.Locked,
		Message: s.Message,
		Now:     s.now(),
	}
}

func (s *Store) GraceCheck(bookings []string) {

	if s.Locked {
		// don't affect bookings but equally, don't do someone out of
		// the autocancellation's lower usage charge compared to just
		// not taking up the booking. So push bookings back, for processing later.
		later := time.Now().Add(s.GraceRebound)
		for _, b := range bookings {
			if s.Checker != nil { //incase checker is being refreshed
				s.Checker.Push(later, b)
			}
		}

	}

	for _, name := range bookings {
		b, err := s.GetBooking(name)
		if err != nil {
			continue //skip this booking - probably cancelled
		}
		p, err := s.GetPolicy(b.Policy)
		if err != nil {
			continue
		}
		if !p.EnforceGracePeriod {
			continue
		}
		if !b.Started {
			s.CancelBooking(b, "auto-grace-check")
		}

	}

}

// HumaniseDuration returns a concise human readable string representing the duration
func HumaniseDuration(t time.Duration) string {
	return t.Round(time.Second).String()
}

// MakeBooking makes bookings for users, according to the policy
// If a user does not exist, one is created.
// APIs for users should call this version
// do not use mutex, because it calls function that handles that
func (s *Store) MakeBooking(slot, user string, when interval.Interval) (Booking, error) {
	where := "store.MakeBooking"
	log.Trace(where + " awaiting lock")
	s.Lock()
	log.Trace(where + " has lock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released lock")
	}()
	name := uuid.New().String()

	b, err := s.makeBookingWithName(slot, user, when, name, true) //check groups

	msg := "successful booking"

	if err != nil {
		msg = "failed booking because " + err.Error()
	}

	log.WithFields(log.Fields{"slot": slot, "user": user, "start": when.Start.String(), "end": when.End.String(), "name": name}).Info(msg)

	return b, err

}

// MakeBookingWithID makes bookings for users, according to the policy
// If a user does not exist, one is created.
// The booking ID is set by the caller, so that bookings can be edited/replaced
// This version should only be called by Admin users
func (s *Store) MakeBookingWithName(slot, user string, when interval.Interval, name string, checkGroup bool) (Booking, error) {
	where := "store.MakeBookingWithName"
	log.Trace(where + " awaiting lock")
	s.Lock()
	log.Trace(where + " has lock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released lock")
	}()

	b, err := s.makeBookingWithName(slot, user, when, name, checkGroup) //leave up to admin to enfore group membership on this booking (e.g. might be a one off booking that is not intended to confer future access to any policies

	msg := "successful booking"

	if err != nil {
		msg = "failed booking because " + err.Error()
	}

	log.WithFields(log.Fields{"slot": slot, "user": user, "start": when.Start.String(), "end": when.End.String(), "name": name}).Info(msg)

	return b, err
}

// MakeBookingWithID makes bookings for users, according to the policy
// If a user does not exist, one is created.
// The booking ID is set by the caller, so that bookings can be edited/replaced
// Internal usage only - no locks
// making a booking does not confer the right to make future bookings
// the access to a policy is determined by the groups associated with a user
// checkGroup is for when replaceBookings is making bookings without reference to groups
// e.g. for pre-making identities that can't be operated by the user, there is no need for groups because that would allow other bookings to be made
// by the student, potentially
func (s *Store) makeBookingWithName(slot, user string, when interval.Interval, name string, checkGroup bool) (Booking, error) {

	sl, ok := s.Slots[slot]

	if !ok {
		return Booking{}, errors.New("slot " + slot + " not found")
	}

	p, ok := s.Policies[sl.Policy]

	if !ok {
		msg := "policy " + sl.Policy + " not found"
		log.Warnf("makebooking: %s %s %s %v %s: %s", sl.Policy, slot, user, when, name, msg)
		return Booking{}, errors.New(msg)
	}

	_, ok = p.SlotMap[slot]

	if !ok {
		return Booking{}, errors.New("slot " + slot + " not in policy " + sl.Policy)
	}

	r, ok := s.Resources[sl.Resource]

	if !ok {
		return Booking{}, errors.New("resource " + sl.Resource + " not found")
	}

	// to avoid replay of policies known to user, that we've revoked, but that still exist,
	// we check if policy is covered by any current groups allocated to the user
	// this doesn't prevent replay of joining the original group again
	// but that can't happen if that group no longer exists.
	u, ok := s.Users[user]

	if ok {

		if checkGroup {

			pm := make(map[string]bool)

			for gn := range u.Groups {
				if g, ok := s.Groups[gn]; ok {
					for _, p := range g.Policies {
						pm[p] = true
					}
				}
			}

			if _, ok := pm[sl.Policy]; !ok {
				return Booking{}, errors.New("user " + user + " belongs to no group that includes this policy")
			}
		}

		// pass if not checking group

	} else {

		if checkGroup {
			//not found, don't create user, because will not be authorised for the group
			return Booking{}, errors.New("user " + user + " not found")
		} else {
			//we're prob doing an admin task like replace bookings, so create a new user (without conferring any further rights to book - we don't know about any groups here anyway)
			u := NewUser()
			s.Users[user] = u
		}

	}

	// get the user again, in case we just created it
	u, ok = s.Users[user]
	if !ok {
		return Booking{}, errors.New("user " + user + " was not found and creation failed")
	}

	// check if too many bookings already
	if p.EnforceMaxBookings {

		// remove stale entries from user's list of current bookings
		s.pruneUserBookings(user)

		// first check how many bookings under this policy already
		cb := []string{}

		for k, v := range u.Bookings {
			if v.Policy == sl.Policy {
				cb = append(cb, k)
			}
		}
		currentBookings := int64(len(cb))

		if currentBookings >= p.MaxBookings {
			return Booking{}, errors.New("you currently have " +
				strconv.FormatInt(currentBookings, 10) +
				" current/future bookings which is at or exceeds the limit of " +
				strconv.FormatInt(p.MaxBookings, 10) +
				" for policy " +
				sl.Policy)
		}

	}

	// check if booking is within slot window
	fp, ok := s.Filters[sl.Window]

	if !ok {
		return Booking{}, errors.New("window filter " + sl.Window + " not found")
	}

	if !fp.Allowed(when) {
		return Booking{}, errors.New("bookings cannot be made outside the window for the slot")
	}

	// check if booking is within bookahead window
	if p.EnforceBookAhead {
		if when.End.After(s.now().Add(p.BookAhead)) {
			return Booking{}, errors.New("bookings cannot be made more than " +
				HumaniseDuration(p.BookAhead) +
				" ahead of the current time")
		}
	}

	// check if booking requested starts in past

	now := s.now()

	if p.EnforceAllowStartInPast { //make allowance for delays in receiving request, if policy permits
		now = now.Add(-1 * p.AllowStartInPastWithin) //adjust the now value to perform the check required by the policy
	}

	if when.Start.Before(now) {
		if p.EnforceAllowStartInPast {
			return Booking{}, errors.New("booking cannot start more than " + HumaniseDuration(p.AllowStartInPastWithin) + " in the past")
		} else {
			return Booking{}, errors.New("booking cannot start in the past (start: " + when.Start.String() + ", now:" + now.String() + ")")
		}
	}

	// check if booking is starting soon enough, if policy enforces StartsWithin
	if p.EnforceStartsWithin {

		now = s.now().Add(p.StartsWithin) //get fresh, undjusted, value of now to avoid incorrect policy decisions, and adjust as required to make the check

		if when.Start.After(now) {
			return Booking{}, errors.New("booking cannot start more than " + HumaniseDuration(p.StartsWithin) + " in the future")
		}
	}

	if p.EnforceNextAvailable {

		// check if booking is starting soon enough after the earliest current booking, or now, if there is no booking, if NextAvailable is enforced
		a, err := s.getAvailability(slot)

		if err != nil {
			return Booking{}, errors.New("enforcing next available policy setting failed because " + err.Error())
		}

		if len(a) < 1 {
			return Booking{}, errors.New("enforcing next available policy setting because availability list was empty")
		}

		latest := a[0].Start.Add(p.NextAvailable)

		if when.Start.After(latest) {
			return Booking{}, errors.New("due to next available policy setting, booking cannot start more than " + HumaniseDuration(p.NextAvailable) + " after the last booking ends, i.e. " + latest.String())
		}

	}

	now = s.now() // return now to the current value in case we use it again and overlook that we have adjusted it in the checks above

	// check for existing usage tracker for this policy?
	_, ok = u.Usage[sl.Policy]

	if !ok { //create usage tracker (always track usage, even if not limited)
		ut, err := time.ParseDuration("0s")
		if err != nil {
			return Booking{}, errors.New("could not initialise user tracker for user " +
				user +
				" because " +
				err.Error())
		}
		u.Usage[sl.Policy] = &ut
	}

	duration := when.End.Sub(when.Start)

	currentUsage := *u.Usage[sl.Policy]
	newUsage := currentUsage + duration

	// Check if usage allowance sufficient
	if p.EnforceMaxUsage && (newUsage > p.MaxUsage) {
		remaining := p.MaxUsage - currentUsage
		return Booking{}, errors.New("requested duration of " +
			HumaniseDuration(duration) +
			" exceeds remaining usage limit of " +
			HumaniseDuration(remaining))
	}

	// Check minimum duration is ok
	if p.EnforceMinDuration && (duration < p.MinDuration) {
		return Booking{}, errors.New("requested duration of " +
			HumaniseDuration(duration) +
			" shorter than minimum permitted duration of " +
			HumaniseDuration(p.MinDuration))
	}

	// check maximum duration is ok
	if p.EnforceMaxDuration && (duration > p.MaxDuration) {
		return Booking{}, errors.New("requested duration of " +
			HumaniseDuration(duration) +
			" longer than maximum permitted duration of " +
			HumaniseDuration(p.MaxDuration))
	}

	// If this is a simulation with no hardware or other resource constraints, we don't make bookings in the diary, we just grant access
	// seeing as other policy aspects have been satisfied

	if !p.EnforceUnlimitedUsers { //skip booking if we allow unlimited users

		// see if the booking can be made ....
		err := r.Diary.Request(when, name)

		if err != nil {
			return Booking{}, err
		}
	}

	// successful (or skipped) booking, so update usage tracker with value we calculated earlier
	u.Usage[sl.Policy] = &newUsage

	booking := Booking{
		Cancelled:   false,
		Name:        name,
		Policy:      sl.Policy,
		Slot:        slot,
		Started:     false,
		Unfulfilled: false,
		User:        user,
		When:        when,
	}

	s.Bookings[name] = &booking
	s.Users[user].Bookings[name] = &booking

	// register for autocancellation if required by policy
	if p.EnforceGracePeriod {
		checkTime := when.Start.Add(p.GracePeriod)
		log.Debugf("makebooking: requesting grace check %s at %s", name, checkTime.String())
		err := s.Checker.Push(checkTime, name)
		if err != nil {
			log.Errorf("makebooking failed to request grace check for %s at %s because %s", name, checkTime.String(), err.Error())
		}
	} else {
		log.Debugf("makebooking: grace period is not being enforced for %s", name)
	}

	return booking, nil

}

// PruneAll is maintenance operation ensuring all bookings are moved
// to the old bookings list, wherever that touches our implementation
func (s *Store) PruneAll() {
	where := "store.PruneAll"
	log.Trace(where + " awaiting lock")
	s.Lock()
	log.Trace(where + " has lock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released lock")
	}()

	s.pruneBookings()
	s.pruneDiaries()
	s.pruneUserBookingsAll()

}

// PruneDiaries is maintenance operation that moves expired bookings from
// the map of current bookings to the map of old bookings
// because they are rebalanced, so don't do too frequently.
// don't use mutex because this is called from other functions
func (s *Store) pruneBookings() {

	stale := make(map[string]*Booking)

	for k, v := range s.Bookings {
		if s.now().After(v.When.End) {
			stale[k] = v
		}
	}

	for k, v := range stale {
		s.OldBookings[k] = v
		delete(s.Bookings, k)
	}

}

// PruneDiaries is a maintenance operation to prune old bookings from diaries
// to make booking decisions faster. There is an overhead to pruning trees
// because they are rebalanced, so don't do too frequently.
// don't use mutex because this is called from other functions
// that already have the mutex
func (s *Store) pruneDiaries() {
	for _, r := range s.Resources {
		r.Diary.ClearBefore(s.now())
	}
}

// PruneUserBookings is a maintenace operation to move
// expired bookings from the map of bookings but only
// to do so for a given user (e.g. ahead of checking
// their policy limits on future bookings).
// don't use mutex because this is called from other functions
// that already have the mutex
func (s *Store) pruneUserBookings(user string) {

	u, ok := s.Users[user]

	if !ok {
		return //do nothing
	}

	stale := make(map[string]*Booking)

	for k, v := range u.Bookings {
		if s.now().After(v.When.End) {
			stale[k] = v
		} else if v.Cancelled { //TODO test we release bookings ok
			stale[k] = v
		}
	}

	for k, v := range stale {
		u.OldBookings[k] = v
		delete(u.Bookings, k)
	}

}

// Prune all user bookings during regular maintenance
// don't use mutex because this is called from other functions
// that already have the mutex
func (s *Store) pruneUserBookingsAll() {

	u := s.Users

	for k := range u {
		s.pruneUserBookings(k)
	}

}

// ReplaceBookings will replace all bookings with a new set
// each booking must be valid for the manifest, i.e. all
// references to other entities must be valid.
// Note that the manifest should be set first
// Diaries need to be cleared by cancelling bookings to refund
// usage to  users before making the replacement bookings through
// the standard method
func (s *Store) ReplaceBookings(bm map[string]Booking) (error, []string) {
	where := "store.ReplaceBookings"
	log.Trace(where + " awaiting lock")
	s.Lock()
	log.Trace(where + " has lock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released lock")
	}()
	// do not take the lock except where we need it below - because we call on functions that take the lock

	// Check bookings are individually sane given our manifest
	msg := []string{}

	for _, v := range bm {
		err, ms := s.checkBooking(v)
		if err != nil {
			for _, m := range ms {
				msg = append(msg, m)
			}
		}
	}

	if len(msg) > 0 {
		return errors.New("malformed booking"), msg
	}

	// bookings are ok, so clean house.

	//Stop our grace period checker, and clean it out
	s.Checker.Clean()

	// we want to refund our users, so go through each booking and cancel

	for k, v := range s.Bookings {
		err := s.cancelBooking(*v, "replaceManifest")
		if err != nil {
			msg = append(msg,
				"could not refund user "+
					v.User+" "+HumaniseDuration(v.When.End.Sub(v.When.Start))+
					" for replaced booking "+k+" on policy "+v.Policy)
		}
	}

	// can't delete bookings as we iterate over map, so just create a fresh map

	s.Bookings = make(map[string]*Booking)

	for k := range s.Resources {
		r := s.Resources[k]
		r.Diary = diary.New(k)
		s.Resources[k] = r
	}

	// Now make the bookings, respecting policy and usage
	for _, v := range bm {
		_, err := s.makeBookingWithName(v.Slot, v.User, v.When, v.Name, false) //ignore group check on bookings

		lm := "successful booking"

		if err != nil {
			lm = "failed booking because " + err.Error()
		}

		log.WithFields(log.Fields{"policy": v.Policy, "slot": v.Slot, "user": v.User, "start": v.When.Start.String(), "end": v.When.End.String(), "name": v.Name}).Info(lm)

		if err != nil {
			msg = append(msg, "booking "+v.Name+" failed because "+err.Error())
		}

		// s.Bookings is updated by MakeBookingWithID so we mustn't update it ourselves
	}

	return nil, []string{}
}

// ReplaceManifest overwrites the existing manifest with a new one i.e. does not retain existing elements from any previous manifests
// but it does retain non-Manifest elements such as bookings.
func (s *Store) ReplaceManifest(m Manifest) error {
	where := "store.ReplaceManifest"
	log.Trace(where + " awaiting lock")
	s.Lock()
	log.Trace(where + " has lock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released lock")
	}()

	// lock is taken after we check whether we need to alter the store (see below)
	err, _ := checkManifest(m)

	if err != nil {
		return err //user can call CheckDescriptions some other way if they want the manifest error details
	}

	// we can get errors making filters, so do that before doing anything destructive
	// even though we checked it with CheckManifest, we have to handle the errors
	fm := make(map[string]*filter.Filter)

	for k, w := range m.Windows {

		f := filter.New()
		err = f.SetAllowed(w.Allowed)
		if err != nil {
			return errors.New("failed to create allowed intervals for window " + k + ":" + err.Error())
		}
		err := f.SetDenied(w.Denied)
		if err != nil {
			return errors.New("failed to create denied intervals for window " + k + ":" + err.Error())
		}

		fm[k] = f
	}

	// we're going to do the replacement now, goodbye old manifest data.
	s.Filters = fm

	// Make new maps for our new entities (note this is m for manifest, not m for swagger models)
	s.Descriptions = m.Descriptions
	s.DisplayGuides = m.DisplayGuides
	s.Policies = m.Policies
	s.Resources = m.Resources
	s.Slots = m.Slots
	s.Streams = m.Streams
	s.UISets = m.UISets
	s.Windows = m.Windows

	status := "Loaded at " + s.now().Format(time.RFC3339)

	// SlotMap is used for checking if slots are listed in policy
	// DisplayGuidesMap is used for exporting complete policies
	for k, v := range s.Policies {
		v.SlotMap = make(map[string]bool)
		for _, k := range v.Slots {
			v.SlotMap[k] = true
		}
		v.DisplayGuidesMap = make(map[string]DisplayGuide)
		for _, k := range v.DisplayGuides {
			v.DisplayGuidesMap[k] = s.DisplayGuides[k]
		}
		s.Policies[k] = v
	}

	for k := range s.Resources {
		r := s.Resources[k]
		r.Diary = diary.New(k)
		s.Resources[k] = r
		// default to available because unavailable kit is the exception
		s.Resources[k].Diary.SetAvailable(status)
	}

	// populate UIs with descriptions now to save doing it repetively later
	s.UIs = make(map[string]UIDescribed)

	for k, v := range m.UIs {

		d, err := s.getDescription(v.Description)

		if err != nil {
			return err
		}

		uid := UIDescribed{
			Description:          d,
			DescriptionReference: v.Description,
			URL:                  m.UIs[k].URL,
			StreamsRequired:      m.UIs[k].StreamsRequired,
		}
		s.UIs[k] = uid
	}

	// populate the groups with the descriptions now, to save repetitively doing it later
	s.Groups = make(map[string]GroupDescribed)

	for k, v := range m.Groups {

		d, err := s.getDescription(v.Description)

		if err != nil {
			return err
		}

		gd := GroupDescribed{
			Description:          d,
			DescriptionReference: v.Description,
			Policies:             v.Policies,
		}
		s.Groups[k] = gd
	}

	return nil

}

// ReplaceOldBookings will replace the map of old bookings with the supplied list or return an error if the bookings have issues. All existing users are deleted, and replaced with users with usages that match the old bookings
// use ReplaceUserGroups to add permissions for users, do not bother with old dummy bookings because these confer
// no future booking privileges (now that we get allowed policies by checking a user's groups).
func (s *Store) ReplaceOldBookings(bm map[string]Booking) (error, []string) {
	where := "store.ReplaceOldBookings"
	log.Trace(where + " awaiting lock")
	s.Lock()
	log.Trace(where + " has lock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released lock")
	}()

	// Check bookings are individually sane given our manifest
	msg := []string{}

	for _, v := range bm {
		err, ms := s.checkBooking(v)
		if err != nil {
			for _, m := range ms {
				msg = append(msg, m)
			}
		}
	}

	if len(msg) > 0 {
		return errors.New("malformed booking"), msg
	}

	// bookings are ok, so clean house.

	// no need to handle any diaries or cancellations, because these are old bookings
	s.OldBookings = make(map[string]*Booking)

	// we don't refund any usages because we are removing all users too (will remake them to match replacemenet old bookings)
	s.Users = make(map[string]*User)

	// Map the bookings, and create new users and usage trackers to reflect the updated "old bookings"
	for k, v := range bm {

		ob := v //make local copy so we can get a pointer detached from the loop variable

		s.OldBookings[k] = &ob

		// add new user if does not exist
		if _, ok := s.Users[ob.User]; !ok {
			s.Users[ob.User] = NewUser()
		}

		// get user from map to allow editing of it
		u := s.Users[ob.User]

		// check for existing usage tracker for this policy?
		_, ok := u.Usage[ob.Policy]

		if !ok { //create usage tracker (always track usage, even if not limited)
			ut, err := time.ParseDuration("0s")

			if err != nil { //in practice, will not throw error because we know the string "0s" is `good`
				return errors.New("could not initialise user tracker"), []string{}

			}

			u.Usage[ob.Policy] = &ut
		}

		duration := ob.When.End.Sub(ob.When.Start)
		currentUsage := *u.Usage[ob.Policy]
		newUsage := currentUsage + duration

		// update usage tracker with duration of this booking
		u.Usage[ob.Policy] = &newUsage

		// replace edited user in map
		s.Users[ob.User] = u

	}

	return nil, []string{}

}

// Replace Users is not implemented because it would allow
// the consistency of the store to be broken (e.g. which users
// were associated with which bookings). As for usage, the
// ReplaceBookings method already handles adjustments to usage
// automatically so there is no need to edit users.
// If a user needs more usage allowance, then they need a new policy,
// rather than an adjustment to their old usage value.
func (s *Store) ReplaceUsers(u map[string]UserExternal) (error, []string) {
	return errors.New("not implemented"), []string{}
}

// ReplaceUsersGroups allows administrators to add and remove policies from
// users, e.g. to add or restrict access to experiments
// A user that does not exist, is created, and the groups added.
// Policies must exist or an error is thrown
func (s *Store) ReplaceUserGroups(u map[string][]string) (error, []string) {
	where := "store.ReplaceUserGroups"
	log.Trace(where + " awaiting lock")
	s.Lock()
	log.Trace(where + " has lock")
	defer func() {
		s.Unlock()
		log.Trace(where + " released lock")
	}()

	msg := []string{}

	for k, v := range u {
		// check all groups exist
		for _, g := range v {
			if _, ok := s.Groups[g]; !ok {
				msg = append(msg, "user "+k+" wants group "+g+" which does not exist")
			}
		}
	}

	if len(msg) > 0 {
		return errors.New("group(s) not found"), msg
	}

	for k, v := range u {
		if _, ok := s.Users[k]; !ok { //create new user if does not exist
			s.Users[k] = NewUser()
		}
		u := s.Users[k]
		gm := make(map[string]bool)
		for _, g := range v {
			gm[g] = true
		}
		u.Groups = gm

	}

	return nil, []string{}
}

// Run handles the regular pruning of bookings and autocancellation checks
func (s *Store) Run(ctx context.Context, pruneEvery time.Duration, checkEvery time.Duration) {

	go s.denyClient.Run(ctx) //setup already done in the With/Set functions

	go func() { //This needs to run more often than the pruning operation, because it frees unused bookings for others. Suggest one minute (balance CPU usage and timeliness of checks)
		defer func() {
			log.Trace("store.Run checking goro stopped")
		}()
		expired := make(chan []string)
		s.Checker.Run(ctx, checkEvery, expired)
		log.Debug("store will grace check bookings every " + checkEvery.String())
		for {
			select {
			case <-ctx.Done():
				log.Trace("store stopped grace checking bookings permanently")
				return
			case bookings := <-expired:
				log.Debugf("Gracechecking %d bookings", len(bookings))
				s.GraceCheck(bookings)
			}
		}
	}()
	go func() { //this is a routine maintenance operation to keep data structures free of stale data, and can run as infrequently, suggest 1 hour if most bookings are 30min+ sessions.
		log.Debug("store will prune bookings & diaries every " + pruneEvery.String())
		defer func() {
			log.Trace("store.Run pruning goro stopped")
		}()
		for {

			select {
			case <-ctx.Done():
				log.Trace("store pruning stopped permanently")
				return
			case <-time.After(pruneEvery):
				log.Trace("store pruning all bookings & diaries at time " + s.Now().String()) //must be mutexed version because Run does not take the lock
				s.PruneAll()
			}
		}
	}()
}

// SetResourceIsAvailable sets the underlying resource's availability
func (s *Store) SetResourceIsAvailable(resource string, available bool, reason string) error {
	s.Lock()
	defer s.Unlock()

	r, ok := s.Resources[resource]

	if !ok {
		return errors.New("resource " + resource + " not found")
	}

	if available {
		r.Diary.SetAvailable(reason)
	} else {
		r.Diary.SetUnavailable(reason)
	}

	return nil

}

// SetSlotIsAvailable sets the underlying resource's availability
func (s *Store) SetSlotIsAvailable(slot string, available bool, reason string) error {
	s.Lock()
	defer s.Unlock()
	sl, ok := s.Slots[slot]

	if !ok {
		return errors.New("slot " + slot + " not found")
	}

	r, ok := s.Resources[sl.Resource]

	if !ok {
		return errors.New("resource " + sl.Resource + " not found")
	}

	if available {
		r.Diary.SetAvailable(reason)
	} else {
		r.Diary.SetUnavailable(reason)
	}

	return nil

}

// ValidateBooking checks if booking exists and details are valid (i.e. must confirm booking contents, not just ID)
// Don't take the lock - rely on calling function(s) to handle that
func (s *Store) validateBooking(booking Booking) error {

	b, ok := s.Bookings[booking.Name]

	if !ok {
		return errors.New("not found")
	}

	// compare the externally relevant fields of the booking (ignore internal boolean fields)
	t1 := Booking{
		Name:   b.Name,
		Policy: b.Policy,
		Slot:   b.Slot,
		User:   b.User,
		When:   b.When,
	}
	t2 := Booking{
		Name:   booking.Name,
		Policy: booking.Policy,
		Slot:   booking.Slot,
		User:   booking.User,
		When:   booking.When,
	}

	if t1 != t2 { //spam submission with non-matching details
		return errors.New("could not verify booking details")
	}

	if b.When.Start.After(s.now()) {
		return errors.New("too early")
	}

	if b.When.End.Before(s.now()) {
		return errors.New("too late")
	}

	if b.Cancelled {
		return errors.New("cancelled")
	}

	// check availability
	ok, reason, err := s.getSlotIsAvailable(b.Slot)

	if err != nil {
		return err
	}

	if !ok {
		return errors.New(reason)
	}

	return nil

}

// Operations not on the store

// availability returns a slice of available intervals between start and end, given a set of unavailable intervals
func availability(bi []interval.Interval, start, end time.Time) []interval.Interval {

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

		// A single large interval could overlap both start and end!

		//trim an interval if it overlaps start
		if i.Start.Before(start) {
			i.Start = start
		}

		//trim an interval if it overlaps end
		if i.End.After(end) {
			i.End = end
		}

		fa = append(fa, i)
	}

	return fa

}

// calculateUsage applies policy to booking to work out the usage that should be charged
func calculateUsage(b Booking, p Policy) (time.Duration, error) {

	// The booking attracts different usage tariffs depending on if/when started and cancelled:
	// cancelled before start, or unfulfilled - no usage
	// cancelled after start but before grace period - grace period
	// cancelled at/after end of grace period due to no-show (not started) - grace_period + grace_penalty
	// cancelled after grace period, but booking started - from booking start to time cancelled

	if !b.Cancelled {
		if b.Unfulfilled {
			return time.Duration(0), nil
		}
		return b.When.End.Sub(b.When.Start), nil
	}

	if b.CancelledAt.Before(b.When.Start) {
		return time.Duration(0), nil
	}

	if b.CancelledAt.After(b.When.End) {
		return b.When.End.Sub(b.When.Start), nil
	}

	if p.EnforceGracePeriod {

		if b.CancelledAt.Before(b.When.Start.Add(p.GracePeriod)) {
			return p.GracePeriod, nil
		}

		if !(b.Started) { //assume autocancellation worked, don't charge user for our mistake if it didn't
			// so don't check whether rest of data is consistent with autocancellation, just leave
			// that data for reporting purposes not calculation purposes
			return p.GracePeriod + p.GracePenalty, nil
		}

	}

	return b.CancelledAt.Sub(b.When.Start), nil

}

func checkDescriptions(items map[string]Description) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.Name == "" {
			msg = append(msg, "missing name field in description "+k)
		}
		if item.Type == "" {
			msg = append(msg, "missing type field in description "+k)
		}
		if item.Short == "" {
			msg = append(msg, "missing short field in description "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func checkDisplayGuides(items map[string]DisplayGuide) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.BookAhead == time.Duration(0) {
			msg = append(msg, "missing book_ahead field in display_guide "+k)
		}
		if item.Duration == time.Duration(0) {
			msg = append(msg, "missing duration field in display_guide "+k)
		}
		if item.MaxSlots == 0 {
			msg = append(msg, "missing max_slots field in display_guide "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func checkGroups(items map[string]Group) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.Description == "" {
			msg = append(msg, "missing description field in group "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

// CheckManifest checks for internal consistency, throwing an error
// if there are any unresolved references by name
func CheckManifest(m Manifest) (error, []string) {
	return checkManifest(m)
}

// checkManifest checks for internal consistency, throwing an error
// if there are any unresolved references by name
func checkManifest(m Manifest) (error, []string) {

	// check if any elements have duplicate or missing names

	err, msg := checkDescriptions(m.Descriptions)

	if err != nil {
		return err, msg
	}

	err, msg = checkDisplayGuides(m.DisplayGuides)

	if err != nil {
		return err, msg
	}

	err, msg = checkGroups(m.Groups)

	if err != nil {
		return err, msg
	}

	err, msg = checkPolicies(m.Policies)

	if err != nil {
		return err, msg
	}

	err, msg = checkResources(m.Resources)

	if err != nil {
		return err, msg
	}

	err, msg = checkStreams(m.Streams)

	if err != nil {
		return err, msg
	}

	err, msg = checkSlots(m.Slots)

	if err != nil {
		return err, msg
	}

	err, msg = checkUIs(m.UIs)

	if err != nil {
		return err, msg
	}

	err, msg = checkUISets(m.UISets)

	if err != nil {
		return err, msg
	}

	err, msg = checkWindows(m.Windows)

	if err != nil {
		return err, msg
	}

	// Check that all references are present

	// Description -> N/A

	// Group -> Description, Policies
	for k, v := range m.Groups {
		if _, ok := m.Descriptions[v.Description]; !ok {
			m := "group " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		for _, p := range v.Policies {
			if _, ok := m.Policies[p]; !ok {
				m := "group " + k + " references non-existent policy: " + p
				msg = append(msg, m)
			}
		}
	}

	// Policy -> Description, DisplayGuide, Slots
	for k, v := range m.Policies {
		if _, ok := m.Descriptions[v.Description]; !ok {
			m := "policy " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		for _, dg := range v.DisplayGuides {
			if _, ok := m.DisplayGuides[dg]; !ok {
				m := "policy " + k + " references non-existent display_guide: " + dg
				msg = append(msg, m)
			}
		}
		for _, s := range v.Slots {
			if _, ok := m.Slots[s]; !ok {
				m := "policy " + k + " references non-existent slot: " + s
				msg = append(msg, m)
			}
		}
	}

	// Resource ->  Description, Stream
	for k, v := range m.Resources {
		if _, ok := m.Descriptions[v.Description]; !ok {
			m := "resource " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		for _, s := range v.Streams {
			if _, ok := m.Streams[s]; !ok {
				m := "resource " + k + " references non-existent stream: " + s
				msg = append(msg, m)
			}
		}
	}

	// Slot -> Description, Policy, Resource, UISet, Window
	for k, v := range m.Slots {
		if _, ok := m.Descriptions[v.Description]; !ok {
			m := "slot " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		if _, ok := m.Policies[v.Policy]; !ok {
			m := "slot " + k + " references non-existent policy: " + v.Policy
			msg = append(msg, m)
		}
		if _, ok := m.Resources[v.Resource]; !ok {
			m := "slot " + k + " references non-existent resource: " + v.Resource
			msg = append(msg, m)
		}
		if _, ok := m.UISets[v.UISet]; !ok {
			m := "slot " + k + " references non-existent ui_set: " + v.UISet
			msg = append(msg, m)
		}
		if _, ok := m.Windows[v.Window]; !ok {
			m := "slot " + k + " references non-existent window: " + v.Window
			msg = append(msg, m)
		}
	}

	// Stream -> N/A

	// UI -> Description, StreamsRequired

	for k, v := range m.UIs {
		if _, ok := m.Descriptions[v.Description]; !ok {
			m := "ui " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		// this check still applies, even though it relates in part to the templating process
		for _, s := range v.StreamsRequired {
			if _, ok := m.Streams[s]; !ok {
				m := "ui " + k + " references non-existent stream: " + s
				msg = append(msg, m)
			}
		}
	}

	// UISet -> UIs
	for k, v := range m.UISets {
		for _, u := range v.UIs {
			if _, ok := m.UIs[u]; !ok {
				m := "ui_set " + k + " references non-existent ui: " + u
				msg = append(msg, m)
			}
		}
	}

	if len(msg) > 0 {
		return errors.New("missing reference(s)"), msg
	}

	return nil, []string{}

}

func checkPolicies(items map[string]Policy) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.Description == "" {
			msg = append(msg, "missing description field in policy "+k)
		}
		if item.Slots == nil {
			msg = append(msg, "missing slots field in policy "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func checkResources(items map[string]Resource) (error, []string) {

	msg := []string{}

	for k, item := range items {
		// ConfigURL is optional
		if item.Description == "" {
			msg = append(msg, "missing description field in resource "+k)
		}
		if item.Streams == nil {
			msg = append(msg, "missing streams field in resource "+k)
		}
		if item.TopicStub == "" {
			msg = append(msg, "missing topic_stub field in resource "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func checkSlots(items map[string]Slot) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.Description == "" {
			msg = append(msg, "missing description field in slot "+k)
		}
		if item.Policy == "" {
			msg = append(msg, "missing policy field in slot "+k)
		}
		if item.Resource == "" {
			msg = append(msg, "missing resource field in slot "+k)
		}
		if item.UISet == "" {
			msg = append(msg, "missing ui_set field in slot "+k)
		}
		if item.Window == "" {
			msg = append(msg, "missing window field in slot "+k)
		}

	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func checkStreams(items map[string]Stream) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.ConnectionType == "" {
			msg = append(msg, "missing connection_type field in stream "+k)
		}
		if item.For == "" {
			msg = append(msg, "missing for field in stream "+k)
		}
		if item.Scopes == nil {
			msg = append(msg, "missing scopes field in stream "+k)
		}
		if item.Topic == "" {
			msg = append(msg, "missing topic field in stream "+k)
		}
		if item.URL == "" {
			msg = append(msg, "missing url field in stream "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func checkUIs(items map[string]UI) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.URL == "" {
			msg = append(msg, "missing url field in ui "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func checkUISets(items map[string]UISet) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.UIs == nil {
			msg = append(msg, "missing uis field in ui_set "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func checkWindows(items map[string]Window) (error, []string) {

	msg := []string{}

	for k, item := range items {
		// a window has to have at least one allowed period to be valid
		// a slot should be deleted rather than have a window with no allowed periods
		if item.Allowed == nil {
			msg = append(msg, "missing allowed field in window "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	// we can get errors making filters, so check that

	for k, w := range items {

		f := filter.New()
		err := f.SetAllowed(w.Allowed)
		if err != nil {
			msg = append(msg, "failed to create allowed intervals for window "+k+": "+err.Error())
		}
		err = f.SetDenied(w.Denied)
		if err != nil {
			msg = append(msg, "failed to create denied intervals for window "+k+": "+err.Error())
		}

	}

	if len(msg) > 0 {
		return errors.New("failed creating filter"), msg
	}

	return nil, []string{}

}

// Unmarshallers for structs with durations
// so that we can handle JSON in our store format during testing
// which makes it easier to read diffs due to the lack
// of pointers unlike the swagger models
// method from https://penkovski.com/post/go-unmarshal-custom-types/
func (p *Policy) UnmarshalJSON(data []byte) (err error) {

	var tmp struct {
		// durations are set to string for now
		AllowStartInPastWithin string `json:"allow_start_in_past_within"  yaml:"allow_start_in_past_within"`
		BookAhead              string `json:"book_ahead"  yaml:"book_ahead"`
		MaxDuration            string `json:"max_duration"  yaml:"max_duration"`
		MinDuration            string `json:"min_duration"  yaml:"min_duration"`
		MaxUsage               string `json:"max_usage"  yaml:"max_usage"`
		NextAvailable          string `json:"next_available"  yaml:"next_available"`
		StartsWithin           string `json:"starts_within"  yaml:"starts_within"`

		// other fields stay the same
		Description             string   `json:"description"  yaml:"description"`
		DisplayGuides           []string `json:"display_guides"  yaml:"display_guides"`
		EnforceAllowStartInPast bool     `json:"enforce_allow_start_in_past"  yaml:"enforce_allow_start_in_past"`
		EnforceBookAhead        bool     `json:"enforce_book_ahead"  yaml:"enforce_book_ahead"`
		EnforceMaxBookings      bool     `json:"enforce_max_bookings"  yaml:"enforce_max_bookings"`
		EnforceMaxDuration      bool     `json:"enforce_max_duration"  yaml:"enforce_max_duration"`
		EnforceMinDuration      bool     `json:"enforce_min_duration"  yaml:"enforce_min_duration"`
		EnforceMaxUsage         bool     `json:"enforce_max_usage"  yaml:"enforce_max_usage"`
		EnforceNextAvailable    bool     `json:"enforce_next_available"  yaml:"enforce_next_available"`
		EnforceStartsWithin     bool     `json:"enforce_starts_within"  yaml:"enforce_starts_within"`
		EnforceUnlimitedUsers   bool     `json:"enforce_unlimited_users"  yaml:"enforce_unlimited_users"`
		MaxBookings             int64    `json:"max_bookings"  yaml:"max_bookings"`
		Slots                   []string `json:"slots" yaml:"slots"`
	}

	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	// insert default durations if required
	if tmp.BookAhead == "" {
		tmp.BookAhead = "0s"
	}
	if tmp.MaxDuration == "" {
		tmp.MaxDuration = "0s"
	}
	if tmp.MinDuration == "" {
		tmp.MinDuration = "0s"
	}
	if tmp.NextAvailable == "" {
		tmp.NextAvailable = "0s"
	}
	if tmp.AllowStartInPastWithin == "" {
		tmp.AllowStartInPastWithin = "0s"
	}
	if tmp.StartsWithin == "" {
		tmp.StartsWithin = "0s"
	}
	if tmp.MaxUsage == "" {
		tmp.MaxUsage = "0s"
	}

	// parse durations
	ba, err := time.ParseDuration(tmp.BookAhead)
	if err != nil {
		return err
	}
	xd, err := time.ParseDuration(tmp.MaxDuration)
	if err != nil {
		return err
	}
	na, err := time.ParseDuration(tmp.NextAvailable)
	if err != nil {
		return err
	}
	nd, err := time.ParseDuration(tmp.MinDuration)
	if err != nil {
		return err
	}
	sp, err := time.ParseDuration(tmp.AllowStartInPastWithin)
	if err != nil {
		return err
	}
	sw, err := time.ParseDuration(tmp.StartsWithin)
	if err != nil {
		return err
	}
	xu, err := time.ParseDuration(tmp.MaxUsage)
	if err != nil {
		return err
	}

	p.AllowStartInPastWithin = sp
	p.BookAhead = ba
	p.MaxDuration = xd
	p.NextAvailable = na
	p.MinDuration = nd
	p.MaxUsage = xu
	p.StartsWithin = sw

	p.Description = tmp.Description
	p.DisplayGuides = tmp.DisplayGuides
	p.EnforceAllowStartInPast = tmp.EnforceAllowStartInPast
	p.EnforceBookAhead = tmp.EnforceBookAhead
	p.EnforceMaxBookings = tmp.EnforceMaxBookings
	p.EnforceMaxDuration = tmp.EnforceMaxDuration
	p.EnforceMinDuration = tmp.EnforceMinDuration
	p.EnforceMaxUsage = tmp.EnforceMaxUsage
	p.EnforceNextAvailable = tmp.EnforceNextAvailable
	p.EnforceStartsWithin = tmp.EnforceStartsWithin
	p.EnforceUnlimitedUsers = tmp.EnforceUnlimitedUsers
	p.MaxBookings = tmp.MaxBookings
	p.Slots = tmp.Slots

	return nil

}

func (d *DisplayGuide) UnmarshalJSON(data []byte) (err error) {

	var tmp struct {
		// durations are set to string for now
		Duration  string `json:"duration" yaml:"duration"`
		BookAhead string `json:"book_ahead" yaml:"book_ahead"`

		//others stay the same
		MaxSlots int    `json:"max_slots" yaml:"max_slots"`
		Label    string `json:"label" yaml:"label"`
	}

	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	// set default durations
	if tmp.BookAhead == "" {
		tmp.BookAhead = "0s"
	}
	if tmp.Duration == "" {
		tmp.Duration = "0s"
	}

	// parse durations

	ba, err := time.ParseDuration(tmp.BookAhead)
	if err != nil {
		return err
	}
	dd, err := time.ParseDuration(tmp.Duration)
	if err != nil {
		return err
	}

	d.BookAhead = ba
	d.Duration = dd
	d.MaxSlots = tmp.MaxSlots
	d.Label = tmp.Label

	return nil

}
