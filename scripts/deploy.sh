#!/bin/sh
set +x

type=${1:-guest}

sam package \
    --template-file template.yaml \
    --output-template-file serverless-output.yaml \
    --s3-bucket ci-cleaner-$type

sam deploy \
    --template-file serverless-output.yaml \
    --stack-name aws-ci-cleaner \
    --capabilities CAPABILITY_IAM
