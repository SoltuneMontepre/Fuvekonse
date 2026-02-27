package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"

	"fuvekonse/sqs-worker/config"

	"github.com/aws/aws-lambda-go/events"
)

const internalAPIKeyHeader = "X-Internal-Api-Key"

func handler(request events.SQSEvent) (events.SQSEventResponse, error) {
	log.Printf("Received %d SQS messages", len(request.Records))

	baseURL := config.GetEnvOr("GENERAL_SERVICE_URL", "")
	apiKey := config.GetEnvOr("INTERNAL_API_KEY", "")
	if baseURL == "" || apiKey == "" {
		log.Printf("GENERAL_SERVICE_URL or INTERNAL_API_KEY not set; skipping ticket job processing")
		return events.SQSEventResponse{}, nil
	}

	jobURL := strings.TrimSuffix(baseURL, "/") + "/internal/jobs/ticket"
	log.Printf("Calling general-service: %s", jobURL)
	client := &http.Client{}

	var batchItemFailures []events.SQSBatchItemFailure
	for _, record := range request.Records {
		body := []byte(record.Body)
		req, err := http.NewRequest(http.MethodPost, jobURL, bytes.NewReader(body))
		if err != nil {
			log.Printf("Failed to create request for message %s: %v", record.MessageId, err)
			batchItemFailures = append(batchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: record.MessageId})
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(internalAPIKeyHeader, apiKey)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to call general-service for message %s: %v", record.MessageId, err)
			batchItemFailures = append(batchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: record.MessageId})
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			log.Printf("general-service returned %d for message %s; body: %s", resp.StatusCode, record.MessageId, string(body))
			batchItemFailures = append(batchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: record.MessageId})
			continue
		}
		resp.Body.Close()
		log.Printf("Processed ticket job message %s successfully", record.MessageId)
	}

	return events.SQSEventResponse{BatchItemFailures: batchItemFailures}, nil
}
