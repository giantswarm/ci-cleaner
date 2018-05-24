AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31
Resources:
  CICleanerFunction:
    Type: AWS::Serverless::Function
    Properties:
      Handler: aws-ci-cleaner
      Runtime: go1.x
      Events:
        Schedule1:
          Type: Schedule
          Properties:
            Schedule: rate(1 hour)
      Environment:
        Variables:
          S3_BUCKET: ci-cleaner-{{type}}
