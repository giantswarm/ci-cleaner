package cmd

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-05-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	pkgazure "github.com/giantswarm/ci-cleaner/pkg/cleaner/azure"
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
	azureInstallations  string
	azureLocation       string
	azureSubscriptionID string
	azureTenantID       string
)

func init() {
	AzureCmd.Flags().StringVar(&azureClientID, "client-id", "", "Client ID.")
	AzureCmd.Flags().StringVar(&azureClientSecret, "client-secret", "", "Client secret.")
	AzureCmd.Flags().StringVar(&azureInstallations, "installations", "ghost,godsmack", "Comma separated list of installation names to cleanup.")
	AzureCmd.Flags().StringVar(&azureLocation, "location", "", "Location.")
	AzureCmd.Flags().StringVar(&azureSubscriptionID, "subscription-id", "", "Subscription ID.")
	AzureCmd.Flags().StringVar(&azureTenantID, "tenant-id", "", "Tenant ID.")
}

func runAzure(cmd *cobra.Command, args []string) error {
	var err error

	var servicePrincipalToken *adal.ServicePrincipalToken
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
		c := pkgazure.CleanerConfig{
			Logger: logger,
			VirtualNetworkPeeringsClient: newVirtualNetworkPeeringsClient(azureSubscriptionID, servicePrincipalToken),
			VirtualNetworksClient:        newVirtualNetworksClient(azureSubscriptionID, servicePrincipalToken),

			Installations: strings.Split(azureInstallations, ","),
		}

		azureCleaner, err = pkgazure.NewCleaner(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = azureCleaner.Clean(context.Background())
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func newVirtualNetworkPeeringsClient(azureSubscriptionID string, servicePrincipalToken *adal.ServicePrincipalToken) *network.VirtualNetworkPeeringsClient {
	c := network.NewVirtualNetworkPeeringsClient(azureSubscriptionID)
	c.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)

	return &c
}

func newVirtualNetworksClient(azureSubscriptionID string, servicePrincipalToken *adal.ServicePrincipalToken) *network.VirtualNetworksClient {
	c := network.NewVirtualNetworksClient(azureSubscriptionID)
	c.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)

	return &c
}
