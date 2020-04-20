package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network"
	"github.com/giantswarm/microerror"
)

// cleanVirtualNetworkPeering delete virtual network peering
// leftover by e2e test on control plane.
func (c Cleaner) cleanVirtualNetworkPeering(ctx context.Context) error {
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

	return nil
}
