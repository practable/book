package interval

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	avl "github.com/timdrysdale/interval/trees/avltree"
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

	// overlap partially with a -> should reject the Put
	c := Interval{Start: now.Add(time.Second), End: now.Add(3 * time.Second)}

	_, err = at.Put(c, "z")

	assert.Error(t, err)
	assert.Equal(t, "conflict with existing", err.Error())
	assert.Equal(t, 2, at.Size())
	v = at.Values()

	assert.Equal(t, "x", v[0])
	assert.Equal(t, "y", v[1])
}
