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
	"testing"
	"time"

	rt "github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/phayes/freeport"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	apiclient "github.com/timdrysdale/interval/internal/client/client"
	"github.com/timdrysdale/interval/internal/client/client/admin"
	"github.com/timdrysdale/interval/internal/client/client/users"
	cmodels "github.com/timdrysdale/interval/internal/client/models"
	"github.com/timdrysdale/interval/internal/config"
	"github.com/timdrysdale/interval/internal/login"
	"github.com/timdrysdale/interval/internal/serve/models"
	"github.com/timdrysdale/interval/internal/store"
	"gopkg.in/yaml.v2"
)

var debug bool
var cfg config.ServerConfig
var currentTime *time.Time
var cs, ch string //client scheme and host
var timeout time.Duration
var aa, ua rt.ClientAuthInfoWriter

// Are thinking about making a models.Manifest object
// to compare responses to? Don't. Tried it.
// Durations don't get populated properly when
// you unmarshal into models.Manifest, so not particularly
// useful for comparing to responses. Better just to use
// strings, and may as well be consistent.
var manifestYAML = []byte(`descriptions:
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

var manifestJSON = []byte(`{"descriptions":{"d-p-a":{"name":"policy-a","type":"policy","short":"a"},"d-p-b":{"name":"policy-b","type":"policy","short":"b"},"d-r-a":{"name":"resource-a","type":"resource","short":"a"},"d-r-b":{"name":"resource-b","type":"resource","short":"b"},"d-sl-a":{"name":"slot-a","type":"slot","short":"a"},"d-sl-b":{"name":"slot-b","type":"slot","short":"b"},"d-ui-a":{"name":"ui-a","type":"ui","short":"a"},"d-ui-b":{"name":"ui-b","type":"ui","short":"b"}},"display_guides":{"1mFor20m":{"book_ahead":"20m","duration":"1m","max_slots":15,"label":"1m"}},"policies":{"p-a":{"book_ahead":"1h","description":"d-p-a","display_guides":["1mFor20m"],"enforce_book_ahead":true,"enforce_max_bookings":false,"enforce_max_duration":false,"enforce_min_duration":false,"enforce_max_usage":false,"max_bookings":0,"max_duration":"0s","min_duration":"0s","max_usage":"0s","slots":["sl-a"]},"p-b":{"book_ahead":"2h0m0s","description":"d-p-b","enforce_book_ahead":true,"enforce_max_bookings":true,"enforce_max_duration":true,"enforce_min_duration":true,"enforce_max_usage":true,"max_bookings":2,"max_duration":"10m0s","min_duration":"5m0s","max_usage":"30m0s","slots":["sl-b"]}},"resources":{"r-a":{"description":"d-r-a","streams":["st-a","st-b"],"topic_stub":"aaaa00"},"r-b":{"description":"d-r-b","streams":["st-a","st-b"],"topic_stub":"bbbb00"}},"slots":{"sl-a":{"description":"d-sl-a","policy":"p-a","resource":"r-a","ui_set":"us-a","window":"w-a"},"sl-b":{"description":"d-sl-b","policy":"p-b","resource":"r-b","ui_set":"us-b","window":"w-b"}},"streams":{"st-a":{"url":"https://relay-access.practable.io","connection_type":"session","for":"data","scopes":["read","write"],"topic":"tbc"},"st-b":{"url":"https://relay-access.practable.io","connection_type":"session","for":"video","scopes":["read"],"topic":"tbc"}},"uis":{"ui-a":{"description":"d-ui-a","url":"a","streams_required":["st-a","st-b"]},"ui-b":{"description":"d-ui-b","url":"b","streams_required":["st-a","st-b"]}},"ui_sets":{"us-a":{"uis":["ui-a"]},"us-b":{"uis":["ui-a","ui-b"]}},"windows":{"w-a":{"allowed":[{"start":"2022-11-04T00:00:00.000Z","end":"2022-11-06T00:00:00.000Z"}],"denied":[]},"w-b":{"allowed":[{"start":"2022-11-04T00:00:00.000Z","end":"2022-11-06T00:00:00.000Z"}],"denied":[]}}}`)

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

var bookingsJSON = []byte(`[{"name":"bk-0","cancelled":false,"policy":"p-a","slot":"sl-a","started":false,"unfulfilled":false,"user":"u-a","when":{"start":"2022-11-05T00:10:00Z","end":"2022-11-05T00:15:00Z"}},{"name":"bk-1","cancelled":false,"policy":"p-b","slot":"sl-b","started":false,"unfulfilled":false,"user":"u-b","when":{"start":"2022-11-05T00:20:00Z","end":"2022-11-05T00:30:00Z"}}]`)

var bookings2JSON = []byte(`[{"cancelled":false,"name":"bk-0","policy":"p-b","slot":"sl-b","started":false,"unfulfilled":false,"user":"user-a","when":{"start":"2022-11-05T00:10:00Z","end":"2022-11-05T00:15:00Z"}},{"cancelled":false,"name":"bk-1","policy":"p-b","slot":"sl-b","started":false,"unfulfilled":false,"user":"user-b","when":{"start":"2022-11-05T00:20:00Z","end":"2022-11-05T00:30:00Z"}},{"cancelled":false,"name":"bk-2","policy":"p-b","slot":"sl-b","started":false,"unfulfilled":false,"user":"user-c","when":{"start":"2022-11-05T00:35:00Z","end":"2022-11-05T00:40:00Z"}},{"cancelled":false,"name":"bk-3","policy":"p-b","slot":"sl-b","started":false,"unfulfilled":false,"user":"user-d","when":{"start":"2022-11-05T00:45:00Z","end":"2022-11-05T00:50:00Z"}},{"cancelled":false,"name":"bk-4","policy":"p-b","slot":"sl-b","started":false,"unfulfilled":false,"user":"user-e","when":{"start":"2022-11-05T00:55:00Z","end":"2022-11-05T01:00:00Z"}},{"cancelled":false,"name":"bk-5","policy":"p-b","slot":"sl-b","started":false,"unfulfilled":false,"user":"user-f","when":{"start":"2022-11-05T01:05:00Z","end":"2022-11-05T01:10:00Z"}},{"cancelled":false,"name":"bk-6","policy":"p-b","slot":"sl-b","started":false,"unfulfilled":false,"user":"user-g","when":{"start":"2022-11-05T01:15:00Z","end":"2022-11-05T01:20:00Z"}},{"cancelled":false,"name":"bk-7","policy":"p-b","slot":"sl-b","started":false,"unfulfilled":false,"user":"user-h","when":{"start":"2022-11-05T01:25:00Z","end":"2022-11-05T01:30:00Z"}}]`)

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
  policies:
  - p-a
  usage:
    p-a: 5m0s
u-b:
  bookings:
  - bk-1
  old_bookings: []
  policies:
  - p-b
  usage:
    p-b: 10m0s
`)
var oldUsersYAML = []byte(`---
u-a:
  bookings: []
  old_bookings: 
  - bk-0
  policies:
  - p-a
  usage:
    p-a: 5m0s
u-b:
  bookings: []
  old_bookings: 
  - bk-1
  policies:
  - p-b
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

}
func TestMain(m *testing.M) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port, err := freeport.GetFreePort()
	if err != nil {
		panic(err)
	}

	host := "http://[::]:" + strconv.Itoa(port)

	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	currentTime = &ct

	cfg = config.ServerConfig{
		Host:                host,
		Port:                port,
		StoreSecret:         []byte("somesecret"),
		RelaySecret:         []byte("anothersecret"),
		MinUserNameLength:   6,
		AccessTokenLifetime: time.Duration(time.Minute),
		// we can update the mock time by changing the value pointed to by currentTime
		Now:        func() time.Time { return *currentTime },
		PruneEvery: time.Duration(10 * time.Millisecond), //short so we convert bookings to old bookings quickly in tests
	}

	// modify the time function used to verify the jwt token
	// this should mean any time we set currentTime, the store and jwt both have the same time
	jwt.TimeFunc = func() time.Time { return *currentTime }

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

	go Run(ctx, cfg)

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
	now := (*currentTime).Unix()
	nbf := now - 1
	iat := nbf
	exp := nbf + 86400 //1 day
	token := login.New(audience, subject, scopes, iat, nbf, exp)
	return login.Sign(token, string(cfg.StoreSecret))
}

func signedUserTokenFor(subject string) (string, error) {

	audience := cfg.Host
	scopes := []string{"booking:user"}
	now := (*currentTime).Unix()
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

func printBookings(t *testing.T, bm cmodels.Bookings) {
	for k, v := range bm {
		fmt.Print(strconv.Itoa(k) + " : " + *v.User + " " + *v.Policy + " " + *v.Slot + " " + v.When.Start.String() + " " + v.When.End.String() + "\n")
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
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
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
	esa := store.StoreStatusAdmin{
		Locked:       false,
		Message:      "Welcome to the interval booking store",
		Now:          time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC),
		Bookings:     0,
		Descriptions: 8,
		Filters:      2,
		OldBookings:  0,
		Policies:     2,
		Resources:    2,
		Slots:        2,
		Streams:      2,
		UIs:          2,
		UISets:       2,
		Users:        0,
		Windows:      2}
	assert.Equal(t, esa, ssa)

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
	assert.Equal(t, expectedUsers, exportedUsers)
	resp.Body.Close()
}

func TestReplaceExportOldBookingsExportUsers(t *testing.T) {

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

	// move time forward
	ct := time.Date(2022, 11, 5, 6, 0, 0, 0, time.UTC)
	currentTime = &ct
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
	body, err := ioutil.ReadAll(resp.Body)
	var ssa store.StoreStatusAdmin
	err = json.Unmarshal(body, &ssa)
	assert.NoError(t, err)
	resp.Body.Close()
	esa := store.StoreStatusAdmin{
		Locked:       false,
		Message:      "Welcome to the interval booking store",
		Now:          time.Date(2022, 11, 5, 6, 0, 0, 0, time.UTC),
		Bookings:     0,
		Descriptions: 8,
		Filters:      2,
		OldBookings:  2,
		Policies:     2,
		Resources:    2,
		Slots:        2,
		Streams:      2,
		UIs:          2,
		UISets:       2,
		Users:        2,
		Windows:      2}
	assert.Equal(t, esa, ssa)

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
	assert.Equal(t, expectedUsers, exportedUsers)
	resp.Body.Close()

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
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &ssa)
	assert.NoError(t, err)
	resp.Body.Close()
	esa = store.StoreStatusAdmin{
		Locked:       false,
		Message:      "Welcome to the interval booking store",
		Now:          time.Date(2022, 11, 5, 6, 0, 0, 0, time.UTC),
		Bookings:     0,
		Descriptions: 8,
		Filters:      2,
		OldBookings:  0, // no old bookings
		Policies:     2,
		Resources:    2,
		Slots:        2,
		Streams:      2,
		UIs:          2,
		UISets:       2,
		Users:        0, // no users - these are wiped when the old bookings are wiped
		Windows:      2}
	assert.Equal(t, esa, ssa)

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
		Descriptions: 8,
		Filters:      2,
		OldBookings:  0,
		Policies:     2,
		Resources:    2,
		Slots:        2,
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
		Descriptions: 8,
		Filters:      2,
		OldBookings:  0,
		Policies:     2,
		Resources:    2,
		Slots:        2,
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
	var ss models.SlotStatus
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
	expected := `{"book_ahead":"1h0m0s","description":{"name":"policy-a","short":"a","type":"policy"},"display_guides":{"1mFor20m":{"book_ahead":"20m0s","duration":"1m0s","label":"1m","max_slots":15}},"enforce_book_ahead":true,"max_duration":"0s","max_usage":"0s","min_duration":"0s","slots":["sl-a"]}` + "\n"
	assert.Equal(t, expected, string(body))
	resp.Body.Close()

}
func TestGetAvailability(t *testing.T) {

	// make sure our pre-prepared bookings are in the future
	// other tests may have advanced time
	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	currentTime = &ct

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
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/policies/p-b/slots/sl-b", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
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
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/policies/p-b/slots/sl-b", nil)
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
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/policies/p-b/slots/sl-b", nil)
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
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/policies/p-b/slots/sl-b", nil)
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
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/policies/p-b/slots/sl-b", nil)
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
	currentTime = &ct

	satoken := loadTestManifest(t)
	removeAllBookings(t)
	bm := getBookings(t)
	assert.Equal(t, 0, len(bm))

	// user access token
	sutoken, err := signedUserToken()
	assert.NoError(t, err)

	client := &http.Client{}
	req, err := http.NewRequest("POST", cfg.Host+"/api/v1/policies/p-b/slots/sl-b", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	// add query params
	q := req.URL.Query()
	q.Add("user_name", "someuser") //must match token
	q.Add("from", "2022-11-05T00:01:00Z")
	q.Add("to", "2022-11-05T00:07:00Z")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode) //should be ok!
	body, err := ioutil.ReadAll(resp.Body)
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
	currentTime = &ct

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
	currentTime = &ct

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
	bookings := `[{"name":"bk-6","policy":"p-b","slot":"sl-b","user":"user-g","when":{"end":"2022-11-05T01:20:00.000Z","start":"2022-11-05T01:15:00.000Z"}}]` + "\n"
	assert.Equal(t, bookings, string(body))

	// cancel booking bk-6
	client = &http.Client{}
	req, err = http.NewRequest("DELETE", cfg.Host+"/api/v1/users/user-g/bookings/bk-6", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode) //not found means deleted

	// get bookings again
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/users/user-g/bookings", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
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
	currentTime = &ct
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
	bookings = `[{"name":"bk-5","policy":"p-b","slot":"sl-b","user":"user-f","when":{"end":"2022-11-05T01:10:00.000Z","start":"2022-11-05T01:05:00.000Z"}}]` + "\n"
	assert.Equal(t, bookings, string(body))

}

func TestGetActivity(t *testing.T) {

	// make sure our pre-prepared bookings are in the future
	// other tests may have advanced time
	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	currentTime = &ct

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
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	bookings := `[{"name":"bk-6","policy":"p-b","slot":"sl-b","user":"user-g","when":{"end":"2022-11-05T01:20:00.000Z","start":"2022-11-05T01:15:00.000Z"}}]` + "\n"
	assert.Equal(t, bookings, string(body))

	// move time forward to within the booked activity
	ct = time.Date(2022, 11, 5, 1, 15, 1, 0, time.UTC)
	currentTime = &ct

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
	//Order might change
	expected1 := `{"description":{"name":"slot-b","short":"b","type":"slot"},"exp":1667611200,"nbf":1667610900,"streams":[{"audience":"https://relay-access.practable.io","connection_type":"session","for":"data","scopes":["read","write"],"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb3BpYyI6ImJiYmIwMC1zdC1hIiwiUHJlZml4Ijoic2Vzc2lvbiIsIlNjb3BlcyI6WyJyZWFkIiwid3JpdGUiXSwic3ViIjoidXNlci1nIiwiYXVkIjpbImh0dHBzOi8vcmVsYXktYWNjZXNzLnByYWN0YWJsZS5pbyJdLCJleHAiOjE2Njc2MTEyMDAsIm5iZiI6MTY2NzYxMDkwMCwiaWF0IjoxNjY3NjEwOTAwfQ.B2jdYIYf6YHV1rSK6RkMyrGX2eQAPFg6QYwc6siVpb4","topic":"bbbb00-st-a","url":"https://relay-access.practable.io"},{"audience":"https://relay-access.practable.io","connection_type":"session","for":"video","scopes":["read"],"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb3BpYyI6ImJiYmIwMC1zdC1iIiwiUHJlZml4Ijoic2Vzc2lvbiIsIlNjb3BlcyI6WyJyZWFkIl0sInN1YiI6InVzZXItZyIsImF1ZCI6WyJodHRwczovL3JlbGF5LWFjY2Vzcy5wcmFjdGFibGUuaW8iXSwiZXhwIjoxNjY3NjExMjAwLCJuYmYiOjE2Njc2MTA5MDAsImlhdCI6MTY2NzYxMDkwMH0.9A-5zGLjB3Dw2PpGHfYNoapfrt-VKa8BmRVaggF4oAk","topic":"bbbb00-st-b","url":"https://relay-access.practable.io"}],"uis":[{"description":{"name":"ui-a","short":"a","type":"ui"},"streams_required":["st-a","st-b"],"url":"a"},{"description":{"name":"ui-b","short":"b","type":"ui"},"streams_required":["st-a","st-b"],"url":"b"}]}` + "\n"
	expected2 := `{"description":{"name":"slot-b","short":"b","type":"slot"},"exp":1667611200,"nbf":1667610900,"streams":[{"audience":"https://relay-access.practable.io","connection_type":"session","for":"video","scopes":["read"],"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb3BpYyI6ImJiYmIwMC1zdC1iIiwiUHJlZml4Ijoic2Vzc2lvbiIsIlNjb3BlcyI6WyJyZWFkIl0sInN1YiI6InVzZXItZyIsImF1ZCI6WyJodHRwczovL3JlbGF5LWFjY2Vzcy5wcmFjdGFibGUuaW8iXSwiZXhwIjoxNjY3NjExMjAwLCJuYmYiOjE2Njc2MTA5MDAsImlhdCI6MTY2NzYxMDkwMH0.9A-5zGLjB3Dw2PpGHfYNoapfrt-VKa8BmRVaggF4oAk","topic":"bbbb00-st-b","url":"https://relay-access.practable.io"},{"audience":"https://relay-access.practable.io","connection_type":"session","for":"data","scopes":["read","write"],"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb3BpYyI6ImJiYmIwMC1zdC1hIiwiUHJlZml4Ijoic2Vzc2lvbiIsIlNjb3BlcyI6WyJyZWFkIiwid3JpdGUiXSwic3ViIjoidXNlci1nIiwiYXVkIjpbImh0dHBzOi8vcmVsYXktYWNjZXNzLnByYWN0YWJsZS5pbyJdLCJleHAiOjE2Njc2MTEyMDAsIm5iZiI6MTY2NzYxMDkwMCwiaWF0IjoxNjY3NjEwOTAwfQ.B2jdYIYf6YHV1rSK6RkMyrGX2eQAPFg6QYwc6siVpb4","topic":"bbbb00-st-a","url":"https://relay-access.practable.io"}],"uis":[{"description":{"name":"ui-a","short":"a","type":"ui"},"streams_required":["st-a","st-b"],"url":"a"},{"description":{"name":"ui-b","short":"b","type":"ui"},"streams_required":["st-a","st-b"],"url":"b"}]}` + "\n"
	assert.True(t, expected1 == string(body) || expected2 == string(body))

}

func TestAddGetPoliciesAndStatus(t *testing.T) {

	// make sure our pre-prepared bookings are in the future
	// other tests may have advanced time
	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	currentTime = &ct

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

	// get policies for user u-g
	sutoken, err := signedUserTokenFor("user-g")
	assert.NoError(t, err)
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/users/user-g/policies", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	// display_guides are omitted if empty
	policies := `[{"book_ahead":"2h0m0s","description":{"name":"policy-b","short":"b","type":"policy"},"enforce_book_ahead":true,"enforce_max_bookings":true,"enforce_max_duration":true,"enforce_max_usage":true,"enforce_min_duration":true,"max_bookings":2,"max_duration":"10m0s","max_usage":"30m0s","min_duration":"5m0s","slots":["sl-b"]}]` + "\n"
	assert.Equal(t, policies, string(body))

	// get policy status for p-b for user u-g
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/users/user-g/policies/p-b", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	status := `{"current_bookings":1,"old_bookings":0,"usage":"5m0s"}` + "\n"
	assert.Equal(t, status, string(body))

	//add policy p-a for user u-g
	client = &http.Client{}
	req, err = http.NewRequest("POST", cfg.Host+"/api/v1/users/user-g/policies/p-a", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, 204, resp.StatusCode) //should be ok!

	// get policies for user u-g
	sutoken, err = signedUserTokenFor("user-g")
	assert.NoError(t, err)
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/users/user-g/policies", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", sutoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	// order can change so check both possibilities
	// unmarshalling can comparing objects did not work reliably because of the use of pointers in struct
	policies1JSON := `[{"book_ahead":"2h0m0s","description":{"name":"policy-b","short":"b","type":"policy"},"enforce_book_ahead":true,"enforce_max_bookings":true,"enforce_max_duration":true,"enforce_max_usage":true,"enforce_min_duration":true,"max_bookings":2,"max_duration":"10m0s","max_usage":"30m0s","min_duration":"5m0s","slots":["sl-b"]},{"book_ahead":"1h0m0s","description":{"name":"policy-a","short":"a","type":"policy"},"display_guides":{"1mFor20m":{"book_ahead":"20m0s","duration":"1m0s","label":"1m","max_slots":15}},"enforce_book_ahead":true,"max_duration":"0s","max_usage":"0s","min_duration":"0s","slots":["sl-a"]}]` + "\n"
	policies2JSON := `[{"book_ahead":"1h0m0s","description":{"name":"policy-a","short":"a","type":"policy"},"display_guides":{"1mFor20m":{"book_ahead":"20m0s","duration":"1m0s","label":"1m","max_slots":15}},"enforce_book_ahead":true,"max_duration":"0s","max_usage":"0s","min_duration":"0s","slots":["sl-a"]},{"book_ahead":"2h0m0s","description":{"name":"policy-b","short":"b","type":"policy"},"enforce_book_ahead":true,"enforce_max_bookings":true,"enforce_max_duration":true,"enforce_max_usage":true,"enforce_min_duration":true,"max_bookings":2,"max_duration":"10m0s","max_usage":"30m0s","min_duration":"5m0s","slots":["sl-b"]}]` + "\n"

	assert.True(t, policies1JSON == string(body) || policies2JSON == string(body))

	if !(policies1JSON == string(body) || policies2JSON == string(body)) {
		t.Error("Policies did not match")
		t.Log("expected either (1):\n " + policies1JSON + " or (2):\n" + policies2JSON)
		t.Log("got: " + string(body))
	}
	//var actualPolicies, expectedPolicies models.PoliciesDescribed
	//err = json.Unmarshal(policiesJSON, &expectedPolicies)
	//assert.NoError(t, err)
	//err = json.Unmarshal(body, &actualPolicies)
	//assert.NoError(t, err)
	//assert.Equal(t, expectedPolicies, actualPolicies)

}

// TestRestrictedToAdmin checks that users cannot access admin endpoints
func TestRestrictedToAdmin(t *testing.T) {

	ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
	currentTime = &ct

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
		currentTime = &ct
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
	currentTime = &ct

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
		setLock(t, false, "unlocked")
		ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
		currentTime = &ct
	}
	locked := func() {
		loadTestManifest(t)
		removeAllBookings(t)
		setLock(t, true, "locked")
		ct := time.Date(2022, 11, 5, 0, 0, 0, 0, time.UTC)
		currentTime = &ct
	}

	getAvailability := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := users.NewGetAvailabilityParams().WithTimeout(timeout).WithPolicyName("p-a").WithSlotName("sl-a")
		return bc.Users.GetAvailability(p, auth)
	}

	getDescription := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := users.NewGetDescriptionParams().WithTimeout(timeout).WithDescriptionName("d-r-a")
		return bc.Users.GetDescription(p, auth)
	}
	getPolicy := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := users.NewGetPolicyParams().WithTimeout(timeout).WithPolicyName("p-a")
		return bc.Users.GetPolicy(p, auth)
	}
	getStoreStatusUser := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := users.NewGetStoreStatusUserParams().WithTimeout(timeout)
		return bc.Users.GetStoreStatusUser(p, auth)
	}
	makeBooking := func(bc *apiclient.Client, auth rt.ClientAuthInfoWriter) (interface{}, error) {
		p := users.NewMakeBookingParams().
			WithTimeout(timeout).
			WithPolicyName("p-a").
			WithSlotName("sl-a").
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
		"GetDescriptionLockedAdminAllowed":       {locked, getDescription, authAdmin, true, `[GET /descriptions/{description_name}][200] getDescriptionOK`},
		"GetDescriptionLockedUserDenied":         {locked, getDescription, authUser, false, `[GET /descriptions/{description_name}][401] getDescriptionUnauthorized`},
		"GetDescriptionUnlockedAdminAllowed":     {unlocked, getDescription, authAdmin, true, `[GET /descriptions/{description_name}][200] getDescriptionOK`},
		"GetDescriptionUnlockedUserAllowed":      {unlocked, getDescription, authUser, true, `[GET /descriptions/{description_name}][200] getDescriptionOK`},
		"GetPolicyLockedAdminAllowed":            {locked, getPolicy, authAdmin, true, `[GET /policies/{policy_name}][200] getPolicyOK`},
		"GetPolicyLockedUserDenied":              {locked, getPolicy, authUser, false, `[GET /policies/{policy_name}][401] getPolicyUnauthorized`},
		"GetPolicyUnlockedAdminAllowed":          {unlocked, getPolicy, authAdmin, true, `[GET /policies/{policy_name}][200] getPolicyOK`},
		"GetPolicyUnlockedUserAllowed":           {unlocked, getPolicy, authUser, true, `[GET /policies/{policy_name}][200] getPolicyOK`},
		"GetAvailabilityLockedAdminAllowed":      {locked, getAvailability, authAdmin, true, `[GET /policies/{policy_name}/slots/{slot_name}][200] getAvailabilityOK`},
		"GetAvailabilityLockedUserDenied":        {locked, getAvailability, authUser, false, `[GET /policies/{policy_name}/slots/{slot_name}][401] getAvailabilityUnauthorized`},
		"GetAvailabilityUnlockedAdminAllowed":    {unlocked, getAvailability, authAdmin, true, `[GET /policies/{policy_name}/slots/{slot_name}][200] getAvailabilityOK`},
		"GetAvailabilityUnlockedUserAllowed":     {unlocked, getAvailability, authUser, true, `[GET /policies/{policy_name}/slots/{slot_name}][200] getAvailabilityOK`},
		"makeBookingLockedAdminAllowed":          {locked, makeBooking, authAdmin, true, `[POST /policies/{policy_name}/slots/{slot_name}][204] makeBookingNoContent`},
		"makeBookingLockedUserDenied":            {locked, makeBooking, authUser, false, `[POST /policies/{policy_name}/slots/{slot_name}][401] makeBookingUnauthorized`},
		"makeBookingUnlockedAdminAllowed":        {unlocked, makeBooking, authAdmin, true, `[POST /policies/{policy_name}/slots/{slot_name}][204] makeBookingNoContent`},
		"makeBookingUnlockedUserAllowed":         {unlocked, makeBooking, authUser, true, `[POST /policies/{policy_name}/slots/{slot_name}][204] makeBookingNoContent`},
		"GetStoreStatusUserLockedAdminAllowed":   {locked, getStoreStatusUser, authAdmin, true, `[GET /users/status][200] getStoreStatusUserOK`},
		"GetStoreStatusUserLockedUserAllowed":    {locked, getStoreStatusUser, authUser, true, `[GET /users/status][200] getStoreStatusUserOK`},
		"GetStoreStatusUserUnlockedAdminAllowed": {unlocked, getStoreStatusUser, authAdmin, true, `[GET /users/status][200] getStoreStatusUserOK`},
		"GetStoreStatusUserUnlockedUserAllowed":  {unlocked, getStoreStatusUser, authUser, true, `[GET /users/status][200] getStoreStatusUserOK`},
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
