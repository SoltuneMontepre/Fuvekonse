#!/bin/bash
set -x

awslocal sqs create-queue --queue-name fuvekon-queue

set +x