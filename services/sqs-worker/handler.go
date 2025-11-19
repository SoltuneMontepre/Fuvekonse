package main

import (
	"log"

	"github.com/aws/aws-lambda-go/events"
)

func handler(request events.SQSEvent) (events.SQSEventResponse, error) {
	log.Println("Running in Lambda...")
	log.Printf("Received %d SQS messages", len(request.Records))
	log.Println("Processing messages...")
	log.Println("Messages processed successfully")
	return events.SQSEventResponse{}, nil
}
