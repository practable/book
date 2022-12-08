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
	Start time.Time `json:"start" yaml:"start"`
	End   time.Time `json:"end" yaml:"end"`
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

// Len finds the length of the slice - needed for sort
func (s byInterval) Len() int {
	return len(s)
}

// Swap swaps two elements of a slice - needed for sort
func (s byInterval) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less finds the earliest start time, or if starts are the same, the earliest end time
// needed for sort
func (s byInterval) Less(i, j int) bool {
	if s[i].Start.Equal(s[j].Start) {
		return s[i].End.Before(s[j].End)
	}
	return s[i].Start.Before(s[j].Start)
}

// Sort orders intervals by start time (and then by end time to split a tie on start time)
func Sort(intervals *[]Interval) {
	sort.Sort(byInterval(*intervals))
}

// Merge combines any overlapping intervals
func Merge(intervals []Interval) []Interval {

	if len(intervals) < 1 {
		return []Interval{}
	}

	Sort(&intervals)

	merged := []Interval{intervals[0]}

	//merge any overlaps
	for _, next := range intervals[1:] {

		last := &merged[len(merged)-1]

		if next.Start.After(last.End) { // normal case
			merged = append(merged, next)

		} else { // overlapping, so extend last interval
			last.End = next.End
		}
	}
	return merged
}

// Invert sorts, merges and then inverts the given intervals
func Invert(intervals []Interval) []Interval {

	if len(intervals) < 1 {
		return []Interval{}
	}

	merged := Merge(intervals)

	// set the start of the first deny interval to zero time
	inverted := []Interval{Interval{Start: ZeroTime}}

	for _, next := range merged {

		last := &inverted[len(inverted)-1]

		// set the end of the last deny interval, minus a nanosecond to prevent overlap
		last.End = next.Start.Add(-time.Nanosecond)

		// set the start time of the next deny interval, plus a nanosecond to prevent overlap
		inverted = append(inverted, Interval{Start: next.End.Add(time.Nanosecond)})

	}

	// set the end of the last deny interval to infinity
	p := &inverted[len(inverted)-1]
	p.End = Infinity

	return inverted

}
