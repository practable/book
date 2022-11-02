package pool

import (
	"sync"

	"github.com/timdrysdale/interval/shared"
)

type Pool struct {
	*sync.RWMutex `json:"-" yaml:"-"`
	Description   shared.Description   `json:"description"`
	Activities    map[string]*Activity `json:"activities"`
	Available     map[string]int64     `json:"available"`
	InUse         map[string]int64     `json:"inUse"`
	Now           func() int64         `json:"-" yaml:"-"`
}
