package cmd

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"
)

var (
	AzureCmd = &cobra.Command{
		Use:   "azure",
		Short: "Cleanup leftover Azure CI resources.",
		RunE:  runAzure,
	}
)

var (
	azureClientID       string
	azureClientSecret   string
	azureSubscriptionID string
	azureTenantID       string
	azureLocation       string
	logger              *micrologger.MicroLogger
)

func init() {
	var err error

	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			panic(fmt.Sprintf("%#v", err))
		}
	}

	RootCmd.AddCommand(AzureCmd)

	AzureCmd.Flags().StringVar(&azureClientID, "client-id", "", "Client ID.")
	AzureCmd.Flags().StringVar(&azureClientSecret, "client-secret", "", "Client secret.")
	AzureCmd.Flags().StringVar(&azureSubscriptionID, "subscription-id", "", "Subscription ID.")
	AzureCmd.Flags().StringVar(&azureTenantID, "tenant-id", "", "Tenant ID.")
	AzureCmd.Flags().StringVar(&azureLocation, "location", "", "Location.")
}

func runAzure(cmd *cobra.Command, args []string) error {
	var err error

	var servicePrincipalToken string
	{
		env, err := azure.EnvironmentFromName(azure.PublicCloud.Name)
		if err != nil {
			return microerror.Mask(err)
		}

		oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, azureTenantID)
		if err != nil {
			return microerror.Mask(err)
		}

		servicePrincipalToken, err = adal.NewServicePrincipalToken(*oauthConfig, azureClientID, azureClientSecret, env.ServiceManagementEndpoint)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureCleaner *pkgazure.Cleaner
	{
		c := &pkgazure.CleanerConfig{
			//TODO:   newTODOClient(azureSubscriptionID, servicePrincipalToken),
			Logger: logger,
		}

		azureCleaner, err = pkgazure.NewCleaner(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = azureCleaner.Clean()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

//func newTODOClient(azureSubscriptionID string, servicePrincipalToken string) *TODO {
//	c := compute.NewVirtualMachineScaleSetsClient(azureSubscriptionID)
//	c.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
//
//	return &c
//}
