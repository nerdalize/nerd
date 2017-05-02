package v1batch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync"

	"github.com/nerdalize/nerd/nerd/client"
	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

const (
	//AuthHeader is the name of the HTTP Authorization header.
	AuthHeader = "Authorization"

	projectsPrefix = "projects"

	tasksEndpoint   = "tasks"
	tokensEndpoint  = "tokens"
	datasetEndpoint = "datasets"
	workersEndpoint = "workers"
	queuesEndpoint  = "queues"
)

//Client is a client for the Nerdalize API.
type Client struct {
	ClientConfig
	cred string
	m    sync.Mutex
}

//ClientConfig provides config details to create a Nerd client.
type ClientConfig struct {
	Doer        Doer
	JWTProvider JWTProvider
	Base        *url.URL
	Logger      client.Logger
}

//ClientInterface is an interface so client calls can be mocked.
type ClientInterface interface {
	ClientDatasetInterface
	ClientPingInterface
	ClientQueueInterface
	ClientTaskInterface
	ClientTokenInterface
}

// Force compile errors when Client doesn't implement ClientInterface.
var _ ClientInterface = &Client{}

// Doer executes http requests.  It is implemented by *http.Client.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

//NewClient creates a new Nerd client from a config object. The http.DefaultClient
//will be used as default Doer.
func NewClient(conf ClientConfig) *Client {
	if conf.Doer == nil {
		conf.Doer = http.DefaultClient
	}
	if conf.Base.Path != "" && conf.Base.Path[len(conf.Base.Path)-1] != '/' {
		conf.Base.Path = conf.Base.Path + "/"
	}
	cl := &Client{
		ClientConfig: conf,
		cred:         "",
		m:            sync.Mutex{},
	}
	return cl
}

//getJWT atomically gets the JWT from the JWT provider.
func (c *Client) getJWT() (string, error) {
	c.m.Lock()
	defer c.m.Unlock()
	if c.JWTProvider == nil {
		return "", fmt.Errorf("No JWT provider found")
	}
	if c.cred == "" || c.JWTProvider.IsExpired() {
		cred, err := c.JWTProvider.Retrieve()
		if err != nil {
			return "", err
		}
		c.cred = cred
	}
	return c.cred, nil
}

//doRequest requests the server and decodes the output into the `output` field.
//
//doRequest will set the Authentication header with the JWT provided by the JWTProvider.
//When a status code >= 400 is returned by the server the returned error will be of type HTTPError.
func (c *Client) doRequest(method, urlPath string, input, output interface{}) (err error) {
	cred, err := c.getJWT()
	if err != nil {
		return err
	}

	path, err := url.Parse(urlPath)
	if err != nil {
		return client.NewError("invalid url path provided", err)
	}

	resolved := c.Base.ResolveReference(path)

	var req *http.Request
	if input != nil {
		buf := bytes.NewBuffer(nil)
		enc := json.NewEncoder(buf)
		err = enc.Encode(input)
		if err != nil {
			return client.NewError("failed to encode the request body", err)
		}
		req, err = http.NewRequest(method, resolved.String(), buf)
		if err != nil {
			return client.NewError("failed to create HTTP request", err)
		}
	} else {
		req, err = http.NewRequest(method, resolved.String(), nil)
		if err != nil {
			return client.NewError("failed to create HTTP request", err)
		}
	}

	req.Header.Set(AuthHeader, "Bearer "+cred)
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
			return client.NewError(fmt.Sprintf("failed to decode successfull HTTP response (%s)", resp.Status), err)
		}
	}

	return nil
}

//createPath is a convienent wrapper for creating a resource path prefixed by the project namespace.
func createPath(projectID string, elem ...string) string {
	return path.Join(projectsPrefix, projectID, path.Join(elem...))
}
