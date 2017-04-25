package v1batch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"

	"github.com/nerdalize/nerd/nerd/client"
	v2payload "github.com/nerdalize/nerd/nerd/payload/v2"
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

//Nerd is a client for the Nerdalize API.
type Client struct {
	ClientConfig
	cred string
}

//NerdConfig provides config details to create a Nerd client.
type ClientConfig struct {
	Doer        Doer
	JWTProvider JWTProvider
	Base        *url.URL
	Logger      Logger
	QueueOps    QueueOps
}

// Doer executes http requests.  It is implemented by *http.Client.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// QueueOps is an interface that includes queue operations.
type QueueOps interface {
	ReceiveMessages(queueURL string, maxNoOfMessages, waitTimeSeconds int) (messages []interface{}, err error)
	UnmarshalMessage(message interface{}, v interface{}) error
	DeleteMessage(queueURL string, message interface{}) error
}

//NewNerdClient creates a new Nerd client from a config object. The http.DefaultClient
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
	}
	return cl
}

func (c *Client) getCredentials() (string, error) {
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

func (c *Client) doRequest(method, urlPath string, input, output interface{}) (err error) {
	cred, err := c.getCredentials()
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	if input != nil {
		enc := json.NewEncoder(buf)
		err = enc.Encode(input)
		if err != nil {
			return &client.Error{"failed to encode the request body", err}
		}
	}

	path, err := url.Parse(urlPath)
	if err != nil {
		return &client.Error{"invalid url path provided", err}
	}

	resolved := c.Base.ResolveReference(path)
	req, err := http.NewRequest(method, resolved.String(), buf)
	logRequest(req, c.Logger)
	if err != nil {
		return &client.Error{"failed to create HTTP request", err}
	}

	req.Header.Set(AuthHeader, "Bearer "+cred)
	resp, err := c.Doer.Do(req)
	if err != nil {
		return &client.Error{"failed to perform HTTP request", err}
	}
	logResponse(resp, c.Logger)

	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		errv := &v2payload.Error{}
		err = dec.Decode(errv)
		if err != nil {
			return &client.Error{fmt.Sprintf("failed to decode unexpected HTTP response (%s)", resp.Status), err}
		}

		return &HTTPError{
			StatusCode: resp.StatusCode,
			Err:        errv,
		}
	}

	if output != nil {
		err = dec.Decode(output)
		if err != nil {
			return &client.Error{fmt.Sprintf("failed to decode successfull HTTP response (%s)", resp.Status), err}
		}
	}

	return nil
}

func createPath(projectID string, elem ...string) string {
	return path.Join(projectsPrefix, projectID, path.Join(elem...))
}

type Logger interface {
	Debugf(format string, args ...interface{})
	Error(args ...interface{})
}

func logRequest(req *http.Request, logger Logger) {
	txt, err := httputil.DumpRequest(req, true)
	// retry without printing the body
	if err != nil {
		txt, err = httputil.DumpRequest(req, false)
	}
	if err == nil {
		logger.Debugf("HTTP Request:\n%s", txt)
	} else {
		logger.Error("Failed to log HTTP request")
	}
}

func logResponse(res *http.Response, logger Logger) {
	txt, err := httputil.DumpResponse(res, true)
	// retry without printing the body
	if err != nil {
		txt, err = httputil.DumpResponse(res, false)
	}
	if err == nil {
		logger.Debugf("HTTP Response:\n%s", txt)
	} else {
		logger.Error("Failed to log HTTP response")
	}
}
