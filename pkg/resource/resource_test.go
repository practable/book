package resource

import (
	"bufio"
	"bytes"
	"context"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	rt "github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/golang-jwt/jwt/v4"
	"github.com/phayes/freeport"
	"github.com/practable/book/internal/config"
	"github.com/practable/book/internal/login"
	"github.com/practable/book/internal/server"
	"github.com/practable/book/internal/store"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml" //see note below
)

// imports: we must use "sigs.k8s.io/yaml" to take advantage of our JSON UnMarshall extensions
// for manifests in the Init function where we generate JSON from the YAML
// if manifest tests are breaking unexpectedly, check that the correct import is being used

var debug bool
var cfg config.ServerConfig
var ctp *time.Time
var ct time.Time
var cs, ch string //client scheme and host
var timeout time.Duration
var aa, ua rt.ClientAuthInfoWriter
var s *server.Server
var host, secret string

// Are you thinking about making a models.Manifest object
// to compare responses to? Don't. Tried it.
// Durations don't get populated properly when
// you unmarshal into models.Manifest, so not particularly
// useful for comparing to responses. Better just to use
// strings, and may as well be consistent.
var manifestYAML = []byte(`descriptions:
  d-g-a:
    name: group-a
    type: group
    short: a
  d-g-b:
    name: group-b
    type: group
    short: b
  d-p-a:
    name: policy-a
    type: policy
    short: a
  d-p-b:
    name: policy-b
    type: policy
    short: b
  d-p-modes:
    name: policy-modes
    type: policy
    short: modes
  d-r-a:
    name: resource-a
    type: resource
    short: a
  d-r-b:
    name: resource-b
    type: resource
    short: b
  d-sl-a:
    name: slot-a
    type: slot
    short: a
  d-sl-b:
    name: slot-b
    type: slot
    short: b
  d-sl-modes:
    name: slot-modes
    type: slot
    short: modes
  d-ui-a:
    name: ui-a
    type: ui
    short: a
  d-ui-b:
    name: ui-b
    type: ui
    short: b
display_guides:
  1mFor20m:
    book_ahead: 20m
    duration: 1m
    max_slots: 15
    label: 1m
groups:
  g-a:
    description: d-g-a
    policies: 
      - p-a
  g-b:
    description: d-g-b
    policies:
      - p-b
policies:
  p-a:
    book_ahead: 1h
    description: d-p-a
    display_guides:
      - 1mFor20m
    enforce_book_ahead: true
    enforce_max_bookings: false
    enforce_max_duration: false
    enforce_min_duration: false
    enforce_max_usage: false
    max_bookings: 0
    max_duration: 0s
    min_duration: 0s
    max_usage: 0s
    slots:
    - sl-a
  p-b:
    book_ahead: 2h0m0s
    description: d-p-b
    enforce_book_ahead: true
    enforce_max_bookings: true
    enforce_max_duration: true
    enforce_min_duration: true
    enforce_max_usage: true
    max_bookings: 2
    max_duration: 10m0s
    min_duration: 5m0s
    max_usage: 30m0s
    slots:
    - sl-b
  p-modes:
    allow_start_in_past_within: 1m0s
    book_ahead: 2h0m0s
    description: d-p-modes
    enforce_allow_start_in_past: true
    enforce_book_ahead: true
    enforce_max_bookings: true
    enforce_max_duration: true
    enforce_min_duration: true
    enforce_max_usage: true
    enforce_next_available: true
    enforce_starts_within: true
    enforce_unlimited_users: true
    max_bookings: 2
    max_duration: 10m0s
    min_duration: 5m0s
    max_usage: 30m0s
    next_available: 1m0s
    slots:
    - sl-modes
    starts_within: 1m0s
resources:
  r-a:
    description: d-r-a
    streams:
    - st-a
    - st-b
    topic_stub: aaaa00
  r-b:
    description: d-r-b
    streams:
    - st-a
    - st-b
    topic_stub: bbbb00
slots:
  sl-a:
    description: d-sl-a
    policy: p-a
    resource: r-a
    ui_set: us-a
    window: w-a
  sl-b:
    description: d-sl-b
    policy: p-b
    resource: r-b
    ui_set: us-b
    window: w-b
  sl-modes:
    description: d-sl-modes
    policy: p-modes
    resource: r-b
    ui_set: us-b
    window: w-b
streams:
  st-a:
    url: https://relay-access.practable.io
    connection_type: session
    for: data
    scopes:
    - read
    - write
    topic: tbc
  st-b:
    url: https://relay-access.practable.io
    connection_type: session
    for: video
    scopes:
    - read
    topic: tbc
uis:
  ui-a:
    description: d-ui-a
    url: a
    streams_required:
    - st-a
    - st-b
  ui-b:
    description: d-ui-b
    url: b
    streams_required:
    - st-a
    - st-b
ui_sets:
  us-a:
    uis:
    - ui-a
  us-b:
    uis:
    - ui-a
    - ui-b
windows:
  w-a:
    allowed:
    - start: 2022-11-04T00:00:00Z
      end: 2022-11-06T00:00:00Z
    denied: []
  w-b:
    allowed:
    - start: 2022-11-04T00:00:00Z
      end: 2022-11-06T00:00:00Z
    denied: []`)

var manifestJSON []byte

func init() {
	debug = true
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

	// create JSON versions of the manifests (using our convert functions which use the JSON tags and Unmarshal functions to handle durations)

	var err error

	manifestJSON, err = yaml.YAMLToJSON(manifestYAML)

	if err != nil {
		panic(err)
	}

}

func setNow(s *server.Server, now time.Time) {
	ct = now //this updates the jwt time function
	s.Store.SetNow(func() time.Time { return now })
}

func TestMain(m *testing.M) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port, err := freeport.GetFreePort()
	if err != nil {
		panic(err)
	}

	host = "[::]:" + strconv.Itoa(port)
	fqdn := "http://" + host

	ct = time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	ctp = &ct
	secret = "somesecret"
	cfg = config.ServerConfig{
		CheckEvery:          time.Duration(10 * time.Millisecond),
		GraceRebound:        time.Duration(10 * time.Millisecond),
		Host:                fqdn,
		Port:                port,
		StoreSecret:         []byte(secret),
		RelaySecret:         []byte("anothersecret"),
		MinUserNameLength:   6,
		AccessTokenLifetime: time.Duration(time.Minute),
		Now:                 func() time.Time { return time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) }, //we will update this later as needed with s.Store.SetNow()
		PruneEvery:          time.Duration(10 * time.Millisecond),                                     //short so we convert bookings to old bookings quickly in tests
	}

	// modify the time function used to verify the jwt token
	// so that we can set it from the current time
	jwt.TimeFunc = func() time.Time { return *ctp }

	// scheme and host that should be used with the autogenerated client
	cs = "http"
	ch = "localhost:" + strconv.Itoa(port)
	timeout = time.Second
	satoken, err := signedAdminToken()
	if err != nil {
		panic(err)
	}
	sutoken, err := signedUserToken()
	if err != nil {
		panic(err)
	}
	aa = httptransport.APIKeyAuth("Authorization", "header", satoken)
	ua = httptransport.APIKeyAuth("Authorization", "header", sutoken)

	s = server.New(cfg) //s is global so we can mock time
	go s.Run(ctx)

	time.Sleep(time.Second)

	exitVal := m.Run()

	os.Exit(exitVal)
}

func signedAdminToken() (string, error) {

	audience := cfg.Host
	subject := "someuser"
	scopes := []string{"booking:admin"}
	now := ct.Unix()
	nbf := now - 1
	iat := nbf
	exp := nbf + 86400 //1 day
	token := login.New(audience, subject, scopes, iat, nbf, exp)
	return login.Sign(token, string(cfg.StoreSecret))
}

func signedUserTokenFor(subject string) (string, error) {

	audience := cfg.Host
	scopes := []string{"booking:user"}
	now := ct.Unix()
	nbf := now - 1
	iat := nbf
	exp := nbf + 86400 //1 day
	token := login.New(audience, subject, scopes, iat, nbf, exp)
	return login.Sign(token, string(cfg.StoreSecret))
}

func signedUserToken() (string, error) {
	return signedUserTokenFor("someuser")
}

func loadTestManifest(t *testing.T) string {
	stoken, err := signedAdminToken()
	assert.NoError(t, err)
	client := &http.Client{}
	bodyReader := bytes.NewReader(manifestJSON)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/manifest", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	resp.Body.Close()
	return stoken //for use by other commands in test
}

// TestManifestOK lets us know if our test manifest is correct
func TestManifestOK(t *testing.T) {

	var m store.Manifest

	err := yaml.Unmarshal(manifestYAML, &m)

	assert.NoError(t, err)

	err, msgs := store.CheckManifest(m)

	assert.NoError(t, err)

	if err != nil {
		t.Log(msgs)
	}
}

func TestGetResources(t *testing.T) {

	loadTestManifest(t)

	audience := cfg.Host
	subject := "someuser"
	scopes := []string{"booking:admin"}
	nbf := ct.Add(time.Second * -1)
	iat := ct
	exp := ct.Add(time.Hour * 24) //1 day
	token, err := NewToken(audience, subject, secret, scopes, iat, nbf, exp)

	assert.NoError(t, err)

	c := Config{
		BasePath: "/api/v1",
		Host:     host,
		Scheme:   "http",
		Token:    token,
		Timeout:  time.Duration(5 * time.Second),
	}

	c.Prepare()

	actual, err := c.GetResources()

	assert.NoError(t, err)

	// make a map for comparison

	am := make(map[string]About)

	for _, v := range actual {
		ab := v
		am[ab.Name] = ab
	}

	expected := make(map[string]About)

	expected["r-a"] = About{
		Name:      "r-a",
		Streams:   []string{"st-a", "st-b"},
		TopicStub: "aaaa00",
	}
	expected["r-b"] = About{
		Name:      "r-b",
		Streams:   []string{"st-a", "st-b"},
		TopicStub: "bbbb00",
	}

	assert.Equal(t, expected, am)

}
