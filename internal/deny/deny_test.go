package deny

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeny(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := New("some_uuid", time.Second).WithScheme("http")

	go client.Run(ctx)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, "/bids/deny?bid=bid0&exp=1674164170", req.URL.String())
		rw.WriteHeader(http.StatusNoContent)
	}))

	r := make(chan string)

	client.Request <- Request{
		URL:       strings.TrimPrefix(server.URL, "http://"),
		BookingID: "bid0",
		ExpiresAt: 1674164170,
		Result:    r,
	}

	result := <-r

	assert.Equal(t, "ok", result)

	// Close the server when test finishes
	defer server.Close()
}
