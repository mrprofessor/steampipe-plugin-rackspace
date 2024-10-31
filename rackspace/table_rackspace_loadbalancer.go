package rackspace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

type LoadBalancer struct {
	ID                int               `json:"id"`
	Name              string            `json:"name"`
	Protocol          string            `json:"protocol"`
	Port              int               `json:"port"`
	Algorithm         string            `json:"algorithm"`
	Status            string            `json:"status"`
	Timeout           int               `json:"timeout"`
	NodeCount         int               `json:"nodeCount"`
	ConnectionLogging ConnectionLogging `json:"connectionLogging"`
	HttpsRedirect     bool              `json:"httpsRedirect"`
	HalfClosed        bool              `json:"halfClosed"`
	ContentCaching    ContentCaching    `json:"contentCaching"`
	Cluster           Cluster           `json:"cluster"`
	Updated           TimeWrapper       `json:"updated"`
	Created           TimeWrapper       `json:"created"`
	VirtualIps        []VirtualIP       `json:"virtualIps"`
	SourceAddresses   SourceAddresses   `json:"sourceAddresses"`
}

type ConnectionLogging struct {
	Enabled bool `json:"enabled"`
}

type ContentCaching struct {
	Enabled bool `json:"enabled"`
}

type Cluster struct {
	Name string `json:"name"`
}

type VirtualIP struct {
	ID        int    `json:"id"`
	IPAddress string `json:"address"`
	IPVersion string `json:"ipVersion"`
	Type      string `json:"type"`
}

type SourceAddresses struct {
	IPv6Public     string `json:"ipv6Public"`
	IPv4Public     string `json:"ipv4Public"`
	IPv4ServiceNet string `json:"ipv4Servicenet"`
}

func tableRackspaceLoadBalancer() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_loadbalancer",
		Description: "Retrieve details of Rackspace Load Balancers.",
		List: &plugin.ListConfig{
			Hydrate: listLoadBalancers,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getLoadBalancer,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "The unique ID of the Load Balancer."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the Load Balancer."},
			{Name: "protocol", Type: proto.ColumnType_STRING, Description: "The protocol used by the Load Balancer."},
			{Name: "port", Type: proto.ColumnType_INT, Description: "The port on which the Load Balancer listens."},
			{Name: "algorithm", Type: proto.ColumnType_STRING, Description: "The load balancing algorithm."},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "The status of the Load Balancer."},
			{Name: "timeout", Type: proto.ColumnType_INT, Description: "The timeout for the Load Balancer."},
			{Name: "node_count", Type: proto.ColumnType_INT, Description: "The number of nodes attached to the Load Balancer.", Transform: transform.FromField("NodeCount")},
			{Name: "updated", Type: proto.ColumnType_TIMESTAMP, Description: "The timestamp when the Load Balancer was last updated.", Transform: transform.FromField("Updated.Time")},
			{Name: "created", Type: proto.ColumnType_TIMESTAMP, Description: "The timestamp when the Load Balancer was created.", Transform: transform.FromField("Created.Time")},
			{Name: "virtual_ips", Type: proto.ColumnType_JSON, Description: "List of virtual IPs associated with the Load Balancer."},
			{Name: "connection_logging", Type: proto.ColumnType_BOOL, Description: "Whether connection logging is enabled for the Load Balancer.", Transform: transform.FromField("ConnectionLogging.Enabled")},
			{Name: "https_redirect", Type: proto.ColumnType_BOOL, Description: "Whether HTTPS redirect is enabled for the Load Balancer.", Transform: transform.FromField("HttpsRedirect")},
			{Name: "half_closed", Type: proto.ColumnType_BOOL, Description: "Whether half-closed connections are enabled for the Load Balancer."},
			{Name: "content_caching", Type: proto.ColumnType_BOOL, Description: "Whether content caching is enabled for the Load Balancer.", Transform: transform.FromField("ContentCaching.Enabled")},
			{Name: "cluster_name", Type: proto.ColumnType_STRING, Description: "The cluster name associated with the Load Balancer.", Transform: transform.FromField("Cluster.Name")},
			{Name: "source_addresses", Type: proto.ColumnType_JSON, Description: "Source IP addresses for the Load Balancer, including IPv4 and IPv6 addresses."},
		},
	}
}

func listLoadBalancers(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)

	// Construct API request URL
	apiUrl := fmt.Sprintf(
		"https://%s.loadbalancers.api.rackspacecloud.com/v1.0/%s/loadbalancers",
		*rackspaceConfig.Region,
		*rackspaceConfig.TenantID,
	)

	// Create an HTTP client and make the request
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	// Set the Authorization header
	req.Header.Add("X-Auth-Token", *rackspaceConfig.TokenID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve load balancers: %s", resp.Status)
	}

	// Parse the response
	var result struct {
		LoadBalancers []LoadBalancer `json:"loadBalancers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Stream each load balancer to the table
	for _, lb := range result.LoadBalancers {
		d.StreamListItem(ctx, lb)
	}

	return nil, nil
}

func getLoadBalancer(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)
	loadBalancerID := d.EqualsQuals["id"].GetInt64Value()

	// Construct the API request URL
	apiUrl := fmt.Sprintf(
		"https://%s.loadbalancers.api.rackspacecloud.com/v1.0/%s/loadbalancers/%d",
		*rackspaceConfig.Region,
		*rackspaceConfig.TenantID,
		loadBalancerID,
	)

	// Create an HTTP request
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	// Set the Authorization header
	req.Header.Add("X-Auth-Token", *rackspaceConfig.TokenID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve load balancer: %s", resp.Status)
	}

	// Parse the response
	var result struct {
		LoadBalancer LoadBalancer `json:"loadBalancer"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.LoadBalancer, nil
}
