package server

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/practable/book/internal/config"
	"github.com/practable/book/internal/operations"
	"github.com/practable/book/internal/serve"
	"github.com/practable/book/internal/store"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	Config config.ServerConfig
	Store  *store.Store
}

// New Creates a new server, and provides a pointer to underlying store
// so as to permit testing, e.g. mocking time in the store
func New(config config.ServerConfig) *Server {
	s, err := NewWithError(config)
	if err != nil {
		panic(err)
	}
	return s
}

// NewWithError constructs a server and reports persistence initialisation
// failures to production callers without changing the established New helper.
func NewWithError(config config.ServerConfig) (*Server, error) {

	st := store.New().
		WithNow(config.Now).
		WithRelaySecret(string(config.RelaySecret)).
		WithRequestTimeout(config.RequestTimeout).
		WithDisableCancelAfterUse(config.DisableCancelAfterUse)

	if config.GraceRebound != time.Duration(0) {
		st.WithGraceRebound(config.GraceRebound)
	}
	if config.Repository != nil {
		if err := st.WithRepository(config.Repository); err != nil {
			return nil, fmt.Errorf("load persistent booking state: %w", err)
		}
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

	return s, nil

}

// Run API server and an interval store to support it

func (s *Server) Run(ctx context.Context) {

	log.Trace("server.Run started")

	defer func() {
		log.Trace("server.Run stopped")

	}()

	// serve.API captures the interrupt signal, so let it cancel other goro
	// provide other goro with new context, and pass the cancel() to serve.API
	// so it can call it when shutdown

	ctxStore, cancelStore := context.WithCancel(context.Background())

	go s.Store.Run(ctxStore, s.Config.PruneEvery, s.Config.CheckEvery)
	if s.Config.OperationsRepository != nil && s.Config.JobRunnerURL != "" && len(s.Config.WebhookSecret) == 32 {
		dispatcher := &operations.Dispatcher{
			Repository: s.Config.OperationsRepository,
			Endpoint:   s.Config.JobRunnerURL,
			Secret:     s.Config.WebhookSecret,
			Owner:      fmt.Sprintf("%s-%d", hostname(), os.Getpid()),
			Now:        s.Config.Now,
		}
		go dispatcher.Run(ctxStore, s.Config.WebhookPollEvery)
	}

	go serve.API(ctx, s.Config, cancelStore)

	log.Trace("server.Runs started, awaiting context cancellation")

	<-ctxStore.Done() //cannot use ctx.Done() because will leave hanging process when used with book/cmd where there is no cancellation of ctx

}

func hostname() string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return "book"
	}
	return name
}
