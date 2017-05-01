package aws

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
)

//QueueClient is a client to AWS' SQS queueing service. The client implements the v1batch.QueueOps interface.
type QueueClient struct {
	Service *sqs.SQS
}

//NewQueueClient creates a new QueueClient
func NewQueueClient(c *credentials.Credentials, region string) (*QueueClient, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: c,
		Region:      aws.String(region),
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not create AWS sessions")
	}
	return &QueueClient{
		Service: sqs.New(sess),
	}, nil
}

//ReceiveMessages receives messages from the queue.
func (c *QueueClient) ReceiveMessages(queueURL string, maxNoOfMessages, waitTimeSeconds int64) (messages []interface{}, err error) {
	out, err := c.Service.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: aws.Int64(maxNoOfMessages),
		WaitTimeSeconds:     aws.Int64(waitTimeSeconds),
	})
	if err != nil {
		return nil, err
	}
	ret := make([]interface{}, len(out.Messages))
	for _, msg := range out.Messages {
		ret = append(ret, msg)
	}
	return ret, nil
}

//UnmarshalMessage decodes a message.
func (c *QueueClient) UnmarshalMessage(message interface{}, v interface{}) error {
	msg, ok := message.(*sqs.Message)
	if !ok {
		return errors.New("message was not of type *sqs.Message")
	}
	return json.Unmarshal([]byte(aws.StringValue(msg.Body)), v)
}

//DeleteMessage deletes a message from the queue.
func (c *QueueClient) DeleteMessage(queueURL string, message interface{}) error {
	msg, ok := message.(*sqs.Message)
	if !ok {
		return errors.New("message was not of type *sqs.Message")
	}
	_, err := c.Service.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})
	return err
}
