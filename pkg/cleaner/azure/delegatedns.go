package azure

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2017-10-01/dns"
	"github.com/giantswarm/microerror"
)

const (
	dnsFailureError    = "Temporary failure in name resolution"
	e2eterraformPrefix = "e2eterraform"
	resourceGroup      = "root_dns_zone_rg"
	zoneName           = "azure.gigantic.io"
)

func (c Cleaner) cleanDelegateDNSRecords(ctx context.Context) error {
	var lastError error

	recordsIter, err := c.dnsRecordSetsClient.ListAllByDNSZoneComplete(ctx, resourceGroup, zoneName, nil, "")
	if err != nil {
		return microerror.Mask(err)
	}

	deadLine := time.Now().Add(-gracePeriod).UTC()

	for ; recordsIter.NotDone(); recordsIter.Next() {
		record := recordsIter.Value()

		del, err := c.dnsRecordShouldBeDeleted(ctx, record, deadLine)
		if err != nil {
			c.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to check DNS record %q", *record.Name), "stack", fmt.Sprintf("%#v", microerror.Mask(err)))
			c.logger.LogCtx(ctx, "level", "error", "message", "skipping")
			lastError = err
			continue
		}

		if del {
			c.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("DNS record %s has to be deleted", *record.Name))
			err := c.deleteRecord(ctx, record)
			if err != nil {
				c.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to check DNS record %q", *record.Name), "stack", fmt.Sprintf("%#v", microerror.Mask(err)))
				c.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to delete DNS record %q", *record.Name), "stack", fmt.Sprintf("%#v", microerror.Mask(err)))
				c.logger.LogCtx(ctx, "level", "error", "message", "skipping")
				lastError = err
				continue
			}

			c.logger.LogCtx(ctx, "level", "debug", "info", fmt.Sprintf("DNS record %s was deleted", *record.Name))
		} else {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("DNS record %s has to be kept", *record.Name))
		}

	}

	if lastError != nil {
		return microerror.Mask(lastError)
	}

	return nil
}

func (c Cleaner) deleteRecord(ctx context.Context, dnsRecord dns.RecordSet) error {
	_, err := c.dnsRecordSetsClient.Delete(ctx, resourceGroup, zoneName, *dnsRecord.Name, dns.NS, *dnsRecord.Etag)

	return err
}

func (c Cleaner) dnsRecordShouldBeDeleted(ctx context.Context, dnsRecord dns.RecordSet, since time.Time) (bool, error) {
	if !isTerraformCIRecord(*dnsRecord.Name) {
		return false, nil
	}

	resolves, err := resolvesApiName(*dnsRecord.Name)
	if err != nil {
		c.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("Unexpected error when trying to resolve %s: %s", dnsRecord.Name, err.Error()))
		return false, nil
	}

	return !resolves, nil
}

// isTerraformCIResourceGroup checks if resource group name was created by Terraform CI.
func isTerraformCIRecord(s string) bool {
	return strings.HasPrefix(s, e2eterraformPrefix)
}

// Tries to resolve the API hostname on the specified delegated zone.
func resolvesApiName(name string) (bool, error) {
	full := fmt.Sprintf("api.%s.%s", name, zoneName)

	addresses, err := net.LookupHost(full)

	if err != nil {
		fmt.Printf(err.Error())
		if !strings.Contains(err.Error(), dnsFailureError) {
			return false, err
		}
	}

	return len(addresses) > 0, nil
}
