package config

import (
	"time"

	"github.com/timdrysdale/interval/internal/store"
)

type ServerConfig struct {
	AccessTokenLifetime time.Duration
	Host                string
	MinUserNameLength   int
	Port                int
	PruneEvery          time.Duration
	RelaySecret         string
	StoreSecret         []byte
	Store               *store.Store
}
