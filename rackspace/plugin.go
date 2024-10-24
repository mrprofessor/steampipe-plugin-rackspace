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
			"rackspace_compute": tableRackspaceCompute(),
			"rackspace_image":   tableRackspaceImage(),
			"rackspace_snapshot": tableRackspaceSnapshot(),
			"rackspace_volume":   tableRackspaceVolume(),
		},
	}
	return p
}
