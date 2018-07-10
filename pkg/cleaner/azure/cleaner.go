package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-05-01/network"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type CleanerConfig struct {
	Logger                       micrologger.Logger
	VirtualNetworkPeeringsClient *network.VirtualNetworkPeeringsClient
	VirtualNetworksClient        *network.VirtualNetworksClient

	Installations []string
}

type Cleaner struct {
	logger                       micrologger.Logger
	virtualNetworkPeeringsClient *network.VirtualNetworkPeeringsClient
	virtualNetworksClient        *network.VirtualNetworksClient

	installations []string
}

func NewCleaner(config CleanerConfig) (*Cleaner, error) {
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
		logger: config.Logger,
		virtualNetworkPeeringsClient: config.VirtualNetworkPeeringsClient,
		virtualNetworksClient:        config.VirtualNetworksClient,

		installations: config.Installations,
	}

	return c, nil
}

func (c *Cleaner) Clean(ctx context.Context) error {
	// TODO cleanup peeringso
	//
	//     https://godoc.org/github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-05-01/network#VirtualNetworksClient
	//     https://godoc.org/github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-05-01/network#VirtualNetworkPeeringsClient
	//

	r, err := c.virtualNetworksClient.ListAll(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, v := range r.Values() {
		// We skip all virtual networks of installations we do not want to cleanup.
		if !containsString(c.installations, *v.Name) {
			continue
		}
		fmt.Printf("\n")
		fmt.Printf("%#v\n", *v.ID)
		fmt.Printf("%#v\n", *v.Name)
		fmt.Printf("\n")
	}

	return nil
}

func containsString(list []string, s string) bool {
	for _, l := range list {
		if l == s {
			return true
		}
	}

	return false
}

func isAnyEmpty(list []string) bool {
	for _, l := range list {
		if l == "" {
			return true
		}
	}

	return false
}
