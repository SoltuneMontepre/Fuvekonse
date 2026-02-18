package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Publisher sends messages to SQS.
type Publisher interface {
	PublishTicketJob(ctx context.Context, msg *TicketJobMessage) error
}

// SQSClient wraps the AWS SQS client for publishing.
type SQSClient struct {
	client   *sqs.Client
	queueURL string
}

// NewSQSClient creates an SQS client. If SQS_QUEUE_URL (or SQS_QUEUE) is empty, returns nil (queue disabled).
func NewSQSClient(ctx context.Context) (*SQSClient, error) {
	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		queueURL = os.Getenv("SQS_QUEUE")
	}
	if queueURL == "" {
		log.Println("SQS_QUEUE_URL and SQS_QUEUE not set; ticket queue disabled")
		return nil, nil
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-southeast-1"
	}

	// Use LocalStack when explicitly set or when queue URL points to localhost/localstack
	useLocalStack := os.Getenv("USE_LOCALSTACK") == "true" ||
		strings.Contains(queueURL, "localhost") || strings.Contains(queueURL, "localstack")
	var cfg aws.Config
	var err error

	if useLocalStack {
		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if accessKey == "" {
			accessKey = "test"
		}
		if secretKey == "" {
			secretKey = "test"
		}
		localEndpoint := os.Getenv("LOCALSTACK_ENDPOINT")
		if localEndpoint == "" {
			localEndpoint = "http://localhost:4566"
		}
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           localEndpoint,
					SigningRegion: region,
				}, nil
			})),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}
	if err != nil {
		return nil, err
	}

	return &SQSClient{
		client:   sqs.NewFromConfig(cfg),
		queueURL: queueURL,
	}, nil
}

// PublishTicketJob sends a ticket job message to the queue.
func (c *SQSClient) PublishTicketJob(ctx context.Context, msg *TicketJobMessage) error {
	if c == nil {
		log.Printf("ERROR: PublishTicketJob called on nil SQSClient (action=%s) — this indicates a nil interface bug", msg.Action)
		return fmt.Errorf("SQS client is nil; cannot publish job")
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = c.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.queueURL),
		MessageBody: aws.String(string(body)),
	})
	return err
}
