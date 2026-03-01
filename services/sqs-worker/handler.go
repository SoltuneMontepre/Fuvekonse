package main

import (
	"context"
	"log"
	"sync"

	"fuvekonse/sqs-worker/config"
	"fuvekonse/sqs-worker/db"
	"fuvekonse/sqs-worker/processor"

	"github.com/aws/aws-lambda-go/events"
	"gorm.io/gorm"
)

var (
	dbOnce sync.Once
	gormDB *gorm.DB
	dbErr  error
)

func getDB() (*gorm.DB, error) {
	dbOnce.Do(func() {
		if dbErr = config.ValidateDBEnv(); dbErr != nil {
			return
		}
		gormDB, dbErr = db.Connect(config.DatabaseDSN())
		if dbErr != nil {
			return
		}
		if dbErr = db.AutoMigrate(gormDB); dbErr != nil {
			return
		}
		dbErr = db.ValidateSchema(context.Background(), gormDB)
	})
	return gormDB, dbErr
}

func handler(request events.SQSEvent) (events.SQSEventResponse, error) {
	log.Printf("Received %d SQS messages", len(request.Records))

	g, err := getDB()
	if err != nil {
		log.Printf("Database init failed: %v", err)
		var batchItemFailures []events.SQSBatchItemFailure
		for _, record := range request.Records {
			batchItemFailures = append(batchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: record.MessageId})
		}
		return events.SQSEventResponse{BatchItemFailures: batchItemFailures}, nil
	}

	ctx := context.Background()
	var batchItemFailures []events.SQSBatchItemFailure

	for _, record := range request.Records {
		body := []byte(record.Body)
		err := processor.ProcessTicketJob(ctx, g, body)
		if err != nil {
			log.Printf("Message %s: %v", record.MessageId, err)
			if processor.IsPermanentError(err) {
				log.Printf("Message %s: permanent error, not retrying", record.MessageId)
				continue
			}
			batchItemFailures = append(batchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: record.MessageId})
			continue
		}
		log.Printf("Processed ticket job message %s successfully", record.MessageId)
	}

	return events.SQSEventResponse{BatchItemFailures: batchItemFailures}, nil
}
