// package interval implements an interval AVL tree with arbitrary payload for each interval

package interval

import (
	"sort"
	"time"
)

// https://stackoverflow.com/questions/25065055/what-is-the-maximum-time-time-in-go
var ZeroTime = time.Unix(0, 0)
var Infinity = time.Unix(1<<63-62135596801, 999999999)

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

func Invert(intervals []Interval) []Interval {

	if len(intervals) < 1 {
		return []Interval{}
	}

	Sort(&intervals)

	inverted := []Interval{}

	// set the start of the first deny interval to zero time
	inverted = append(inverted, Interval{Start: ZeroTime})

	for _, a := range intervals {

		d := &inverted[len(inverted)-1]

		if d.Start.Before(a.Start) { // normal case

			// set the end of the last deny interval
			d.End = a.Start

			// set the start time of the next deny interval
			inverted = append(inverted, Interval{Start: a.End})

		} else { // overlapping allow periods

			// delay start time of deny interval until end of allow interval
			d.Start = a.End
		}

	}

	// set the end of the last deny interval to infinity
	p := &inverted[len(inverted)-1]
	p.End = Infinity

	return inverted

}
