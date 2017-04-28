package v1batch

import "net/http"

//ClientPingInterface is an interface so client ping calls can be mocked.
type ClientPingInterface interface {
	Ping() error
}

//Ping will error if there are connection issues
func (c *Client) Ping() error {
	return c.doRequest(http.MethodGet, "/ping", nil, nil)
}
