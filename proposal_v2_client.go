package client

type Client struct {
	Auth AuthClient
	Nerd NerdClient
	AWS  AWSClient
}

type NerdClient struct{}

type AuthClient struct{}

type AWSStorageBackend interface {
	DownloadObject()
	UploadObject()
}

type AWSQueueBackend interface {
	ReceiveMessages() [][]byte
}

type AWSBackend interface {
	AWSQueueBackend
	AWSStorageBackend
}

func NewRealAWSBackend(sess aws.Session)
