package nerd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

type User struct {
	// Username  string
	// Password  string
	NlzToken  string
	AWSBucket string
}

func GetCurrentUser() *User {
	//TODO: Get nlz token from env variables
	return &User{
		// Username:  "",
		// Password:  "",
		NlzToken:  "",
		AWSBucket: "boris.nerdalize.net",
	}
}

func (user *User) GetAWSCredentials() (*credentials.Credentials, error) {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKeyID == "" {
		return credentials.AnonymousCredentials, fmt.Errorf("please set AWS Access Key in AWS_ACCESS_KEY_ID environment variable")
	}
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretAccessKey == "" {
		return credentials.AnonymousCredentials, fmt.Errorf("please set AWS Access Key Secret in AWS_SECRET_ACCESS_KEY environment variable")
	}

	return credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""), nil
}
