#!/bin/sh
set +x

. scripts/template.sh $1

sam package \
    --template-file function.yaml \
    --output-template-file serverless-output.yaml \
    --s3-bucket aws-ci-cleaner

sam deploy \
    --template-file serverless-output.yaml \
    --stack-name aws-ci-cleaner \
    --capabilities CAPABILITY_IAM
