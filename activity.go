package activity

import (
	"sync"

	"github.com/timdrysdale/interval/filter"
)

// Activity represents an individual activity that can be booked
type Activity struct {
	*sync.RWMutex `json:"-"`
	Config        Config             `json:"config"`
	Description   shared.Description `json:"description"`
	Available     filter.Filter      `json:"available"`
	Streams       map[string]*Stream `json:"streams"`
	UI            []*UI              `json:"ui"`
}
