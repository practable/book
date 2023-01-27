package interval

import (
	"testing"
	"time"

	avl "github.com/practable/book/internal/trees/avltree"

	"github.com/stretchr/testify/assert"
)

func TestComparator(t *testing.T) {

	now := time.Now()

	a := Interval{Start: now, End: now.Add(2 * time.Second)}

	b := Interval{Start: now.Add(3 * time.Second), End: now.Add(4 * time.Second)}

	assert.Equal(t, -1, Comparator(a, b))

	assert.Equal(t, 1, Comparator(b, a))

	assert.Equal(t, 0, Comparator(a, a))

	// overlap partially with a
	c := Interval{Start: now.Add(time.Second), End: now.Add(3 * time.Second)}

	assert.Equal(t, 0, Comparator(a, c))

}

func TestAVL(t *testing.T) {

	at := avl.NewWith(Comparator)

	now := time.Now()
	a := Interval{Start: now, End: now.Add(2 * time.Second)}
	b := Interval{Start: now.Add(3 * time.Second), End: now.Add(4 * time.Second)}

	_, err := at.Put(a, "x")
	assert.NoError(t, err)
	_, err = at.Put(b, "y")
	assert.NoError(t, err)

	v := at.Values()
	assert.Equal(t, 2, at.Size())
	assert.Equal(t, "x", v[0])
	assert.Equal(t, "y", v[1])

	// overlap partially with a -> Put should throw an error
	c := Interval{Start: now.Add(time.Second), End: now.Add(3 * time.Second)}
	_, err = at.Put(c, "z")
	assert.Error(t, err)
	assert.Equal(t, "conflict with existing", err.Error())
	assert.Equal(t, 2, at.Size())
	v = at.Values()

	assert.Equal(t, "x", v[0])
	assert.Equal(t, "y", v[1])

	assert.Equal(t, 2, at.Size())
	at.Remove(a)
	assert.Equal(t, 1, at.Size())

}

func TestSort(t *testing.T) {

	now := time.Now()

	a := Interval{Start: now, End: now.Add(2 * time.Second)}
	b := Interval{Start: now.Add(3 * time.Second), End: now.Add(4 * time.Second)}
	c := Interval{Start: now.Add(5 * time.Second), End: now.Add(6 * time.Second)}
	d := Interval{Start: now.Add(7 * time.Second), End: now.Add(8 * time.Second)}

	intervals := []Interval{c, a, d, b}

	// check intervals are out of order to start with
	assert.Equal(t, intervals[0], c)
	assert.Equal(t, intervals[1], a)
	assert.Equal(t, intervals[2], d)
	assert.Equal(t, intervals[3], b)

	Sort(&intervals)

	// check order is now correct
	assert.Equal(t, intervals[0], a)
	assert.Equal(t, intervals[1], b)
	assert.Equal(t, intervals[2], c)
	assert.Equal(t, intervals[3], d)

}
func TestSortSameStart(t *testing.T) {

	now := time.Now()

	a := Interval{Start: now, End: now.Add(2 * time.Second)}
	b := Interval{Start: now.Add(3 * time.Second), End: now.Add(4 * time.Second)}
	c := Interval{Start: now.Add(3 * time.Second), End: now.Add(8 * time.Second)}
	d := Interval{Start: now.Add(7 * time.Second), End: now.Add(8 * time.Second)}

	intervals := []Interval{c, a, d, b}

	// check intervals are out of order to start with
	assert.Equal(t, intervals[0], c)
	assert.Equal(t, intervals[1], a)
	assert.Equal(t, intervals[2], d)
	assert.Equal(t, intervals[3], b)

	Sort(&intervals)

	// check order is now correct
	assert.Equal(t, intervals[0], a)
	assert.Equal(t, intervals[1], b)
	assert.Equal(t, intervals[2], c)
	assert.Equal(t, intervals[3], d)

}

func TestInvert(t *testing.T) {

	now := time.Now()

	a := Interval{Start: now, End: now.Add(2 * time.Second)}
	b := Interval{Start: now.Add(3 * time.Second), End: now.Add(4 * time.Second)}
	c := Interval{Start: now.Add(5 * time.Second), End: now.Add(6 * time.Second)}
	d := Interval{Start: now.Add(7 * time.Second), End: now.Add(8 * time.Second)}

	intervals := []Interval{c, a, d, b}

	// check intervals are out of order to start with
	assert.Equal(t, intervals[0], c)
	assert.Equal(t, intervals[1], a)
	assert.Equal(t, intervals[2], d)
	assert.Equal(t, intervals[3], b)

	inverted := Invert(intervals)

	// check order is now correct, with inverted intervals

	expected := []Interval{
		Interval{Start: ZeroTime, End: a.Start.Add(-time.Nanosecond)},
		Interval{Start: a.End.Add(time.Nanosecond), End: b.Start.Add(-time.Nanosecond)},
		Interval{Start: b.End.Add(time.Nanosecond), End: c.Start.Add(-time.Nanosecond)},
		Interval{Start: c.End.Add(time.Nanosecond), End: d.Start.Add(-time.Nanosecond)},
		Interval{Start: d.End.Add(time.Nanosecond), End: Infinity},
	}

	assert.Equal(t, inverted, expected)

}

func TestInvertOverlapping(t *testing.T) {

	now := time.Now()

	// c overlaps b
	a := Interval{Start: now, End: now.Add(2 * time.Second)}
	b := Interval{Start: now.Add(3 * time.Second), End: now.Add(5 * time.Second)}
	c := Interval{Start: now.Add(4 * time.Second), End: now.Add(6 * time.Second)}
	d := Interval{Start: now.Add(7 * time.Second), End: now.Add(9 * time.Second)}

	intervals := []Interval{c, a, d, b}

	// check intervals are out of order to start with
	assert.Equal(t, intervals[0], c)
	assert.Equal(t, intervals[1], a)
	assert.Equal(t, intervals[2], d)
	assert.Equal(t, intervals[3], b)

	inverted := Invert(intervals)

	// check order is now correct, with inverted intervals

	expected := []Interval{
		Interval{Start: ZeroTime, End: a.Start.Add(-time.Nanosecond)},
		Interval{Start: a.End.Add(time.Nanosecond), End: b.Start.Add(-time.Nanosecond)},
		//skip b.End, c.Start because within overlapped allow intervals
		Interval{Start: c.End.Add(time.Nanosecond), End: d.Start.Add(-time.Nanosecond)},
		Interval{Start: d.End.Add(time.Nanosecond), End: Infinity},
	}

	assert.Equal(t, inverted, expected)

}

func TestMergeNone(t *testing.T) {

	now := time.Now()

	a := Interval{Start: now, End: now.Add(2 * time.Second)}
	b := Interval{Start: now.Add(3 * time.Second), End: now.Add(4 * time.Second)}

	intervals := []Interval{a, b}

	merged := Merge(intervals)

	// check order is now correct, with inverted intervals

	expected := []Interval{
		Interval{Start: now, End: now.Add(2 * time.Second)},
		Interval{Start: now.Add(3 * time.Second), End: now.Add(4 * time.Second)},
	}

	assert.Equal(t, expected, merged)
}

func TestMergeSimple(t *testing.T) {

	now := time.Now()

	a := Interval{Start: now, End: now.Add(2 * time.Second)}
	b := Interval{Start: now.Add(1 * time.Second), End: now.Add(4 * time.Second)}

	intervals := []Interval{a, b}

	merged := Merge(intervals)

	// check order is now correct, with inverted intervals

	expected := []Interval{
		Interval{Start: now, End: now.Add(4 * time.Second)},
	}

	assert.Equal(t, expected, merged)
}

func TestMergeOverlapping(t *testing.T) {

	now := time.Now()

	// c overlaps b
	a := Interval{Start: now, End: now.Add(2 * time.Second)}
	b := Interval{Start: now.Add(3 * time.Second), End: now.Add(5 * time.Second)}
	c := Interval{Start: now.Add(4 * time.Second), End: now.Add(6 * time.Second)}
	d := Interval{Start: now.Add(7 * time.Second), End: now.Add(9 * time.Second)}

	intervals := []Interval{c, a, d, b}

	// check intervals are out of order to start with
	assert.Equal(t, intervals[0], c)
	assert.Equal(t, intervals[1], a)
	assert.Equal(t, intervals[2], d)
	assert.Equal(t, intervals[3], b)

	merged := Merge(intervals)

	// check order is now correct, with inverted intervals

	expected := []Interval{
		Interval{Start: now, End: now.Add(2 * time.Second)},
		Interval{Start: now.Add(3 * time.Second), End: now.Add(6 * time.Second)},
		Interval{Start: now.Add(7 * time.Second), End: now.Add(9 * time.Second)},
	}

	assert.Equal(t, expected, merged)
}

func TestMergeSameStart(t *testing.T) {

	now := time.Now()

	// c overlaps b
	a := Interval{Start: now, End: now.Add(2 * time.Second)}
	b := Interval{Start: now.Add(2 * time.Second), End: now.Add(9 * time.Second)}
	c := Interval{Start: now.Add(2 * time.Second), End: now.Add(6 * time.Second)}
	d := Interval{Start: now.Add(2 * time.Second), End: now.Add(5 * time.Second)}
	// ensure no cheating by just finding smallest and largest times - this interval is non-contiguous so must be retained as separate interval
	e := Interval{Start: now.Add(11 * time.Second), End: now.Add(12 * time.Second)}

	intervals := []Interval{a, b, c, d, e}

	merged := Merge(intervals)

	// check order is now correct, with inverted intervals

	expected := []Interval{
		Interval{Start: now, End: now.Add(9 * time.Second)},
		Interval{Start: now.Add(11 * time.Second), End: now.Add(12 * time.Second)},
	}

	assert.Equal(t, merged, expected)
}
