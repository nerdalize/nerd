package v1auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/nerdalize/nerd/nerd/client"
	v1payload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"
)

//TokenClient is used for a bunch of operations on the auth API that don't require oauth authentication.
type TokenClient struct {
	TokenClientConfig
}

//TokenClientConfig is the config for TokenClient
type TokenClientConfig struct {
	Doer   Doer
	Base   *url.URL
	Logger client.Logger
}

//NewTokenClient creates a new TokenClient.
func NewTokenClient(c TokenClientConfig) *TokenClient {
	if c.Doer == nil {
		c.Doer = http.DefaultClient
	}
	if c.Base.Path != "" && c.Base.Path[len(c.Base.Path)-1] != '/' {
		c.Base.Path = c.Base.Path + "/"
	}
	return &TokenClient{c}
}

//doRequest requests the server and decodes the output into the `output` field.
//
//When a status code >= 400 is returned by the server the returned error will be of type HTTPError.
func (c *TokenClient) doRequest(method, urlPath string, input, output interface{}) (err error) {
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

	req.Header.Set("Content-Type", "application/json")
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
			return client.NewError(fmt.Sprintf("failed to decode successfull HTTP response (%s)", resp.Status), err)
		}
	}

	return nil
}

//RefreshJWT refreshes a JWT with a refresh token
func (c *TokenClient) RefreshJWT(projectID, jwt, secret string) (output *v1payload.RefreshWorkerJWTOutput, err error) {
	output = &v1payload.RefreshWorkerJWTOutput{}
	input := &v1payload.RefreshWorkerJWTInput{
		WorkerJWT: v1payload.WorkerJWT{
			Token:  jwt,
			Secret: secret,
		},
	}
	path := fmt.Sprintf("%v/%v/refresh", tokenEndpoint, projectID)
	return output, c.doRequest(http.MethodPost, path, input, output)
}

//RevokeJWT revokes a JWT
func (c *TokenClient) RevokeJWT(projectID, jwt, secret string) (output *v1payload.RefreshWorkerJWTOutput, err error) {
	output = &v1payload.RefreshWorkerJWTOutput{}
	input := &v1payload.RefreshWorkerJWTInput{
		WorkerJWT: v1payload.WorkerJWT{
			Token:  jwt,
			Secret: secret,
		},
	}
	path := fmt.Sprintf("%v/%v/revoke", tokenEndpoint, projectID)
	return output, c.doRequest(http.MethodPost, path, input, output)
}
