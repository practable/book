// package interval implements an interval AVL tree with arbitrary payload for each interval

package interval

import "time"

type Interval struct {
	Start time.Time
	End   time.Time
}

// Comparator function (sort by IDs)
func byInterval(a, b interface{}) int {

	// Type assertion, program will panic if this is not respected
	t1 := a.(Interval)
	t2 := b.(Interval)

	switch {
	case t1.Start.After(t2.End):
		return 1
	case t1.End.Before(t2.Start):
		return -1
	default:
		return 0 //this implies some degree of overlap
	}
}
