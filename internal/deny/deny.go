package deny

import (
	"context"
	"sync"
	"time"

	"github.com/timdrysdale/interval/internal/login"

	httptransport "github.com/go-openapi/runtime/client"
	ac "github.com/timdrysdale/interval/internal/ac/client"
	ao "github.com/timdrysdale/interval/internal/ac/client/operations"
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
	Scheme  string
	Secret  string
	Timeout time.Duration
}

func New(secret string, timeout time.Duration) *Client {

	return &Client{
		&sync.Mutex{},
		func() time.Time { return time.Now() },
		make(chan Request, 64),
		"https",
		secret,
		timeout,
	}
}

func (c *Client) SetNow(now func() time.Time) *Client {
	c.Lock()
	defer c.Unlock()
	c.now = now
	return c
}

func (c *Client) WithScheme(scheme string) *Client {
	c.Lock()
	defer c.Unlock()
	c.Scheme = scheme
	return c
}

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
				req.Result <- "signing admin token failed because " + err.Error()
			}

			auth := httptransport.APIKeyAuth("Authorization", "header", stoken)
			trans := ac.DefaultTransportConfig().WithHost(req.URL).WithSchemes([]string{c.Scheme})
			client := ac.NewHTTPClientWithConfig(nil, trans)
			param := ao.NewDenyParams().WithTimeout(c.Timeout).WithBid(req.BookingID).WithExp(req.ExpiresAt)
			payload, err := client.Operations.Deny(param, auth)

			if err != nil {
				req.Result <- "relay deny request failed because " + err.Error()
			}

			if !payload.IsSuccess() {
				req.Result <- "relay deny request failed because" + payload.String()
			}

			req.Result <- "ok"
		}

	}
}
