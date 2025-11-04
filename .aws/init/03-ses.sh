#!/bin/bash
set -x

awslocal ses verify-email-identity --email fuve.vietnam@gmail.com

set +x