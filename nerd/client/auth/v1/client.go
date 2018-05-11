package v1auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/nerdalize/nerd/nerd/client"
	v1payload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"
	log "github.com/sirupsen/logrus"
)

const (
	authHeader = "Authorization"
	//TokenEndpoint is the endpoint from where to fetch the JWT.
	tokenEndpoint    = "token"
	projectsEndpoint = "projects"
	clustersEndpoint = "clusters"
	//NCEScope is the JWT scope for the NCE service
	NCEScope = "nce.nerdalize.com"
)

//Client is the client for the nerdalize authentication server.
type Client struct {
	ClientConfig
	cred string
	m    sync.Mutex
}

//ClientConfig provides config details to create an Auth client.
type ClientConfig struct {
	Doer               Doer
	Base               *url.URL
	OAuthTokenProvider OAuthTokenProvider
	Logger             *log.Logger
}

// Doer executes http requests.  It is implemented by *http.Client.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

//NewClient creates a new Auth.
func NewClient(c ClientConfig) *Client {
	if c.Doer == nil {
		c.Doer = http.DefaultClient
	}
	if c.Base.Path != "" && c.Base.Path[len(c.Base.Path)-1] != '/' {
		c.Base.Path = c.Base.Path + "/"
	}
	return &Client{
		ClientConfig: c,
		cred:         "",
		m:            sync.Mutex{},
	}
}

//getOAuthAccessToken gets the oauth access token to authenticate
func (c *Client) getOAuthAccessToken() (string, error) {
	c.m.Lock()
	defer c.m.Unlock()
	if c.OAuthTokenProvider == nil {
		return "", fmt.Errorf("no oauth token provider provider found")
	}
	if c.cred == "" || c.OAuthTokenProvider.IsExpired() {
		cred, err := c.OAuthTokenProvider.Retrieve()
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
	cred, err := c.getOAuthAccessToken()
	if err != nil {
		return err
	}

	path, err := url.Parse(urlPath)
	if err != nil {
		return client.NewError("invalid url path provided", err)
	}

	resolved := c.Base.ResolveReference(path)

	var req *http.Request
	if input != nil && method != http.MethodGet {
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

	req.Header.Set(authHeader, "Bearer "+cred)
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

//ListClusters lists clusters
func (c *Client) ListClusters() (output *v1payload.ListClustersOutput, err error) {
	output = &v1payload.ListClustersOutput{}
	return output, c.doRequest(http.MethodGet, clustersEndpoint, nil, &output.Clusters)
}

//GetCluster retrieve a precised cluster so we can get its authentication details.
func (c *Client) GetCluster(url string) (output *v1payload.GetClusterOutput, err error) {
	output = &v1payload.GetClusterOutput{}
	return output, c.doRequest(http.MethodGet, url, nil, output)
}

//ListProjects lists projects
func (c *Client) ListProjects() (output *v1payload.ListProjectsOutput, err error) {
	output = &v1payload.ListProjectsOutput{}
	return output, c.doRequest(http.MethodGet, projectsEndpoint, nil, &output.Projects)
}

//GetProject retrieve a precised project so we can validate its existence and find on which cluster it's living.
func (c *Client) GetProject(id string) (output *v1payload.GetProjectOutput, err error) {
	output = &v1payload.GetProjectOutput{}
	return output, c.doRequest(http.MethodGet, fmt.Sprintf("%s/%s", projectsEndpoint, id), nil, output)
}

//GetJWT gets a JWT for a given scope
func (c *Client) GetJWT(scope string) (output *v1payload.GetJWTOutput, err error) {
	output = &v1payload.GetJWTOutput{}
	path := tokenEndpoint + "?service=" + scope
	return output, c.doRequest(http.MethodGet, path, nil, output)
}

//GetWorkerJWT gets a new worker JWT
func (c *Client) GetWorkerJWT(project, scope string) (output *v1payload.GetWorkerJWTOutput, err error) {
	output = &v1payload.GetWorkerJWTOutput{}
	path := fmt.Sprintf("%v/%v/worker/?service=%v", tokenEndpoint, project, scope)
	return output, c.doRequest(http.MethodPost, path, nil, output)
}
