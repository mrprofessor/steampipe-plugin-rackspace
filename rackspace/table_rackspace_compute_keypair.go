package rackspace

import (
	"context"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/keypairs"
	"github.com/gophercloud/gophercloud/v2/pagination"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func tableRackspaceComputeKeyPair() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_compute_keypair",
		Description: "Retrieve details of Rackspace Compute keypairs.",
		List: &plugin.ListConfig{
			Hydrate: listComputeKeypairs,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("name"),
			Hydrate:    getComputeKeypair,
		},
		Columns: []*plugin.Column{
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the keypair."},
			{Name: "public_key", Type: proto.ColumnType_STRING, Description: "The public key of the keypair."},
			{Name: "fingerprint", Type: proto.ColumnType_STRING, Description: "The fingerprint of the keypair."},
			{Name: "user_id", Type: proto.ColumnType_STRING, Description: "The user ID associated with the keypair."},
			{Name: "type", Type: proto.ColumnType_STRING, Description: "The type of keypair, such as ssh or x509."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Description: "The date and time when the keypair was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "The date and time when the keypair was last updated."},
			{Name: "deleted_at", Type: proto.ColumnType_TIMESTAMP, Description: "The date and time when the keypair was deleted."},
			{Name: "deleted", Type: proto.ColumnType_BOOL, Description: "Whether the keypair is deleted."},
			

		},
	}
}

func listComputeKeypairs(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
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

	// List keypairs
	listOpts := keypairs.ListOpts{}
	pager := keypairs.List(client, listOpts)
	err = pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		keypairList, err := keypairs.ExtractKeyPairs(page)
		if err != nil {
			return false, err
		}

		for _, kp := range keypairList {
			d.StreamListItem(ctx, kp)
		}
		return true, nil
	})

	return nil, err
}

func getComputeKeypair(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	name := d.EqualsQuals["name"].GetStringValue()

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

	getOpts := keypairs.GetOpts{}
	logger := plugin.Logger(ctx)
	keypair, err := keypairs.Get(ctx, client, name, getOpts).Extract()
	logger.Info("getComputeKeypair @@@", keypair)

	if err != nil {
		return nil, err
	}

	return keypair, nil
}
