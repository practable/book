package datetime

import (
	"errors"
	"time"
)

const (
	format = time.RFC3339
)

var notParsedErr = errors.New("could not parse datetime")

func Parse(dt string) (time.Time, error) {
	t, err := time.Parse(format, dt)

	if err != nil {
		err = notParsedErr
	}

	return t, err
}
