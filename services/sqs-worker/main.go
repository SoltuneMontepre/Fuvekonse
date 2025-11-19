package main

import (
	"worker/config"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	if config.IsLambdaEnv() {
		lambda.Start(handler)
	} else {
		Local()
	}
}
