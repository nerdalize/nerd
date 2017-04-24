package client

type Client struct {
	Auth AuthClient
	Nerd NerdClient
	AWS  AWSClient
}
