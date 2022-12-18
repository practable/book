package config

import (
	"time"

	"github.com/timdrysdale/interval/internal/store"
)

type ServerConfig struct {
	AccessTokenLifetime time.Duration
	CheckEvery          time.Duration
	GraceRebound        time.Duration
	Host                string
	MinUserNameLength   int
	Now                 func() time.Time
	Port                int
	PruneEvery          time.Duration
	RelaySecret         []byte
	StoreSecret         []byte
	Store               *store.Store
}
