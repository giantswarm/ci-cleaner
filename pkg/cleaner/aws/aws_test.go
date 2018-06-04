package aws

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func TestShouldBeDeleted(t *testing.T) {
	now := time.Now()
	twoHoursAgo := now.Add(-2 * time.Hour)

	tcs := []struct {
		stack       *cloudformation.Stack
		expected    bool
		description string
	}{
		{
			description: "stack without creation time should be deleted",
			stack: &cloudformation.Stack{
				StackName: aws.String("blblalal"),
			},
			expected: true,
		},
		{
			description: "recent host peer stack should not be deleted",
			stack: &cloudformation.Stack{
				StackName:    aws.String("host-peer-ci-blblalal"),
				CreationTime: &now,
			},
			expected: false,
		},
		{
			description: "old host peer stack should be deleted",
			stack: &cloudformation.Stack{
				StackName:    aws.String("host-peer-ci-blblalal"),
				CreationTime: &twoHoursAgo,
			},
			expected: true,
		},
		{
			description: "recent cluster ci stack should not be deleted",
			stack: &cloudformation.Stack{
				StackName:    aws.String("cluster-ci-blblalal"),
				CreationTime: &now,
			},
			expected: false,
		},
		{
			description: "old cluster ci stack should be deleted",
			stack: &cloudformation.Stack{
				StackName:    aws.String("cluster-ci-blblalal"),
				CreationTime: &twoHoursAgo,
			},
			expected: true,
		},
		{
			description: "recent cluster e2e stack should not be deleted",
			stack: &cloudformation.Stack{
				StackName:    aws.String("e2e-blblalal"),
				CreationTime: &now,
			},
			expected: false,
		},
		{
			description: "old cluster e2e stack should be deleted",
			stack: &cloudformation.Stack{
				StackName:    aws.String("e2e-blblalal"),
				CreationTime: &twoHoursAgo,
			},
			expected: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			actual := shouldBeDeleted(tc.stack)

			if actual != tc.expected {
				t.Errorf("checking if %q should be deleted, want %t, got %t", *tc.stack.StackName, tc.expected, actual)
			}
		})
	}
}
