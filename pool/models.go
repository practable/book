package pool

import (
	"sync"
	"time"

	"github.com/timdrysdale/interval/filter"
	"github.com/timdrysdale/interval/permission"
	"github.com/timdrysdale/interval/resource"
)

// Note that an activity combines a UISet and Equipment - here we are only concerned
// with the Equipment, because that is what is real, and limited in supply.
// The choice of UISet is handled over on the booking side, to allow different UISet with
// the same Equipment depending on who is booking.

// Store represents Equipments stored according to Pool
type Store struct {
	*sync.RWMutex `json:"-" yaml:"-"`

	// Pools maps all equipment in the store
	Equipments map[string]*Equipment `json:"equipments"`

	// Pools maps all pools in the store
	Pools map[string]*Pool `json:"pools"`

	// Now is a function for getting the time - useful for mocking in test
	Now func() time.Time `json:"-" yaml:"-"`
}

// Pool represents the booking status of the activities in a pool
// Note that each pool can have a different minSession / MaxSession duration
// but that users are limited to fixed maximum number of sessions they can book
// across the system to prevent users with access with more pools booking even more
// experiments simultaneously.
type Pool struct {
	*sync.RWMutex `json:"-" yaml:"-"`
	Description   shared.Description `json:"description"`
	Equipments    map[string]bool    `json:"equipments"`
	Now           func() time.Time   `json:"-" yaml:"-"`
}

// Activity represents an individual activity that can be booked
type Equipment struct {
	*sync.RWMutex `json:"-"`
	Config        Config `json:"config"`
	Description   `json:"description"`
	Available     filter.Filter `json:"available"`
	Bookings      resource.Resource
	Streams       map[string]*Stream `json:"streams"`
}

// Config represents a hardware configuration file URL
// that may be useful to a UI
type Config struct {
	URL string `json:"url"`
}

// Stream represents a data or video stream from a relay
// typically accessed via POST with bearer token
type Stream struct {
	*sync.RWMutex `json:"-"`

	// For is the key in the UI's URL in which the client puts
	// the relay (wss) address and code after getting them
	// from the relay
	For string `json:"for,omitempty"`

	// URL of the relay access point for this stream
	URL string `json:"url"`

	// signed bearer token for accessing the stream
	// submit token in the header
	Token string `json:"token,omitempty"`

	// Verb is the HTTP method, typically post
	Verb string `json:"verb,omitempty"`

	// Permission is a prototype for the permission token that the booking system
	// generates and puts into the Token field
	Permission permission.Token `json:"permission,omitempty"`
}
