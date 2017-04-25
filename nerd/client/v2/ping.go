package v2client

import "net/http"

//Ping will error if there are connection issues
func (c *Nerd) Ping() error {
	return c.doRequest(http.MethodGet, "/ping", nil, nil)
}
