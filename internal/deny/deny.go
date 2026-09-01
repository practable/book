package deny

import (
	"context"
	"net/url"
	"sync"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	ac "github.com/practable/book/internal/ac/client"
	ao "github.com/practable/book/internal/ac/client/operations"
	"github.com/practable/book/internal/login"
	log "github.com/sirupsen/logrus"
)

type Request struct {
	Result    chan string
	URL       string
	BookingID string
	ExpiresAt int64
}

type Client struct {
	*sync.Mutex
	now     func() time.Time
	Request chan Request
	Secret  string
	Timeout time.Duration
}

// reply reports a terminal result without allowing a caller that has timed out
// to block the single denial worker indefinitely.
func reply(ctx context.Context, result chan string, message string) {
	select {
	case result <- message:
	case <-ctx.Done():
	}
}

func New() *Client {

	return &Client{
		&sync.Mutex{},
		func() time.Time { return time.Now() },
		make(chan Request, 64),
		"replaceme",
		time.Minute,
	}
}

func (c *Client) SetNow(now func() time.Time) *Client {
	c.Lock()
	defer c.Unlock()
	c.now = now
	return c
}

func (c *Client) SetSecret(secret string) *Client {
	c.Lock()
	defer c.Unlock()
	c.Secret = secret
	return c
}

func (c *Client) SetTimeout(d time.Duration) *Client {
	c.Lock()
	defer c.Unlock()
	c.Timeout = d
	return c
}

func (c *Client) WithNow(now func() time.Time) *Client {
	c.Lock()
	defer c.Unlock()
	c.now = now
	return c
}

func (c *Client) WithSecret(secret string) *Client {
	c.Lock()
	defer c.Unlock()
	c.Secret = secret
	return c
}

func (c *Client) WithTimeout(d time.Duration) *Client {
	c.Lock()
	defer c.Unlock()
	c.Timeout = d
	return c
}

/*
func (c *Client) WithScheme(scheme string) *Client {
	c.Lock()
	defer c.Unlock()
	c.Scheme = scheme
	return c
}*/

func (c *Client) Run(ctx context.Context) {
	log.Trace("deny.Run started")
	defer func() {
		log.Trace("deny.Run stopped")
	}()
	for {
	NEXT:
		select {

		case <-ctx.Done():
			log.Trace("deny.Run context cancelled")
			return
		case req, ok := <-c.Request:

			log.WithFields(log.Fields{"request": req}).Debug("deny request received")

			if !ok {
				log.Info("deny stopping permanently because request channel closed")
				return //our request channel is closed, so no more to do
			}

			if req.Result == nil {
				log.WithFields(log.Fields{"request": req}).Error("no results channel supplied to deny")
				break NEXT //user forgot to send us a result channel, so do nothing
			}

			// prep the auth (don't cache, in case using multiple relays)
			audience := req.URL
			subject := "admin"
			scopes := []string{"relay:admin"}
			now := c.now().Unix()
			nbf := now - 1
			iat := nbf
			exp := nbf + 300

			token := login.New(audience, subject, scopes, iat, nbf, exp)
			stoken, err := login.Sign(token, c.Secret)

			if err != nil { //token should generate ok, unless secret is blank?
				msg := "signing admin token failed because" + err.Error()
				log.WithFields(log.Fields{"request": req}).Error("deny error is" + msg)
				reply(ctx, req.Result, msg)
				break NEXT
			}

			auth := httptransport.APIKeyAuth("Authorization", "header", stoken)
			URL, err := url.ParseRequestURI(req.URL)
			if err != nil || URL.Scheme == "" || URL.Host == "" {
				detail := "missing scheme or host"
				if err != nil {
					detail = err.Error()
				}
				msg := "relay deny request failed because url parsing error " + detail
				log.WithFields(log.Fields{"request": req}).Error("deny error is" + msg)
				reply(ctx, req.Result, msg)
				break NEXT
			}

			trans := ac.DefaultTransportConfig().WithSchemes([]string{URL.Scheme}).WithHost(URL.Host)
			if URL.Path != "" && URL.Path != "/" {
				trans = trans.WithBasePath(URL.EscapedPath())
			}

			client := ac.NewHTTPClientWithConfig(nil, trans)
			param := ao.NewDenyParams().WithTimeout(c.Timeout).WithBid(req.BookingID).WithExp(req.ExpiresAt)
			payload, err := client.Operations.Deny(param, auth)

			if err != nil {
				msg := "relay deny request failed with error because" + err.Error()
				log.WithFields(log.Fields{"request": req}).Error("deny error is" + msg)
				reply(ctx, req.Result, msg)
				break NEXT
			}

			if !payload.IsSuccess() {
				msg := "relay deny request failed with no success because" + payload.String()
				log.WithFields(log.Fields{"request": req}).Error("deny error is" + msg)
				reply(ctx, req.Result, msg)
				break NEXT
			}

			log.WithFields(log.Fields{"request": req}).Info("deny successful at cancelling session at relay")
			reply(ctx, req.Result, "ok")
		}

	}
}
