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

// NetworkPort represents a Rackspace network port
type NetworkPort struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	AdminStateUp   bool      `json:"admin_state_up"`
	NetworkID      string    `json:"network_id"`
	TenantID       string    `json:"tenant_id"`
	DeviceOwner    string    `json:"device_owner"`
	MACAddress     string    `json:"mac_address"`
	FixedIPs       []FixedIP `json:"fixed_ips"`
	SecurityGroups []string  `json:"security_groups"`
	DeviceID       string    `json:"device_id"`
}

// FixedIP represents a fixed IP associated with a network port
type FixedIP struct {
	SubnetID  string `json:"subnet_id"`
	IPAddress string `json:"ip_address"`
}

func tableRackspaceNetworkPort() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_network_port",
		Description: "Retrieve details of Rackspace network ports.",
		List: &plugin.ListConfig{
			Hydrate: listNetworkPorts,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_STRING, Description: "The unique identifier of the network port."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the network port."},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "The status of the network port."},
			{Name: "admin_state_up", Type: proto.ColumnType_BOOL, Description: "Indicates if the administrative state of the network port is up."},
			{Name: "network_id", Type: proto.ColumnType_STRING, Description: "The ID of the network to which the port is associated."},
			{Name: "tenant_id", Type: proto.ColumnType_STRING, Description: "The ID of the tenant that owns the network port."},
			{Name: "device_owner", Type: proto.ColumnType_STRING, Description: "The entity using this port, such as `compute:None`."},
			{Name: "mac_address", Type: proto.ColumnType_STRING, Description: "The MAC address of the network port.", Transform: transform.FromField("MACAddress")},
			{Name: "fixed_ips", Type: proto.ColumnType_JSON, Description: "List of fixed IP addresses associated with the network port.", Transform: transform.FromField("FixedIPs")},
			{Name: "security_groups", Type: proto.ColumnType_JSON, Description: "List of security groups associated with the network port."},
			{Name: "device_id", Type: proto.ColumnType_STRING, Description: "The ID of the device using this network port."},
		},
	}
}

func listNetworkPorts(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)

	// Construct API request URL
	apiUrl := fmt.Sprintf(
		"https://%s.networks.api.rackspacecloud.com/v2.0/ports",
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
		return nil, fmt.Errorf("failed to retrieve network ports: %s", resp.Status)
	}

	// Parse the response
	var result struct {
		Ports []NetworkPort `json:"ports"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Stream each port to the table
	for _, port := range result.Ports {
		d.StreamListItem(ctx, port)
	}

	return nil, nil
}
