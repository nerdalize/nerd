package nerd

//Session returns session information for authorization purposes
type Session struct {
	AWSAccessKeyID     string `json:"aws_access_key_id"`
	AWSSecretAccessKey string `json:"aws_secret_access_key"`
	AWSSQSQueueURL     string `json:"aws_sqs_queue_url"`
	AWSRegion          string `json:"aws_region"`
}
