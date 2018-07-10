package azure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-05-01/network"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type CleanerConfig struct {
	GroupsClient                 *resources.GroupsClient
	Logger                       micrologger.Logger
	VirtualNetworkPeeringsClient *network.VirtualNetworkPeeringsClient
	VirtualNetworksClient        *network.VirtualNetworksClient

	Installations []string
}

type Cleaner struct {
	groupsClient                 *resources.GroupsClient
	logger                       micrologger.Logger
	virtualNetworkPeeringsClient *network.VirtualNetworkPeeringsClient
	virtualNetworksClient        *network.VirtualNetworksClient

	installations []string
}

func NewCleaner(config CleanerConfig) (*Cleaner, error) {
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
		groupsClient: config.GroupsClient,
		logger:       config.Logger,
		virtualNetworkPeeringsClient: config.VirtualNetworkPeeringsClient,
		virtualNetworksClient:        config.VirtualNetworksClient,

		installations: config.Installations,
	}

	return c, nil
}

func (c *Cleaner) Clean(ctx context.Context) error {
	c.logger.LogCtx(ctx, "level", "debug", "message", "starting Azure CI cleanup")

	for _, i := range c.installations {
		r, err := c.virtualNetworksClient.List(ctx, i)
		if err != nil {
			return microerror.Mask(err)
		}

		for {
			for _, v := range r.Values() {
				for _, p := range *v.VirtualNetworkPeerings {
					if !isCIResource(*p.Name) {
						continue
					}

					_, err = c.groupsClient.Get(ctx, *p.Name)
					if IsResourceGroupNotFound(err) && p.PeeringState == network.VirtualNetworkPeeringStateDisconnected {
						c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting vnet peering '%s'", *p.Name))

						_, err := c.virtualNetworkPeeringsClient.Delete(ctx, i, *v.Name, *p.Name)
						if err != nil {
							return microerror.Mask(err)
						}

						c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted vnet peering '%s'", *p.Name))

						time.Sleep(1 * time.Second)
						continue
					} else if err != nil {
						return microerror.Mask(err)
					}
				}
			}

			if r.NotDone() {
				err = r.Next()
				if err != nil {
					return microerror.Mask(err)
				}
				continue
			}

			break
		}
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
