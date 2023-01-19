package deny

import (
	"bufio"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var debug bool

func init() {
	debug = true
	if debug {
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{FullTimestamp: false, DisableColors: true})
		defer log.SetOutput(os.Stdout)

	} else {
		log.SetLevel(log.WarnLevel)
		var ignore bytes.Buffer
		logignore := bufio.NewWriter(&ignore)
		log.SetOutput(logignore)
	}

}

func TestDeny(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := New()

	go client.Run(ctx)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, "/bids/deny?bid=bid0&exp=1674164170", req.URL.String())
		rw.WriteHeader(http.StatusNoContent)
	}))

	r := make(chan string)

	client.Request <- Request{
		URL:       server.URL,
		BookingID: "bid0",
		ExpiresAt: 1674164170,
		Result:    r,
	}

	result := <-r

	assert.Equal(t, "ok", result)

	// Close the server when test finishes
	defer server.Close()
}
