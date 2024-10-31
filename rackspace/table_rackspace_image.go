package rackspace

import (
	"context"
	"fmt"
	"reflect"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

func tableRackspaceImage() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_image",
		Description: "Rackspace Images (Glance)",
		List: &plugin.ListConfig{
			Hydrate: listImages,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getImage,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_STRING, Description: "The ID of the image"},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the image"},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "Status of the image, e.g., 'queued', 'active'."},
			{Name: "tags", Type: proto.ColumnType_JSON, Description: "Tags associated with the image."},
			{Name: "container_format", Type: proto.ColumnType_STRING, Description: "Container format of the image, e.g., 'ami', 'bare', 'ovf'."},
			{Name: "disk_format", Type: proto.ColumnType_STRING, Description: "Disk format of the image, e.g., 'raw', 'vmdk', 'qcow2'."},
			{Name: "min_disk", Type: proto.ColumnType_INT, Description: "Minimum disk size required to boot the image, in GB.", Transform: transform.FromField("MinDiskGigabytes")},
			{Name: "min_ram", Type: proto.ColumnType_INT, Description: "Minimum RAM required to boot the image, in MB.", Transform: transform.FromField("MinRAMMegabytes")},
			{Name: "owner", Type: proto.ColumnType_STRING, Description: "Tenant ID the image belongs to.", Transform: transform.FromField("Owner")},
			{Name: "protected", Type: proto.ColumnType_BOOL, Description: "Indicates whether the image is protected from deletion."},
			{Name: "visibility", Type: proto.ColumnType_STRING, Description: "Visibility of the image, e.g., 'public' or 'private'."},
			{Name: "hidden", Type: proto.ColumnType_BOOL, Description: "Indicates if the image is hidden from default listings."},
			{Name: "checksum", Type: proto.ColumnType_STRING, Description: "Checksum of the image data."},
			{Name: "size_bytes", Type: proto.ColumnType_INT, Description: "Size of the image data, in bytes."},
			{Name: "metadata", Type: proto.ColumnType_JSON, Description: "Metadata associated with the image.", Transform: transform.FromField("Metadata")},
			{Name: "properties", Type: proto.ColumnType_JSON, Description: "Additional properties associated with the image."},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the image was created."},
			{Name: "updated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the image was last updated."},
			{Name: "file", Type: proto.ColumnType_STRING, Description: "Location of the image file."},
			{Name: "schema", Type: proto.ColumnType_STRING, Description: "Path to the JSON schema representing the image."},
			{Name: "virtual_size", Type: proto.ColumnType_INT, Description: "Virtual size of the image, in bytes.", Transform: transform.FromField("VirtualSize")},
		},
	}
}

// listImages fetches all available images
func listImages(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)

	provider, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	// Fetch region from config
	region, err := getRegion(ctx, d)
	if err != nil {
		return nil, err
	}

	client, err :=
		openstack.NewImageV2(provider, gophercloud.EndpointOpts{
			Region: *region,
		})

	if err != nil {
		return nil, err
	}

	pager := images.List(client, images.ListOpts{})

	var imageList []images.Image
	pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		imageList, err = images.ExtractImages(page)

		if err != nil {
			return false, err
		}

		for _, img := range imageList {
			logger.Info("image id", img.ID)
			logger.Info("\n\n@@@ image", img)

			// Use reflection to print all fields and their values
			val := reflect.ValueOf(img)
			typ := reflect.TypeOf(img)

			for i := 0; i < val.NumField(); i++ {
				field := typ.Field(i).Name
				value := val.Field(i).Interface()
				logger.Info(fmt.Sprintf("%s: %v", field, value))
			}

			d.StreamListItem(ctx, img)
		}
		return true, nil
	})

	return nil, nil
}

// getImage fetches a single image by its ID
func getImage(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	quals := d.EqualsQuals
	id := quals["id"].GetStringValue()
	provider, err := connect(ctx, d)

	if err != nil {
		logger.Error("AuthenticatedClient error")
		return nil, err
	}

	// Fetch region from config
	region, err := getRegion(ctx, d)
	if err != nil {
		return nil, err
	}

	client, err :=
		openstack.NewImageV2(provider, gophercloud.EndpointOpts{
			Region: *region,
		})
	if err != nil {
		return nil, err
	}

	// Add ctx to the images.Get call
	image, err := images.Get(ctx, client, id).Extract()
	if err != nil {
		return nil, err
	}
	return image, nil
}
