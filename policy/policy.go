package Policy

import (
	"time"

	"github.com/timdrysdale/interval/filter"
	"github.com/timdrysdale/interval/shared"
)

//  https://stackoverflow.com/questions/51774563/yaml-unmarshal-errors-cannot-unmarshal-string-into-time-duration-in-golang
// Unmarshaling of time.Duration works in yaml.v3, https://play.golang.org/p/-6y0zq96gVz"

type Policy struct {
	Description        shared.Description
	EnforceMaxBookings bool
	EnforceMaxDuration bool
	EnforceMinDuration bool
	EnforceMaxUsage    bool
	Filter             filter.Filter
	MaxBookings        int64
	MaxDuration        time.Duration
	MinDuration        time.Duration
	Name               string
	MaxUsage           time.Duration
}
