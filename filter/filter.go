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
	avl "interval/internal/trees/avltree"

	"github.com/google/uuid"
	"github.com/timdrysdale/interval/interval"
)

// Filter represents an allowed interval, with a list of denied sub-intervals
type Filter struct {
	notAllowed *avl.tree // a deny list calculated by inverting the allow list
	denied     *avl.Tree // the deny list
}

// New creates a new filter with an empty deny list and no allowed interval
func New() *Filter {
	return &Resource{
		notAllowed: avl.NewWith(interval.Comparator),
		denied:     avl.NewWith(interval.Comparator),
	}
}

// SetAllowed adds the allowed intervals to the `allowed list`
func (f *Filter) SetAllowed(allowed []interval.Interval) error {

	// invert the intervals to become notAllowed intervals
	notAllowed := interval.Invert(allowed)

	for _, na := range notAllowed {

		u := uuid.New()

		_, err := r.notAllowed.Put(when, u)

		if err != nil {
			return err
		}

	}

	return nil
}

// SetDenied adds an interval to the `denied list`
func (r *Resource) SetDenied(denied []interval.Interval) error {

	for _, d := range denied {

		u := uuid.New()

		_, err := r.denied.Put(when, u)

		if err != nil {
			return err
		}

	}

	return nil

}

// Allowed returns true if the interval is allowed, which means
// it falls completely within an allowed interval
// does not intersect even partially with a denied interval
func (r *Resource) Allowed(when interval.Interval) bool {

	u := uuid.New()

	_, err := r.notAllowed.Put(when, u)

	if err != nil {
		//return a zero-value UUID if there is an error
		return [16]byte{}, err
	}

	return u, err

}
