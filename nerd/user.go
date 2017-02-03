package nerd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

//User holds a reference to a user session
type User struct {
	// Username  string
	// Password  string
	JWT       string
	AWSBucket string
	Region    string
}

//GetCurrentUser returns the current user session
func GetCurrentUser() *User {
	//TODO: Get jwt from env variables
	return &User{
		JWT: "",
		//TODO: This should not be hardcoded.
		AWSBucket: "boris.nerdalize.net",
		Region:    "eu-west-1",
	}
}

//GetAWSCredentials fetches the user's AWS token.
//At the moment this is a mock function that reads the credentials from
//environment variables, in the future this function would interface with the NCE API.
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
