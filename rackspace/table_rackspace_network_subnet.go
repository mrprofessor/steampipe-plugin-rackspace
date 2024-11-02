package rackspace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

const listNetworkSubnetsUrl = "https://%s.networks.api.rackspacecloud.com/v2.0/subnets"

// NetworkSubnet represents a Rackspace network subnet
type NetworkSubnet struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	EnableDHCP      bool             `json:"enable_dhcp"`
	NetworkID       string           `json:"network_id"`
	TenantID        string           `json:"tenant_id"`
	DNSNameservers  []string         `json:"dns_nameservers"`
	AllocationPools []AllocationPool `json:"allocation_pools"`
	HostRoutes      []string         `json:"host_routes"`
	IPVersion       int              `json:"ip_version"`
	GatewayIP       *string          `json:"gateway_ip"`
	CIDR            string           `json:"cidr"`
}

// AllocationPool represents the IP allocation pool for a subnet
type AllocationPool struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

func tableRackspaceNetworkSubnet() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_network_subnet",
		Description: "Retrieve details of Rackspace network subnets.",
		List: &plugin.ListConfig{
			Hydrate: listNetworkSubnets,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_STRING, Description: "The unique identifier of the subnet."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the subnet."},
			{Name: "enable_dhcp", Type: proto.ColumnType_BOOL, Description: "Indicates if DHCP is enabled on the subnet.", Transform: transform.FromField("EnableDHCP")},
			{Name: "network_id", Type: proto.ColumnType_STRING, Description: "The ID of the network to which the subnet belongs."},
			{Name: "tenant_id", Type: proto.ColumnType_STRING, Description: "The ID of the tenant that owns the subnet."},
			{Name: "dns_nameservers", Type: proto.ColumnType_JSON, Description: "List of DNS nameservers associated with the subnet.", Transform: transform.FromField("DNSNameservers")},
			{Name: "allocation_pools", Type: proto.ColumnType_JSON, Description: "IP allocation pools for the subnet.", Transform: transform.FromField("AllocationPools")},
			{Name: "host_routes", Type: proto.ColumnType_JSON, Description: "List of host routes associated with the subnet."},
			{Name: "ip_version", Type: proto.ColumnType_INT, Description: "IP version used by the subnet (e.g., 4 for IPv4).", Transform: transform.FromField("IPVersion")},
			{Name: "gateway_ip", Type: proto.ColumnType_STRING, Description: "The IP address of the subnet gateway.", Transform: transform.FromField("GatewayIP")},
			{Name: "cidr", Type: proto.ColumnType_STRING, Description: "The CIDR of the subnet.", Transform: transform.FromField("CIDR")},
		},
	}
}

func listNetworkSubnets(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)

	// Construct API request URL
	apiUrl := fmt.Sprintf(
		listNetworkSubnetsUrl,
		*rackspaceConfig.Region,
	)

	// Create an HTTP request
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	// Set the Authorization header
	req.Header.Add("X-Auth-Token", *rackspaceConfig.TokenID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve network subnets: %s", resp.Status)
	}

	// Parse the response
	var result struct {
		Subnets []NetworkSubnet `json:"subnets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Stream each subnet to the table
	for _, subnet := range result.Subnets {
		d.StreamListItem(ctx, subnet)
	}

	return nil, nil
}
