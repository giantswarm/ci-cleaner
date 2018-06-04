package aws

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type Config struct {
	CFClient CFClient
	Logger   *micrologger.MicroLogger
}

type Cleaner struct {
	cfClient CFClient
	logger   *micrologger.MicroLogger
}

func New(config *Config) (*Cleaner, error) {
	if config.CFClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "CFClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	cleaner := &Cleaner{
		cfClient: config.CFClient,
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
		"e2e-",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(*stack.StackName, prefix) {
			return true
		}
	}

	return false
}
