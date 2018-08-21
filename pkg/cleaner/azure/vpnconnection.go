package azure

import (
	"context"
	"fmt"
	"net/http"

	"github.com/giantswarm/microerror"
)

// cleanVPNConnection clean up left of vpn connections
// which do not have a corresponding resource group.
func (c Cleaner) cleanVPNConnection(ctx context.Context) error {
	// Create a map of resource groups name.
	groupMap := make(map[string]bool)
	groupIter, err := c.groupsClient.ListComplete(ctx, "", nil)
	if err != nil {
		return microerror.Mask(err)
	}

	for ; groupIter.NotDone(); groupIter.Next() {
		group := groupIter.Value()

		if isCIResource(*group.Name) {
			groupMap[*group.Name] = true
		}
	}

	// Clean vpn connections in every installation.
	var lastError error
	for _, i := range c.installations {
		iter, err := c.virtualNetworkGatewayConnectionsClient.ListComplete(ctx, i)
		if err != nil {
			return microerror.Mask(err)
		}

		for ; iter.NotDone(); iter.Next() {
			connection := iter.Value()

			if !isCIResource(*connection.Name) {
				// Skip non CI vpn connections.
				continue
			}

			// Delete vpn connection which do not have a corresponding resource group.
			_, exist := groupMap[*connection.Name]
			if !exist {
				c.logger.Log("level", "error", "message", fmt.Sprintf("ensuring deletion of vpn connection %q", *connection.Name))

				resFuture, err := c.virtualNetworkGatewayConnectionsClient.Delete(ctx, i, *connection.Name)
				if err != nil {
					c.logger.Log("level", "error", "message", fmt.Sprintf("did not ensure deletion of vpn connection %q", *connection.Name), "error", microerror.Mask(err))
					lastError = err
					continue
				}

				res, err := c.virtualNetworkGatewayConnectionsClient.DeleteResponder(resFuture.Response())
				if res.Response != nil && res.StatusCode == http.StatusNotFound {
					// fall through
				} else if err != nil {
					c.logger.Log("level", "error", "message", fmt.Sprintf("did not ensure deletion of vpn connection %q", *connection.Name), "error", microerror.Mask(err))
					lastError = err
					continue
				}

				c.logger.Log("level", "error", "message", fmt.Sprintf("ensured deletion of vpn connection %q", *connection.Name))
			}
		}
	}

	if lastError != nil {
		return microerror.Mask(lastError)
	}

	return nil
}
