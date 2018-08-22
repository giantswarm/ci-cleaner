package azure

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-05-01/network"
	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2018-03-01/insights"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type CleanerConfig struct {
	ActivityLogsClient           *insights.ActivityLogsClient
	GroupsClient                 *resources.GroupsClient
	Logger                       micrologger.Logger
	VirtualNetworkPeeringsClient *network.VirtualNetworkPeeringsClient
	VirtualNetworksClient        *network.VirtualNetworksClient

	Installations []string
}

type Cleaner struct {
	activityLogsClient           *insights.ActivityLogsClient
	groupsClient                 *resources.GroupsClient
	logger                       micrologger.Logger
	virtualNetworkPeeringsClient *network.VirtualNetworkPeeringsClient
	virtualNetworksClient        *network.VirtualNetworksClient

	installations []string
}

func NewCleaner(config CleanerConfig) (*Cleaner, error) {
	if config.ActivityLogsClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ActivityLogsClient must not be empty", config)
	}
	if config.GroupsClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GroupsClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.VirtualNetworkPeeringsClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VirtualNetworkPeeringsClient must not be empty", config)
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

	c := &Cleaner{
		activityLogsClient: config.ActivityLogsClient,
		groupsClient:       config.GroupsClient,
		logger:             config.Logger,
		virtualNetworkPeeringsClient: config.VirtualNetworkPeeringsClient,
		virtualNetworksClient:        config.VirtualNetworksClient,

		installations: config.Installations,
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

	c.logger.LogCtx(ctx, "level", "debug", "message", "finished Azure CI cleanup")

	return nil
}

func isCIResource(s string) bool {
	return strings.HasPrefix(s, "ci-cur-") || strings.HasPrefix(s, "ci-wip-")
}

func isAnyEmpty(list []string) bool {
	for _, l := range list {
		if l == "" {
			return true
		}
	}

	return false
}
