package deny

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	log "github.com/sirupsen/logrus"
	ac "github.com/timdrysdale/interval/internal/ac/client"
	ao "github.com/timdrysdale/interval/internal/ac/client/operations"
	"github.com/timdrysdale/interval/internal/login"
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

	for {
	NEXT:
		select {

		case <-ctx.Done():
			return
		case req, ok := <-c.Request:

			if !ok {
				return //our request channel is closed, so no more to do
			}

			if req.Result == nil {
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
				req.Result <- "signing admin token failed because" + err.Error()
			}

			auth := httptransport.APIKeyAuth("Authorization", "header", stoken)
			URL, err := url.Parse(req.URL)
			if err != nil {
				req.Result <- "relay deny request failed because url parsing error" + err.Error()
			}

			host := strings.TrimPrefix(req.URL, URL.Scheme+"://")

			log.Debugf("scheme: %s, host: %s", URL.Scheme, host)

			trans := ac.DefaultTransportConfig().WithHost(host).WithSchemes([]string{URL.Scheme})
			client := ac.NewHTTPClientWithConfig(nil, trans)
			param := ao.NewDenyParams().WithTimeout(c.Timeout).WithBid(req.BookingID).WithExp(req.ExpiresAt)
			payload, err := client.Operations.Deny(param, auth)

			if err != nil {
				req.Result <- "relay deny request failed because" + err.Error()
			}

			if !payload.IsSuccess() {
				req.Result <- "relay deny request failed because" + payload.String()
			}

			req.Result <- "ok"
		}

	}
}
