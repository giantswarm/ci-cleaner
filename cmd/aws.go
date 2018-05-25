package cmd

import (
	"fmt"
	"os"

	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/ci-cleaner/pkg/cleaner/aws"
)

var (
	AwsCmd = &cobra.Command{
		Use:   "aws",
		Short: "cleanup leftover AWS CI resources",
		Run:   runAws,
	}
)

var (
	accessKeyID     string
	secretAccessKey string
	region          string
	logger          *micrologger.MicroLogger
)

func init() {
	var err error
	logger, err = micrologger.New(micrologger.Config{})
	if err != nil {
		panic(fmt.Sprintf("%#v", err))
	}

	RootCmd.AddCommand(AwsCmd)

	AwsCmd.Flags().StringVar(&accessKeyID, "access-key-id", "", "Access key ID.")
	AwsCmd.Flags().StringVar(&secretAccessKey, "secret-access-key", "", "Secret access key.")
	AwsCmd.Flags().StringVar(&region, "region", "", "Region.")
}

func runAws(cmd *cobra.Command, args []string) {
	c := &aws.Config{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		Region:          region,

		Logger: logger,
	}

	a, err := aws.New(c)
	if err != nil {
		logger.Log("level", "error", "message", "exiting with status 1 due to error", "stack", fmt.Sprintf("%#v", err))
		os.Exit(1)
	}

	err = a.Clean()
	if err != nil {
		logger.Log("level", "error", "message", "exiting with status 1 due to error", "stack", fmt.Sprintf("%#v", err))
		os.Exit(1)
	}
}
