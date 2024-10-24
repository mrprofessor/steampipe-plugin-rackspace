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
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

func tableRackspaceCompute() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_compute",
		Description: "Rackspace Compute (Nova)",
		List: &plugin.ListConfig{
			Hydrate: listCompute,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getCompute,
		},
		Columns: []*plugin.Column{

			// Basic fields
			{Name: "id", Type: proto.ColumnType_STRING, Description: "The ID of the server"},
			{Name: "tenant_id", Type: proto.ColumnType_STRING, Description: "TenantID identifies the tenant owning this server resource"},
			{Name: "user_id", Type: proto.ColumnType_STRING, Description: "UserID uniquely identifies the user account owning the tenant"},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name contains the human-readable name for the server"},
			{Name: "updated", Type: proto.ColumnType_TIMESTAMP, Description: "Last updated timestamp"},
			{Name: "created", Type: proto.ColumnType_TIMESTAMP, Description: "Server creation timestamp"},
			{Name: "host_id", Type: proto.ColumnType_STRING, Description: "HostID is the host where the server is located in the cloud"},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "Status contains the current operational status of the server, such as IN_PROGRESS or ACTIVE."},
			{Name: "progress", Type: proto.ColumnType_INT, Description: "Server progress ranges from 0..100"},
			{Name: "accessIPv4", Type: proto.ColumnType_STRING, Description: "IPv4 address of the server"},
			{Name: "accessIPv6", Type: proto.ColumnType_STRING, Description: "IPv6 address of the server"},
			{Name: "image", Type: proto.ColumnType_JSON, Description: "Image refers to a JSON object, which itself indicates the OS image used to deploy the server"},
			{Name: "flavor", Type: proto.ColumnType_JSON, Description: "Hardware configuration of the deployed server"},
			{Name: "addresses", Type: proto.ColumnType_JSON, Description: "List of IP addresses assigned to the server"},
			{Name: "metadata", Type: proto.ColumnType_JSON, Description: "User-specified key-value pairs attached to the server"},
			{Name: "links", Type: proto.ColumnType_JSON, Description: "Links includes HTTP references to itself"},
			{Name: "key_name", Type: proto.ColumnType_STRING, Description: "Public key name used for the server's SSH login", Transform: transform.FromField("KeyName")},
			{Name: "admin_pass", Type: proto.ColumnType_STRING, Description: "Administrative password", Transform: transform.FromField("AdminPass")},
			{Name: "security_groups", Type: proto.ColumnType_JSON, Description: "Security groups applied to the server", Transform: transform.FromField("SecurityGroups")},
			{Name: "attached_volumes", Type: proto.ColumnType_JSON, Description: "Volume attachments for the server", Transform: transform.FromField("AttachedVolumes")},

			// State and configuration fields
			{Name: "vm_state", Type: proto.ColumnType_STRING, Description: "Virtual machine state, such as 'active'.", Transform: transform.FromField("VmState")},
			{Name: "disk_config", Type: proto.ColumnType_STRING, Description: "Disk configuration, such as 'AUTO' or 'MANUAL'.", Transform: transform.FromField("DiskConfig")},
			{Name: "power_state", Type: proto.ColumnType_INT, Description: "Power state of the server (e.g., 1 for running).", Transform: transform.FromField("PowerState")},
			{Name: "task_state", Type: proto.ColumnType_STRING, Description: "Task state of the server.", Transform: transform.FromField("TaskState")},
			{Name: "fault", Type: proto.ColumnType_JSON, Description: "Information about server failures", Transform: transform.FromField("Fault")},
			{Name: "tags", Type: proto.ColumnType_JSON, Description: "Tags attached to the server", Transform: transform.FromField("Tags")},
			{Name: "server_groups", Type: proto.ColumnType_JSON, Description: "UUIDs of server groups to which the server belongs", Transform: transform.FromField("ServerGroups")},
			{Name: "availability_zone", Type: proto.ColumnType_STRING, Description: "The availability zone of the server", Transform: transform.FromField("AvailabilityZone")},

			// Extended attributes
			{Name: "host", Type: proto.ColumnType_STRING, Description: "The host or hypervisor where the instance is running", Transform: transform.FromField("Host")},
			{Name: "instance_name", Type: proto.ColumnType_STRING, Description: "The name of the instance", Transform: transform.FromField("InstanceName")},
			{Name: "hypervisor_hostname", Type: proto.ColumnType_STRING, Description: "The hostname of the hypervisor", Transform: transform.FromField("HypervisorHostname")},
			{Name: "reservation_id", Type: proto.ColumnType_STRING, Description: "The reservation ID of the instance", Transform: transform.FromField("ReservationID")},
			{Name: "launch_index", Type: proto.ColumnType_INT, Description: "The launch index of the instance", Transform: transform.FromField("LaunchIndex")},
			{Name: "ramdisk_id", Type: proto.ColumnType_STRING, Description: "The ID of the RAM disk image of the instance", Transform: transform.FromField("RAMDiskID")},
			{Name: "kernel_id", Type: proto.ColumnType_STRING, Description: "The ID of the kernel image of the instance", Transform: transform.FromField("KernelID")},
			{Name: "hostname", Type: proto.ColumnType_STRING, Description: "The hostname of the instance", Transform: transform.FromField("Hostname")},
			{Name: "root_device_name", Type: proto.ColumnType_STRING, Description: "The name of the root device of the instance", Transform: transform.FromField("RootDeviceName")},
			{Name: "userdata", Type: proto.ColumnType_STRING, Description: "User data associated with the instance", Transform: transform.FromField("Userdata")},

			// Time fields
			{Name: "launched_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the instance was launched", Transform: transform.FromField("LaunchedAt")},
			{Name: "terminated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the instance was terminated", Transform: transform.FromField("TerminatedAt")},

			// Rackspace-specific fields (Not supported by gohercloud)
			{Name: "public_ip_zone_id", Type: proto.ColumnType_STRING, Description: "Rackspace-specific public IP zone ID."},
		},
	}
}

func listCompute(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {

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
		openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
			Region: *region,
		})

	if err != nil {
		return nil, err
	}

	opts2 := servers.ListOpts{}
	pager := servers.List(client, opts2)

	var serverList []servers.Server
	pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		serverList, err = servers.ExtractServers(page)

		if err != nil {
			return false, err
		}

		for _, s := range serverList {
			logger.Info("server id", s.ID)
			logger.Info("\n\n@@@ server server", s)

			// Use reflection to print all fields and their values
			val := reflect.ValueOf(s)
			typ := reflect.TypeOf(s)

			for i := 0; i < val.NumField(); i++ {
				field := typ.Field(i).Name
				value := val.Field(i).Interface()
				logger.Info(fmt.Sprintf("%s: %v", field, value))
			}

			logger.Info("\n\n\n@@@ server server", s)

			d.StreamListItem(ctx, s)
		}
		return true, nil
	})

	return nil, nil
}

func getCompute(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	quals := d.EqualsQuals
	id := quals["id"].GetStringValue()
	provider, err := connect(ctx, d)

	logger.Warn("AuthenticatedClient print something")

	if err != nil {
		logger.Error("AuthenticatedClient error")
		fmt.Println(err)
		return nil, err
	}

	// Fetch region from config
	region, err := getRegion(ctx, d)
	if err != nil {
		return nil, err
	}

	logger.Info("Region @@@", region)

	client, err :=
		openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
			Region: *region,
		})
	fmt.Println(client)
	if err != nil {
		fmt.Println("NewComputeV2 Error:")
		fmt.Println(err)
		return nil, err
	}

	server, err := servers.Get(ctx, client, id).Extract()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return server, nil
}
