package azure

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/giantswarm/microerror"
)

const (
	// gracePeriod represents the maximum time the CI resources are allowed to
	// remain up. CI resources older than gracePeriod will be deleted.
	gracePeriod = 90 * time.Minute
)

func (c Cleaner) cleanResourceGroup(ctx context.Context) error {
	var lastError error

	// It would be more efficient here to use a filter like "startswith(name,'ci-') or startswith(name,'e2e')"
	// but this does not seems to work now, see https://github.com/Azure/azure-sdk-for-go/issues/2480.
	groupIter, err := c.groupsClient.ListComplete(ctx, "", nil)
	if err != nil {
		return microerror.Mask(err)
	}

	deadLine := time.Now().Add(-gracePeriod).UTC()

	for ; groupIter.NotDone(); groupIter.Next() {
		group := groupIter.Value()

		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("check resource group %q", *group.Name))

		shouldBeDeleted, err := c.groupShouldBeDeleted(ctx, group, deadLine)
		if err != nil {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("failed to check resource group %q", *group.Name), "error", err.Error())
			r.logger.LogCtx(ctx, "level", "debug", "message", "skipping")
			lastError = err
			continue
		}

		if shouldBeDeleted {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring deletion of resource group %q", *group.Name))

			respFuture, err := c.groupsClient.Delete(ctx, *group.Name)
			if err != nil {
				c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not ensure deletion for resource group %q ", *group.Name), "error", err.Error())
				lastError = err
				continue
			}

			res, err := c.groupsClient.DeleteResponder(respFuture.Response())
			if res.Response != nil && res.StatusCode == http.StatusNotFound {
				// fall through
			} else if err != nil {
				c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not ensure deletion for resource group %q ", *group.Name), "error", err.Error())
				lastError = err
				continue
			}

			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured deletion of resource group %q", *group.Name))
		}
	}

	if lastError != nil {
		return microerror.Mask(lastError)
	}

	return nil
}

func (c Cleaner) groupShouldBeDeleted(ctx context.Context, group resources.Group, since time.Time) (bool, error) {
	if !isCIResource(*group.Name) {
		return false, nil
	}

	hasActivity, err := c.groupHasActivity(ctx, group, since)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return !hasActivity, nil
}

// groupHasActivity checks if groupName resource group had activity since given time argument.
func (c Cleaner) groupHasActivity(ctx context.Context, group resources.Group, since time.Time) (bool, error) {
	filter := fmt.Sprintf("eventTimestamp ge '%s' and resourceGroupName eq '%s'", since.Format(time.RFC3339Nano), *group.Name)
	eventIter, err := c.activityLogsClient.ListComplete(ctx, filter, "")
	if err != nil {
		return false, microerror.Mask(err)
	}

	// event := eventIter.Value()
	// c.logger.Log("level", "debug", "message", fmt.Sprintf("resource group event: %s %s at %s", *event.OperationName.LocalizedValue, *event.Status.LocalizedValue, event.EventTimestamp.String()))

	// NotDone returns true when eventIter contains events.
	return eventIter.NotDone(), nil
}
