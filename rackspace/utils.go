package rackspace

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// TimeWrapper is a struct to handle time.Time in JSON
type TimeWrapper struct {
	Time time.Time `json:"time"`
}

// Global httpClient for reuse
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func connect(ctx context.Context, d *plugin.QueryData) (*gophercloud.ProviderClient, error) {

	// Load connection from cache, which preserves throttling protection etc
	cacheKey := "rackspace"
	if cachedData, ok := d.ConnectionManager.Cache.Get(cacheKey); ok {
		return cachedData.(*gophercloud.ProviderClient), nil
	}

	// Start with an empty config
	connectionOpts := gophercloud.AuthOptions{}
	var identityEndpoint, tenantID, tokenID string

	// Prefer config options given in Steampipe
	rackspaceConfig := GetConfig(d.Connection)

	if rackspaceConfig.IdentityEndpoint != nil {
		identityEndpoint = *rackspaceConfig.IdentityEndpoint
	}
	if rackspaceConfig.TenantID != nil {
		tenantID = *rackspaceConfig.TenantID
	}
	if rackspaceConfig.TokenID != nil {
		tokenID = *rackspaceConfig.TokenID
	}

	if identityEndpoint == "" {
		return nil, errors.New("'identity_endpoint' must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}
	connectionOpts.IdentityEndpoint = identityEndpoint
	if tenantID == "" {
		return nil, errors.New("'tenant_id' must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}
	connectionOpts.TenantID = tenantID

	if tokenID == "" {
		return nil, errors.New("'token_id' must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}
	connectionOpts.TokenID = tokenID

	// Create the client
	client, err := openstack.AuthenticatedClient(ctx, connectionOpts)

	if err != nil {
		return nil, fmt.Errorf("error creating Openstack client: %s", err.Error())
	}

	// Save to cache
	d.ConnectionManager.Cache.Set(cacheKey, client)

	// Done
	return client, nil
}

func getRegion(_ context.Context, d *plugin.QueryData) (*string, error) {
	// Prefer config options given in Steampipe
	rackspaceConfig := GetConfig(d.Connection)
	var region *string
	if rackspaceConfig.Region != nil {
		region = rackspaceConfig.Region
	}

	if *region == "" {
		return nil, errors.New("'region' must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}
	return region, nil
}
