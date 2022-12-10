package datetime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	//https://dave.cheney.net/2019/05/07/prefer-table-driven-tests

	var noTime time.Time //nil time for comparing against result in tests expected to produce error

	tests := map[string]struct {
		input string
		want  time.Time
		error error
	}{
		"ok_fractional_seconds": {input: "2022-12-06T01:34:15.125Z", want: time.Date(2022, 12, 6, 1, 34, 15, 125000000, time.UTC), error: nil},
		"ok_notimezone":         {input: "2022-12-06T01:34:15Z", want: time.Date(2022, 12, 6, 1, 34, 15, 0, time.UTC), error: nil},
		"er_no_z":               {input: "2022-12-06T01:34:15", want: noTime, error: notParsedErr},
		"er_no_t":               {input: "2022-12-06 01:34:15Z", want: noTime, error: notParsedErr},
		"er_not_time":           {input: "dogs and cats", want: noTime, error: notParsedErr},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Parse(tc.input)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.error, err)
		})
	}

}
