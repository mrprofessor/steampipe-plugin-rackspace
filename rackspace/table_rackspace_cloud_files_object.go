package rackspace

import (
	"context"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/objectstorage/v1/objects"
	"github.com/gophercloud/gophercloud/v2/pagination"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableRackspaceCloudFilesObject() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_cloud_files_object",
		Description: "Rackspace Cloud Files objects.",
		List: &plugin.ListConfig{
			Hydrate: listDummy,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "container_name", Require: plugin.Required},
			},
		},
		Columns: []*plugin.Column{
			{Name: "container_name", Type: proto.ColumnType_STRING, Description: "The name of the container holding the object.", Transform: transform.FromQual("container_name")},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the object."},
			{Name: "content_type", Type: proto.ColumnType_STRING, Description: "The content type of the object."},
			{Name: "bytes", Type: proto.ColumnType_INT, Description: "The size of the object in bytes."},
			{Name: "last_modified", Type: proto.ColumnType_TIMESTAMP, Description: "The timestamp of the last modification of the object."},
			{Name: "hash", Type: proto.ColumnType_STRING, Description: "The hash of the object."},
			{Name: "subdir", Type: proto.ColumnType_STRING, Description: "Whether the object contains a subdir."},
			{Name: "is_latest", Type: proto.ColumnType_BOOL, Description: "Whether the object version is the latest one."},
			{Name: "version_id", Type: proto.ColumnType_STRING, Description: "The version ID of the object, when versioning is enabled."},
		},
	}
}

func listDummy(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	containerName := d.EqualsQuals["container_name"].GetStringValue()

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

	listOpts := objects.ListOpts{}
	pager := objects.List(client, containerName, listOpts)
	pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		objectList, err := objects.ExtractInfo(page)
		if err != nil {
			return false, err
		}

		for _, object := range objectList {
			d.StreamListItem(ctx, object)
		}
		return true, nil
	})

	return nil, nil
}
