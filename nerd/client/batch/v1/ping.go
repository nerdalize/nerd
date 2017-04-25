package v1batch

import "net/http"

//Ping will error if there are connection issues
func (c *Client) Ping() error {
	return c.doRequest(http.MethodGet, "/ping", nil, nil)
}
