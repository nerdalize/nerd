package client

import (
	"path"

	"github.com/dghubble/sling"
	v2payload "github.com/nerdalize/nerd/nerd/payload/v2"
)

//CreateQueue creates registers this client as workable capacity
func (nerdapi *NerdAPIClient) CreateQueue() (worker *v2payload.CreateQueueOutput, err error) {
	worker = &v2payload.CreateQueueOutput{}
	url := nerdapi.url(path.Join(queuesEndpoint))
	s := sling.New().Post(url)
	err = nerdapi.doRequest(s, worker)
	return
}
