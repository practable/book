// Package resource implements a client that obtains resource
// information, and can get/set resource availability
package resource

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	rt "github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/golang-jwt/jwt/v4"
	apiclient "github.com/practable/book/internal/client/client"
	log "github.com/sirupsen/logrus"
)

//rt "github.com/go-openapi/runtime"
//	httptransport "github.com/go-openapi/runtime/client"
//	apiclient "github.com/practable/book/internal/client/client"

type About struct {
	Name    string
	Streams []string
	Tests   []string
}

type Status struct {
	Available bool
	Reason    string
}

type Config struct {
	BasePath  string
	Host      string
	Scheme    string
	Token     string
	Timeout   time.Duration
	auth      rt.ClientAuthInfoWriter
	transport *apiclient.TransportConfig
	//client    *apiclient.Client
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

func (c *Config) Prepare() {

	c.auth = httptransport.APIKeyAuth("Authorization", "header", c.Token)
	c.transport = apiclient.DefaultTransportConfig().WithBasePath(c.BasePath).WithHost(c.Host).WithSchemes([]string{c.Scheme})
	//c.client = apiclient.NewHTTPClientWithConfig(nil, c.transport)
}

func (c *Config) GetResources() ([]About, error) {

	client := &http.Client{}
	url := c.Scheme + "://" + c.Host + c.BasePath + "/admin/resources"
	log.Tracef("url:%s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", c.Token)

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Status code was %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)

	results := make(map[string]About)
	log.Tracef("body:%s", string(body))
	err = json.Unmarshal(body, &results)

	if err != nil {
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

	/*
		bc := apiclient.NewHTTPClientWithConfig(nil, c.transport)
		p := admin.NewGetResourcesParams().WithTimeout(c.Timeout)
		resp, err := bc.Admin.GetResources(p, c.auth)

		if err != nil {
			fmt.Println(err.Error())
			log.Error(err.Error())
			return []About{}, err
		}

		r := []About{}

		for _, v := range resp.Payload { //models.Resources

			a := About{
				Name:    *v.TopicStub, //or k?
				Streams: v.Streams,
				Tests:   v.Tests,
			}
			r = append(r, a)

		}
		// todo check return code
		return r, err
	*/
}

func (c *Config) GetResourceAvailability(name string) Status {
	return Status{}
}

func (c *Config) SetResourceAvailability(name string, available bool, reason string) Status {
	return Status{}
}

/*
	status, err := bc.Admin.ExportBookings(
		admin.NewExportBookingsParams().WithTimeout(timeout),
		aa)
		p := admin.NewSetSlotIsAvailableParams().WithTimeout(timeout).WithSlotName("sl-a").WithAvailable(true).WithReason("test")
		return bc.Admin.SetSlotIsAvailable(p, auth)
*/
