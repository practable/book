package store

/*

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

*/
