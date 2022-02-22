package history

import "time"

func (h *history) Add(action store.Action) {

}

// how many replay pointers do we need?
func (h *history) Replay() {
	// reset replay pointer to start of history
}

// NewReplayAll returns a pointer to a Replay object
// which holds a pointer to the history, to
// save copying it. Messages will be replayed
// until there are no more messages. Messages
// added to the history after creating the object
// but before the message preceding them has been replayed
// will be replayed too.
func (h *history) NewReplayAll() *Replay {
	// return a replay object,
	// pointer to history to save copying it
	// replay pointer set to beginning of stack
}

// NewReplayAllUntilNow returns a pointer to a Replay object
// which holds a pointer to the history, to
// save copying it.
// All messages received up until the current time
// will be replayed
func (h *history) NewReplayAllUntilNow() *Replay {
	// return a replay object,
	// pointer to history to save copying it
	// replay pointer set to beginning of stack
}

// NewReplayInterval returns a pointer to a Replay object
// which holds a pointer to the history, to
// save copying it.  All messages received within the
// interval from -> to will be replayed.
// Intervals are exclusive, due to the use of time.Before, and
// time.After.
func (h *history) NewReplayInterval(from, to time.Time) *Replay {
	// return a replay object,
	// pointer to history to save copying it
	// replay pointer set to beginning of stack
}

//
func (r *Replay) GetNext() (interval.Action, error) {

}
