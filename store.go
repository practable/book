// package store holds bookings with arbitrary durations
package interval

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/timdrysdale/interval/diary"
	"github.com/timdrysdale/interval/filter"
	"github.com/timdrysdale/interval/interval"
)

// Activity represents connection details for a live booking
type Activity struct {
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
	Cancelled   bool
	Name        string // booking uid
	Policy      string // reference to policy it was booked under
	Slot        string // slot name
	Started     bool
	Unfulfilled bool   //when the resource was unavailable
	User        string // user pseudo id
	When        interval.Interval
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

// Policy represents what a user can book, and any limits on bookings/usage
// Unmarshaling of time.Duration works in yaml.v3, https://play.golang.org/p/-6y0zq96gVz"
type Policy struct {
	BookAhead          time.Duration   `json:"book_ahead"  yaml:"book_ahead"`
	Description        string          `json:"description"  yaml:"description"`
	EnforceBookAhead   bool            `json:"enforce_book_ahead"  yaml:"enforce_book_ahead"`
	EnforceMaxBookings bool            `json:"enforce_max_bookings"  yaml:"enforce_max_bookings"`
	EnforceMaxDuration bool            `json:"enforce_max_duration"  yaml:"enforce_max_duration"`
	EnforceMinDuration bool            `json:"enforce_min_duration"  yaml:"enforce_min_duration"`
	EnforceMaxUsage    bool            `json:"enforce_max_usage"  yaml:"enforce_max_usage"`
	MaxBookings        int64           `json:"max_bookings"  yaml:"max_bookings"`
	MaxDuration        time.Duration   `json:"max_duration"  yaml:"max_duration"`
	MinDuration        time.Duration   `json:"min_duration"  yaml:"min_duration"`
	MaxUsage           time.Duration   `json:"max_usage"  yaml:"max_usage"`
	Slots              []string        `json:"slots" yaml:"slots"`
	SlotMap            map[string]bool `json:"-" yaml:"-"` // internal usage, do not populate from file
}

type PolicyStatus struct {
	OldBookings     int64         `json:"old_bookings"  yaml:"old_bookings"`
	CurrentBookings int64         `json:"current_bookings"  yaml:"current_bookings"`
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

	// Locked is true when we want to stop making bookings or getting info while we do uploads/maintenance
	// The API handler has to check this, e.g. if locked, do not make bookings or check availability on
	// behalf of users. We can't do this automatically in the methods because then we'd need some sort
	// of admin override, to permit maintenance when locked (which is the whole point of locking the system)
	Locked bool

	// Message represents our message of the day, to send to users (e.g. to explain system is locked)
	Message string

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
	UIs map[string]UIDescribed

	// UISets represents the lists of user interfaces for particular slots
	UISets map[string]UISet

	// UsagePolicies represents all the UsagePolicy(ies) in use
	Users map[string]*User

	// Window represents allowed and denied time periods for slots
	Windows map[string]Window
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
	// URL with moustache {{key}} templating for stream connections
	URL             string   `json:"url"  yaml:"url"`
	StreamsRequired []string `json:"streams_required"  yaml:"streams_required"`
}

// UISet represents UIs that can be used with a slot
type UISet struct {
	UIs []string
}

// User represents bookings and usage information associated with a single user
// remembering policies allows us to direct a user to link to a policy for a course just once, and then have that remembered
// at least until a system restart -> should be logged as a transaction
type User struct {
	Bookings    map[string]*Booking       //map by id for retrieval
	OldBookings map[string]*Booking       //map by id, for admin dashboards
	Policies    map[string]bool           //map of policies that apply to the user
	Usage       map[string]*time.Duration //map by policy for checking usage
}

type UserExternal struct {
	Bookings    []string
	OldBookings []string
	Policies    []string
	Usage       map[string]string //map humanised durations by policy name
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
		false,
		"Welcome to the interval booking store",
		func() time.Time { return time.Now() },
		make(map[string]*Booking),
		make(map[string]Policy),
		make(map[string]Resource),
		make(map[string]Slot),
		make(map[string]Stream),
		make(map[string]UIDescribed),
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
		if s.Now().After(v.When.End) {
			stale[k] = v
		}
	}

	for k, v := range stale {
		s.OldBookings[k] = v
		delete(s.Bookings, k)
	}

}

// PruneUserBookings is a maintenace operation to move
// expired bookings from the map of bookings but only
// to do so for a given user (e.g. ahead of checking
// their policy limits on future bookings).
func (s *Store) PruneUserBookings(user string) {

	u, ok := s.Users[user]

	if !ok {
		return //do nothing
	}

	stale := make(map[string]*Booking)

	for k, v := range u.Bookings {
		if s.Now().After(v.When.End) {
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

func HumaniseDuration(t time.Duration) string {
	return t.Round(time.Second).String()
}

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

// GetAvailability returns a list of intervals for which a given slot is available under a given policy, or an error if the slot or policy is not found. The policy contains aspects such as look-ahead which may limit the window of availability.
func (s *Store) GetAvailability(policy, slot string) ([]interval.Interval, error) {

	p, ok := s.Policies[policy]

	if !ok {
		return []interval.Interval{}, errors.New("policy " + policy + " not found")
	}

	_, ok = p.SlotMap[slot]

	if !ok {
		return []interval.Interval{}, errors.New("slot " + slot + " not in policy " + policy)
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

	if len(bk) == 0 { // no bookings
		a := []interval.Interval{
			interval.Interval{
				Start: start,
				End:   end,
			},
		}

		return a, nil
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

// SetSlotIsAvailable sets the underlying resource's availability
func (s *Store) SetSlotIsAvailable(slot string, available bool, reason string) error {

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

func NewUser() *User {
	return &User{
		make(map[string]*Booking),
		make(map[string]*Booking),
		make(map[string]bool),
		make(map[string]*time.Duration),
	}
}

// CancelBooking cancels a booking or returns an error if not found
func (s *Store) CancelBooking(booking Booking) error {

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

	if b.When.End.Before(s.Now()) {
		return errors.New("cannot cancel booking that has already ended")
	}

	if b.Started { //TODO - allow cancelling started bookings by deny-listing the stream tokens
		return errors.New("cannot cancel booking that has already been used")
	}

	delete(s.Bookings, b.Name)

	b.Cancelled = true

	s.OldBookings[b.Name] = b

	// adjust usage for user

	refund := b.When.End.Sub(b.When.Start)

	if b.When.Start.Before(s.Now()) {
		refund = b.When.End.Sub(s.Now()) //only refund portion after cancellation
	}

	u, ok := s.Users[b.User]

	if !ok { //might happen if server is restarted, old booking restored but user has not made any new bookings yet
		// could be a prompt to create users for restored bookings ....
		return errors.New("cancelled but could not refund usage to unknown user " + b.User)
	}

	*u.Usage[b.Policy] = *u.Usage[b.Policy] - refund //refund reduces usage

	s.Users[b.User] = u

	return nil

}

// MakeBooking makes bookings for users, according to the policy
// If a user does not exist, one is created.
// APIs for users should call this version
func (s *Store) MakeBooking(policy, slot, user string, when interval.Interval) (Booking, error) {

	name := uuid.New().String()
	return s.MakeBookingWithName(policy, slot, user, when, name)

}

// MakeBookingWithID makes bookings for users, according to the policy
// If a user does not exist, one is created.
// The booking ID is set by the caller, so that bookings can be edited/replaced
// This version should only be called by Admin users
func (s *Store) MakeBookingWithName(policy, slot, user string, when interval.Interval, name string) (Booking, error) {

	p, ok := s.Policies[policy]

	if !ok {
		return Booking{}, errors.New("policy " + policy + " not found")
	}

	_, ok = p.SlotMap[slot]

	if !ok {
		return Booking{}, errors.New("slot " + slot + " not in policy " + policy)
	}

	sl, ok := s.Slots[slot]

	if !ok {
		return Booking{}, errors.New("slot " + slot + " not found")
	}

	r, ok := s.Resources[sl.Resource]

	if !ok {
		return Booking{}, errors.New("resource " + sl.Resource + " not found")
	}

	u, ok := s.Users[user]

	if !ok { //not found, create new user
		u = NewUser()
		s.Users[user] = u
	}

	// (re-)add policy to user's list
	u.Policies[policy] = true

	// check if too many bookings already
	if p.EnforceMaxBookings {

		// remove stale entries from user's list of current bookings
		s.PruneUserBookings(user)

		// first check how many bookings under this policy already
		cb := []string{}

		for k, v := range u.Bookings {
			if v.Policy == policy {
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
				policy)
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
		if when.End.After(s.Now().Add(p.BookAhead)) {
			return Booking{}, errors.New("bookings cannot be made more than " +
				HumaniseDuration(p.BookAhead) +
				" ahead of the current time")
		}
	}

	// check for existing usage tracker for this policy?
	_, ok = u.Usage[policy]

	if !ok { //create usage tracker (always track usage, even if not limited)
		ut, err := time.ParseDuration("0s")
		if err != nil {
			return Booking{}, errors.New("could not initialise user tracker for user " +
				user +
				" because " +
				err.Error())
		}
		u.Usage[policy] = &ut
	}

	duration := when.End.Sub(when.Start)

	currentUsage := *u.Usage[policy]
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

	// see if the booking can be made ....

	err := r.Diary.Request(when, name)

	if err != nil {
		return Booking{}, err
	}

	// successful, so update usage tracker with value we calculated earlier
	u.Usage[policy] = &newUsage

	booking := Booking{
		Cancelled:   false,
		Name:        name,
		Policy:      policy,
		Slot:        slot,
		Started:     false,
		Unfulfilled: false,
		User:        user,
		When:        when,
	}

	s.Bookings[name] = &booking
	s.Users[user].Bookings[name] = &booking

	return booking, nil

}

// ValidateBooking checks if booking exists and details are valid (i.e. must confirm booking contents, not just ID)
func (s *Store) ValidateBooking(booking Booking) error {

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

	if b.When.Start.After(s.Now()) {
		return errors.New("too early")
	}

	if b.When.End.Before(s.Now()) {
		return errors.New("too late")
	}

	if b.Cancelled {
		return errors.New("cancelled")
	}

	// check availability
	ok, reason, err := s.GetSlotIsAvailable(b.Slot)

	if err != nil {
		return err
	}

	if !ok {
		return errors.New(reason)
	}

	return nil

}

// GetActivity returns an activity associated with a booking, or an error
// if the booking is invalid in some way
func (s *Store) GetActivity(booking Booking) (Activity, error) {

	err := s.ValidateBooking(booking)

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

		a.UIs = append(a.UIs, ui)
	}

	return a, nil
}

func (s *Store) GetBookingsFor(user string) ([]Booking, error) {

	if _, ok := s.Users[user]; !ok {
		return []Booking{}, errors.New("user not found")
	}

	b := []Booking{}

	for _, v := range s.Bookings {
		if user == v.User {
			b = append(b, *v)
		}
	}

	return b, nil

}

func (s *Store) GetOldBookingsFor(user string) ([]Booking, error) {

	if _, ok := s.Users[user]; !ok {
		return []Booking{}, errors.New("user not found")
	}

	b := []Booking{}

	for _, v := range s.OldBookings {
		if user == v.User {
			b = append(b, *v)
		}
	}

	return b, nil
}

func (s *Store) GetPolicyStatusFor(user, policy string) (PolicyStatus, error) {

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

	b, err := s.GetBookingsFor(user)
	if err != nil {
		return PolicyStatus{}, err
	}

	for _, v := range b {
		if policy == v.Policy {
			bp = append(bp, v)
		}
	}

	ob, err := s.GetOldBookingsFor(user)
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

// CheckBooking returns nil error if booking is ok, or an error and a slice of messages describing issues
func (s *Store) CheckBooking(b Booking) (error, []string) {

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

// ExportBookings returns a map of all current/future bookings
func (s *Store) ExportBookings() map[string]Booking {

	s.Lock()
	defer s.Unlock()

	bm := make(map[string]Booking)

	for k, v := range s.Bookings {
		bm[k] = *v
	}

	return bm
}

// ReplaceBookings will replace all bookings with a new set
// each booking must be valid for the manifest, i.e. all
// references to other entities must be valid.
// Note that the manifest should be set first
// Diaries need to be cleared by cancelling bookings to refund
// usage to  users before making the replacement bookings through
// the standard method
func (s *Store) ReplaceBookings(bm map[string]Booking) (error, []string) {
	s.Lock()
	defer s.Unlock()

	// Check bookings are individually sane given our manifest
	msg := []string{}

	for _, v := range bm {
		err, ms := s.CheckBooking(v)
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
	// we want to refund our users, so go through each booking and cancel

	for k, v := range s.Bookings {
		err := s.CancelBooking(*v)
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
		_, err := s.MakeBookingWithName(v.Policy, v.Slot, v.User, v.When, v.Name)

		if err != nil {
			msg = append(msg, "booking "+v.Name+" failed because "+err.Error())
		}

		// s.Bookings is updated by MakeBookingWithID so we mustn't update it ourselves
	}

	return nil, []string{}
}

// ExportOldBookings returns a map by name of old bookings
func (s *Store) ExportOldBookings() map[string]Booking {
	s.Lock()
	defer s.Unlock()

	bm := make(map[string]Booking)

	for k, v := range s.OldBookings {
		bm[k] = *v
	}

	return bm
}

// ReplaceOldBookings will replace the map of old bookings with the supplied list or return an error if the bookings have issues. All existing users are deleted, and replaced with users with usages that match the old bookings
func (s *Store) ReplaceOldBookings(bm map[string]Booking) (error, []string) {
	s.Lock()
	defer s.Unlock()

	// Check bookings are individually sane given our manifest
	msg := []string{}

	for _, v := range bm {
		err, ms := s.CheckBooking(v)
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

		// update user policies to include policy of this booking
		u.Policies[ob.Policy] = true

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

// ExportUsers returns a map of users, listing the names of bookings, old bookings, policies and
// their usage to date by policy name
func (s *Store) ExportUsers() map[string]UserExternal {

	s.Lock()
	defer s.Unlock()

	um := make(map[string]UserExternal)

	for k, v := range s.Users {

		bs := []string{}
		obs := []string{}
		ps := []string{}
		ds := make(map[string]string)

		for k := range v.Bookings {
			bs = append(bs, k)
		}

		for k := range v.OldBookings {
			obs = append(obs, k)
		}
		for k := range v.Policies {
			ps = append(ps, k)
		}
		for k, v := range v.Usage {
			ds[k] = HumaniseDuration(*v)
		}

		um[k] = UserExternal{
			Bookings:    bs,
			OldBookings: obs,
			Policies:    ps,
			Usage:       ds,
		}
	}

	return um
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

// ReplaceUsersPolicies allows administrators to add and remove policies from
// users, e.g. to add or restrict access to experiments
// A user that does not exist, is created, and the policies added.
// Policies must exist or an error is thrown
func (s *Store) ReplaceUserPolicies(u map[string][]string) (error, []string) {

	s.Lock()
	defer s.Unlock()

	msg := []string{}

	for k, v := range u {
		// check all policies exist
		for _, p := range v {
			if _, ok := s.Policies[p]; !ok {
				msg = append(msg, "user "+k+" policy "+p+" does not exist")
			}
		}
	}

	if len(msg) > 0 {
		return errors.New("policy not found"), msg
	}

	for k, v := range u {
		u := s.Users[k]
		pm := make(map[string]bool)
		for _, p := range v {
			pm[p] = true
		}
		u.Policies = pm

	}

	return nil, []string{}
}
