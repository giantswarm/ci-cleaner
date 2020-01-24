package aws

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	// gracePeriod represents the maximum time the CI resources are allowed to
	// remain up. CI resources older than gracePeriod will be deleted.
	gracePeriod = 90 * time.Minute
)

// EC2Client describes the methods required to be implemented by a EC2
// AWS client.
type EC2Client interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	ModifyInstanceAttribute(*ec2.ModifyInstanceAttributeInput) (*ec2.ModifyInstanceAttributeOutput, error)
}

// CFClient describes the methods required to be implemented by a CloudFormation
// AWS client.
type CFClient interface {
	DeleteStack(*cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error)
	DescribeStacks(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
	UpdateTerminationProtection(*cloudformation.UpdateTerminationProtectionInput) (*cloudformation.UpdateTerminationProtectionOutput, error)
}

type Route53Client interface {
	ListHostedZones(input *route53.ListHostedZonesInput) (*route53.ListHostedZonesOutput, error)
}

// S3Client describes the methods required to be implemented by a S3 AWS
// client.
type S3Client interface {
	ListBuckets(*s3.ListBucketsInput) (*s3.ListBucketsOutput, error)
	DeleteBucket(*s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error)
	ListObjectsV2(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
	DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	DeleteObjects(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error)
}
