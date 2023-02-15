// package filter represents intervals that are allowed, and denied
// so an interval can be checked for
// (a) falling completely within an allowed interval AND
// (b) also not even partially overlapping a denied interval
// Note that our avl trees best supports finding clashing intervals
// so turn our allow list into something we can search for clashes against
// i.e. invert it to make denied periods between allowed periods
// and take note of the left most and right most periods allowed to complete the
// check without using tree functions (to avoid setting  arbitrary zero and infinite time)
// additional check
// but not finding intervals that fall completely within other intervals
//
package filter

import (
	"sync"

	avl "github.com/practable/book/internal/trees/avltree"

	"github.com/google/uuid"
	"github.com/practable/book/internal/interval"
)

// Filter represents an allowed interval, with a list of denied sub-intervals
type Filter struct {
	*sync.RWMutex  `json:"-"`
	notAllowed     *avl.Tree // a deny list calculated by inverting the allow list
	denied         *avl.Tree // the deny list
	allowedList    []interval.Interval
	notAllowedList []interval.Interval
	deniedList     []interval.Interval
	combinedList   []interval.Interval
}

// New creates a new filter with an empty deny list and no allowed interval
func New() *Filter {
	return &Filter{
		&sync.RWMutex{},
		avl.NewWith(interval.Comparator),
		avl.NewWith(interval.Comparator),
		[]interval.Interval{},
		[]interval.Interval{
			interval.Interval{
				Start: interval.ZeroTime,
				End:   interval.DistantFuture,
			}},
		[]interval.Interval{},
		[]interval.Interval{
			interval.Interval{
				Start: interval.ZeroTime,
				End:   interval.DistantFuture,
			},
		},
	}
}

// SetAllowed adds the allowed intervals to the `allowed list`
func (f *Filter) SetAllowed(allowed []interval.Interval) error {
	f.Lock()
	defer f.Unlock()

	// add to any existing allowed intervals we have
	f.allowedList = interval.Merge(append(f.allowedList, allowed...))

	// invert the intervals to become notAllowed intervals
	f.notAllowedList = interval.Invert(f.allowedList)

	// update the combined list
	f.combinedList = interval.Merge(append(f.notAllowedList, f.deniedList...))

	// make a new AVL tree (rather than try to modify the existing one)
	// it's easier to start again because we are storing inverted values in this tree
	// so calling this function more than once would require removing intervals from the notAllowed tree
	// easier just to track all the allowed regions we've had so far and recalculate the AVL tree
	f.notAllowed = avl.NewWith(interval.Comparator)

	// add to the AVL trees
	for _, na := range f.notAllowedList {

		u := uuid.New()

		_, err := f.notAllowed.Put(na, u)

		if err != nil {
			return err
		}

	}

	return nil
}

// SetDenied adds intervals to the `denied list`
func (f *Filter) SetDenied(denied []interval.Interval) error {
	f.Lock()
	defer f.Unlock()

	denied = interval.Merge(denied)

	// update the lists
	f.deniedList = interval.Merge(append(f.deniedList, denied...))

	f.combinedList = interval.Merge(append(f.notAllowedList, f.deniedList...))

	for _, d := range denied {

		u := uuid.New()

		_, err := f.denied.Put(d, u)

		if err != nil {
			return err
		}

	}

	return nil

}

// Allowed returns true if the interval is allowed
// It must  not conflict with notAllowed
// It must not conflict with denied
// This does not write to the filter so only needs
// a read lock
func (f *Filter) Allowed(when interval.Interval) bool {
	f.RLock()
	defer f.RUnlock()

	u := uuid.New()

	_, err := f.notAllowed.CouldPut(when, u)

	if err != nil {
		return false
	}

	_, err = f.denied.CouldPut(when, u)

	if err != nil {
		return false
	}

	return true

}

// Export exports the filter as a list of denied intervals, accounting for the combined effects of allowed and denied intervals
func (f *Filter) Export() []interval.Interval {

	return f.combinedList
}
