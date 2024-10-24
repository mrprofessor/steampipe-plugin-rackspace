package main

import (
	"steampipe-plugin-rackspace/rackspace"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{PluginFunc: rackspace.Plugin})
}
