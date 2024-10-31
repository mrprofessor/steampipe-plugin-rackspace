package rackspace

import (
	"context"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/v2/pagination"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableRackspaceComputeFlavor() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_compute_flavor",
		Description: "Retrieve details of Rackspace Compute flavors.",
		List: &plugin.ListConfig{
			Hydrate: listComputeFlavors,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getComputeFlavor,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_STRING, Description: "The unique identifier of the flavor."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the flavor."},
			{Name: "ram", Type: proto.ColumnType_INT, Description: "The amount of RAM in MB."},
			{Name: "vcpus", Type: proto.ColumnType_INT, Description: "The number of virtual CPUs."},
			{Name: "disk", Type: proto.ColumnType_INT, Description: "The disk size in GB."},
			{Name: "swap", Type: proto.ColumnType_INT, Description: "The amount of swap space in MB."},
			{Name: "rxtx_factor", Type: proto.ColumnType_DOUBLE, Description: "The RX/TX factor used for bandwidth calculations."},
			{Name: "is_public", Type: proto.ColumnType_BOOL, Description: "Whether the flavor is public or private."},
			{Name: "ephemeral", Type: proto.ColumnType_INT, Description: "The amount of ephemeral storage in GB.", Transform: transform.FromField("Ephemeral")},
			// INFO: gophercloud can't unmarshal the extra_specs field ("OS-FLV-WITH-EXT-SPECS:extra_specs")
			// Link: https://github.com/gophercloud/gophercloud/blob/c697dbb84b05feb3ccc8aa0e422306c205baf1de/openstack/compute/v2/flavors/results.go#L88
			{Name: "extra_specs", Type: proto.ColumnType_JSON, Description: "The extra specifications of the flavor.", Transform: transform.FromField("ExtraSpecs")},
			{Name: "description", Type: proto.ColumnType_STRING, Description: "The description of the flavor.", Transform: transform.FromField("Description")},
		},
	}
}

func listComputeFlavors(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
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

	// List flavors
	pager := flavors.ListDetail(client, flavors.ListOpts{})
	err = pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		flavorList, err := flavors.ExtractFlavors(page)
		if err != nil {
			return false, err
		}

		for _, flavor := range flavorList {
			d.StreamListItem(ctx, flavor)
		}
		return true, nil
	})

	return nil, err
}

func getComputeFlavor(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Get flavor ID from the query qualifiers
	flavorID := d.EqualsQuals["id"].GetStringValue()
	if flavorID == "" {
		return nil, nil
	}

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

	// Get the specific flavor by ID
	flavor, err := flavors.Get(ctx, client, flavorID).Extract()
	if err != nil {
		return nil, err
	}

	return flavor, nil
}
