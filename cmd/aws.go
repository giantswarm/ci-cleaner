package cmd

import (
	"fmt"
	"os"

	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"

	"github.com/giantswarm/ci-cleaner/pkg/cleaner/aws"
	"github.com/giantswarm/ci-cleaner/pkg/errorcollection"
)

var (
	AwsCmd = &cobra.Command{
		Use:   "aws",
		Short: "Cleanup leftover AWS CI resources.",
		Run:   runAws,
	}
)

var (
	accessKeyID     string
	secretAccessKey string
	region          string
)

func init() {
	AwsCmd.Flags().StringVar(&accessKeyID, "access-key-id", "", "Access key ID.")
	AwsCmd.Flags().StringVar(&secretAccessKey, "secret-access-key", "", "Secret access key.")
	AwsCmd.Flags().StringVar(&region, "region", "", "Region.")
}

// runAws runs the AWS related cleaner jobs, prints error output
// and exits with a non-zero exit case when errors occur.
func runAws(cmd *cobra.Command, args []string) {
	awsCfg := &awsSDK.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Region:      awsSDK.String(region),
	}
	s, err := session.NewSession(awsCfg)
	if err != nil {
		fmt.Printf("Problem setting up a new AWS session: %#v\n", err)
		os.Exit(1)
	}
	cfClient := cloudformation.New(s)
	s3Client := s3.New(s)

	c := &aws.Config{
		CFClient: cfClient,
		S3Client: s3Client,
		Logger:   logger,
	}

	a, err := aws.New(c)
	if err != nil {
		fmt.Printf("Problem creating the AWS cleaner: %#v\n", err)
		os.Exit(1)
	}

	err = a.Clean()
	if err != nil {
		// Print our collected errors
		if errors, ok := err.(*errorcollection.ErrorCollection); ok {
			fmt.Println("\nErrors:")
			fmt.Println(errors.Dump())
		}

		os.Exit(1)
	}

}
