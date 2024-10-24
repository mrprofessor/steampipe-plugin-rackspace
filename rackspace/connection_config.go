package rackspace

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

type rackspaceConfig struct {
	IdentityEndpoint *string `cty:"identity_endpoint"`
	TenantID         *string `cty:"tenant_id"`
	TokenID          *string `cty:"token_id"`
	Region           *string `cty:"region"`
}

var ConfigSchema = map[string]*schema.Attribute{
	"identity_endpoint": {
		Type: schema.TypeString,
	},
	"tenant_id": {
		Type: schema.TypeString,
	},
	"token_id": {
		Type: schema.TypeString,
	},
	"region": {
		Type: schema.TypeString,
	},
}

func ConfigInstance() interface{} {
	return &rackspaceConfig{}
}

// GetConfig :: retrieve and cast connection config from query data
func GetConfig(connection *plugin.Connection) rackspaceConfig {
	if connection == nil || connection.Config == nil {
		return rackspaceConfig{}
	}
	config, _ := connection.Config.(rackspaceConfig)
	return config
}
