package aws

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestStackShouldBeDeleted(t *testing.T) {
	now := time.Now()
	twoHoursAgo := now.Add(-2 * time.Hour)
	dummyStatus := "FOO"
	statusDeleteInProgress := "DELETE_IN_PROGRESS"

	tcs := []struct {
		stack       *cloudformation.Stack
		expected    bool
		description string
	}{
		{
			description: "stack without creation time should be deleted",
			stack: &cloudformation.Stack{
				StackName:   aws.String("blblalal"),
				StackStatus: &dummyStatus,
			},
			expected: true,
		},
		{
			description: "recent host peer stack should not be deleted",
			stack: &cloudformation.Stack{
				StackName:    aws.String("host-peer-ci-blblalal"),
				CreationTime: &now,
				StackStatus:  &dummyStatus,
			},
			expected: false,
		},
		{
			description: "old host peer stack should be deleted",
			stack: &cloudformation.Stack{
				StackName:    aws.String("host-peer-ci-blblalal"),
				CreationTime: &twoHoursAgo,
				StackStatus:  &dummyStatus,
			},
			expected: true,
		},
		{
			description: "recent cluster ci stack should not be deleted",
			stack: &cloudformation.Stack{
				StackName:    aws.String("cluster-ci-blblalal"),
				CreationTime: &now,
				StackStatus:  &dummyStatus,
			},
			expected: false,
		},
		{
			description: "old cluster ci stack should be deleted",
			stack: &cloudformation.Stack{
				StackName:    aws.String("cluster-ci-blblalal"),
				CreationTime: &twoHoursAgo,
				StackStatus:  &dummyStatus,
			},
			expected: true,
		},
		{
			description: "recent cluster e2e stack should not be deleted",
			stack: &cloudformation.Stack{
				StackName:    aws.String("e2e-blblalal"),
				CreationTime: &now,
				StackStatus:  &dummyStatus,
			},
			expected: false,
		},
		{
			description: "old cluster e2e stack should be deleted",
			stack: &cloudformation.Stack{
				StackName:    aws.String("e2e-blblalal"),
				CreationTime: &twoHoursAgo,
				StackStatus:  &dummyStatus,
			},
			expected: true,
		},
		{
			description: "stack that is already being deleted",
			stack: &cloudformation.Stack{
				StackName:    aws.String("e2e-blabla"),
				CreationTime: &twoHoursAgo,
				StackStatus:  &statusDeleteInProgress,
			},
			expected: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			actual := stackShouldBeDeleted(tc.stack)

			if actual != tc.expected {
				t.Errorf("checking if %q should be deleted, want %t, got %t", *tc.stack.StackName, tc.expected, actual)
			}
		})
	}
}

func TestBucketShouldBeDeleted(t *testing.T) {
	now := time.Now()
	twoHoursAgo := now.Add(-2 * time.Hour)

	tcs := []struct {
		bucket      *s3.Bucket
		expected    bool
		description string
	}{
		{
			description: "bucket without creation time should be deleted",
			bucket: &s3.Bucket{
				Name: aws.String("blblalal"),
			},
			expected: true,
		},
		{
			description: "recent ci wip bucket should not be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("270935918670-g8s-ci-wip-50a83-d4f51"),
				CreationDate: &now,
			},
			expected: false,
		},
		{
			description: "recent ci wip log bucket should not be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("ci-wip-ac84b-7a52e-g8s-access-logs"),
				CreationDate: &now,
			},
			expected: false,
		},
		{
			description: "recent ci cur bucket should not be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("270935918670-g8s-ci-cur-50a83-d4f51"),
				CreationDate: &now,
			},
			expected: false,
		},
		{
			description: "recent ci cur log bucket should not be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("ci-cur-ac84b-7a52e-g8s-access-logs"),
				CreationDate: &now,
			},
			expected: false,
		},
		{
			description: "recent ci clop bucket should not be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("270935918670-g8s-ci-clop-50a83-d4f51"),
				CreationDate: &now,
			},
			expected: false,
		},
		{
			description: "recent ci clop log bucket should not be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("ci-clop-ac84b-7a52e-g8s-access-logs"),
				CreationDate: &now,
			},
			expected: false,
		},
		{
			description: "old ci wip bucket should be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("270935918670-g8s-ci-wip-50a83-d4f51"),
				CreationDate: &twoHoursAgo,
			},
			expected: true,
		},
		{
			description: "old ci wip log bucket should be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("ci-wip-ac84b-7a52e-g8s-access-logs"),
				CreationDate: &twoHoursAgo,
			},
			expected: true,
		},
		{
			description: "old ci cur bucket should be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("270935918670-g8s-ci-cur-50a83-d4f51"),
				CreationDate: &twoHoursAgo,
			},
			expected: true,
		},
		{
			description: "old ci cur log bucket should be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("ci-cur-ac84b-7a52e-g8s-access-logs"),
				CreationDate: &twoHoursAgo,
			},
			expected: true,
		},
		{
			description: "old ci clop bucket should be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("270935918670-g8s-ci-clop-50a83-d4f51"),
				CreationDate: &twoHoursAgo,
			},
			expected: true,
		},
		{
			description: "old ci clop log bucket should be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("ci-clop-ac84b-7a52e-g8s-access-logs"),
				CreationDate: &twoHoursAgo,
			},
			expected: true,
		},
		{
			description: "recent general bucket should not be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("270935918670-g8s-84ar8-ci-5555-clop-blabla"),
				CreationDate: &now,
			},
			expected: false,
		},
		{
			description: "old general bucket should not be deleted",
			bucket: &s3.Bucket{
				Name:         aws.String("270935918670-g8s-84ar8-ci-5555-clop-blabla"),
				CreationDate: &twoHoursAgo,
			},
			expected: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			actual := bucketShouldBeDeleted(tc.bucket)

			if actual != tc.expected {
				t.Errorf("checking if %q should be deleted, want %t, got %t", *tc.bucket.Name, tc.expected, actual)
			}
		})
	}
}
