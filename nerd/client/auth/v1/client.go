package v1auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/nerdalize/nerd/nerd/client"
	v1payload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"
)

const (
	//TokenEndpoint is the endpoint from where to fetch the JWT.
	TokenEndpoint = "token/?service=nce.nerdalize.com"
)

//Auth is the client for the nerdalize authentication server.
type Client struct {
	ClientConfig
}

//NerdConfig provides config details to create a Nerd client.
type ClientConfig struct {
	Doer   Doer
	Base   *url.URL
	Logger client.Logger
}

// Doer executes http requests.  It is implemented by *http.Client.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

//NewAuthAPI creates a new Auth.
func NewClient(c ClientConfig) *Client {
	if c.Doer == nil {
		c.Doer = http.DefaultClient
	}
	if c.Base.Path != "" && c.Base.Path[len(c.Base.Path)-1] != '/' {
		c.Base.Path = c.Base.Path + "/"
	}
	return &Client{c}
}

//GetToken fetches a JWT for a given user.
func (c *Client) GetToken(user, pass string) (string, error) {
	path, err := url.Parse(TokenEndpoint)
	if err != nil {
		return "", client.NewError("invalid url path provided", err)
	}
	resolved := c.Base.ResolveReference(path)

	req, err := http.NewRequest(http.MethodGet, resolved.String(), nil)
	if err != nil {
		return "", client.NewError("invalid url path provided", err)
	}
	req.SetBasicAuth(user, pass)
	client.LogRequest(req, c.Logger)
	resp, err := c.Doer.Do(req)
	if err != nil {
		return "", client.NewError("failed to perform HTTP request", err)
	}
	client.LogResponse(resp, c.Logger)

	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		errv := &v1payload.Error{}
		err = dec.Decode(errv)
		if err != nil {
			return "", client.NewError(fmt.Sprintf("failed to decode unexpected HTTP response (%s)", resp.Status), err)
		}

		return "", errv
	}

	type body struct {
		Token string `json:"token"`
	}
	b := &body{}
	err = dec.Decode(b)
	if err != nil {
		return "", client.NewError(fmt.Sprintf("failed to decode successfull HTTP response (%s)", resp.Status), err)
	}
	return b.Token, nil
}
