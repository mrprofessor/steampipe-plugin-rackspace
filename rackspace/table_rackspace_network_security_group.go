package rackspace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// SecurityGroup represents a Rackspace network security group
type SecurityGroup struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	TenantID           string              `json:"tenant_id"`
	Description        string              `json:"description"`
	SecurityGroupRules []SecurityGroupRule `json:"security_group_rules"`
	ExternalServiceID  *string             `json:"external_service_id"`
	ExternalService    *string             `json:"external_service"`
}

// SecurityGroupRule represents a rule within a security group
type SecurityGroupRule struct {
	ID                string  `json:"id"`
	Direction         string  `json:"direction"`
	Protocol          *string `json:"protocol"`
	Description       *string `json:"description"`
	PortRangeMin      *int    `json:"port_range_min"`
	PortRangeMax      *int    `json:"port_range_max"`
	EtherType         string  `json:"ethertype"`
	RemoteGroupID     *string `json:"remote_group_id"`
	RemoteIPPrefix    *string `json:"remote_ip_prefix"`
	SecurityGroupID   string  `json:"security_group_id"`
	TenantID          string  `json:"tenant_id"`
	ExternalServiceID *string `json:"external_service_id"`
	ExternalService   *string `json:"external_service"`
}

func tableRackspaceNetworkSecurityGroup() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_network_security_group",
		Description: "Retrieve details of Rackspace network security groups.",
		List: &plugin.ListConfig{
			Hydrate: listSecurityGroups,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_STRING, Description: "The unique identifier of the security group."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the security group."},
			{Name: "tenant_id", Type: proto.ColumnType_STRING, Description: "The ID of the tenant associated with the security group."},
			{Name: "description", Type: proto.ColumnType_STRING, Description: "The description of the security group."},
			{Name: "external_service_id", Type: proto.ColumnType_STRING, Description: "External service ID associated with the security group, if any."},
			{Name: "external_service", Type: proto.ColumnType_STRING, Description: "External service name associated with the security group, if any."},
			{Name: "security_group_rules", Type: proto.ColumnType_JSON, Description: "List of rules associated with the security group."},
		},
	}
}

func listSecurityGroups(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)

	// Construct API request URL
	apiUrl := fmt.Sprintf(
		"https://%s.networks.api.rackspacecloud.com/v2.0/security-groups",
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
		return nil, fmt.Errorf("failed to retrieve security groups: %s", resp.Status)
	}

	// Parse the response
	var result struct {
		SecurityGroups []SecurityGroup `json:"security_groups"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Stream each security group to the table
	for _, group := range result.SecurityGroups {
		d.StreamListItem(ctx, group)
	}

	return nil, nil
}
