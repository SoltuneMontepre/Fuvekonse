#!/bin/bash
set -x

awslocal ses verify-email-identity --email fuveSupport@fuve.com

set +x