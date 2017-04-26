package v1payload

import "time"

//CreateTokenInput is input for token creation
type CreateTokenInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

//CreateTokenOutput is output for token creation
type CreateTokenOutput struct {
	AWSAccessKeyID     string    `json:"aws_access_key_id"`
	AWSExpiration      time.Time `json:"aws_expiration"`
	AWSSecretAccessKey string    `json:"aws_secret_access_key"`
	AWSSessionToken    string    `json:"aws_session_token"`
}
