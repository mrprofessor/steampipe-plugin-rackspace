package rackspace

import (
	"context"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/limits"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableRackspaceComputeLimit() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_compute_limit",
		Description: "Retrieves Rackspace Compute resource limits and usage information for a tenant.",
		List: &plugin.ListConfig{
			Hydrate: getComputeLimit,
		},
		Columns: []*plugin.Column{
			{Name: "max_total_cores", Type: proto.ColumnType_INT, Description: "The maximum number of cores available to a tenant.", Transform: transform.FromField("MaxTotalCores")},
			{Name: "max_image_meta", Type: proto.ColumnType_INT, Description: "The maximum amount of image metadata available to a tenant.", Transform: transform.FromField("MaxImageMeta")},
			{Name: "max_server_meta", Type: proto.ColumnType_INT, Description: "The maximum amount of server metadata available to a tenant.", Transform: transform.FromField("MaxServerMeta")},
			{Name: "max_personality", Type: proto.ColumnType_INT, Description: "The maximum number of personality files available to a tenant.", Transform: transform.FromField("MaxPersonality")},
			{Name: "max_personality_size", Type: proto.ColumnType_INT, Description: "The maximum size of each personality file in bytes.", Transform: transform.FromField("MaxPersonalitySize")},
			{Name: "max_total_keypairs", Type: proto.ColumnType_INT, Description: "The maximum number of keypairs available to a tenant.", Transform: transform.FromField("MaxTotalKeypairs")},
			{Name: "max_security_groups", Type: proto.ColumnType_INT, Description: "The maximum number of security groups available to a tenant.", Transform: transform.FromField("MaxSecurityGroups")},
			{Name: "max_security_group_rules", Type: proto.ColumnType_INT, Description: "The maximum number of security group rules available to a tenant.", Transform: transform.FromField("MaxSecurityGroupRules")},
			{Name: "max_server_groups", Type: proto.ColumnType_INT, Description: "The maximum number of server groups available to a tenant.", Transform: transform.FromField("MaxServerGroups")},
			{Name: "max_server_group_members", Type: proto.ColumnType_INT, Description: "The maximum number of members in each server group.", Transform: transform.FromField("MaxServerGroupMembers")},
			{Name: "max_total_floating_ips", Type: proto.ColumnType_INT, Description: "The maximum number of floating IPs available to a tenant.", Transform: transform.FromField("MaxTotalFloatingIps")},
			{Name: "max_total_instances", Type: proto.ColumnType_INT, Description: "The maximum number of instances available to a tenant.", Transform: transform.FromField("MaxTotalInstances")},
			{Name: "max_total_ram_size", Type: proto.ColumnType_INT, Description: "The total amount of RAM available to a tenant in megabytes (MB).", Transform: transform.FromField("MaxTotalRAMSize")},
			{Name: "total_cores_used", Type: proto.ColumnType_INT, Description: "The total number of cores currently in use.", Transform: transform.FromField("TotalCoresUsed")},
			{Name: "total_instances_used", Type: proto.ColumnType_INT, Description: "The total number of instances currently in use.", Transform: transform.FromField("TotalInstancesUsed")},
			{Name: "total_floating_ips_used", Type: proto.ColumnType_INT, Description: "The total number of floating IPs currently in use.", Transform: transform.FromField("TotalFloatingIpsUsed")},
			{Name: "total_ram_used", Type: proto.ColumnType_INT, Description: "The total amount of RAM currently in use in megabytes (MB).", Transform: transform.FromField("TotalRAMUsed")},
			{Name: "total_security_groups_used", Type: proto.ColumnType_INT, Description: "The total number of security groups currently in use.", Transform: transform.FromField("TotalSecurityGroupsUsed")},
			{Name: "total_server_groups_used", Type: proto.ColumnType_INT, Description: "The total number of server groups currently in use.", Transform: transform.FromField("TotalServerGroupsUsed")},
		},
	}
}

// TODO: Use APIs to get the Rate data which are missing from gohercloud
func getComputeLimit(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Authenticate with Rackspace
	provider, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	// Retrieve the region information
	region, err := getRegion(ctx, d)
	if err != nil {
		return nil, err
	}

	// Create a Compute client
	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: *region,
	})
	if err != nil {
		return nil, err
	}

	// Retrieve the limits
	limitsData, err := limits.Get(ctx, client, limits.GetOpts{}).Extract()
	if err != nil {
		return nil, err
	}

	// Stream limits data as a single row
	d.StreamListItem(ctx, limitsData.Absolute)
	return limitsData.Absolute, nil
}
