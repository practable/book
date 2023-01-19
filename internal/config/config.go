package config

import (
	"time"

	"github.com/timdrysdale/interval/internal/deny"
	"github.com/timdrysdale/interval/internal/store"
)

type ServerConfig struct {
	AccessTokenLifetime   time.Duration
	CheckEvery            time.Duration
	DenyRequests          chan deny.Request
	DisableCancelAfterUse bool
	GraceRebound          time.Duration
	Host                  string
	MinUserNameLength     int
	Now                   func() time.Time
	Port                  int
	PruneEvery            time.Duration
	RelaySecret           []byte //TODO update to string to suit internal/login.Sign()
	RequestTimeout        time.Duration
	StoreSecret           []byte //TODO update to string to suit internal/login.Sign()?
	Store                 *store.Store
}
