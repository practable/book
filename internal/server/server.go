package server

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/interval/internal/config"
	"github.com/timdrysdale/interval/internal/serve"
	"github.com/timdrysdale/interval/internal/store"
)

type Server struct {
	Config config.ServerConfig
	Store  *store.Store
}

// New Creates a new server, and provides a pointer to underlying store
// so as to permit testing, e.g. mocking time in the store
func New(config config.ServerConfig) *Server {

	st := store.New().
		WithNow(config.Now).
		WithRelaySecret(string(config.RelaySecret)).
		WithRequestTimeout(config.RequestTimeout).
		WithDisableCancelAfterUse(config.DisableCancelAfterUse)

	if config.GraceRebound != time.Duration(0) {
		st.WithGraceRebound(config.GraceRebound)
	}

	if config.Now == nil {
		config.Now = func() time.Time { return time.Now() }
	}

	if config.PruneEvery == time.Duration(0) {
		log.Warning("pruneEvery not set, setting to 1h")
		config.PruneEvery = time.Duration(time.Hour)
	}
	if config.CheckEvery == time.Duration(0) {
		log.Warning("checkEvery not set, setting to 1h")
		config.CheckEvery = time.Duration(time.Hour)
	}

	config.Store = st

	s := &Server{
		Config: config,
		Store:  st,
	}

	return s

}

// Run API server and an interval store to support it

func (s *Server) Run(ctx context.Context) {

	go s.Store.Run(ctx, s.Config.PruneEvery, s.Config.CheckEvery)

	go serve.API(ctx, s.Config)

	<-ctx.Done()
}
