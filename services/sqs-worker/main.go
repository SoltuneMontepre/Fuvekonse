package main

import (
	"fuvekonse/sqs-worker/config"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	config.LoadEnv()

	if config.IsLambdaEnv() {
		lambda.Start(handler)
	} else {
		Local()
	}
}
