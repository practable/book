package book

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	debug := false
	if debug {
		log.SetReportCaller(true)
		log.SetLevel(log.TraceLevel)
		log.SetFormatter(&log.TextFormatter{FullTimestamp: false, DisableColors: true})
		defer log.SetOutput(os.Stdout)

	} else {
		log.SetLevel(log.WarnLevel)
		var ignore bytes.Buffer
		logignore := bufio.NewWriter(&ignore)
		log.SetOutput(logignore)
	}

}

// TestLogin checks there is a book service running, by logging in
// We assume that if book is passing other tests, then this is sufficient
// to confirm that this package is functioning
func TestLogin(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := DefaultConfig()
	go Run(ctx, cfg)

	time.Sleep(time.Second) // let book server start

	client := &http.Client{}
	req, err := http.NewRequest("POST", cfg.Host+"/api/v1/login/someuser", nil)
	assert.NoError(t, err)
	resp, err := client.Do(req)
	assert.NoError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	atr := &AccessToken{}
	err = json.Unmarshal(body, atr)
	assert.NoError(t, err)
	assert.Equal(t, float64(3601), *(atr.Exp)-*(atr.Nbf)) // one hour and one second
	assert.Equal(t, "someuser", *(atr.Sub))
	assert.Equal(t, []string{"booking:user"}, atr.Scopes)
	assert.Equal(t, "ey", (*(atr.Token))[0:2]) //necessary but not sufficient!
}

// TestDefaultConfig checks that known sensible configuration params are provided
// Together with TestLogin, this should be sufficient to check that this method
// of providing book is valid
func TestDefaultConfig(t *testing.T) {

	c := DefaultConfig()

	assert.Equal(t, time.Hour, c.AccessTokenLifetime)
	assert.Equal(t, time.Minute, c.CheckEvery)
	assert.Equal(t, false, c.DisableCancelAfterUse)

	u, err := url.Parse(c.Host)
	assert.Equal(t, "http", u.Scheme)
	host, port, err := net.SplitHostPort(u.Host)
	assert.NoError(t, err)
	assert.Equal(t, "::", host)
	pint, err := strconv.Atoi(port)
	assert.NoError(t, err)
	assert.True(t, pint > 0) //should be non-zero port number

	assert.Equal(t, 6, c.MinUserNameLength)
	assert.Equal(t, time.Minute, c.PruneEvery)
	assert.Equal(t, time.Duration(30*time.Second), c.RequestTimeout)
	assert.True(t, len(c.StoreSecret) > 0)
	assert.Equal(t, "", c.RelaySecret) //user needs to supply this, we cannot know it

}
