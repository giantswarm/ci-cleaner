package aws

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/ci-cleaner/pkg/errorcollection"
)

type Config struct {
	CFClient CFClient
	Logger   micrologger.Logger
	S3Client S3Client
}

type Cleaner struct {
	cfClient CFClient
	logger   micrologger.Logger
	s3Client S3Client
}

func New(config *Config) (*Cleaner, error) {
	if config.CFClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "CFClient must not be empty")
	}
	if config.S3Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "S3Client must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	cleaner := &Cleaner{
		cfClient: config.CFClient,
		logger:   config.Logger,
		s3Client: config.S3Client,
	}

	return cleaner, nil
}

// getFunctionName returns the name of the function passed as argument.
func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// Clean calls our cleaner functions and logs errors if they happen.
// We don't return errors as we want all cleaners to be called.
func (a *Cleaner) Clean() error {
	type cleanerFn func() error

	cleaners := []cleanerFn{
		a.cleanStacks,
		a.cleanBuckets,
	}

	errors := &errorcollection.ErrorCollection{}

	for _, f := range cleaners {
		a.logger.Log("level", "debug", "message", fmt.Sprintf("running cleaner %s", getFunctionName(f)))
		err := f()
		if err != nil {
			a.logger.Log("level", "error", "message", fmt.Sprintf("%d error(s) in cleaner %s", getFunctionName(f)))

			if val, ok := err.(*errorcollection.ErrorCollection); ok {
				errors.Append(val)
			}
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

func (a *Cleaner) cleanStacks() error {
	errors := &errorcollection.ErrorCollection{}

	input := &cloudformation.DescribeStacksInput{}
	output, err := a.cfClient.DescribeStacks(input)
	if err != nil {
		errors.Append(microerror.Mask(err))
		return errors
	}

	for _, stack := range output.Stacks {
		if !stackShouldBeDeleted(stack) {
			a.logger.Log("level", "debug", "message", fmt.Sprintf("leaving stack %#q untouched: %#v", *stack.StackName, *stack))
			continue
		}
		a.logger.Log("level", "debug", "message", fmt.Sprintf("found that stack %#q should be deleted", *stack.StackName))

		if *stack.EnableTerminationProtection {
			a.logger.Log("level", "debug", "message", fmt.Sprintf("disabling termination protection for stack %#q", *stack.StackName))
			enableTerminationProtection := false
			updateTerminationProtection := &cloudformation.UpdateTerminationProtectionInput{
				EnableTerminationProtection: &enableTerminationProtection,
				StackName:                   stack.StackName,
			}
			_, err = a.cfClient.UpdateTerminationProtection(updateTerminationProtection)
			if err != nil {
				errors.Append(microerror.Mask(err))
				// do not return on error, try to continue deleting.
				a.logger.Log("level", "error", "message", fmt.Sprintf("failed disabling stack protection %#q: %#v. Skipping deletion.", *stack.StackName, err))
				continue
			}
		}

		deleteStackInput := &cloudformation.DeleteStackInput{
			StackName: stack.StackName,
		}
		_, err := a.cfClient.DeleteStack(deleteStackInput)
		if err != nil {
			errors.Append(microerror.Mask(err))
			// do not return on error, try to continue deleting.
			a.logger.Log("level", "error", "message", fmt.Sprintf("failed deleting stack %#q: %#v", *stack.StackName, err))
		} else {
			a.logger.Log("level", "info", "message", fmt.Sprintf("deleted stack %#q", *stack.StackName))
		}
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

func (a *Cleaner) cleanBuckets() error {
	errors := &errorcollection.ErrorCollection{}

	input := &s3.ListBucketsInput{}
	output, err := a.s3Client.ListBuckets(input)
	if err != nil {
		errors.Append(microerror.Mask(err))
		return errors
	}

	for _, bucket := range output.Buckets {
		if !bucketShouldBeDeleted(bucket) {
			a.logger.Log("level", "debug", "message", fmt.Sprintf("leaving bucket %#q untouched", *bucket.Name))
			continue
		}
		a.logger.Log("level", "debug", "message", fmt.Sprintf("found that bucket %#q should be deleted", *bucket.Name))
		err := a.deleteBucket(bucket.Name)
		if err != nil {
			errors.Append(microerror.Mask(err))
			a.logger.Log("level", "error", "message", fmt.Sprintf("failed deleting bucket %#q: %#v", *bucket.Name, err))
		} else {
			a.logger.Log("level", "info", "message", fmt.Sprintf("deleted bucket %#q", *bucket.Name))
		}
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

func stackShouldBeDeleted(stack *cloudformation.Stack) bool {
	if stack.CreationTime == nil {
		// bad formed stack, should be deleted
		return true
	}

	now := time.Now().UTC()
	timeDiff := now.Sub(*stack.CreationTime)

	// do not delete recent stacks.
	if timeDiff < gracePeriod {
		return false
	}

	// do not delete stacks that are already being deleted
	if *stack.StackStatus == "DELETE_IN_PROGRESS" || *stack.StackStatus == "DELETE_COMPLETE" {
		return false
	}

	prefixes := []string{
		"cluster-ci-",
		"host-peer-ci-",
		"e2e-",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(*stack.StackName, prefix) {
			return true
		}
	}

	return false
}

func bucketShouldBeDeleted(bucket *s3.Bucket) bool {
	if bucket.CreationDate == nil {
		// bad formed bucket, should be deleted
		return true
	}

	now := time.Now().UTC()
	timeDiff := now.Sub(*bucket.CreationDate)

	// do not delete recent buckets.
	if timeDiff < gracePeriod {
		return false
	}

	prefixes := []string{
		"ci-cur-",
		"ci-wip-",
		"ci-clop-",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(*bucket.Name, prefix) {
			return true
		}
	}
	substrings := []string{
		"g8s-ci-cur-",
		"g8s-ci-wip-",
		"g8s-ci-clop-",
	}
	for _, substring := range substrings {
		if strings.Contains(*bucket.Name, substring) {
			return true
		}
	}

	return false
}

func (a *Cleaner) deleteBucket(name *string) error {
	var repeat bool
	for {
		i := &s3.ListObjectsV2Input{
			Bucket: name,
		}
		o, err := a.s3Client.ListObjectsV2(i)
		if err != nil {
			return microerror.Mask(err)
		}
		if o.IsTruncated != nil && *o.IsTruncated {
			repeat = true
		}
		if len(o.Contents) == 0 {
			break
		}

		for _, o := range o.Contents {
			i := &s3.DeleteObjectInput{
				Bucket: name,
				Key:    o.Key,
			}
			_, err := a.s3Client.DeleteObject(i)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		if !repeat {
			break
		}
	}
	deleteBucketInput := &s3.DeleteBucketInput{
		Bucket: name,
	}
	_, err := a.s3Client.DeleteBucket(deleteBucketInput)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}
