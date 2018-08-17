package aws

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	// gracePeriod represents the maximum time the CI resources are allowed to
	// remain up. CI resources older than gracePeriod will be deleted.
	gracePeriod = 90 * time.Minute
)

// CFClient describes the methods required to be implemented by a CloudFormation
// AWS client.
type CFClient interface {
	DeleteStack(*cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error)
	DescribeStacks(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
}

// S3Client describes the methods required to be implemented by a S3 AWS
// client.
type S3Client interface {
	ListBuckets(*s3.ListBucketsInput) (*s3.ListBucketsOutput, error)
	DeleteBucket(*s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error)
	ListObjectsV2(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
	DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
}
