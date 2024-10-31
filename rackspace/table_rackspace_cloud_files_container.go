package rackspace

import (
	"context"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/objectstorage/v1/containers"
	"github.com/gophercloud/gophercloud/v2/pagination"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func tableRackspaceCloudFilesContainer() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_cloud_files_container",
		Description: "Rackspace Cloud Files containers.",
		List: &plugin.ListConfig{
			Hydrate: listContainers,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("name"),
			Hydrate:    getContainer,
		},
		Columns: []*plugin.Column{
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the container."},
			{Name: "bytes", Type: proto.ColumnType_INT, Description: "Total bytes stored in the container."},
			{Name: "count", Type: proto.ColumnType_INT, Description: "Number of objects stored in the container."},
			{Name: "metadata", Type: proto.ColumnType_JSON, Description: "Metadata associated with the container."},
		},
	}
}

func listContainers(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	provider, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	region, err := getRegion(ctx, d)
	if err != nil {
		return nil, err
	}

	client, err := openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{
		Region: *region,
	})
	if err != nil {
		return nil, err
	}

	opts := containers.ListOpts{}
	pager := containers.List(client, opts)
	pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		containerList, err := containers.ExtractInfo(page)
		if err != nil {
			return false, err
		}

		for _, container := range containerList {
			d.StreamListItem(ctx, container)
		}
		return true, nil
	})

	return nil, nil
}

// Thanks to the absurdist nature of the Gophercloud library, we have to define a struct that can accommodate the
// Container data as well as the metadata. The getContainer function only returns the metadata if the container is found.
// So we have to first find the container using the List function and then use the Get function to retrieve the metadata.
// https://github.com/gophercloud/gophercloud/blob/master/openstack/objectstorage/v1/containers/requests.go
// TODO
// Remove getContainer function had use a getMetadata function as a hydrate function for the metadata column
func getContainer(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	name := d.EqualsQuals["name"].GetStringValue()

	provider, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	region, err := getRegion(ctx, d)
	if err != nil {
		return nil, err
	}

	client, err := openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{
		Region: *region,
	})
	if err != nil {
		return nil, err
	}

	// Define struct with int64 for Count to match Gophercloud
	var containerData struct {
		Name     string
		Bytes    int64
		Count    int64
		Metadata map[string]string
	}

	// Use List with Prefix option to find the exact container by name
	opts := containers.ListOpts{Prefix: name}
	err = containers.List(client, opts).EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		containerList, err := containers.ExtractInfo(page)
		if err != nil {
			return false, err
		}

		for _, container := range containerList {
			if container.Name == name {
				containerData.Name = container.Name
				containerData.Bytes = container.Bytes
				containerData.Count = container.Count
				return false, nil // Stop iterating once we find the matching container
			}
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	// Retrieve metadata if needed
	getOpts := containers.GetOpts{}
	getResult := containers.Get(ctx, client, name, getOpts)
	if getResult.Err != nil {
		return nil, getResult.Err
	}
	containerData.Metadata, err = getResult.ExtractMetadata()
	if err != nil {
		return nil, err
	}

	return containerData, nil
}
