package azure

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2017-10-01/dns"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-05-01/network"
	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2018-03-01/insights"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type CleanerConfig struct {
	Logger micrologger.Logger

	ActivityLogsClient                     *insights.ActivityLogsClient
	DNSRecordSetsClient                    *dns.RecordSetsClient
	GroupsClient                           *resources.GroupsClient
	VirtualNetworkGatewayConnectionsClient *network.VirtualNetworkGatewayConnectionsClient
	VirtualNetworkPeeringsClient           *network.VirtualNetworkPeeringsClient
	VirtualNetworksClient                  *network.VirtualNetworksClient

	Installations []string
	AzureLocation string
}

type Cleaner struct {
	logger micrologger.Logger

	activityLogsClient                     *insights.ActivityLogsClient
	dnsRecordSetsClient                    *dns.RecordSetsClient
	groupsClient                           *resources.GroupsClient
	virtualNetworkGatewayConnectionsClient *network.VirtualNetworkGatewayConnectionsClient
	virtualNetworkPeeringsClient           *network.VirtualNetworkPeeringsClient
	virtualNetworksClient                  *network.VirtualNetworksClient

	installations []string
	azureLocation string
}

func NewCleaner(config CleanerConfig) (*Cleaner, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ActivityLogsClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ActivityLogsClient must not be empty", config)
	}
	if config.DNSRecordSetsClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.DNSRecordSetsClient must not be empty", config)
	}
	if config.GroupsClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GroupsClient must not be empty", config)
	}
	if config.VirtualNetworkPeeringsClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VirtualNetworkPeeringsClient must not be empty", config)
	}
	if config.VirtualNetworkGatewayConnectionsClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VirtualNetworkGatewayConnectionsClient must not be empty", config)
	}
	if config.VirtualNetworksClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VirtualNetworksClient must not be empty", config)
	}

	if len(config.Installations) == 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.Installations must not be empty", config)
	}
	if isAnyEmpty(config.Installations) {
		return nil, microerror.Maskf(invalidConfigError, "%T.Installations must contain non empty items", config)
	}
	if len(config.AzureLocation) == 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.AzureLocation must not be empty", config)
	}

	c := &Cleaner{
		logger: config.Logger,

		activityLogsClient:                     config.ActivityLogsClient,
		dnsRecordSetsClient:                    config.DNSRecordSetsClient,
		groupsClient:                           config.GroupsClient,
		virtualNetworkPeeringsClient:           config.VirtualNetworkPeeringsClient,
		virtualNetworkGatewayConnectionsClient: config.VirtualNetworkGatewayConnectionsClient,
		virtualNetworksClient:                  config.VirtualNetworksClient,

		installations: config.Installations,
		azureLocation: config.AzureLocation,
	}

	return c, nil
}

func (c *Cleaner) Clean(ctx context.Context) error {
	c.logger.LogCtx(ctx, "level", "debug", "message", "starting Azure CI cleanup")

	err := c.cleanVirtualNetworkPeering(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = c.cleanResourceGroup(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = c.cleanDelegateDNSRecords(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = c.cleanDelegateDNSRecords(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	c.logger.LogCtx(ctx, "level", "debug", "message", "finished Azure CI cleanup")

	return nil
}

func isCIResource(s string) bool {
	r := false
	r = r || strings.HasPrefix(s, "ci-last-")
	r = r || strings.HasPrefix(s, "ci-prev-")
	r = r || strings.HasPrefix(s, "ci-cur-")
	r = r || strings.HasPrefix(s, "ci-wip-")

	return r
}

func isAnyEmpty(list []string) bool {
	for _, l := range list {
		if l == "" {
			return true
		}
	}

	return false
}
