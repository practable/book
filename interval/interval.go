// package interval implements an interval AVL tree with arbitrary payload for each interval

package interval

import (
	"sort"
	"time"
)

type Interval struct {
	Start time.Time
	End   time.Time
}

// Comparator function (sort by IDs)
func Comparator(a, b interface{}) int {

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

// https://gobyexample.com/sorting-by-functions

type byInterval []Interval

func (s byInterval) Len() int {
	return len(s)
}

func (s byInterval) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byInterval) Less(i, j int) bool {
	return s[i].Start.Before(s[j].Start)
}

func Sort(intervals *[]Interval) {

	sort.Sort(byInterval(*intervals))

}
