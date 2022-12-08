package server

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/interval/internal/config"
	"github.com/timdrysdale/interval/internal/serve"
	"github.com/timdrysdale/interval/internal/store"
)

// Run starts API server and an interval store to support it
func Run(ctx context.Context, config config.ServerConfig) {

	log.Debug("store will prune bookings & diaries every " + config.PruneEvery.String())

	s := store.New().WithNow(config.Now)

	if config.Now == nil {
		config.Now = func() time.Time { return time.Now() }
	}

	if config.PruneEvery == time.Duration(0) {
		log.Warning("pruneEvery not set, setting to 1h")
		config.PruneEvery = time.Duration(time.Hour)
	}

	go s.Run(ctx, config.PruneEvery)

	config.Store = s

	go serve.API(ctx, config)

	<-ctx.Done()
}
