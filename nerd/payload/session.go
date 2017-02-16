package payload

import "time"

//SessionCreateOutput is returned when a user created a valid session
type SessionCreateOutput struct {
	AWSAccessKeyID     string    `json:"aws_access_key_id"`
	AWSExpiration      time.Time `json:"aws_expiration"`
	AWSSecretAccessKey string    `json:"aws_secret_access_key"`
	AWSSessionToken    string    `json:"aws_session_token"`

	AWSQueueURL      string `json:"aws_queue_url"`
	AWSStorageBucket string `json:"aws_storage_bucket"`
	AWSStorageRoot   string `json:"aws_storage_root"`
}
