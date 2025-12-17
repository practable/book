package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	rt "github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/phayes/freeport"
	apiclient "github.com/practable/book/internal/client/client"
	"github.com/practable/book/internal/client/client/admin"
	"github.com/practable/book/internal/client/client/users"
	cmodels "github.com/practable/book/internal/client/models"
	"github.com/practable/book/internal/config"
	"github.com/practable/book/internal/login"
	"github.com/practable/book/internal/serve/models"
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
var s *Server

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

var manifestGraceYAML = []byte(`descriptions:
  d-g-a:
    name: group-a
    type: group
    short: a
  d-p-a:
    name: policy-a
    type: policy
    short: a
  d-p-b:
    name: policy-b
    type: policy
    short: b
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
policies:
  p-a:
    book_ahead: 1h
    description: d-p-a
    display_guides:
      - 1mFor20m
    enforce_grace_period: true
    grace_period: 2m
    grace_penalty: 3m
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

var manifestGraceJSON = []byte{}

var bookingsYAML = []byte(`---
- name: bk-0
  cancelled: false
  policy: p-a
  slot: sl-a
  started: false
  unfulfilled: false
  user: u-a
  when:
    start: '2022-11-05T00:10:00Z'
    end: '2022-11-05T00:15:00Z'
- name: bk-1
  cancelled: false
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: u-b
  when:
    start: '2022-11-05T00:20:00Z'
    end: '2022-11-05T00:30:00Z'
`)
var oneBookingYAML = []byte(`---
- name: bk-0
  cancelled: false
  policy: p-a
  slot: sl-a
  started: false
  unfulfilled: false
  user: u-a
  when:
    start: '2022-11-05T00:10:00Z'
    end: '2022-11-05T00:15:00Z'
`)

var bookingsJSON = []byte{}
var bookings2JSON = []byte{}
var bookingsGraceJSON = []byte{}

var bookingsGraceYAML = []byte(`---
- cancelled: false
  name: bk-0
  policy: p-a
  slot: sl-a
  started: false
  unfulfilled: false
  user: user-a
  when:
    start: '2022-11-05T00:01:00Z'
    end: '2022-11-05T00:05:00Z'
- cancelled: false
  name: bk-1
  policy: p-a
  slot: sl-a
  started: false
  unfulfilled: false
  user: user-b
  when:
    start: '2022-11-05T00:06:00Z'
    end: '2022-11-05T00:10:00Z'
- cancelled: false
  name: bk-2
  policy: p-a
  slot: sl-a
  started: false
  unfulfilled: false
  user: user-c
  when:
    start: '2022-11-05T00:11:00Z'
    end: '2022-11-05T00:15:00Z'
`)

var bookings2YAML = []byte(`---
- cancelled: false
  name: bk-0
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: user-a
  when:
    start: '2022-11-05T00:10:00Z'
    end: '2022-11-05T00:15:00Z'
- cancelled: false
  name: bk-1
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: user-b
  when:
    start: '2022-11-05T00:20:00Z'
    end: '2022-11-05T00:30:00Z'
- cancelled: false
  name: bk-2
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: user-c
  when:
    start: '2022-11-05T00:35:00Z'
    end: '2022-11-05T00:40:00Z'
- cancelled: false
  name: bk-3
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: user-d
  when:
    start: '2022-11-05T00:45:00Z'
    end: '2022-11-05T00:50:00Z'
- cancelled: false
  name: bk-4
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: user-e
  when:
    start: '2022-11-05T00:55:00Z'
    end: '2022-11-05T01:00:00Z'
- cancelled: false
  name: bk-5
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: user-f
  when:
    start: '2022-11-05T01:05:00Z'
    end: '2022-11-05T01:10:00Z'
- cancelled: false
  name: bk-6
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: user-g
  when:
    start: '2022-11-05T01:15:00Z'
    end: '2022-11-05T01:20:00Z'
- cancelled: false
  name: bk-7
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: user-h
  when:
    start: '2022-11-05T01:25:00Z'
    end: '2022-11-05T01:30:00Z'
`)

var noBookingsYAML = []byte(`[]`)
var noBookingsJSON = []byte(`[]`) //yes this is the same as the YAML, no {}

var usersYAML = []byte(`---
u-a:
  bookings:
  - bk-0
  old_bookings: []
  groups: []
  usage:
    p-a: 5m0s
u-b:
  bookings:
  - bk-1
  old_bookings: []
  groups: []
  usage:
    p-b: 10m0s
`)
var oldUsersYAML = []byte(`---
u-a:
  bookings: []
  groups: []
  old_bookings: 
  - bk-0
  usage:
    p-a: 5m0s
u-b:
  bookings: []
  groups: []
  old_bookings: 
  - bk-1
  usage:
    p-b: 10m0s
`)

func init() {
	debug = false
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

	// create JSON versions of the manifests (using our convert functions which use the JSON tags and Unmarshal functions to handle durations)

	var err error

	manifestJSON, err = yaml.YAMLToJSON(manifestYAML)

	if err != nil {
		panic(err)
	}

	manifestGraceJSON, err = yaml.YAMLToJSON(manifestGraceYAML)

	if err != nil {
		panic(err)
	}
	bookingsJSON, err = yaml.YAMLToJSON(bookingsYAML)

	if err != nil {
		panic(err)
	}
	bookings2JSON, err = yaml.YAMLToJSON(bookings2YAML)

	if err != nil {
		panic(err)
	}
	bookingsGraceJSON, err = yaml.YAMLToJSON(bookingsGraceYAML)

	if err != nil {
		panic(err)
	}

}

func setNow(s *Server, now time.Time) {
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

	host := "http://[::]:" + strconv.Itoa(port)

	ct = time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	ctp = &ct

	cfg = config.ServerConfig{
		CheckEvery:          time.Duration(10 * time.Millisecond),
		GraceRebound:        time.Duration(10 * time.Millisecond),
		Host:                host,
		Port:                port,
		StoreSecret:         []byte("somesecret"),
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

	s = New(cfg) //s is global so we can mock time
	go s.Run(ctx)

	time.Sleep(time.Second)

	exitVal := m.Run()

	os.Exit(exitVal)
}

// compareClientModelBookings compares two structs, ignoring the pointer values
// and using the values of the individual entries
func equalClientModelBookings(a, b cmodels.Bookings) bool {

	am := make(map[string]cmodels.Booking)
	bm := make(map[string]cmodels.Booking)

	for _, v := range a {
		am[*v.Name] = *v
	}
	for _, v := range b {
		bm[*v.Name] = *v
	}

	return reflect.DeepEqual(am, bm)

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

func setLock(t *testing.T, locked bool, message string) {
	satoken, err := signedAdminToken()
	assert.NoError(t, err)
	auth := httptransport.APIKeyAuth("Authorization", "header", satoken)
	timeout := 1 * time.Second
	c := apiclient.DefaultTransportConfig().WithHost(ch).WithSchemes([]string{cs})
	bc := apiclient.NewHTTPClientWithConfig(nil, c)
	p := admin.NewSetLockParams().WithTimeout(timeout).WithLock(locked).WithMsg(&message)
	_, err = bc.Admin.SetLock(p, auth)
	assert.NoError(t, err)
}

func newBc() *apiclient.Client {
	c := apiclient.DefaultTransportConfig().WithHost(ch).WithSchemes([]string{cs})
	return apiclient.NewHTTPClientWithConfig(nil, c)
}

func getBookings(t *testing.T) cmodels.Bookings {
	stoken, err := signedAdminToken()
	assert.NoError(t, err)
	client := &http.Client{}
	req, err := http.NewRequest("GET", cfg.Host+"/api/v1/admin/bookings", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err := ioutil.ReadAll(resp.Body)
	var exportedBookings cmodels.Bookings
	err = yaml.Unmarshal(body, &exportedBookings)
	assert.NoError(t, err)
	resp.Body.Close()
	return exportedBookings
}
func getOldBookings(t *testing.T) cmodels.Bookings {
	stoken, err := signedAdminToken()
	assert.NoError(t, err)
	client := &http.Client{}
	req, err := http.NewRequest("GET", cfg.Host+"/api/v1/admin/oldbookings", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err := ioutil.ReadAll(resp.Body)
	var exportedBookings cmodels.Bookings
	err = yaml.Unmarshal(body, &exportedBookings)
	assert.NoError(t, err)
	resp.Body.Close()
	return exportedBookings
}
func printBookings(t *testing.T, bm cmodels.Bookings) {
	for k, v := range bm {
		fmt.Print(strconv.Itoa(k) + " : " + *v.User + " " + *v.Policy + " " + *v.Slot + " " + v.When.Start.String() + " " + v.When.End.String() + " " + fmt.Sprintf(" cancelled: %t  started: %t \n", v.Cancelled, v.Started))
	}

}

func printIntervals(t *testing.T, im []*models.Interval) {
	for k, v := range im {
		fmt.Print(strconv.Itoa(k) + " : " + v.Start.String() + " " + v.End.String() + "\n")
	}
}

func removeAllBookings(t *testing.T) {
	stoken, err := signedAdminToken()
	assert.NoError(t, err)
	client := &http.Client{}
	bodyReader := bytes.NewReader(noBookingsJSON)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!

	client = &http.Client{}
	bodyReader = bytes.NewReader(noBookingsJSON)
	req, err = http.NewRequest("PUT", cfg.Host+"/api/v1/admin/oldbookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "application/json")
	resp.Body.Close()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	resp.Body.Close()
}

func addBookings(t *testing.T) {
	stoken, err := signedAdminToken()
	assert.NoError(t, err)
	client := &http.Client{}
	bodyReader := bytes.NewReader(bookings2JSON)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	resp.Body.Close()
	b := getBookings(t)
	assert.Equal(t, 8, len(b))
	if debug {
		fmt.Printf("BOOKINGS: %+v\n", b)
	}

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

func TestReplaceManifestWithClient(t *testing.T) {

	var manifest cmodels.Manifest
	err := json.Unmarshal(manifestJSON, &manifest)
	assert.NoError(t, err)

	satoken, err := signedAdminToken()
	assert.NoError(t, err)
	auth := httptransport.APIKeyAuth("Authorization", "header", satoken)
	timeout := 1 * time.Second
	c := apiclient.DefaultTransportConfig().WithHost(ch).WithSchemes([]string{cs})
	bc := apiclient.NewHTTPClientWithConfig(nil, c)
	p := admin.NewReplaceManifestParams().WithTimeout(timeout).WithManifest(&manifest)
	_, err = bc.Admin.ReplaceManifest(p, auth)
	assert.NoError(t, err)
}

func TestLogin(t *testing.T) {

	client := &http.Client{}
	req, err := http.NewRequest("POST", cfg.Host+"/api/v1/login/someuser", nil)
	assert.NoError(t, err)
	resp, err := client.Do(req)
	assert.NoError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	atr := &models.AccessToken{}
	err = json.Unmarshal(body, atr)
	assert.NoError(t, err)
	assert.Equal(t, float64(61), *(atr.Exp)-*(atr.Nbf)) //61 seconds
	assert.Equal(t, "someuser", *(atr.Sub))
	assert.Equal(t, []string{"booking:user"}, atr.Scopes)
	assert.Equal(t, "ey", (*(atr.Token))[0:2]) //necessary but not sufficient!

	// login as user u-g (too short)
	client = &http.Client{}
	req, err = http.NewRequest("POST", cfg.Host+"/api/v1/login/u-g", nil)
	assert.NoError(t, err)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.NoError(t, err)
	expected := `{"code":"404","message":"user name must be 6 or more alphanumeric characters"}` + "\n"
	assert.Equal(t, expected, string(body))

}

func TestCheckReplaceExportManifest(t *testing.T) {

	// make admin token
	stoken, err := signedAdminToken()
	assert.NoError(t, err)

	//check manifest
	client := &http.Client{}
	bodyReader := bytes.NewReader(manifestJSON)
	req, err := http.NewRequest("GET", cfg.Host+"/api/v1/admin/manifest/check", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, "204 No Content", resp.Status) //should be ok!
	resp.Body.Close()
	//replace manifest
	client = &http.Client{}
	bodyReader = bytes.NewReader(manifestJSON)
	req, err = http.NewRequest("PUT", cfg.Host+"/api/v1/admin/manifest", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err := ioutil.ReadAll(resp.Body)
	var ssa store.StoreStatusAdmin
	err = json.Unmarshal(body, &ssa)
	assert.NoError(t, err)
	resp.Body.Close()

	// the manifest does not reset all aspects of the store status
	// avoid errors when go test -count=N with N >1 by only checking
	// what uploading the manifest affects
	assert.Equal(t, int64(12), ssa.Descriptions)
	assert.Equal(t, int64(2), ssa.Filters)
	assert.Equal(t, int64(3), ssa.Policies)
	assert.Equal(t, int64(2), ssa.Resources)
	assert.Equal(t, int64(3), ssa.Slots)
	assert.Equal(t, int64(2), ssa.Streams)
	assert.Equal(t, int64(2), ssa.UIs)
	assert.Equal(t, int64(2), ssa.UISets)
	assert.Equal(t, int64(2), ssa.Windows)

	// export manifest
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/manifest", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err = client.Do(req)
	body, err = ioutil.ReadAll(resp.Body)
	var expectedManifest, exportedManifest store.Manifest
	err = yaml.Unmarshal(manifestYAML, &expectedManifest)
	assert.NoError(t, err)
	err = json.Unmarshal(body, &exportedManifest)
	assert.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, expectedManifest, exportedManifest)

	// for troubleshooting when adding p-modes to manifest
	if !reflect.DeepEqual(expectedManifest, exportedManifest) {

		t.Log(string(body))

		auth := httptransport.APIKeyAuth("Authorization", "header", stoken)
		c := apiclient.DefaultTransportConfig().WithHost(ch).WithSchemes([]string{cs})
		bc := apiclient.NewHTTPClientWithConfig(nil, c)
		p := users.NewGetPolicyParams().WithTimeout(timeout).WithPolicyName("p-modes")
		policy, err := bc.Users.GetPolicy(p, auth)
		assert.NoError(t, err)
		pretty, err := json.Marshal(policy)
		assert.NoError(t, err)
		t.Log(string(pretty))
		expected := `{"Payload":{"allow_start_in_past_within":"1m0s","book_ahead":"2h0m0s","description":{"name":"policy-modes","short":"modes","type":"policy"},"enforce_allow_start_in_past":true,"enforce_book_ahead":true,"enforce_max_bookings":true,"enforce_max_duration":true,"enforce_max_usage":true,"enforce_min_duration":true,"enforce_next_available":true,"enforce_starts_within":true,"enforce_unlimited_users":true,"max_bookings":2,"max_duration":"10m0s","max_usage":"30m0s","min_duration":"5m0s","next_available":"1m0s","slots":["sl-modes"],"starts_within":"1m0s"}}`
		assert.Equal(t, expected, string(pretty))

		if expected == string(pretty) {
			t.Log("policy is correctly returned via GetPolicy, hence is read from manifest ok - check ExportManifest")
		}

	}

}

func TestReplaceExportBookingsExportUsers(t *testing.T) {

	stoken := loadTestManifest(t)

	// replace bookings
	client := &http.Client{}
	bodyReader := bytes.NewReader(bookingsJSON)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!

	// export bookings
	bc := newBc()
	status, err := bc.Admin.ExportBookings(
		admin.NewExportBookingsParams().WithTimeout(timeout),
		aa)
	assert.NoError(t, err)

	var expectedBookings cmodels.Bookings
	err = yaml.Unmarshal(bookingsYAML, &expectedBookings)
	assert.NoError(t, err)
	assert.True(t, equalClientModelBookings(expectedBookings, status.Payload))

	// check our comparison is working - this test should fail
	err = yaml.Unmarshal(oneBookingYAML, &expectedBookings)
	assert.False(t, reflect.DeepEqual(expectedBookings, status.Payload))
	resp.Body.Close()

	// export users (now there are bookings we will have users)
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/users", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err := ioutil.ReadAll(resp.Body)
	var expectedUsers, exportedUsers map[string]store.UserExternal
	err = yaml.Unmarshal(usersYAML, &expectedUsers)
	err = yaml.Unmarshal(body, &exportedUsers)
	assert.NoError(t, err)
	// we have extra users if we run tests over and over,
	// so only check for the presence of users that we know MUST exist
	// and do not fail if other users are present
	for n, ue := range expectedUsers {
		if ua, ok := exportedUsers[n]; ok {
			assert.Equal(t, ue, ua)
		} else {
			t.Errorf("missing user " + n)
		}
	}

	resp.Body.Close()

}

func TestReplaceExportOldBookingsExportUsers(t *testing.T) {

	stoken := loadTestManifest(t)

	// Replace our old bookings (in this case, remove them)
	// so that other old bookings are not causing test fails when
	// we count how many old bookings we have
	client := &http.Client{}
	bodyReader := bytes.NewReader(noBookingsJSON)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/oldbookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	resp.Body.Close()

	// replace bookings
	client = &http.Client{}
	bodyReader = bytes.NewReader(bookingsJSON)
	req, err = http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	resp.Body.Close()

	// move time forward
	ct := time.Date(2022, 11, 5, 6, 0, 0, 0, time.UTC)
	setNow(s, ct)
	time.Sleep(50 * time.Millisecond) //wait for pruning to happen

	// export bookings
	bc := newBc()
	status, err := bc.Admin.ExportBookings(
		admin.NewExportBookingsParams().WithTimeout(timeout),
		aa)
	assert.NoError(t, err)
	var expectedBookings cmodels.Bookings
	err = yaml.Unmarshal(noBookingsYAML, &expectedBookings)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(expectedBookings, status.Payload))

	// export old bookings
	bc = newBc()
	status1, err := bc.Admin.ExportOldBookings(
		admin.NewExportOldBookingsParams().WithTimeout(timeout),
		aa)
	assert.NoError(t, err)
	err = yaml.Unmarshal(bookingsYAML, &expectedBookings)
	assert.NoError(t, err)
	//reflect.DeepEqual gives false negatives sometimes, perhaps due to pointers
	assert.True(t, equalClientModelBookings(expectedBookings, status1.Payload))

	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/status", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	var ssa store.StoreStatusAdmin
	err = json.Unmarshal(body, &ssa)
	assert.NoError(t, err)
	resp.Body.Close()

	// check only the store status elements that are side-effects of this test,
	assert.Equal(t, int64(0), ssa.Bookings)
	assert.Equal(t, int64(2), ssa.OldBookings)
	assert.Equal(t, int64(2), ssa.Users)

	// export users (now there are bookings we will have users)
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/users", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err = ioutil.ReadAll(resp.Body)
	var expectedUsers, exportedUsers map[string]store.UserExternal
	err = yaml.Unmarshal(oldUsersYAML, &expectedUsers)
	err = yaml.Unmarshal(body, &exportedUsers)
	assert.NoError(t, err)
	resp.Body.Close()

	// we have extra users if we run tests over and over,
	// so only check for the presence of users that we know MUST exist
	// and do not fail if other users are present
	// although TODO perhaps in this test it should be strict equality
	// since we remove the oldBookings at the start...
	for n, ue := range expectedUsers {
		if ua, ok := exportedUsers[n]; ok {
			assert.Equal(t, ue, ua)
		} else {
			t.Errorf("missing user " + n)
		}
	}

	// Replace our old bookings (in this case, remove them)
	client = &http.Client{}
	bodyReader = bytes.NewReader(noBookingsJSON)
	req, err = http.NewRequest("PUT", cfg.Host+"/api/v1/admin/oldbookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!

	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/status", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp.Body.Close()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &ssa)
	assert.NoError(t, err)
	resp.Body.Close()

	// check only the store status elements that are side-effects of this test,
	assert.Equal(t, int64(0), ssa.Bookings)
	assert.Equal(t, int64(0), ssa.OldBookings)
	assert.Equal(t, int64(0), ssa.Users)

}

func TestSetLock(t *testing.T) {
	stoken := loadTestManifest(t)
	removeAllBookings(t) // ensure consistent state

	// lock the store
	client := &http.Client{}
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/status", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)

	// add query params
	q := req.URL.Query()
	q.Add("lock", "true")
	q.Add("msg", "Locked for maintenance")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!

	// check store is locked
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/status", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp.Body.Close()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	var ssa store.StoreStatusAdmin
	err = json.Unmarshal(body, &ssa)
	assert.NoError(t, err)
	resp.Body.Close()
	esa := store.StoreStatusAdmin{
		Locked:       true,
		Message:      "Locked for maintenance",
		Now:          time.Date(2022, 11, 5, 6, 0, 0, 0, time.UTC),
		Bookings:     0,
		Descriptions: 12,
		Filters:      2,
		OldBookings:  0,
		Policies:     3,
		Resources:    2,
		Slots:        3,
		Streams:      2,
		UIs:          2,
		UISets:       2,
		Users:        0,
		Windows:      2}
	assert.Equal(t, esa, ssa)

	// unlock the store
	client = &http.Client{}
	req, err = http.NewRequest("PUT", cfg.Host+"/api/v1/admin/status", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)

	// add query params
	q = req.URL.Query()
	q.Add("lock", "false")
	q.Add("msg", "Open for bookings")
	req.URL.RawQuery = q.Encode()

	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!

	// check store is unlocked
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/status", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp.Body.Close()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &ssa)
	assert.NoError(t, err)
	resp.Body.Close()
	esa = store.StoreStatusAdmin{
		Locked:       false,
		Message:      "Open for bookings",
		Now:          time.Date(2022, 11, 5, 6, 0, 0, 0, time.UTC),
		Bookings:     0,
		Descriptions: 12,
		Filters:      2,
		OldBookings:  0,
		Policies:     3,
		Resources:    2,
		Slots:        3,
		Streams:      2,
		UIs:          2,
		UISets:       2,
		Users:        0,
		Windows:      2}
	assert.Equal(t, esa, ssa)

}

func TestSetGetSlotIsAvailable(t *testing.T) {

	stoken := loadTestManifest(t)

	// make unavailable slot sl-a
	client := &http.Client{}
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/slots/sl-a", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)

	// add query params
	q := req.URL.Query()
	q.Add("available", "false")
	q.Add("reason", "failed self-test")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode) //should be ok!
	resp.Body.Close()

	// check unavailable slot sl-a
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/slots/sl-a", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	var ss models.ResourceStatus
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &ss)
	assert.Equal(t, false, *(ss.Available))
	assert.Equal(t, "unavailable because failed self-test", *(ss.Reason))
	resp.Body.Close()

	// check available slot sl-b
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/slots/sl-b", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &ss)
	assert.Equal(t, true, *(ss.Available))
	assert.Equal(t, "Loaded at 2022-11-05T06:00:00Z", *(ss.Reason))
	resp.Body.Close()

	// make available again slot sl-a
	client = &http.Client{}
	req, err = http.NewRequest("PUT", cfg.Host+"/api/v1/admin/slots/sl-a", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)

	// add query params
	q = req.URL.Query()
	q.Add("available", "true")
	q.Add("reason", "passed self-test")
	req.URL.RawQuery = q.Encode()

	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode) //should be ok!
	resp.Body.Close()

	// check available again
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/slots/sl-a", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &ss)
	assert.Equal(t, true, *(ss.Available))
	assert.Equal(t, "passed self-test", *(ss.Reason))
	resp.Body.Close()
}

func TestGetDescription(t *testing.T) {

	loadTestManifest(t)

	stoken, err := signedUserToken()
	assert.NoError(t, err)

	// get description
	client := &http.Client{}
	req, err := http.NewRequest("GET", cfg.Host+"/api/v1/descriptions/d-r-a", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err := ioutil.ReadAll(resp.Body)
	expected := `{"name":"resource-a","short":"a","type":"resource"}` + "\n"
	assert.Equal(t, expected, string(body))
	resp.Body.Close()

}

func TestGetPolicy(t *testing.T) {

	loadTestManifest(t)

	stoken, err := signedUserToken()
	assert.NoError(t, err)

	// get description
	client := &http.Client{}
	req, err := http.NewRequest("GET", cfg.Host+"/api/v1/policies/p-a", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err := ioutil.ReadAll(resp.Body)
	expected := `{"allow_start_in_past_within":"0s","book_ahead":"1h0m0s","description":{"name":"policy-a","short":"a","type":"policy"},"display_guides":{"1mFor20m":{"book_ahead":"20m0s","duration":"1m0s","label":"1m","max_slots":15}},"enforce_book_ahead":true,"max_duration":"0s","max_usage":"0s","min_duration":"0s","next_available":"0s","slots":{"sl-a":{"description":{"name":"slot-a","short":"a","type":"slot"},"policy":"p-a"}},"starts_within":"0s"}` + "\n"
	assert.Equal(t, expected, string(body))
	resp.Body.Close()

}
func TestGetAvailability(t *testing.T) {
	// make sure our pre-prepared bookings are in the future
	// other tests may have advanced time
	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	setNow(s, ct)
	satoken := loadTestManifest(t)
	removeAllBookings(t)
	bm := getBookings(t)
	assert.Equal(t, 0, len(bm))

	// load some bookings to break up the future availability in discrete intervals
	client := &http.Client{}
	bodyReader := bytes.NewReader(bookings2JSON)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", satoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!

	bm = getBookings(t)

	if debug {
		printBookings(t, bm)
	}

	// get availability as user
	sutoken, err := signedUserToken()
	assert.NoError(t, err)

	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/slots/sl-b", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp.Body.Close()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err := ioutil.ReadAll(resp.Body)
	expected := `[{"end":"2022-11-05T00:09:59.999Z","start":"2022-11-05T00:00:00.000Z"},{"end":"2022-11-05T00:19:59.999Z","start":"2022-11-05T00:15:00.000Z"},{"end":"2022-11-05T00:34:59.999Z","start":"2022-11-05T00:30:00.000Z"},{"end":"2022-11-05T00:44:59.999Z","start":"2022-11-05T00:40:00.000Z"},{"end":"2022-11-05T00:54:59.999Z","start":"2022-11-05T00:50:00.000Z"},{"end":"2022-11-05T01:04:59.999Z","start":"2022-11-05T01:00:00.000Z"},{"end":"2022-11-05T01:14:59.999Z","start":"2022-11-05T01:10:00.000Z"},{"end":"2022-11-05T01:24:59.999Z","start":"2022-11-05T01:20:00.000Z"},{"end":"2022-11-05T02:00:00.000Z","start":"2022-11-05T01:30:00.000Z"}]` + "\n"
	assert.Equal(t, expected, string(body))
	im := []*models.Interval{}
	err = json.Unmarshal(body, &im)
	assert.NoError(t, err)

	if debug {
		print("GET\n")
		printIntervals(t, im)
	}
	resp.Body.Close()

	// Try pagination using limit
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/slots/sl-b", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	// add query params
	q := req.URL.Query()
	q.Add("limit", "3")
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err = ioutil.ReadAll(resp.Body)
	expected = `[{"end":"2022-11-05T00:09:59.999Z","start":"2022-11-05T00:00:00.000Z"},{"end":"2022-11-05T00:19:59.999Z","start":"2022-11-05T00:15:00.000Z"},{"end":"2022-11-05T00:34:59.999Z","start":"2022-11-05T00:30:00.000Z"}]` + "\n"
	assert.Equal(t, expected, string(body))
	err = json.Unmarshal(body, &im)
	assert.NoError(t, err)
	if debug {
		print("GET?limit=3\n")
		printIntervals(t, im)
	}
	resp.Body.Close()

	// Try pagination using limit, checking specifying offset=0 works as expected
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/slots/sl-b", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	// add query params
	q = req.URL.Query()
	q.Add("limit", "3")
	q.Add("offset", "0")
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err = ioutil.ReadAll(resp.Body)
	expected = `[{"end":"2022-11-05T00:09:59.999Z","start":"2022-11-05T00:00:00.000Z"},{"end":"2022-11-05T00:19:59.999Z","start":"2022-11-05T00:15:00.000Z"},{"end":"2022-11-05T00:34:59.999Z","start":"2022-11-05T00:30:00.000Z"}]` + "\n"
	assert.Equal(t, expected, string(body))
	err = json.Unmarshal(body, &im)
	assert.NoError(t, err)
	if debug {
		print("GET?limit=3&offset=0\n")
		printIntervals(t, im)
	}
	resp.Body.Close()

	// Try pagination using limit, checking specifying offset=<n*limit> works as expected
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/slots/sl-b", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	// add query params
	q = req.URL.Query()
	q.Add("limit", "3")
	q.Add("offset", "3")
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err = ioutil.ReadAll(resp.Body)
	expected = `[{"end":"2022-11-05T00:44:59.999Z","start":"2022-11-05T00:40:00.000Z"},{"end":"2022-11-05T00:54:59.999Z","start":"2022-11-05T00:50:00.000Z"},{"end":"2022-11-05T01:04:59.999Z","start":"2022-11-05T01:00:00.000Z"}]` + "\n"
	assert.Equal(t, expected, string(body))
	err = json.Unmarshal(body, &im)
	assert.NoError(t, err)
	if debug {
		print("GET?limit=3&offset=3\n")
		printIntervals(t, im)
	}
	resp.Body.Close()

	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/slots/sl-b", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	// add query params
	q = req.URL.Query()
	q.Add("limit", "3")
	q.Add("offset", "6")
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err = ioutil.ReadAll(resp.Body)
	expected = `[{"end":"2022-11-05T01:14:59.999Z","start":"2022-11-05T01:10:00.000Z"},{"end":"2022-11-05T01:24:59.999Z","start":"2022-11-05T01:20:00.000Z"},{"end":"2022-11-05T02:00:00.000Z","start":"2022-11-05T01:30:00.000Z"}]` + "\n"
	assert.Equal(t, expected, string(body))
	err = json.Unmarshal(body, &im)
	assert.NoError(t, err)
	if debug {
		print("GET?limit=3&offset=6\n")
		printIntervals(t, im)
	}
	resp.Body.Close()
}

func TestMakeBooking(t *testing.T) {

	// make sure our pre-prepared bookings are in the future
	// other tests may have advanced time
	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	setNow(s, ct)
	satoken := loadTestManifest(t)
	removeAllBookings(t)
	bm := getBookings(t)
	assert.Equal(t, 0, len(bm))

	// user access token
	sutoken, err := signedUserToken()
	assert.NoError(t, err)

	// TODO add group for user!
	client := &http.Client{}
	req, err := http.NewRequest("POST", cfg.Host+"/api/v1/users/someuser/groups/g-b", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode) //should be ok!
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	client = &http.Client{}
	req, err = http.NewRequest("POST", cfg.Host+"/api/v1/slots/sl-b", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	// add query params
	q := req.URL.Query()
	q.Add("user_name", "someuser") //must match token
	q.Add("from", "2022-11-05T00:01:00Z")
	q.Add("to", "2022-11-05T00:07:00Z")
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode) //should be ok!
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	//export Bookings to check...
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/bookings", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", satoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err = ioutil.ReadAll(resp.Body)
	expectedJSON := []byte(`[{"cancelled":false,"name":"cc85c042-4f9f-42d6-8a37-1a1e6b501640","policy":"p-b","slot":"sl-b","started":false,"unfulfilled":false,"user":"someuser","when":{"start":"2022-11-05T00:01:00Z","end":"2022-11-05T00:07:00Z"}}]`)
	var expected, actual cmodels.Bookings
	err = json.Unmarshal(body, &actual)
	resp.Body.Close()
	assert.NoError(t, err)
	err = json.Unmarshal(expectedJSON, &expected)
	assert.NoError(t, err)

	// names are autogenerated so cannot compare their values but should be same length
	// because both are UUID
	assert.Equal(t, len(*(expected[0].Name)), len(*(actual[0].Name))) //expect a UUID-length random name
	*(expected[0].Name) = ""
	*(actual[0].Name) = ""
	assert.Equal(t, expected[0], actual[0]) //compared bookings omitting the names

}

func TestGetStoreStatus(t *testing.T) {

	// make sure our pre-prepared bookings are in the future
	// other tests may have advanced time
	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	setNow(s, ct)

	satoken := loadTestManifest(t)
	removeAllBookings(t)
	bm := getBookings(t)
	assert.Equal(t, 0, len(bm))

	// unlock the store
	client := &http.Client{}
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/status", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", satoken)

	// add query params
	q := req.URL.Query()
	q.Add("lock", "false")
	q.Add("msg", "Open for bookings")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!

	// check status
	sutoken, err := signedUserToken()
	assert.NoError(t, err)
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/users/status", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp.Body.Close()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	var ssa store.StoreStatusAdmin
	err = json.Unmarshal(body, &ssa)
	assert.NoError(t, err)
	resp.Body.Close()
	esa := store.StoreStatusAdmin{
		Locked:  false,
		Message: "Open for bookings",
		Now:     ct,
	}

	assert.Equal(t, esa, ssa)
}

// Test GetBookings, CancelBookings, GetOldBookings
func TestGetCancelBookingsGetOldBookings(t *testing.T) {

	// make sure our pre-prepared bookings are in the future
	// other tests may have advanced time
	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	setNow(s, ct)
	satoken := loadTestManifest(t)
	removeAllBookings(t)
	bm := getBookings(t)
	assert.Equal(t, 0, len(bm))

	// load some bookings to break up the future availability in discrete intervals
	client := &http.Client{}
	bodyReader := bytes.NewReader(bookings2JSON)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", satoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!

	bm = getBookings(t)

	if debug {
		printBookings(t, bm)
	}

	// login as user-g
	client = &http.Client{}
	req, err = http.NewRequest("POST", cfg.Host+"/api/v1/login/user-g", nil)
	assert.NoError(t, err)
	resp.Body.Close()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.NoError(t, err)
	resp.Body.Close()
	atr := &models.AccessToken{}
	err = json.Unmarshal(body, atr)
	assert.NoError(t, err)
	sutoken := *(atr.Token)

	// get bookings for user u-g
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/users/user-g/bookings", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	bookings := `[{"cancelled_at":"0001-01-01T00:00:00.000Z","name":"bk-6","policy":"p-b","slot":"sl-b","started_at":"0001-01-01T00:00:00.000Z","user":"user-g","when":{"end":"2022-11-05T01:20:00.000Z","start":"2022-11-05T01:15:00.000Z"}}]` + "\n"

	assert.Equal(t, bookings, string(body))

	// cancel booking bk-6
	client = &http.Client{}
	req, err = http.NewRequest("DELETE", cfg.Host+"/api/v1/users/user-g/bookings/bk-6", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode) //NotFound if successful deletion

	// get bookings again
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/users/user-g/bookings", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp.Body.Close()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	bookings = `[]` + "\n"
	assert.Equal(t, bookings, string(body))

	// get old bookings for user (should be none at this time)
	sutoken, err = signedUserTokenFor("user-f")
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/users/user-f/oldbookings", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	bookings = `[]` + "\n"
	assert.Equal(t, bookings, string(body))

	//move time on so that bookings become old
	ct = time.Date(2022, 12, 5, 0, 0, 0, 0, time.UTC)
	setNow(s, ct)
	time.Sleep(50 * time.Millisecond) //allow pruning to take place

	// get old bookings for user
	sutoken, err = signedUserTokenFor("user-f") //need new token that is valid for current time
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/users/user-f/oldbookings", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	bookings = `[{"cancelled_at":"0001-01-01T00:00:00.000Z","name":"bk-5","policy":"p-b","slot":"sl-b","started_at":"0001-01-01T00:00:00.000Z","user":"user-f","when":{"end":"2022-11-05T01:10:00.000Z","start":"2022-11-05T01:05:00.000Z"}}]` + "\n"

	assert.Equal(t, bookings, string(body))

}

func TestGetActivity(t *testing.T) {

	// make sure our pre-prepared bookings are in the future
	// other tests may have advanced time
	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	setNow(s, ct)
	satoken := loadTestManifest(t)
	removeAllBookings(t)
	bm := getBookings(t)
	assert.Equal(t, 0, len(bm))

	// load some bookings to break up the future availability in discrete intervals
	client := &http.Client{}
	bodyReader := bytes.NewReader(bookings2JSON)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", satoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!

	// get bookings for user u-g
	sutoken, err := signedUserTokenFor("user-g")
	assert.NoError(t, err)
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/users/user-g/bookings", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp.Body.Close()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	bookings := `[{"cancelled_at":"0001-01-01T00:00:00.000Z","name":"bk-6","policy":"p-b","slot":"sl-b","started_at":"0001-01-01T00:00:00.000Z","user":"user-g","when":{"end":"2022-11-05T01:20:00.000Z","start":"2022-11-05T01:15:00.000Z"}}]` + "\n"

	assert.Equal(t, bookings, string(body))

	// move time forward to within the booked activity
	ct = time.Date(2022, 11, 5, 1, 15, 1, 0, time.UTC)
	setNow(s, ct)

	// getActivity for booking
	client = &http.Client{}
	req, err = http.NewRequest("PUT", cfg.Host+"/api/v1/users/user-g/bookings/bk-6", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	//Order might change. Note tokens now include BookingID field

	var a models.Activity

	err = json.Unmarshal(body, &a)

	assert.NoError(t, err)

	assert.Equal(t, "slot-b", *a.Description.Name)
	assert.Equal(t, "b", a.Description.Short)
	assert.Equal(t, "slot", *a.Description.Type)

	assert.Equal(t, float64(1667611200), *a.Exp)
	assert.Equal(t, float64(1667610900), *a.Nbf)

	streams := make(map[string]models.ActivityStream)

	for _, s := range a.Streams {
		streams[*s.For] = *s
	}

	sd := streams["data"]
	sv := streams["video"]

	assert.Equal(t, "https://relay-access.practable.io", *sd.Audience)
	assert.Equal(t, "https://relay-access.practable.io", *sv.Audience)
	assert.Equal(t, "session", *sd.ConnectionType)
	assert.Equal(t, "session", *sv.ConnectionType)

	sm := make(map[string]bool)

	for _, scope := range sv.Scopes {
		sm[scope] = true
	}

	assert.Equal(t, 1, len(sm))
	assert.Equal(t, sm["read"], true)

	sm = make(map[string]bool)

	for _, scope := range sd.Scopes {
		sm[scope] = true
	}

	assert.Equal(t, 2, len(sm))
	assert.Equal(t, sm["read"], true)
	assert.Equal(t, sm["write"], true)

	assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJib29raW5nX2lkIjoiYmstNiIsInRvcGljIjoiYmJiYjAwLXN0LWEiLCJwcmVmaXgiOiJzZXNzaW9uIiwic2NvcGVzIjpbInJlYWQiLCJ3cml0ZSJdLCJzdWIiOiJ1c2VyLWciLCJhdWQiOlsiaHR0cHM6Ly9yZWxheS1hY2Nlc3MucHJhY3RhYmxlLmlvIl0sImV4cCI6MTY2NzYxMTIwMCwibmJmIjoxNjY3NjEwOTAwLCJpYXQiOjE2Njc2MTA5MDB9.Y_6UhVu1roW-rIKPlLce7qNUHek6dQ0WXwO4boxFyFQ", sd.Token)

	assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJib29raW5nX2lkIjoiYmstNiIsInRvcGljIjoiYmJiYjAwLXN0LWIiLCJwcmVmaXgiOiJzZXNzaW9uIiwic2NvcGVzIjpbInJlYWQiXSwic3ViIjoidXNlci1nIiwiYXVkIjpbImh0dHBzOi8vcmVsYXktYWNjZXNzLnByYWN0YWJsZS5pbyJdLCJleHAiOjE2Njc2MTEyMDAsIm5iZiI6MTY2NzYxMDkwMCwiaWF0IjoxNjY3NjEwOTAwfQ.uu76zhbEw0ycSuUMEYkgeeADev2GTR-NNW3O2ulx6ZQ", sv.Token)

	assert.Equal(t, "bbbb00-st-a", *sd.Topic)
	assert.Equal(t, "bbbb00-st-b", *sv.Topic)

	assert.Equal(t, "https://relay-access.practable.io/session/bbbb00-st-a", *sd.URL)
	assert.Equal(t, "https://relay-access.practable.io/session/bbbb00-st-b", *sv.URL)

	uim := make(map[string]models.UIDescribed)

	for _, ui := range a.Uis {
		uim[*ui.Description.Name] = *ui
	}

	uia := uim["ui-a"]
	uib := uim["ui-b"]

	assert.Equal(t, "a", uia.Description.Short)
	assert.Equal(t, "b", uib.Description.Short)
	assert.Equal(t, "ui", *uia.Description.Type)
	assert.Equal(t, "ui", *uib.Description.Type)

	sre := make(map[string]bool)
	sre["st-a"] = true
	sre["st-b"] = true

	sr := make(map[string]bool)

	for _, r := range uia.StreamsRequired {
		sr[r] = true
	}

	assert.Equal(t, sre, sr)

	sr = make(map[string]bool)

	for _, r := range uib.StreamsRequired {
		sr[r] = true
	}

	assert.Equal(t, sre, sr)

	//	Example body string from new system {"description":{"name":"slot-b","short":"b","type":"slot"},"exp":1667611200,"nbf":1667610900,"streams":[{"audience":"https://relay-access.practable.io","connection_type":"session","for":"video","scopes":["read"],"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJib29raW5nX2lkIjoiYmstNiIsInRvcGljIjoiYmJiYjAwLXN0LWIiLCJwcmVmaXgiOiJzZXNzaW9uIiwic2NvcGVzIjpbInJlYWQiXSwic3ViIjoidXNlci1nIiwiYXVkIjpbImh0dHBzOi8vcmVsYXktYWNjZXNzLnByYWN0YWJsZS5pbyJdLCJleHAiOjE2Njc2MTEyMDAsIm5iZiI6MTY2NzYxMDkwMCwiaWF0IjoxNjY3NjEwOTAwfQ.uu76zhbEw0ycSuUMEYkgeeADev2GTR-NNW3O2ulx6ZQ","topic":"bbbb00-st-b","url":"https://relay-access.practable.io/session/bbbb00-st-b"},{"audience":"https://relay-access.practable.io","connection_type":"session","for":"data","scopes":["read","write"],"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJib29raW5nX2lkIjoiYmstNiIsInRvcGljIjoiYmJiYjAwLXN0LWEiLCJwcmVmaXgiOiJzZXNzaW9uIiwic2NvcGVzIjpbInJlYWQiLCJ3cml0ZSJdLCJzdWIiOiJ1c2VyLWciLCJhdWQiOlsiaHR0cHM6Ly9yZWxheS1hY2Nlc3MucHJhY3RhYmxlLmlvIl0sImV4cCI6MTY2NzYxMTIwMCwibmJmIjoxNjY3NjEwOTAwLCJpYXQiOjE2Njc2MTA5MDB9.Y_6UhVu1roW-rIKPlLce7qNUHek6dQ0WXwO4boxFyFQ","topic":"bbbb00-st-a","url":"https://relay-access.practable.io/session/bbbb00-st-a"}],"uis":[{"description":{"name":"ui-a","short":"a","type":"ui"},"streams_required":["st-a","st-b"],"url":"a"},{"description":{"name":"ui-b","short":"b","type":"ui"},"streams_required":["st-a","st-b"],"url":"b"}]

}

/* Example URL from booking system v0.2.2
https://static.practable.io/ui/penduino-1.0/?streams=%5B%7B%22for%22%3A%22video%22%2C%22permission%22%3A%7B%22audience%22%3A%22https%3A%2F%2Frelay-access.practable.io%22%2C%22connection_type%22%3A%22session%22%2C%22scopes%22%3A%5B%22read%22%5D%2C%22topic%22%3A%22pend13-video%22%7D%2C%22token%22%3A%22eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b3BpYyI6InBlbmQxMy12aWRlbyIsInByZWZpeCI6InNlc3Npb24iLCJzY29wZXMiOlsicmVhZCJdLCJhdWQiOlsiaHR0cHM6Ly9yZWxheS1hY2Nlc3MucHJhY3RhYmxlLmlvIl0sImV4cCI6MTY3NDE2Njc3MiwibmJmIjoxNjc0MTY2NDcxLCJpYXQiOjE2NzQxNjY0NzF9.0VFqicdsTobjNnYwg8wkETmbR0YhC8Mw4lfss4iqF1c%22%2C%22url%22%3A%22https%3A%2F%2Frelay-access.practable.io%2Fsession%2Fpend13-video%22%2C%22verb%22%3A%22POST%22%7D%2C%7B%22for%22%3A%22data%22%2C%22permission%22%3A%7B%22audience%22%3A%22https%3A%2F%2Frelay-access.practable.io%22%2C%22connection_type%22%3A%22session%22%2C%22scopes%22%3A%5B%22read%22%2C%22write%22%5D%2C%22topic%22%3A%22pend13-data%22%7D%2C%22token%22%3A%22eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b3BpYyI6InBlbmQxMy1kYXRhIiwicHJlZml4Ijoic2Vzc2lvbiIsInNjb3BlcyI6WyJyZWFkIiwid3JpdGUiXSwiYXVkIjpbImh0dHBzOi8vcmVsYXktYWNjZXNzLnByYWN0YWJsZS5pbyJdLCJleHAiOjE2NzQxNjY3NzIsIm5iZiI6MTY3NDE2NjQ3MSwiaWF0IjoxNjc0MTY2NDcxfQ.HQb-E6HpZheFwN9b_oD6nJmWktKZx5PfFdVBX9BuJfQ%22%2C%22url%22%3A%22https%3A%2F%2Frelay-access.practable.io%2Fsession%2Fpend13-data%22%2C%22verb%22%3A%22POST%22%7D%5D&exp=1674166772

One of the stream tokens is:

eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b3BpYyI6InBlbmQxMy12aWRlbyIsInByZWZpeCI6InNlc3Npb24iLCJzY29wZXMiOlsicmVhZCJdLCJhdWQiOlsiaHR0cHM6Ly9yZWxheS1hY2Nlc3MucHJhY3RhYmxlLmlvIl0sImV4cCI6MTY3NDE2Njc3MiwibmJmIjoxNjc0MTY2NDcxLCJpYXQiOjE2NzQxNjY0NzF9.0VFqicdsTobjNnYwg8wkETmbR0YhC8Mw4lfss4iqF1c

It decodes as:

  "alg": "HS256",
  "typ": "JWT"
}
{
  "topic": "pend13-video",
  "prefix": "session",
  "scopes": [
    "read"
  ],
  "aud": [
    "https://relay-access.practable.io"
  ],
  "exp": 1674166772,
  "nbf": 1674166471,
  "iat": 1674166471
}
exp: Thu 19 Jan 22:19:32 GMT 2023
iat: Thu 19 Jan 22:14:31 GMT 2023
nbf: Thu 19 Jan 22:14:31 GMT 2023


Note the lower case topic, prefix, scopes

*/

func TestAddGetPoliciesAndStatus(t *testing.T) {

	// make sure our pre-prepared bookings are in the future
	// other tests may have advanced time
	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	setNow(s, ct)

	satoken := loadTestManifest(t)
	removeAllBookings(t)
	bm := getBookings(t)
	assert.Equal(t, 0, len(bm))

	// load some bookings to break up the future availability in discrete intervals
	client := &http.Client{}
	bodyReader := bytes.NewReader(bookings2JSON)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", satoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!

	sutoken, err := signedUserTokenFor("user-g")
	assert.NoError(t, err)

	// get policy status for p-b for user u-g
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/users/user-g/policies/p-b", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	status := `{"current_bookings":1,"old_bookings":0,"usage":"5m0s"}` + "\n"
	assert.Equal(t, status, string(body))

}

// TestRestrictedToAdmin checks that users cannot access admin endpoints
func TestRestrictedToAdmin(t *testing.T) {

	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	setNow(s, ct)

	satoken := loadTestManifest(t)
	authAdmin := httptransport.APIKeyAuth("Authorization", "header", satoken)
	timeout := 1 * time.Second

	sutoken, err := signedUserToken()
	assert.NoError(t, err)
	authUser := httptransport.APIKeyAuth("Authorization", "header", sutoken)

	var bookings cmodels.Bookings

	err = yaml.Unmarshal(bookingsYAML, &bookings)
	assert.NoError(t, err)

	unlocked := func() {
		loadTestManifest(t)
		removeAllBookings(t)
		setLock(t, false, "open")
		ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
		setNow(s, ct)
	}

	setLock := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		lock := true
		message := "locked for test"
		p := admin.NewSetLockParams().WithTimeout(timeout).WithLock(lock).WithMsg(&message)
		return bc.Admin.SetLock(p, auth)
	}
	exportBookings := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := admin.NewExportBookingsParams().WithTimeout(timeout)
		return bc.Admin.ExportBookings(p, auth)
	}
	exportManifest := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := admin.NewExportManifestParams().WithTimeout(timeout)
		return bc.Admin.ExportManifest(p, auth)
	}
	exportOldBookings := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := admin.NewExportOldBookingsParams().WithTimeout(timeout)
		return bc.Admin.ExportOldBookings(p, auth)
	}

	exportUsers := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := admin.NewExportUsersParams().WithTimeout(timeout)
		return bc.Admin.ExportUsers(p, auth)
	}

	getSlotIsAvailable := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := admin.NewGetSlotIsAvailableParams().WithTimeout(timeout).WithSlotName("sl-a")
		return bc.Admin.GetSlotIsAvailable(p, auth)
	}

	replaceBookings := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := admin.NewReplaceBookingsParams().WithTimeout(timeout).WithBookings(bookings)
		return bc.Admin.ReplaceBookings(p, auth)
	}
	replaceManifest := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		var manifest cmodels.Manifest
		err := json.Unmarshal(manifestJSON, &manifest)
		assert.NoError(t, err)

		p := admin.NewReplaceManifestParams().WithTimeout(timeout).WithManifest(&manifest)
		return bc.Admin.ReplaceManifest(p, auth)
	}
	replaceOldBookings := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := admin.NewReplaceOldBookingsParams().WithTimeout(timeout).WithBookings(bookings)
		return bc.Admin.ReplaceOldBookings(p, auth)
	}

	setSlotIsAvailable := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := admin.NewSetSlotIsAvailableParams().WithTimeout(timeout).WithSlotName("sl-a").WithAvailable(true).WithReason("test")
		return bc.Admin.SetSlotIsAvailable(p, auth)
	}

	tests := map[string]struct {
		setup   func()
		command func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error)
		auth    rt.ClientAuthInfoWriter
		ok      bool
		want    string
	}{
		"exportBookingsAdmin":     {unlocked, exportBookings, authAdmin, true, `[GET /admin/bookings][200] exportBookingsOK`},
		"exportBookingsUser":      {unlocked, exportBookings, authUser, false, `[GET /admin/bookings][401] exportBookingsUnauthorized`},
		"exportManifestAdmin":     {unlocked, exportManifest, authAdmin, true, `[GET /admin/manifest][200] exportManifestOK`},
		"exportManifestUser":      {unlocked, exportManifest, authUser, false, `[GET /admin/manifest][401] exportManifestUnauthorized`},
		"exportOldBookingsAdmin":  {unlocked, exportOldBookings, authAdmin, true, `[GET /admin/oldbookings][200] exportOldBookingsOK`},
		"exportOldBookingsUser":   {unlocked, exportOldBookings, authUser, false, `[GET /admin/oldbookings][401] exportOldBookingsUnauthorized`},
		"exportUsersUser":         {unlocked, exportUsers, authUser, false, `[GET /admin/users][401] exportUsersUnauthorized`},
		"exportUsersAdmin":        {unlocked, exportUsers, authAdmin, true, `[GET /admin/users][200] exportUsersOK`},
		"getSlotIsAvailableUser":  {unlocked, getSlotIsAvailable, authUser, false, `[GET /admin/slots/{slot_name}][401] getSlotIsAvailableUnauthorized`},
		"getSlotIsAvailableAdmin": {unlocked, getSlotIsAvailable, authAdmin, true, `[GET /admin/slots/{slot_name}][200] getSlotIsAvailableOK`},
		"replaceBookingsAdmin":    {unlocked, replaceBookings, authAdmin, true, `[PUT /admin/bookings][200] replaceBookingsOK`},
		"replaceBookingsUser":     {unlocked, replaceBookings, authUser, false, `[PUT /admin/bookings][401] replaceBookingsUnauthorized`},
		"replaceManifestAdmin":    {unlocked, replaceManifest, authAdmin, true, `[PUT /admin/manifest][200] replaceManifestOK`},
		"replaceManifestUser":     {unlocked, replaceManifest, authUser, false, `[PUT /admin/manifest][401] replaceManifestUnauthorized`},
		"replaceOldBookingsAdmin": {unlocked, replaceOldBookings, authAdmin, true, `[PUT /admin/oldbookings][200] replaceOldBookingsOK`},
		"replaceOldBookingsUser":  {unlocked, replaceOldBookings, authUser, false, `[PUT /admin/oldbookings][401] replaceOldBookingsUnauthorized`},
		"setLockAdmin":            {unlocked, setLock, authAdmin, true, `[PUT /admin/status][200] setLockOK`},
		"setLockUser":             {unlocked, setLock, authUser, false, `[PUT /admin/status][401] setLockUnauthorized`},
		"setSlotIsAvailableUser":  {unlocked, setSlotIsAvailable, authUser, false, `[PUT /admin/slots/{slot_name}][401] setSlotIsAvailableUnauthorized`},
		"setSlotIsAvailableAdmin": {unlocked, setSlotIsAvailable, authAdmin, true, `[PUT /admin/slots/{slot_name}][204] setSlotIsAvailableNoContent`},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			tc.setup()

			c := apiclient.DefaultTransportConfig().WithHost(ch).WithSchemes([]string{cs})
			bc := apiclient.NewHTTPClientWithConfig(nil, c)
			got, err := tc.command(bc, tc.auth)
			if debug {
				gots := fmt.Sprintf("%+v", got)
				fmt.Println(gots)
			}
			if len(tc.want) == 0 {
				t.Error("test should check against non-zero length string")
			}
			var s string
			if tc.ok {
				s = fmt.Sprintf("%+v\n", got)
			} else {
				s = fmt.Sprintf("%+v\n", err)
			}
			if len(tc.want) > len(s) {
				t.Error("output too short")
				t.Log(fmt.Sprintf("%+v // %+v\n", got, err))

			} else {
				// don't check this if already an error, throws out of range error
				if tc.want != s[:len(tc.want)] {
					t.Error("Unexpected response")
					t.Log("want: " + tc.want)
					t.Log("got:  " + s)
					fmt.Printf("%+v %+v", got, err)
				}
			}
		})
	}

}

// TestLockedToUser checks that the lock prevents user access to routes but does not block admin
func TestLockedToUser(t *testing.T) {

	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	setNow(s, ct)

	satoken := loadTestManifest(t)
	authAdmin := httptransport.APIKeyAuth("Authorization", "header", satoken)
	timeout := 1 * time.Second

	sutoken, err := signedUserTokenFor("user-a") //to match bookings2JSON
	assert.NoError(t, err)
	authUser := httptransport.APIKeyAuth("Authorization", "header", sutoken)

	var bookings cmodels.Bookings

	err = yaml.Unmarshal(bookingsYAML, &bookings)
	assert.NoError(t, err)

	unlocked := func() {
		ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) //set time before adding bookings
		setNow(s, ct)
		loadTestManifest(t)
		removeAllBookings(t) //else tests fail if run many times
		addBookings(t)
		setLock(t, false, "unlocked")
	}
	locked := func() {
		ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC) //set time before adding bookings
		setNow(s, ct)
		loadTestManifest(t)
		removeAllBookings(t) //else tests fail if run many times
		addBookings(t)
		setLock(t, true, "locked")

	}

	cancelBooking := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := users.NewCancelBookingParams().
			WithTimeout(timeout).
			WithBookingName("bk-0").
			WithUserName("user-a") //get a json error about not unmarshalling number into string in models.Error if omit this
		return nil, bc.Users.CancelBooking(p, auth) //this method only returns error codes
	}
	getActivity := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {

		ct := time.Date(2022, 11, 5, 0, 10, 1, 0, time.UTC)
		setNow(s, ct)
		p := users.NewGetActivityParams().
			WithTimeout(timeout).
			WithBookingName("bk-0").
			WithUserName("user-a")
		return bc.Users.GetActivity(p, auth)
	}
	getAvailability := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := users.NewGetAvailabilityParams().WithTimeout(timeout).WithSlotName("sl-a")
		return bc.Users.GetAvailability(p, auth)
	}

	getBookingsForUser := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {

		if debug {
			p3 := admin.NewExportBookingsParams().WithTimeout(timeout)
			status3, err := bc.Admin.ExportBookings(p3, authAdmin)
			assert.NoError(t, err)
			if err == nil {
				fmt.Printf("BOOKINGS: %+v\n", status3.Payload)
			}
			p2 := admin.NewExportUsersParams().WithTimeout(timeout)
			status2, _ := bc.Admin.ExportUsers(p2, authAdmin)

			fmt.Printf("USERS: %+v\n", status2.Payload)

			// export users (now there are bookings we will have users)
			client := &http.Client{}
			req, err := http.NewRequest("GET", cfg.Host+"/api/v1/admin/users", nil)
			assert.NoError(t, err)
			req.Header.Add("Authorization", satoken)
			resp, err := client.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode) //should be ok!
			body, err := ioutil.ReadAll(resp.Body)
			fmt.Println(string(body))
			resp.Body.Close()
		}

		p := users.NewGetBookingsForUserParams().WithTimeout(timeout).WithUserName("user-a")
		return bc.Users.GetBookingsForUser(p, auth)

	}

	getDescription := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := users.NewGetDescriptionParams().WithTimeout(timeout).WithDescriptionName("d-r-a")
		return bc.Users.GetDescription(p, auth)
	}

	getOldBookingsForUser := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {

		useClient := true

		if !useClient { // for debug purposes (check server side is ok)
			client := &http.Client{}
			req, err := http.NewRequest("GET", cfg.Host+"/api/v1/users/user-a/oldbookings", nil)
			assert.NoError(t, err)
			req.Header.Add("Authorization", sutoken)
			// add query params
			q := req.URL.Query()
			q.Add("user_name", "user-a")

			req.URL.RawQuery = q.Encode()
			resp, err := client.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode) //should be ok!
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			return string(body), err

		} else {
			//ct := time.Date(2022, 11, 7, 0, 0, 0, 0, time.UTC)
			//setNow(s, ct)
			// seems to be a race condition - setting this to the future breaks GetActivity test
			p := users.NewGetOldBookingsForUserParams().WithTimeout(timeout).WithUserName("user-a")
			return bc.Users.GetOldBookingsForUser(p, auth)
		}

	}

	getGroupsForUser := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {

		g := users.NewGetGroupsForUserParams().WithTimeout(timeout).WithUserName("user-a")
		return bc.Users.GetGroupsForUser(g, auth)

	}

	getPolicy := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := users.NewGetPolicyParams().WithTimeout(timeout).WithPolicyName("p-a")
		return bc.Users.GetPolicy(p, auth)
	}
	getPolicyStatusForUser := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := users.NewGetPolicyStatusForUserParams().WithTimeout(timeout).WithPolicyName("p-a").WithUserName("user-a")
		return bc.Users.GetPolicyStatusForUser(p, auth)
	}
	addGroupForUser := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := users.NewAddGroupForUserParams().WithTimeout(timeout).WithGroupName("g-a").WithUserName("user-a")
		return bc.Users.AddGroupForUser(p, auth)
	}
	getStoreStatusUser := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := users.NewGetStoreStatusUserParams().WithTimeout(timeout)
		return bc.Users.GetStoreStatusUser(p, auth)
	}

	makeBooking := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := users.NewMakeBookingParams().
			WithTimeout(timeout).
			WithSlotName("sl-b").
			WithUserName("someuser").
			WithFrom(strfmt.DateTime(time.Date(2022, 11, 5, 1, 0, 0, 0, time.UTC))).
			WithTo(strfmt.DateTime(time.Date(2022, 11, 5, 1, 5, 0, 0, time.UTC)))
		return bc.Users.MakeBooking(p, auth)
	}

	tests := map[string]struct {
		setup   func()
		command func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error)
		auth    rt.ClientAuthInfoWriter
		ok      bool
		want    string
	}{
		"GetDescriptionLockedAdminAllowed":           {locked, getDescription, authAdmin, true, `[GET /descriptions/{description_name}][200] getDescriptionOK`},
		"GetDescriptionLockedUserDenied":             {locked, getDescription, authUser, false, `[GET /descriptions/{description_name}][401] getDescriptionUnauthorized`},
		"GetDescriptionUnlockedAdminAllowed":         {unlocked, getDescription, authAdmin, true, `[GET /descriptions/{description_name}][200] getDescriptionOK`},
		"GetDescriptionUnlockedUserAllowed":          {unlocked, getDescription, authUser, true, `[GET /descriptions/{description_name}][200] getDescriptionOK`},
		"GetPolicyLockedAdminAllowed":                {locked, getPolicy, authAdmin, true, `[GET /policies/{policy_name}][200] getPolicyOK`},
		"GetPolicyLockedUserDenied":                  {locked, getPolicy, authUser, false, `[GET /policies/{policy_name}][401] getPolicyUnauthorized`},
		"GetPolicyUnlockedAdminAllowed":              {unlocked, getPolicy, authAdmin, true, `[GET /policies/{policy_name}][200] getPolicyOK`},
		"GetPolicyUnlockedUserAllowed":               {unlocked, getPolicy, authUser, true, `[GET /policies/{policy_name}][200] getPolicyOK`},
		"GetAvailabilityLockedAdminAllowed":          {locked, getAvailability, authAdmin, true, `[GET /slots/{slot_name}][200] getAvailabilityOK`},
		"GetAvailabilityLockedUserDenied":            {locked, getAvailability, authUser, false, `[GET /slots/{slot_name}][401] getAvailabilityUnauthorized`},
		"GetAvailabilityUnlockedAdminAllowed":        {unlocked, getAvailability, authAdmin, true, `[GET /slots/{slot_name}][200] getAvailabilityOK`},
		"GetAvailabilityUnlockedUserAllowed":         {unlocked, getAvailability, authUser, true, `[GET /slots/{slot_name}][200] getAvailabilityOK`},
		"MakeBookingLockedAdminAllowed":              {locked, makeBooking, authAdmin, true, `[POST /slots/{slot_name}][204] makeBookingNoContent`},
		"MakeBookingLockedUserDenied":                {locked, makeBooking, authUser, false, `[POST /slots/{slot_name}][401] makeBookingUnauthorized`},
		"MakeBookingUnlockedAdminAllowed":            {unlocked, makeBooking, authAdmin, true, `[POST /slots/{slot_name}][204] makeBookingNoContent`},
		"MakeBookingUnlockedUserAllowed":             {unlocked, makeBooking, authUser, true, `[POST /slots/{slot_name}][204] makeBookingNoContent`},
		"GetStoreStatusUserLockedAdminAllowed":       {locked, getStoreStatusUser, authAdmin, true, `[GET /users/status][200] getStoreStatusUserOK`},
		"GetStoreStatusUserLockedUserAllowed":        {locked, getStoreStatusUser, authUser, true, `[GET /users/status][200] getStoreStatusUserOK`},
		"GetStoreStatusUserUnlockedAdminAllowed":     {unlocked, getStoreStatusUser, authAdmin, true, `[GET /users/status][200] getStoreStatusUserOK`},
		"GetStoreStatusUserUnlockedUserAllowed":      {unlocked, getStoreStatusUser, authUser, true, `[GET /users/status][200] getStoreStatusUserOK`},
		"GetBookingsForUserLockedAdminAllowed":       {locked, getBookingsForUser, authAdmin, true, `[GET /users/{user_name}/bookings][200] getBookingsForUserOK`},
		"GetBookingsForUserLockedUserDenied":         {locked, getBookingsForUser, authUser, false, `[GET /users/{user_name}/bookings][401] getBookingsForUserUnauthorized`},
		"GetBookingsForUserUnlockedAdminAllowed":     {unlocked, getBookingsForUser, authAdmin, true, `[GET /users/{user_name}/bookings][200] getBookingsForUserOK`},
		"GetBookingsForUserUnlockedUserAllowed":      {unlocked, getBookingsForUser, authUser, true, `[GET /users/{user_name}/bookings][200] getBookingsForUserOK`},
		"CancelBookingLockedAdminAllowed":            {locked, cancelBooking, authAdmin, false, `[DELETE /users/{user_name}/bookings/{booking_name}][404] cancelBookingNotFound`},
		"CancelBookingLockedUserDenied":              {locked, cancelBooking, authUser, false, `[DELETE /users/{user_name}/bookings/{booking_name}][401] cancelBookingUnauthorized`},
		"CancelBookingUnlockedAdminAllowed":          {unlocked, cancelBooking, authAdmin, false, `[DELETE /users/{user_name}/bookings/{booking_name}][404] cancelBookingNotFound`},
		"CancelBookingUnlockedUserAllowed":           {unlocked, cancelBooking, authUser, false, `[DELETE /users/{user_name}/bookings/{booking_name}][404] cancelBookingNotFound`},
		"GetActivityLockedAdminAllowed":              {locked, getActivity, authAdmin, true, `[PUT /users/{user_name}/bookings/{booking_name}][200] getActivityOK`},
		"GetActivityLockedUserDenied":                {locked, getActivity, authUser, false, `[PUT /users/{user_name}/bookings/{booking_name}][401] getActivityUnauthorized`},
		"GetActivityUnlockedAdminAllowed":            {unlocked, getActivity, authAdmin, true, `[PUT /users/{user_name}/bookings/{booking_name}][200] getActivityOK`},
		"GetActivityUnlockedUserAllowed":             {unlocked, getActivity, authUser, true, `[PUT /users/{user_name}/bookings/{booking_name}][200] getActivityOK`},
		"GetOldBookingsForUserLockedAdminAllowed":    {locked, getOldBookingsForUser, authAdmin, true, `[GET /users/{user_name}/oldbookings][200] getOldBookingsForUserOK`},
		"GetOldBookingsForUserLockedUserDenied":      {locked, getOldBookingsForUser, authUser, false, `[GET /users/{user_name}/oldbookings][401] getOldBookingsForUserUnauthorized`},
		"GetOldBookingsForUserUnlockedAdminAllowed":  {unlocked, getOldBookingsForUser, authAdmin, true, `[GET /users/{user_name}/oldbookings][200] getOldBookingsForUserOK`},
		"GetOldBookingsForUserUnlockedUserAllowed":   {unlocked, getOldBookingsForUser, authUser, true, `[GET /users/{user_name}/oldbookings][200] getOldBookingsForUserOK`},
		"GetGroupsForUserLockedAdminAllowed":         {locked, getGroupsForUser, authAdmin, true, `[GET /users/{user_name}/groups][200] getGroupsForUserOK`},
		"GetGroupsForUserLockedUserDenied":           {locked, getGroupsForUser, authUser, false, `[GET /users/{user_name}/groups][401] getGroupsForUserUnauthorized`},
		"GetGroupsForUserUnlockedAdminAllowed":       {unlocked, getGroupsForUser, authAdmin, true, `[GET /users/{user_name}/groups][200] getGroupsForUserOK`},
		"GetGroupsForUserUnlockedUserAllowed":        {unlocked, getGroupsForUser, authUser, true, `[GET /users/{user_name}/groups][200] getGroupsForUserOK`},
		"GetPolicyStatusForUserLockedAdminAllowed":   {locked, getPolicyStatusForUser, authAdmin, true, `[GET /users/{user_name}/policies/{policy_name}][200] getPolicyStatusForUserOK`},
		"GetPolicyStatusForUserLockedUserDenied":     {locked, getPolicyStatusForUser, authUser, false, `[GET /users/{user_name}/policies/{policy_name}][401] getPolicyStatusForUserUnauthorized`},
		"GetPolicyStatusForUserUnlockedAdminAllowed": {unlocked, getPolicyStatusForUser, authAdmin, true, `[GET /users/{user_name}/policies/{policy_name}][200] getPolicyStatusForUserOK`},
		"GetPolicyStatusForUserUnlockedUserAllowed":  {unlocked, getPolicyStatusForUser, authUser, true, `[GET /users/{user_name}/policies/{policy_name}][200] getPolicyStatusForUserOK`},
		"AddGroupForUserLockedAdminAllowed":          {locked, addGroupForUser, authAdmin, true, `[POST /users/{user_name}/groups/{group_name}][204] addGroupForUserNoContent`},
		"AddGroupForUserLockedUserDenied":            {locked, addGroupForUser, authUser, false, `[POST /users/{user_name}/groups/{group_name}][401] addGroupForUserUnauthorized`},
		"AddGroupForUserUnlockedAdminAllowed":        {unlocked, addGroupForUser, authAdmin, true, `[POST /users/{user_name}/groups/{group_name}][204] addGroupForUserNoContent`},
		"AddGroupForUserUnlockedUserAllowed":         {unlocked, addGroupForUser, authUser, true, `[POST /users/{user_name}/groups/{group_name}][204] addGroupForUserNoContent`},
	}

	//enforce sequential runs of these tests (note, did not remove issue with getActivity sometimes failing to get it's booking)
	mm := &sync.Mutex{}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mm.Lock()
			defer mm.Unlock()

			tc.setup()

			c := apiclient.DefaultTransportConfig().WithHost(ch).WithSchemes([]string{cs})
			bc := apiclient.NewHTTPClientWithConfig(nil, c)
			got, err := tc.command(bc, tc.auth)
			if debug {
				gots := fmt.Sprintf("%+v", got)
				fmt.Println(gots)
			}
			if len(tc.want) == 0 {
				t.Error("test should check against non-zero length string")
			}
			var s string
			if tc.ok {
				s = fmt.Sprintf("%+v\n", got)
			} else {
				s = fmt.Sprintf("%+v\n", err)
			}
			if len(tc.want) > len(s) {
				t.Error("output too short")
				t.Log(fmt.Sprintf("%+v // %+v\n", got, err))

			} else {
				// don't check this if already an error, throws out of range error
				if tc.want != s[:len(tc.want)] {
					t.Error("Unexpected response")
					t.Log("want: " + tc.want)
					t.Log("got:  " + s)
					fmt.Printf("%+v %+v", got, err)

				}
			}
		})
	}

}

func TestAutoCancellation(t *testing.T) {
	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	setNow(s, ct)

	satoken, err := signedAdminToken()
	assert.NoError(t, err)
	client := &http.Client{}
	bodyReader := bytes.NewReader(manifestGraceJSON)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/manifest", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", satoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	resp.Body.Close()

	removeAllBookings(t)

	b := getBookings(t)
	printBookings(t, b)

	assert.NoError(t, err)
	client = &http.Client{}
	bodyReader = bytes.NewReader(bookingsGraceJSON)
	req, err = http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", satoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	resp.Body.Close()
	ct = time.Date(2022, 11, 5, 0, 4, 0, 0, time.UTC)
	setNow(s, ct)

	time.Sleep(50 * time.Millisecond) //let the GracePeriod Check expire

	b = getBookings(t)
	if debug {
		printBookings(t, b)
	}
	ob := getOldBookings(t)
	if debug {
		printBookings(t, ob)
	}
	obn := make(map[string]bool)

	for _, obk := range ob {
		obn[*obk.Name] = true
		assert.True(t, obk.Cancelled) //only cancelled booking can be an old booking here)

	}

	bn := make(map[string]bool)

	for _, bk := range b {
		bn[*bk.Name] = true

	}

	eob := map[string]bool{"bk-0": true}

	assert.Equal(t, eob, obn)

	eb := map[string]bool{"bk-1": true, "bk-2": true}
	assert.Equal(t, eb, bn)

}

func TestUniqueName(t *testing.T) {

	// test does not depend on store state

	// get first username

	client := &http.Client{}
	req, err := http.NewRequest("POST", cfg.Host+"/api/v1/users/unique", nil)
	assert.NoError(t, err)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if debug {
		t.Log(string(body))
	}

	var u0 models.UserName
	err = json.Unmarshal(body, &u0)
	assert.NoError(t, err)

	// get second username
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if debug {
		t.Log(string(body))
	}

	var u1 models.UserName
	err = json.Unmarshal(body, &u1)
	assert.NoError(t, err)

	// check they are 20 chars long (as expected of xid) but different to each other

	assert.Equal(t, 20, len(u0.UserName))
	assert.Equal(t, 20, len(u1.UserName))

	assert.NotEqual(t, u0.UserName, u1.UserName)

	// Note that xids are quite similar to each other
	// different test runs a few seconds apart
	// cf6lmusbig7in9ndtht0
	// cf6ln7cbig7ivol56uo0
	// cf6lnfsbig7iom1croqg
	// cf6loacbig7jff43q4v0

	// Within the same test run:
	// server_test.go:2434: {"user_name":"cf6lnfsbig7iom1croq0"}
	// server_test.go:2447: {"user_name":"cf6lnfsbig7iom1croqg"}
	// And again later:
	// server_test.go:2434: {"user_name":"cf6loacbig7jff43q4v0"}
	// server_test.go:2447: {"user_name":"cf6loacbig7jff43q4vg"}

}

func TestGroups(t *testing.T) {

	// test DOES depend on store state (whether user present)
	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	setNow(s, ct)

	_ = loadTestManifest(t)
	removeAllBookings(t)
	setLock(t, false, "unlocked")
	sutoken, err := signedUserTokenFor("user-a") //to match bookings2JSON
	assert.NoError(t, err)

	// getGroups
	client := &http.Client{}
	req, err := http.NewRequest("GET", cfg.Host+"/api/v1/users/user-a/groups", nil)
	req.Header.Add("Authorization", sutoken)
	assert.NoError(t, err)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode) //should be "not found"
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if debug {
		t.Log(string(body))
	}

	// AddGroupForUser
	client = &http.Client{}
	req, err = http.NewRequest("POST", cfg.Host+"/api/v1/users/user-a/groups/g-a", nil)
	req.Header.Add("Authorization", sutoken)
	assert.NoError(t, err)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode) //should be OK No Content
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if debug {
		t.Log(string(body))
	}

	// get groups again
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/users/user-a/groups", nil)
	req.Header.Add("Authorization", sutoken)
	assert.NoError(t, err)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be OK
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if debug {
		t.Log(string(body))
	}
	assert.Equal(t, `{"g-a":{"description":{"name":"group-a","short":"a","type":"group"}}}`+"\n", string(body))

	// Get group details
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/groups/g-a", nil)
	req.Header.Add("Authorization", sutoken)
	assert.NoError(t, err)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be OK
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if debug {
		t.Log(string(body))
	}
	assert.Equal(t, `{"description":{"name":"group-a","short":"a","type":"group"},"policies":{"p-a":{"allow_start_in_past_within":"0s","book_ahead":"1h0m0s","description":{"name":"policy-a","short":"a","type":"policy"},"display_guides":{"1mFor20m":{"book_ahead":"20m0s","duration":"1m0s","label":"1m","max_slots":15}},"enforce_book_ahead":true,"max_duration":"0s","max_usage":"0s","min_duration":"0s","next_available":"0s","slots":{"sl-a":{"description":{"name":"slot-a","short":"a","type":"slot"},"policy":"p-a"}},"starts_within":"0s"}}}`+"\n", string(body))

	// Add nonexistent group - needs to return a 404 not 500
	client = &http.Client{}
	req, err = http.NewRequest("POST", cfg.Host+"/api/v1/users/user-a/groups/neverheardofit", nil)
	req.Header.Add("Authorization", sutoken)
	assert.NoError(t, err)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode) //should be OK No Content
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if debug {
		t.Log(string(body))
	}

	// TODO test DeleteGroup

}

func TestRebookCancelledSlot(t *testing.T) {

	// Must be able to make a new booking that cuts across two cancelled bookings,
	// one cancelled by a grace period and the other manually

	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	setNow(s, ct)

	satoken, err := signedAdminToken()
	assert.NoError(t, err)
	client := &http.Client{}
	bodyReader := bytes.NewReader(manifestGraceJSON)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/manifest", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", satoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	if debug {
		t.Log(string(body))
	}

	assert.Equal(t, 200, resp.StatusCode)

	resp.Body.Close()

	removeAllBookings(t)

	b := getBookings(t)
	printBookings(t, b)

	assert.NoError(t, err)
	client = &http.Client{}
	bodyReader = bytes.NewReader(bookingsGraceJSON)
	req, err = http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", satoken)
	req.Header.Add("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	resp.Body.Close()
	ct = time.Date(2022, 11, 5, 0, 4, 0, 0, time.UTC)
	setNow(s, ct)

	time.Sleep(50 * time.Millisecond) //let the GracePeriod Check expire

	b = getBookings(t)
	if debug {
		printBookings(t, b)
	}
	ob := getOldBookings(t)
	if debug {
		printBookings(t, ob)
	}
	obn := make(map[string]bool)

	for _, obk := range ob {
		obn[*obk.Name] = true
		assert.True(t, obk.Cancelled) //only cancelled booking can be an old booking here)

	}

	bn := make(map[string]bool)

	for _, bk := range b {
		bn[*bk.Name] = true

	}

	eob := map[string]bool{"bk-0": true} //the booking expired due to not starting within grace period

	assert.Equal(t, eob, obn)

	eb := map[string]bool{"bk-1": true, "bk-2": true}
	assert.Equal(t, eb, bn)

	// cancel the second booking on this resource (we'll book across time taken by two slots previously)
	sutoken, err := signedUserTokenFor("user-b") // different user to the one who cancelled (but still a known user)
	assert.NoError(t, err)

	client = &http.Client{}
	req, err = http.NewRequest("DELETE", cfg.Host+"/api/v1/users/user-b/bookings/bk-1", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode) //NotFound if successful deletion

	sutoken, err = signedUserTokenFor("anotheruser") // different user to the ones who are cancelled (but still a known user)
	assert.NoError(t, err)

	// AddGroupForUser
	client = &http.Client{}
	req, err = http.NewRequest("POST", cfg.Host+"/api/v1/users/anotheruser/groups/g-a", nil)
	req.Header.Add("Authorization", sutoken)
	assert.NoError(t, err)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode) //should be OK No Content
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if debug {
		t.Log(string(body))
	}

	// Main part of the test - now rebook the slot!
	client = &http.Client{}
	req, err = http.NewRequest("POST", cfg.Host+"/api/v1/slots/sl-a", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	// add query params
	q := req.URL.Query()
	q.Add("user_name", "anotheruser")    //must match token
	q.Add("from", "2022-11-05T00:04:05") //overlaps with cancelled booking by just under a minute
	q.Add("to", "2022-11-05T00:09:05Z")  //5min duration
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode) //should be ok!
	body, err = ioutil.ReadAll(resp.Body)
	if debug {
		t.Log(string(body))
	}
	resp.Body.Close()

}

func TestGetResources(t *testing.T) {

	stoken := loadTestManifest(t)

	// make unavailable slot sl-a
	client := &http.Client{}
	req, err := http.NewRequest("GET", cfg.Host+"/api/v1/admin/resources", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	expected := `{"r-a":{"description":"d-r-a","streams":["st-a","st-b"],"tests":null,"topic_stub":"aaaa00"},"r-b":{"description":"d-r-b","streams":["st-a","st-b"],"tests":null,"topic_stub":"bbbb00"}}` + "\n"
	assert.Equal(t, expected, string(body))

	resp.Body.Close()

}
