package aws

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
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
