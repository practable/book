package config

import (
	"time"

	"github.com/practable/book/internal/deny"
	"github.com/practable/book/internal/operations"
	"github.com/practable/book/internal/store"
)

type ServerConfig struct {
	AccessTokenLifetime        time.Duration
	CheckEvery                 time.Duration
	DenyRequests               chan deny.Request
	DisableCancelAfterUse      bool
	GraceRebound               time.Duration
	Host                       string
	MinUserNameLength          int
	Now                        func() time.Time
	Port                       int
	PruneEvery                 time.Duration
	RelaySecret                []byte //TODO update to string to suit internal/login.Sign()
	RequestTimeout             time.Duration
	Repository                 store.BookingRepository
	OperationsRepository       operations.Repository
	JobRunnerURL               string
	WebhookSecret              []byte
	WebhookTolerance           time.Duration
	WebhookPollEvery           time.Duration
	OperationalScheduleEvery   time.Duration
	OperationalScheduleHorizon time.Duration
	StoreSecret                []byte //TODO update to string to suit internal/login.Sign()?
	Store                      *store.Store
}
