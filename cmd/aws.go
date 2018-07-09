package cmd

import (
	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/ci-cleaner/pkg/cleaner/aws"
)

var (
	AwsCmd = &cobra.Command{
		Use:   "aws",
		Short: "Cleanup leftover AWS CI resources.",
		RunE:  runAws,
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

func runAws(cmd *cobra.Command, args []string) error {
	awsCfg := &awsSDK.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Region:      awsSDK.String(region),
	}
	s, err := session.NewSession(awsCfg)
	if err != nil {
		return microerror.Mask(err)
	}
	cfClient := cloudformation.New(s)

	c := &aws.Config{
		CFClient: cfClient,
		Logger:   logger,
	}

	a, err := aws.New(c)
	if err != nil {
		return microerror.Mask(err)
	}

	err = a.Clean()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
