package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
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
    book_ahead: 0s
    description: d-p-a
    enforce_book_ahead: false
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

	if debug {
		t.Log(string(body))
		t.Log(*(atr.Token))
		//atr := &models.Error{}
		//err = json.Unmarshal(body, atr)
		//assert.NoError(t, err)
		//t.Log(*(atr.Code) + *(atr.Message))
	}

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

	/* add query params
	q := req.URL.Query()
	q.Add("lock", "true")
	req.URL.RawQuery = q.Encode()
	*/

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
	t.Log(string(body))
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
	t.Log(string(body))
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
	t.Log(string(body))
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
	t.Log(string(body))
	var expectedUsers, exportedUsers map[string]store.UserExternal
	err = yaml.Unmarshal(oldUsersYAML, &expectedUsers)
	err = yaml.Unmarshal(body, &exportedUsers)
	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, exportedUsers)
	resp.Body.Close()
}
