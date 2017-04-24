package client

type Client struct {
	Auth AuthClient
	Nerd NerdClient
	AWS  AWSClient
}

type NerdClient struct {
	Config NerdClientConfig
}

type NerdClientConfig struct {
	Credentials NerdClientCredentials
}

type NerdClientCredentialProvider interface {
	GetToken() (JWT string)
	IsExpired() bool
}

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
