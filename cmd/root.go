package cmd

import (
	"fmt"

	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"
)

var (
	RootCmd = &cobra.Command{
		Use:   "ci-cleaner",
		Short: "Clean CI resources",
	}
)

var (
	logger micrologger.Logger
)

func init() {
	var err error

	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			panic(fmt.Sprintf("Error creating micrologger instance: %#v", err))
		}
	}

	RootCmd.AddCommand(AwsCmd)
	RootCmd.AddCommand(AzureCmd)
}
