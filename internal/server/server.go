package server

import (
	"context"
	"time"

	"github.com/timdrysdale/interval/internal/config"
	"github.com/timdrysdale/interval/internal/serve"
	"github.com/timdrysdale/interval/internal/store"
)

// Run starts API server and an interval store to support it
func Run(ctx context.Context, config config.ServerConfig) {

	s := store.New().WithNow(config.Now)

	if config.Now == nil {
		config.Now = func() time.Time { return time.Now() }
	}

	go s.Run(ctx, config.PruneEvery)

	config.Store = s

	go serve.API(ctx, config)

	<-ctx.Done()
}
