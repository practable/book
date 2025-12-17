// Package resource implements a client that obtains resource
// information, and can get/set resource availability
package resource

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

//rt "github.com/go-openapi/runtime"
//	httptransport "github.com/go-openapi/runtime/client"
//	apiclient "github.com/practable/book/internal/client/client"

type About struct {
	Name      string   `json:"name"`
	Streams   []string `json:"streams"`
	Tests     []string `json:"tests"`
	TopicStub string   `json:"topic_stub"`
}

type Status struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason"`
}

type Config struct {
	BasePath string
	Host     string
	Scheme   string
	Token    string
	Timeout  time.Duration
}

// Token represents a token used for login or booking
type Token struct {

	// Scopes controlling access booking system
	Scopes []string `json:"scopes"`

	jwt.RegisteredClaims
}

// Token creates and signs a token
func NewToken(audience, subject, secret string, scopes []string, iat, nbf, exp time.Time) (string, error) {
	token := Token{
		Scopes: scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(iat),
			NotBefore: jwt.NewNumericDate(nbf),
			ExpiresAt: jwt.NewNumericDate(exp),
			Audience:  jwt.ClaimStrings{audience},
			Subject:   subject,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, token).SignedString([]byte(secret))
}

func (c *Config) GetResources() ([]About, error) {

	client := &http.Client{}
	url := c.Scheme + "://" + c.Host + c.BasePath + "/admin/resources"
	log.Tracef("GetResource: url is %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("GetResource: new request error was %s", err.Error())
		return nil, err
	}
	req.Header.Add("Authorization", c.Token)

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("GetResource: do request error was %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Errorf("GetResource: Status code was %d", resp.StatusCode)
		return nil, fmt.Errorf("Status code was %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("GetResource: ioutil.ReadAll error is %s", err.Error())
		return nil, err
	}

	results := make(map[string]About)

	log.Tracef("GetResource:  body is %s", string(body))

	err = json.Unmarshal(body, &results)

	if err != nil {
		log.Errorf("GetResource: unmarshal error is %s", err.Error())
		return nil, err
	}

	as := []About{}

	for k, v := range results {
		about := v
		name := k
		about.Name = name
		as = append(as, about)
	}

	return as, nil

}

func (c *Config) GetResourceAvailability(name string) (Status, error) {

	client := &http.Client{}
	url := c.Scheme + "://" + c.Host + c.BasePath + "/admin/resources/" + name
	log.Tracef("GetResourceAvailability: url is %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("GetResourceAvailability: new request error was %s", err.Error())
		return Status{}, err
	}
	req.Header.Add("Authorization", c.Token)

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("GetResourceAvailability: do request error was %s", err.Error())
		return Status{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Errorf("GetResourceAvailability: Status code was %d", resp.StatusCode)
		return Status{}, fmt.Errorf("Status code was %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("GetResourceAvailability: ioutil.ReadAll error is %s", err.Error())
		return Status{}, err
	}

	log.Tracef("GetResourceAvailability:  body is %s", string(body))

	result := Status{}
	err = json.Unmarshal(body, &result)

	if err != nil {
		log.Errorf("GetResourceAvailability: unmarshal error is %s", err.Error())
		return Status{}, err
	}

	return result, nil

}

func (c *Config) SetResourceAvailability(name string, available bool, reason string) error {

	client := &http.Client{}
	url := c.Scheme + "://" + c.Host + c.BasePath + "/admin/resources/" + name
	log.Tracef("SetResourceAvailability: url is %s", url)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		log.Errorf("SetResourceAvailability: new request error was %s", err.Error())
		return err
	}
	req.Header.Add("Authorization", c.Token)

	// add query params
	q := req.URL.Query()
	q.Add("available", strconv.FormatBool(available))
	q.Add("reason", reason)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("SetResourceAvailability: do request error was %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		log.Errorf("SetResourceAvailability: Status code was %d", resp.StatusCode)
		return fmt.Errorf("Status code was %d", resp.StatusCode)
	}

	return nil

}
