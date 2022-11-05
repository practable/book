package store

import (
	"time"

	"github.com/google/uuid"
)

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
