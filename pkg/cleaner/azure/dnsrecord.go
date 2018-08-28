package azure

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2017-10-01/dns"
	"github.com/giantswarm/microerror"
)

const (
	zoneNameFormat      = "%s.%s.azure.gigantic.io"
	recordSetNameSuffix = ".k8s"
)

// cleanDNSRecordSet clean up left over DNS record set
// which do not have a corresponding resource group.
func (c Cleaner) cleanDNSRecordSet(ctx context.Context) error {
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

	// Clean dns record set in every installation.
	var lastError error
	for _, i := range c.installations {
		zoneName := fmt.Sprintf(zoneNameFormat, i, c.azureLocation)
		// List
		iter, err := c.dnsRecordSetsClient.ListByTypeComplete(ctx, i, zoneName, dns.NS, nil, recordSetNameSuffix)
		if err != nil {
			return microerror.Mask(err)
		}

		for ; iter.NotDone(); iter.Next() {
			recordSet := iter.Value()

			if !isCIResource(*recordSet.Name) {
				// Skip non CI dns record set.
				continue
			}

			// Delete dns record set which do not have a corresponding resource group.
			recordSetNameNoSuffix := strings.TrimSuffix(*recordSet.Name, recordSetNameSuffix)
			_, exist := groupMap[recordSetNameNoSuffix]
			if !exist {
				c.logger.Log("level", "error", "message", fmt.Sprintf("ensuring deletion of record set %q", *recordSet.Name))

				res, err := c.dnsRecordSetsClient.Delete(ctx, i, zoneName, *recordSet.Name, dns.NS, "")
				if res.Response != nil && res.StatusCode == http.StatusNotFound {
					// fall through
				} else if err != nil {
					c.logger.Log("level", "error", "message", fmt.Sprintf("did not ensure deletion of record set %q", *recordSet.Name), "stack", fmt.Sprintf("%#v", microerror.Mask(err)))
					lastError = err
					continue
				}

				c.logger.Log("level", "error", "message", fmt.Sprintf("ensured deletion of record set %q", *recordSet.Name))
			}
		}
	}

	if lastError != nil {
		return microerror.Mask(lastError)
	}

	return nil
}
