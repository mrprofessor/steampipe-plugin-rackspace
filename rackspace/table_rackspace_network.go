package rackspace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// Gophercloud creates a wrong URL for the Rackspace network API. It
// appends the version number twice.
// https://hkg.networks.api.rackspacecloud.com/v2.0/v2.0/networks instead of
// https://hkg.networks.api.rackspacecloud.com/v2.0/networks
// ref: https://github.com/gophercloud/gophercloud/blob/c697dbb84b05feb3ccc8aa0e422306c205baf1de/openstack/client.go#L392

type Network struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	AdminStateUp bool     `json:"admin_state_up"`
	Status       string   `json:"status"`
	Subnets      []string `json:"subnets"`
	TenantID     string   `json:"tenant_id"`
	Shared       bool     `json:"shared"`
}

// Link represents pagination links for the networks response.
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

func tableRackspaceNetwork() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_network",
		Description: "Retrieve details of Rackspace Networks.",
		List: &plugin.ListConfig{
			Hydrate: listNetworks,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_STRING, Description: "The unique ID of the network."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the network."},
			{Name: "admin_state_up", Type: proto.ColumnType_BOOL, Description: "The administrative state of the network."},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "The operational status of the network."},
			{Name: "subnets", Type: proto.ColumnType_JSON, Description: "List of subnets associated with the network."},
			{Name: "tenant_id", Type: proto.ColumnType_STRING, Description: "The ID of the tenant that owns the network."},
			{Name: "shared", Type: proto.ColumnType_BOOL, Description: "Whether the network is shared across tenants."},
		},
	}
}

func listNetworks(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	rackspaceConfig := GetConfig(d.Connection)
	client := &http.Client{}
	nextPage := fmt.Sprintf("https://%s.networks.api.rackspacecloud.com/v2.0/networks", *rackspaceConfig.Region)

	// Continue paging through until no further pages are available.
	for nextPage != "" {
		req, err := http.NewRequest("GET", nextPage, nil)
		if err != nil {
			return nil, err
		}

		// Set the Authorization header
		req.Header.Add("X-Auth-Token", *rackspaceConfig.TokenID)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to retrieve networks: %s", resp.Status)
		}

		// Parse the response
		var result struct {
			Networks      []Network `json:"networks"`
			NetworksLinks []Link    `json:"networks_links"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		// Stream each network to the table
		for _, network := range result.Networks {
			d.StreamListItem(ctx, network)
		}

		// Check if there is a next page link
		nextPage = ""
		for _, link := range result.NetworksLinks {
			if link.Rel == "next" {
				nextPage = link.Href
				break
			}
		}
	}

	return nil, nil
}
