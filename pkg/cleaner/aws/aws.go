package aws

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/ci-cleaner/pkg/errorcollection"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type Config struct {
	EC2Client     EC2Client
	CFClient      CFClient
	Logger        micrologger.Logger
	Route53Client Route53Client
	S3Client      S3Client
}

type Cleaner struct {
	ec2Client     EC2Client
	cfClient      CFClient
	logger        micrologger.Logger
	route53Client Route53Client
	s3Client      S3Client
}

func New(config *Config) (*Cleaner, error) {
	if config.CFClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CFClient must not be empty", config)
	}
	if config.EC2Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ec2lient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Route53Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Route53Client must not be empty", config)
	}
	if config.S3Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.S3Client must not be empty", config)
	}

	cleaner := &Cleaner{
		ec2Client:     config.EC2Client,
		cfClient:      config.CFClient,
		logger:        config.Logger,
		route53Client: config.Route53Client,
		s3Client:      config.S3Client,
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
		// NOTE this can be enable when needed for further cleanups.
		// a.cleanHostedZones,
	}

	errors := &errorcollection.ErrorCollection{}

	for _, f := range cleaners {
		a.logger.Log("level", "info", "message", fmt.Sprintf("running cleaner %s", getFunctionName(f)))
		err := f()
		if err != nil {
			a.logger.Log("level", "error", "message", fmt.Sprintf("running cleaner %s", getFunctionName(f)), "stack", fmt.Sprintf("%#v", err))
			errors.Append(err)
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
			continue
		}

		a.logger.Log("level", "info", "message", fmt.Sprintf("found that stack %#q should be deleted", *stack.StackName))

		if isTenantStack(stack) {
			a.logger.Log("level", "debug", "message", fmt.Sprintf("disabling termination protection for EC2 instance belonging to the stack %#q", *stack.StackName))
			err = a.disableMasterTerminationProtection(*stack.StackName)
			if err != nil {
				errors.Append(microerror.Mask(err))
				// do not return on error, try to continue deleting.
				a.logger.Log("level", "error", "message", fmt.Sprintf("failed disabling termination protection for EC2 instance belonging to the stack %#q: %#v. Skipping deletion.", *stack.StackName, err))
				continue
			}
		}

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
			a.logger.Log("level", "error", "message", fmt.Sprintf("failed disabling termination protection for %#q: %#v. Skipping deletion.", *stack.StackName, err))
			continue
		}

		deleteStackInput := &cloudformation.DeleteStackInput{
			StackName: stack.StackName,
		}
		_, err := a.cfClient.DeleteStack(deleteStackInput)
		if err != nil {
			errors.Append(microerror.Mask(err))
			// do not return on error, try to continue deleting.
			a.logger.Log("level", "error", "message", fmt.Sprintf("failed deleting stack %#q: %s", *stack.StackName, err.Error()), "stack", fmt.Sprintf("%#v", err))
			a.logger.Log("level", "debug", "message", fmt.Sprintf("stack details: %#v", stack))
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
			continue
		}
		a.logger.Log("level", "debug", "message", fmt.Sprintf("found that bucket %#q should be deleted", *bucket.Name))
		err := a.deleteBucket(bucket.Name)
		if err != nil {
			errors.Append(microerror.Mask(err))
			a.logger.Log("level", "error", "message", fmt.Sprintf("failed deleting bucket %#q: %#v", *bucket.Name, err), "stack", fmt.Sprintf("%#v", err))
		} else {
			a.logger.Log("level", "info", "message", fmt.Sprintf("deleted bucket %#q", *bucket.Name))
		}
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

func (a *Cleaner) cleanHostedZones() error {
	var marker *string
	for {
		in := &route53.ListHostedZonesInput{
			Marker: marker,
		}

		out, err := a.route53Client.ListHostedZones(in)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, hz := range out.HostedZones {
			if hz.Name == nil || hz.Id == nil {
				continue
			}

			fmt.Printf("\n")
			fmt.Printf("%#v\n", *hz.Id)
			fmt.Printf("%#v\n", *hz.Name)
			fmt.Printf("%#v\n", hz)
			fmt.Printf("\n")
		}

		if out.IsTruncated == nil || !*out.IsTruncated {
			break
		} else {
			marker = out.Marker
		}
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
		"ci-",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(*stack.StackName, prefix) {
			return true
		}
	}

	return false
}

func isTenantStack(stack *cloudformation.Stack) bool {
	outputs := stack.Outputs
	for _, o := range outputs {
		if *o.OutputKey == "MasterImageID" {
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

	patterns := []string{
		`\Aci-last-.*`,
		`\Aci-prev-.*`,
		`\Aci-cur-.*`,
		`\Aci-wip-.*`,
		`g8s-ci-cur-.*`,
		`g8s-ci-wip-.*`,
		`g8s-ci-clop-.*`,
		`\Aci-.*-g8s-access-logs\z`,
		`.*-g8s-ci-.*`,
	}
	for _, pattern := range patterns {
		matches, _ := regexp.MatchString(pattern, *bucket.Name)
		if matches {
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

		//batch up the objects for deletion
		var objects []*s3.ObjectIdentifier
		for _, o := range o.Contents {
			objects = append(objects, &s3.ObjectIdentifier{
				Key: o.Key,
			})
		}
		di := &s3.DeleteObjectsInput{
			Bucket: name,
			Delete: &s3.Delete{
				Objects: objects,
				Quiet:   aws.Bool(true),
			},
		}
		//delete the batch
		_, err = a.s3Client.DeleteObjects(di)
		if err != nil {
			return microerror.Mask(err)
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

func (a *Cleaner) disableMasterTerminationProtection(stackName string) error {

	i := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:aws:cloudformation:stack-name"),
				Values: []*string{
					aws.String(stackName),
				},
			},
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String("*-master"),
				},
			},
		},
	}
	o, err := a.ec2Client.DescribeInstances(i)
	if err != nil {
		return microerror.Mask(err)
	}

	// If there are no masters we can stop here.
	if len(o.Reservations) == 0 {
		return nil
	}

	for _, reservation := range o.Reservations {

		if len(reservation.Instances) != 1 {
			return microerror.Newf("Expected one master instance, got %d", len(reservation.Instances))
		}

		for _, instance := range reservation.Instances {
			i := &ec2.ModifyInstanceAttributeInput{
				DisableApiTermination: &ec2.AttributeBooleanValue{
					Value: aws.Bool(false),
				},
				InstanceId: aws.String(*instance.InstanceId),
			}

			_, err = a.ec2Client.ModifyInstanceAttribute(i)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}
