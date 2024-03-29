package filter

import (
	"testing"
	"time"

	"github.com/practable/book/internal/interval"
	"github.com/stretchr/testify/assert"
)

// Graphical representation of the intervals used in this test (a = allowed, d = denied, s = session to try)
//
//               20                      50                            120                          180
//               |-------a0------------|                               |-----------a2------------------|
//                             |-------------a1------|                 |-------------a3---------|
//                            40	                  60              120                      161
//                                                                     |---------a4-----|
//                                                                    120             150  //force a same-time slot at least once
//
//
//	5	  10                35    42                       80      90             150       161             200       220
//  |--d0-|                 |--d2-|                        |---d3--|              |---d4----|               |---d5----|
//          |--d1----|                                                            |---d6--|                 |----d7------|
//          15      30                                                                    159                           230
//
//1    8     18   22     34     43  44     55  56       80              125  130       160      162            201  205   230     240
//|-s0-|     |-s1-|      |--s2---|  |--s4--|   |---s5---|               |-s7-|         |---s8---|              |-s9-|     |--s10--|
//
//                       34 38                              82    86                           163     168
//                       |s3|                               |--s6-|                            |--s11--|
//
// The resulting list of denied regions (Dn) is
//0               30  35     42                60+1ns                120      150           161      180+1ns
//|-----D0--------|   |--D1--|                 |---------D2-----------|        |----D3-------|       |----D4-----> infinity
var w = time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)

var a0 = interval.Interval{
	Start: w.Add(20 * time.Second),
	End:   w.Add(50 * time.Second),
}

var a1 = interval.Interval{
	Start: w.Add(40 * time.Second),
	End:   w.Add(60 * time.Second),
}

var a2 = interval.Interval{
	Start: w.Add(120 * time.Second),
	End:   w.Add(180 * time.Second),
}

var a3 = interval.Interval{
	Start: w.Add(120 * time.Second),
	End:   w.Add(161 * time.Second),
}

var a4 = interval.Interval{
	Start: w.Add(120 * time.Second),
	End:   w.Add(150 * time.Second),
}

var d0 = interval.Interval{
	Start: w.Add(5 * time.Second),
	End:   w.Add(10 * time.Second),
}

var d1 = interval.Interval{
	Start: w.Add(15 * time.Second),
	End:   w.Add(30 * time.Second),
}

var d2 = interval.Interval{
	Start: w.Add(35 * time.Second),
	End:   w.Add(42 * time.Second),
}

var d3 = interval.Interval{
	Start: w.Add(80 * time.Second),
	End:   w.Add(90 * time.Second),
}

var d4 = interval.Interval{
	Start: w.Add(150 * time.Second),
	End:   w.Add(161 * time.Second),
}

var d5 = interval.Interval{
	Start: w.Add(200 * time.Second),
	End:   w.Add(230 * time.Second),
}
var d6 = interval.Interval{
	Start: w.Add(150 * time.Second),
	End:   w.Add(159 * time.Second),
}

var d7 = interval.Interval{
	Start: w.Add(200 * time.Second),
	End:   w.Add(220 * time.Second),
}

var s0 = interval.Interval{
	Start: w.Add(1 * time.Second),
	End:   w.Add(8 * time.Second),
}

var s1 = interval.Interval{
	Start: w.Add(18 * time.Second),
	End:   w.Add(22 * time.Second),
}

var s2 = interval.Interval{
	Start: w.Add(34 * time.Second),
	End:   w.Add(43 * time.Second),
}

var s3 = interval.Interval{
	Start: w.Add(34 * time.Second),
	End:   w.Add(38 * time.Second),
}

var s4 = interval.Interval{
	Start: w.Add(44 * time.Second),
	End:   w.Add(55 * time.Second),
}

var s5 = interval.Interval{
	Start: w.Add(56 * time.Second),
	End:   w.Add(90 * time.Second),
}

var s6 = interval.Interval{
	Start: w.Add(82 * time.Second),
	End:   w.Add(86 * time.Second),
}

var s7 = interval.Interval{
	Start: w.Add(125 * time.Second),
	End:   w.Add(130 * time.Second),
}

var s8 = interval.Interval{
	Start: w.Add(160 * time.Second),
	End:   w.Add(162 * time.Second),
}

var s9 = interval.Interval{
	Start: w.Add(201 * time.Second),
	End:   w.Add(205 * time.Second),
}

var s10 = interval.Interval{
	Start: w.Add(230 * time.Second),
	End:   w.Add(240 * time.Second),
}
var s11 = interval.Interval{
	Start: w.Add(163 * time.Second),
	End:   w.Add(168 * time.Second),
}

func TestFilter(t *testing.T) {

	f := New()

	err := f.SetAllowed([]interval.Interval{a0, a1, a2, a3, a4})
	assert.NoError(t, err)

	err = f.SetDenied([]interval.Interval{d0, d1, d2, d3, d4, d5, d6, d7})
	assert.NoError(t, err)

	assert.False(t, f.Allowed(s0))
	assert.False(t, f.Allowed(s1))
	assert.False(t, f.Allowed(s2))
	assert.False(t, f.Allowed(s3))
	assert.True(t, f.Allowed(s4))
	assert.False(t, f.Allowed(s5))
	assert.False(t, f.Allowed(s6))
	assert.True(t, f.Allowed(s7))
	assert.False(t, f.Allowed(s8))
	assert.False(t, f.Allowed(s9))
	assert.False(t, f.Allowed(s10))
	assert.True(t, f.Allowed(s11))

	dl := f.Export()
	exp := []interval.Interval{
		interval.Interval{
			Start: interval.ZeroTime,
			End:   w.Add(30 * time.Second),
		},
		interval.Interval{
			Start: w.Add(35 * time.Second),
			End:   w.Add(42 * time.Second),
		},
		interval.Interval{
			Start: w.Add(60 * time.Second).Add(time.Nanosecond),
			End:   w.Add(120 * time.Second).Add(-time.Nanosecond),
		},
		interval.Interval{
			Start: w.Add(150 * time.Second),
			End:   w.Add(161 * time.Second),
		},
		interval.Interval{
			Start: w.Add(180 * time.Second).Add(time.Nanosecond),
			End:   interval.DistantFuture,
		},
	}

	// to help with debugging test
	assert.Equal(t, exp[0], dl[0])
	assert.Equal(t, exp[1], dl[1])
	assert.Equal(t, exp[2], dl[2])
	assert.Equal(t, exp[3], dl[3])
	assert.Equal(t, exp[4], dl[4])

	assert.Equal(t, exp, dl)
}

// The resulting list of denied regions (Dn) is
//                         30  35     42     60+1ns                120       150           161      180+1ns
//ZeroTime <-----D0--------|   |--D1--|      |---------D2-----------|        |----D3-------|       |----D4-----> infinity

func TestExport(t *testing.T) {

	f := New()

	d := f.Export()

	forever := []interval.Interval{interval.Interval{
		Start: interval.ZeroTime,
		End:   interval.DistantFuture,
	}}

	assert.Equal(t, forever, d)

	err := f.SetAllowed([]interval.Interval{a0})
	assert.NoError(t, err)

	d = f.Export()

	assert.Equal(t, []interval.Interval{interval.Interval{
		Start: interval.ZeroTime,
		End:   a0.Start.Add(-time.Nanosecond),
	}, interval.Interval{
		Start: a0.End.Add(time.Nanosecond),
		End:   interval.DistantFuture,
	}}, d)

	err = f.SetDenied([]interval.Interval{d2})
	assert.NoError(t, err)
	d = f.Export()
	assert.Equal(t, []interval.Interval{interval.Interval{
		Start: interval.ZeroTime,
		End:   a0.Start.Add(-time.Nanosecond),
	}, interval.Interval{
		Start: d2.Start,
		End:   d2.End,
	}, interval.Interval{
		Start: a0.End.Add(time.Nanosecond),
		End:   interval.DistantFuture,
	}}, d)

	f = New()

	err = f.SetDenied([]interval.Interval{d0, d1, d2}) //same as before, except some redundant deny periods
	assert.NoError(t, err)
	d = f.Export()
	assert.Equal(t, forever, d) //adding deny intervals in the absence of allowed intervals just results in no change to the default "forever" deny interval

	f = New()

	err = f.SetAllowed([]interval.Interval{a0})
	assert.NoError(t, err)

	err = f.SetDenied([]interval.Interval{d0, d2}) //same as before, except a redundant deny periods
	assert.NoError(t, err)
	d = f.Export()
	assert.Equal(t, []interval.Interval{interval.Interval{
		Start: interval.ZeroTime,
		End:   a0.Start.Add(-time.Nanosecond),
	}, interval.Interval{
		Start: d2.Start,
		End:   d2.End,
	}, interval.Interval{
		Start: a0.End.Add(time.Nanosecond),
		End:   interval.DistantFuture,
	}}, d)

}
