package client

import (
	"path"

	"github.com/dghubble/sling"
	"github.com/nerdalize/nerd/nerd/payload"
)

//CreateWorker creates registers this client as workable capacity
func (nerdapi *NerdAPIClient) CreateWorker() (worker *payload.WorkerCreateOutput, err error) {
	worker = &payload.WorkerCreateOutput{}
	url := nerdapi.url(path.Join(queuesEndpoint))
	s := sling.New().Post(url)
	err = nerdapi.doRequest(s, worker)
	return
}
