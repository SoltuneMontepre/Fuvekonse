#!/bin/bash
set -x

sleep 2

ROLE_ARN="arn:aws:iam::000000000000:role/lambda-execution-role"

if [ -f "/tmp/lambda/general-service.zip" ]; then
  echo "Deploying general-service..."
  awslocal lambda create-function \
    --function-name general-service \
    --runtime provided.al2 \
    --role $ROLE_ARN \
    --handler bootstrap \
    --zip-file fileb:///tmp/lambda/general-service.zip \
    --timeout 30 \
    --memory-size 512 \
    --environment "Variables={DB_HOST=fuvekon-db,REDIS_HOST=fuvekon-cache,S3_BUCKET_URL=http://localstack:4566/fuvekon-bucket,SQS_QUEUE_URL=http://localstack:4566/000000000000/fuvekon-queue,PORT=8085}" \
    || echo "general-service Lambda already exists or failed to create"
  
else
  echo "general-service.zip not found - skipping"
fi

# Deploy ticket-service Lambda
if [ -f "/tmp/lambda/ticket-service.zip" ]; then
  echo "Deploying ticket-service..."
  awslocal lambda create-function \
    --function-name ticket-service \
    --runtime provided.al2 \
    --role $ROLE_ARN \
    --handler bootstrap \
    --zip-file fileb:///tmp/lambda/ticket-service.zip \
    --timeout 30 \
    --memory-size 512 \
    --environment "Variables={DB_HOST=fuvekon-db,REDIS_HOST=fuvekon-cache,S3_BUCKET_URL=http://localstack:4566/fuvekon-bucket,SQS_QUEUE_URL=http://localstack:4566/000000000000/fuvekon-queue,PORT=8085}" \
    || echo "ticket-service Lambda already exists or failed to create"

  echo "ticket-service deployed"
else
  echo "ticket-service.zip not found - skipping"
fi

echo "Lambda deployment completed"

awslocal lambda list-functions --query 'Functions[*].[FunctionName,Runtime,LastModified]' --output table

set +x
