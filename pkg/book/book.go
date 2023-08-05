// Package book provides a method starting the book server
// within another golang code, so as to support the testing
// of other services within the practable ecosystem, like status
// it is NOT intended for production usage
package book

import (
	"context"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/phayes/freeport"
	"github.com/practable/book/internal/config"
	"github.com/practable/book/internal/login"
	"github.com/practable/book/internal/server"
)

type AccessToken struct {

	// Audience
	// Required: true
	Aud *string `json:"aud"`

	// Expires At
	// Required: true
	Exp *float64 `json:"exp"`

	// Issued At
	Iat float64 `json:"iat,omitempty"`

	// Not before
	// Required: true
	Nbf *float64 `json:"nbf"`

	// List of scopes
	// Required: true
	Scopes []string `json:"scopes"`

	// Subject
	// Required: true
	Sub *string `json:"sub"`

	// token
	// Required: true
	Token *string `json:"token"`
}

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
		Host:                  "http://[::]:" + strconv.Itoa(port),
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

	jwt.TimeFunc = cfg.Now

	s := server.New(c)
	s.Run(ctx)
	<-ctx.Done()
}

// AdminAuth provides a pre-pared authorization header for testing purposes
func AdminToken(cfg Config, ttl int64, subject string) (string, error) {

	audience := cfg.Host
	scopes := []string{"booking:admin"}
	now := time.Now().Unix()
	nbf := now - 1
	iat := nbf
	exp := nbf + ttl
	t := login.New(audience, subject, scopes, iat, nbf, exp)

	return login.Sign(t, cfg.StoreSecret)

}
