package v1auth

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/google/go-querystring/query"
	"github.com/nerdalize/nerd/nerd/client"
	v1payload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"
)

//OpsClient is used for a bunch of operations on the auth API that don't require oauth authentication.
type OpsClient struct {
	OpsClientConfig
}

var _ OpsClientInterface = &OpsClient{}

//OpsClientInterface is an interface so client calls can be mocked.
type OpsClientInterface interface {
	GetOAuthCredentials(code, clientID, clientSecret, localServerURL string) (output *v1payload.GetOAuthCredentialsOutput, err error)
	RefreshOAuthCredentials(refreshToken, clientID, clientSecret string) (output *v1payload.RefreshOAuthCredentialsOutput, err error)
}

//OpsClientConfig is the config for OpsClient
type OpsClientConfig struct {
	Doer   Doer
	Base   *url.URL
	Logger *log.Logger
}

//NewOpsClient creates a new OpsClient.
func NewOpsClient(c OpsClientConfig) *OpsClient {
	if c.Doer == nil {
		c.Doer = http.DefaultClient
		if os.Getenv("NERD_ENV") == "dev" {
			c.Doer = &http.Client{Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}}
		}
	}
	if c.Base.Path != "" && c.Base.Path[len(c.Base.Path)-1] != '/' {
		c.Base.Path = c.Base.Path + "/"
	}
	return &OpsClient{c}
}

//doRequest requests the server and decodes the output into the `output` field.
//
//When a status code >= 400 is returned by the server the returned error will be of type HTTPError.
func (c *OpsClient) doRequest(method, urlPath string, input, output interface{}) (err error) {
	path, err := url.Parse(urlPath)
	if err != nil {
		return client.NewError("invalid url path provided", err)
	}

	resolved := c.Base.ResolveReference(path)

	v, err := query.Values(input)
	if err != nil {
		return client.NewError("failed to encode the request body", err)
	}
	query := v.Encode()
	var req *http.Request
	if method == http.MethodGet {
		resolved.RawQuery = query
		req, err = http.NewRequest(method, resolved.String(), nil)
		if err != nil {
			return client.NewError("failed to create HTTP request", err)
		}
	} else {
		req, err = http.NewRequest(method, resolved.String(), strings.NewReader(query))
		if err != nil {
			return client.NewError("failed to create HTTP request", err)
		}
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client.LogRequest(req, c.Logger)
	resp, err := c.Doer.Do(req)
	if err != nil {
		return client.NewError("failed to create HTTP request", err)
	}
	client.LogResponse(resp, c.Logger)

	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		errv := &v1payload.Error{}
		err = dec.Decode(errv)
		if err != nil {
			return client.NewError(fmt.Sprintf("failed to decode unexpected HTTP response (%s)", resp.Status), err)
		}

		return &HTTPError{
			StatusCode: resp.StatusCode,
			Err:        errv,
		}
	}

	if output != nil {
		err = dec.Decode(output)
		if err != nil {
			return client.NewError(fmt.Sprintf("failed to decode successful HTTP response (%s)", resp.Status), err)
		}
	}

	return nil
}

//GetOAuthCredentials gets oauth credentials based on a 'session' code
func (c *OpsClient) GetOAuthCredentials(code, clientID, clientSecret, localServerURL string) (output *v1payload.GetOAuthCredentialsOutput, err error) {
	output = &v1payload.GetOAuthCredentialsOutput{}
	input := &v1payload.GetOAuthCredentialsInput{
		Code:        code,
		ClientID:    clientID,
		GrantType:   "authorization_code",
		RedirectURI: localServerURL,
	}
	return output, c.doRequest(http.MethodPost, "o/token/", input, output)
}

//RefreshOAuthCredentials refreshes an oauth access token
func (c *OpsClient) RefreshOAuthCredentials(refreshToken, clientID, clientSecret string) (output *v1payload.RefreshOAuthCredentialsOutput, err error) {
	output = &v1payload.RefreshOAuthCredentialsOutput{}
	input := &v1payload.RefreshOAuthCredentialsInput{
		RefreshToken: refreshToken,
		ClientID:     clientID,
		GrantType:    "refresh_token",
	}
	return output, c.doRequest(http.MethodPost, "o/token/", input, output)
}
