package azure

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type CleanerConfig struct {
	Logger micrologger.Logger
}

type Cleaner struct {
	logger micrologger.Logger
}

func NewCleaner(config CleanerConfig) (*Cleaner, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	return c, nil
}

func (c *Cleaner) Clean(ctx context.Context) error {
	// TODO cleanup peeringso
	//
	//     https://godoc.org/github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-05-01/network#VirtualNetworksClient
	//     https://godoc.org/github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-05-01/network#VirtualNetworkPeeringsClient
	//
}
