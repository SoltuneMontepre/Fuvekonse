package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"fuvekonse/sqs-worker/config"
	"fuvekonse/sqs-worker/processor"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const pollInterval = 5 * time.Second

func Local() {
	queueURL := config.GetEnvOr("SQS_QUEUE_URL", os.Getenv("SQS_QUEUE"))
	if queueURL == "" {
		log.Fatal("SQS_QUEUE_URL or SQS_QUEUE is required for local mode")
	}

	if err := config.ValidateDBEnv(); err != nil {
		log.Fatalf("DB config: %v (use same DB_* as general-service)", err)
	}

	g, err := getDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Local SQS worker started. Queue: %s (writing to database directly)", queueURL)

	client, err := newSQSClientForLocal(queueURL)
	if err != nil {
		log.Fatalf("Failed to create SQS client: %v", err)
	}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		output, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     5,
			VisibilityTimeout:   30,
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
			processCtx, processCancel := context.WithTimeout(context.Background(), 25*time.Second)
			err := processor.ProcessTicketJob(processCtx, g, body)
			processCancel()

			if err != nil {
				log.Printf("[%s] Process failed: %v", *msg.MessageId, err)
				if processor.IsPermanentError(err) {
					log.Printf("[%s] Permanent error, deleting message to avoid infinite retry", *msg.MessageId)
					delCtx, delCancel := context.WithTimeout(context.Background(), 10*time.Second)
					_, _ = client.DeleteMessage(delCtx, &sqs.DeleteMessageInput{
						QueueUrl:      aws.String(queueURL),
						ReceiptHandle: msg.ReceiptHandle,
					})
					delCancel()
				}
				continue
			}

			log.Printf("[%s] Processed ticket job successfully", *msg.MessageId)
			delCtx, delCancel := context.WithTimeout(context.Background(), 10*time.Second)
			_, err = client.DeleteMessage(delCtx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: msg.ReceiptHandle,
			})
			delCancel()
			if err != nil {
				log.Printf("DeleteMessage %s failed: %v (message may be processed again)", *msg.MessageId, err)
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
