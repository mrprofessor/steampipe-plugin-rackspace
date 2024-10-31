package rackspace

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name: "steampipe-plugin-rackspace",
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
			Schema:      ConfigSchema,
		},
		DefaultTransform: transform.FromGo().NullIfZero(),
		TableMap: map[string]*plugin.Table{
			"rackspace_compute_server":        tableRackspaceComputeServer(),
			"rackspace_image":                 tableRackspaceImage(),
			"rackspace_snapshot":              tableRackspaceSnapshot(),
			"rackspace_volume":                tableRackspaceVolume(),
			"rackspace_cloud_files_container": tableRackspaceCloudFilesContainer(),
			"rackspace_cloud_files_object":    tableRackspaceCloudFilesObject(),
			"rackspace_compute_keypair":       tableRackspaceComputeKeyPair(),
			"rackspace_compute_flavor":        tableRackspaceComputeFlavor(),
			"rackspace_compute_limit":         tableRackspaceComputeLimit(),
			"rackspace_loadbalancer":          tableRackspaceLoadBalancer(),
			"rackspace_network":               tableRackspaceNetwork(),
			"rackspace_message_queue":         tableRackspaceMessageQueue(),
			"rackspace_dns_domain":            tableRackspaceDNSDomain(),
		},
	}
	return p
}
