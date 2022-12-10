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
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/phayes/freeport"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/timdrysdale/interval/internal/config"
	"github.com/timdrysdale/interval/internal/login"
	"github.com/timdrysdale/interval/internal/serve/models"
	"github.com/timdrysdale/interval/internal/store"
	"gopkg.in/yaml.v2"
)

var debug bool
var cfg config.ServerConfig
var currentTime *time.Time

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
policies:
  p-a:
    book_ahead: 1h
    description: d-p-a
    display_guides:
      1m:
        book_ahead: 20m
        duration: 1m
        max_slots: 15
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
    audience: a
    connection_type: a
    for: a
    scopes:
    - r
    - w
    topic: a
    url: a
  st-b:
    audience: b
    connection_type: b
    for: b
    scopes:
    - r
    - w
    topic: b
    url: b
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

var bookingsYAML = []byte(`---
bk-0:
  cancelled: false
  name: bk-0
  policy: p-a
  slot: sl-a
  started: false
  unfulfilled: false
  user: u-a
  when:
    start: '2022-11-05T00:10:00Z'
    end: '2022-11-05T00:15:00Z'
bk-1:
  cancelled: false
  name: bk-1
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: u-b
  when:
    start: '2022-11-05T00:20:00Z'
    end: '2022-11-05T00:30:00Z'
`)

var bookings2YAML = []byte(`---
bk-0:
  cancelled: false
  name: bk-0
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: u-a
  when:
    start: '2022-11-05T00:10:00Z'
    end: '2022-11-05T00:15:00Z'
bk-1:
  cancelled: false
  name: bk-1
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: u-b
  when:
    start: '2022-11-05T00:20:00Z'
    end: '2022-11-05T00:30:00Z'
bk-2:
  cancelled: false
  name: bk-2
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: u-c
  when:
    start: '2022-11-05T00:35:00Z'
    end: '2022-11-05T00:40:00Z'
bk-3:
  cancelled: false
  name: bk-3
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: u-d
  when:
    start: '2022-11-05T00:45:00Z'
    end: '2022-11-05T00:50:00Z'
bk-4:
  cancelled: false
  name: bk-4
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: u-e
  when:
    start: '2022-11-05T00:55:00Z'
    end: '2022-11-05T01:00:00Z'
bk-5:
  cancelled: false
  name: bk-5
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: u-f
  when:
    start: '2022-11-05T01:05:00Z'
    end: '2022-11-05T01:10:00Z'
bk-6:
  cancelled: false
  name: bk-6
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: u-a
  when:
    start: '2022-11-05T01:15:00Z'
    end: '2022-11-05T01:20:00Z'
bk-7:
  cancelled: false
  name: bk-7
  policy: p-b
  slot: sl-b
  started: false
  unfulfilled: false
  user: u-g
  when:
    start: '2022-11-05T01:25:00Z'
    end: '2022-11-05T01:30:00Z'
`)

var noBookingsYAML = []byte(`{}`)

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
		MinUserNameLength:   6,
		AccessTokenLifetime: time.Duration(time.Minute),
		// we can update the mock time by changing the value pointed to by currentTime
		Now:        func() time.Time { return *currentTime },
		PruneEvery: time.Duration(10 * time.Millisecond), //short so we convert bookings to old bookings quickly in tests
	}

	// modify the time function used to verify the jwt token
	// this should mean any time we set currentTime, the store and jwt both have the same time
	jwt.TimeFunc = func() time.Time { return *currentTime }

	go Run(ctx, cfg)

	time.Sleep(time.Second)

	exitVal := m.Run()

	os.Exit(exitVal)
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

func signedUserToken() (string, error) {

	audience := cfg.Host
	subject := "someuser"
	scopes := []string{"booking:user"}
	now := (*currentTime).Unix()
	nbf := now - 1
	iat := nbf
	exp := nbf + 86400 //1 day
	token := login.New(audience, subject, scopes, iat, nbf, exp)
	return login.Sign(token, string(cfg.StoreSecret))
}

func loadTestManifest(t *testing.T) string {
	stoken, err := signedAdminToken()
	assert.NoError(t, err)
	client := &http.Client{}
	bodyReader := bytes.NewReader(manifestYAML)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/manifest", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "text/plain")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	resp.Body.Close()
	return stoken //for use by other commands in test
}

func getBookings(t *testing.T) map[string]store.Booking {
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
	var exportedBookings map[string]store.Booking
	err = yaml.Unmarshal(body, &exportedBookings)
	assert.NoError(t, err)
	resp.Body.Close()
	return exportedBookings
}

func printBookings(t *testing.T, bm map[string]store.Booking) {
	for k, v := range bm {
		fmt.Print(k + " : " + v.User + " " + v.Policy + " " + v.Slot + " " + v.When.Start.String() + " " + v.When.End.String() + "\n")
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
	bodyReader := bytes.NewReader(noBookingsYAML)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "text/plain")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!

	client = &http.Client{}
	bodyReader = bytes.NewReader(noBookingsYAML)
	req, err = http.NewRequest("PUT", cfg.Host+"/api/v1/admin/oldbookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "text/plain")
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
}

func TestLogin(t *testing.T) {

	client := &http.Client{}
	req, err := http.NewRequest("POST", cfg.Host+"/api/v1/login/someuser", nil)
	assert.NoError(t, err)
	resp, err := client.Do(req)
	assert.NoError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	atr := &models.AccessToken{}
	err = json.Unmarshal(body, atr)
	assert.NoError(t, err)
	assert.Equal(t, float64(61), *(atr.Exp)-*(atr.Nbf)) //61 seconds
	assert.Equal(t, "someuser", *(atr.Sub))
	assert.Equal(t, []string{"booking:user"}, atr.Scopes)
	assert.Equal(t, "ey", (*(atr.Token))[0:2]) //necessary but not sufficient!

}

func TestCheckReplaceExportManifest(t *testing.T) {

	// make admin token
	stoken, err := signedAdminToken()
	assert.NoError(t, err)

	//check manifest
	client := &http.Client{}
	bodyReader := bytes.NewReader(manifestYAML)
	req, err := http.NewRequest("GET", cfg.Host+"/api/v1/admin/manifest/check", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "text/plain")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, "204 No Content", resp.Status) //should be ok!
	resp.Body.Close()

	//replace manifest
	client = &http.Client{}
	bodyReader = bytes.NewReader(manifestYAML)
	req, err = http.NewRequest("PUT", cfg.Host+"/api/v1/admin/manifest", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "text/plain")
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
	err = yaml.Unmarshal(body, &exportedManifest)
	assert.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, expectedManifest, exportedManifest)

}

func TestReplaceExportBookingsExportUsers(t *testing.T) {

	stoken := loadTestManifest(t)

	// replace bookings
	client := &http.Client{}
	bodyReader := bytes.NewReader(bookingsYAML)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "text/plain")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!

	// export bookings
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/bookings", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err := ioutil.ReadAll(resp.Body)
	var expectedBookings, exportedBookings map[string]store.Booking
	err = yaml.Unmarshal(bookingsYAML, &expectedBookings)
	err = yaml.Unmarshal(body, &exportedBookings)
	assert.NoError(t, err)
	assert.Equal(t, expectedBookings, exportedBookings)
	resp.Body.Close()

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
	bodyReader := bytes.NewReader(bookingsYAML)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "text/plain")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!

	// move time forward
	ct := time.Date(2022, 11, 5, 6, 0, 0, 0, time.UTC)
	currentTime = &ct
	time.Sleep(50 * time.Millisecond) //wait for pruning to happen

	// export bookings
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/bookings", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err := ioutil.ReadAll(resp.Body)
	var expectedBookings, exportedBookings map[string]store.Booking
	err = yaml.Unmarshal(noBookingsYAML, &expectedBookings)
	err = yaml.Unmarshal(body, &exportedBookings)
	assert.NoError(t, err)
	assert.Equal(t, expectedBookings, exportedBookings)
	resp.Body.Close()

	// export old bookings
	client = &http.Client{}
	req, err = http.NewRequest("GET", cfg.Host+"/api/v1/admin/oldbookings", nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) //should be ok!
	body, err = ioutil.ReadAll(resp.Body)
	err = yaml.Unmarshal(bookingsYAML, &expectedBookings)
	err = yaml.Unmarshal(body, &exportedBookings)
	assert.NoError(t, err)
	assert.Equal(t, expectedBookings, exportedBookings)
	resp.Body.Close()

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
	bodyReader = bytes.NewReader(noBookingsYAML)
	req, err = http.NewRequest("PUT", cfg.Host+"/api/v1/admin/oldbookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", stoken)
	req.Header.Add("Content-Type", "text/plain")
	resp, err = client.Do(req)
	assert.NoError(t, err)
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
	expected := `{"book_ahead":"1h0m0s","description":"d-p-a","display_guides":[{"book_ahead":"20m0s","duration":"1m0s","max_slots":15}],"enforce_book_ahead":true,"max_duration":"0s","max_usage":"0s","min_duration":"0s","slots":["sl-a"]}` + "\n"
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
	bodyReader := bytes.NewReader(bookings2YAML)
	req, err := http.NewRequest("PUT", cfg.Host+"/api/v1/admin/bookings", bodyReader)
	assert.NoError(t, err)
	req.Header.Add("Authorization", satoken)
	req.Header.Add("Content-Type", "text/plain")
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
