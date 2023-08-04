package book

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/phayes/freeport"
	"github.com/practable/book/internal/config"
	"github.com/practable/book/internal/server"
)

type Config struct {
	AccessTokenLifetime   time.Duration
	CheckEvery            time.Duration
	DisableCancelAfterUse bool
	Host                  string
	MinUserNameLength     int
	Now                   func() time.Time
	Port                  int
	PruneEvery            time.Duration
	RelaySecret           string
	RequestTimeout        time.Duration
	StoreSecret           string
}

func DefaultConfig() Config {

	port, err := freeport.GetFreePort()
	if err != nil {
		panic(err)
	}

	return Config{
		AccessTokenLifetime:   time.Hour,
		CheckEvery:            time.Minute,
		DisableCancelAfterUse: false,
		Host:                  "[::]",
		MinUserNameLength:     6,
		Now:                   func() time.Time { return time.Now() },
		Port:                  port,
		PruneEvery:            time.Minute,
		RelaySecret:           "",
		RequestTimeout:        time.Duration(30 * time.Second),
		StoreSecret:           uuid.New().String(),
	}
}

func Run(ctx context.Context, cfg Config) {

	c := config.ServerConfig{
		AccessTokenLifetime:   cfg.AccessTokenLifetime,
		CheckEvery:            cfg.CheckEvery,
		DisableCancelAfterUse: cfg.DisableCancelAfterUse,
		Host:                  cfg.Host,
		MinUserNameLength:     cfg.MinUserNameLength,
		Now:                   cfg.Now,
		Port:                  cfg.Port,
		PruneEvery:            cfg.PruneEvery,
		RelaySecret:           []byte(cfg.RelaySecret),
		RequestTimeout:        cfg.RequestTimeout,
		StoreSecret:           []byte(cfg.StoreSecret),
	}

	s := server.New(c)
	s.Run(ctx)
	<-ctx.Done()
}
