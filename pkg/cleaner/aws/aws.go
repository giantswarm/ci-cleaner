package aws

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	// gracePeriod represents the maximum time the CI resources are allowed to
	// remain up. CI resources older than gracePeriod will be deleted.
	gracePeriod = 90 * time.Minute
)

type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string

	Logger *micrologger.MicroLogger
}

type Cleaner struct {
	cfClient *cloudformation.CloudFormation
	logger   *micrologger.MicroLogger
}

func New(config *Config) (*Cleaner, error) {
	if config.AccessKeyID == "" {
		return nil, microerror.Maskf(invalidConfigError, "Access Key ID must not be empty")
	}
	if config.SecretAccessKey == "" {
		return nil, microerror.Maskf(invalidConfigError, "Secret Access Key  must not be empty")
	}
	if config.Region == "" {
		return nil, microerror.Maskf(invalidConfigError, "Region must not be empty")
	}

	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	awsCfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
		Region:      aws.String(config.Region),
	}
	s, err := session.NewSession(awsCfg)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cleaner := &Cleaner{
		cfClient: cloudformation.New(s),
		logger:   config.Logger,
	}

	return cleaner, nil
}

func (a *Cleaner) Clean() error {
	input := &cloudformation.DescribeStacksInput{}
	output, err := a.cfClient.DescribeStacks(input)
	if err != nil {
		return err
	}

	for _, stack := range output.Stacks {
		a.logger.Log("level", "debug", "message", fmt.Sprintf("checking stack %q", *stack.StackName))
		if shouldBeDeleted(stack) {
			a.logger.Log("level", "debug", "message", fmt.Sprintf("stack %q should be deleted", *stack.StackName))
			deleteStackInput := &cloudformation.DeleteStackInput{
				StackName: stack.StackName,
			}
			_, err := a.cfClient.DeleteStack(deleteStackInput)
			if err != nil {
				// do not return on error, try to continue deleting.
				a.logger.Log("level", "debug", "message", fmt.Sprintf("could not delete stack %q: %#v", *stack.StackName, err))
			} else {
				a.logger.Log("level", "debug", "message", fmt.Sprintf("stack %q was deleted", *stack.StackName))
			}
		}
	}
	return nil
}

func shouldBeDeleted(stack *cloudformation.Stack) bool {
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

	prefixes := []string{
		"cluster-ci-",
		"host-peer-ci-",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(*stack.StackName, prefix) {
			return true
		}
	}

	return false
}
