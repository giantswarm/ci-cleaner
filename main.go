package main

import (
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func lambdaHandler() error {
	svc := cloudformation.New(session.New())
	input := &cloudformation.DescribeStacksInput{}

	output, err := svc.DescribeStacks(input)
	if err != nil {
		return err
	}

	for _, stack := range output.Stacks {
		if shouldBeDeleted(stack) {
			log.Printf("stack %q should be deleted", *stack.StackName)
			deleteStackInput := &cloudformation.DeleteStackInput{
				StackName: stack.StackName,
			}
			_, err := svc.DeleteStack(deleteStackInput)
			if err != nil {
				// do not return on error, try to continue deleting.
				log.Printf("could not delete stack %q: %#v", *stack.StackName, err)
			} else {
				log.Printf("stack %q was deleted", *stack.StackName)
			}
		}
	}
	return nil
}

func shouldBeDeleted(stack *cloudformation.Stack) bool {
	now := time.Now().UTC()
	timeDiff := now.Sub(*stack.CreationTime)

	// do not delete recent stacks.
	if timeDiff < 90*time.Minute {
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

func main() {
	lambda.Start(lambdaHandler)
}
