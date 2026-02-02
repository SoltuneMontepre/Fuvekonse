package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"fuvekonse/sqs-worker/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const pollInterval = 5 * time.Second

func Local() {
	queueURL := config.GetEnvOr("SQS_QUEUE_URL", os.Getenv("SQS_QUEUE"))
	baseURL := config.GetEnvOr("GENERAL_SERVICE_URL", "")
	apiKey := config.GetEnvOr("INTERNAL_API_KEY", "")

	if queueURL == "" {
		log.Fatal("SQS_QUEUE_URL or SQS_QUEUE is required for local mode")
	}
	if baseURL == "" || apiKey == "" {
		log.Fatal("GENERAL_SERVICE_URL and INTERNAL_API_KEY are required for local mode (to process ticket jobs)")
	}

	jobURL := strings.TrimSuffix(baseURL, "/") + "/internal/jobs/ticket"
	log.Printf("Local SQS worker started. Queue: %s, Job URL: %s", queueURL, jobURL)

	client, err := newSQSClientForLocal(queueURL)
	if err != nil {
		log.Fatalf("Failed to create SQS client: %v", err)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		output, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     5,
			VisibilityTimeout:    30,
		})
		cancel()

		if err != nil {
			log.Printf("ReceiveMessage error: %v", err)
			time.Sleep(pollInterval)
			continue
		}

		for _, msg := range output.Messages {
			if msg.MessageId == nil || msg.Body == nil || msg.ReceiptHandle == nil {
				continue
			}
			body := []byte(*msg.Body)
			ok := processOneMessage(httpClient, jobURL, apiKey, body, *msg.MessageId)
			if ok {
				delCtx, delCancel := context.WithTimeout(context.Background(), 10*time.Second)
				_, err := client.DeleteMessage(delCtx, &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(queueURL),
					ReceiptHandle: msg.ReceiptHandle,
				})
				delCancel()
				if err != nil {
					log.Printf("DeleteMessage %s failed: %v (message may be processed again)", *msg.MessageId, err)
				}
			}
		}

		time.Sleep(time.Second)
	}
}

func newSQSClientForLocal(queueURL string) (*sqs.Client, error) {
	region := config.GetEnvOr("AWS_REGION", "ap-southeast-1")
	useLocalStack := config.GetEnvOr("USE_LOCALSTACK", "") == "true" ||
		strings.Contains(queueURL, "localhost") || strings.Contains(queueURL, "localstack")

	ctx := context.Background()
	var cfg aws.Config
	var err error

	if useLocalStack {
		endpoint := config.GetEnvOr("LOCALSTACK_ENDPOINT", "http://localhost:4566")
		cfg, err = awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(region),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
			awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpoint, SigningRegion: region}, nil
			})),
		)
	} else {
		cfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	}
	if err != nil {
		return nil, err
	}

	return sqs.NewFromConfig(cfg), nil
}

func processOneMessage(client *http.Client, jobURL, apiKey string, body []byte, messageID string) bool {
	req, err := http.NewRequest(http.MethodPost, jobURL, bytes.NewReader(body))
	if err != nil {
		log.Printf("[%s] Failed to create request: %v", messageID, err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(internalAPIKeyHeader, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[%s] Failed to call general-service: %v", messageID, err)
		return false
	}
	resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[%s] general-service returned %d", messageID, resp.StatusCode)
		return false
	}
	log.Printf("[%s] Processed ticket job successfully", messageID)
	return true
}
