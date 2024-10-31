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

type MessageQueue struct {
	Name     string                 `json:"name"`
	Href     string                 `json:"href"`
	Stats    MessageQueueStats      `json:"stats"`
	Metadata map[string]interface{} `json:"metadata"`
}

// MessageQueueStats represents the structure for queue statistics
type MessageQueueStats struct {
	Messages MessageStats `json:"messages"`
}

// MessageStats holds the details of the queue messages statistics
type MessageStats struct {
	Claimed int                 `json:"claimed"`
	Oldest  MessageQueueMessage `json:"oldest"`
	Total   int                 `json:"total"`
	Newest  MessageQueueMessage `json:"newest"`
	Free    int                 `json:"free"`
}

// MessageQueueMessage holds details of a specific message within the stats
type MessageQueueMessage struct {
	Age     int    `json:"age"`
	Href    string `json:"href"`
	Created string `json:"created"`
}

func tableRackspaceMessageQueue() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_message_queue",
		Description: "Retrieve details of Rackspace Message Queues.",
		List: &plugin.ListConfig{
			Hydrate: listQueues,
		},
		Columns: []*plugin.Column{
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the queue."},
			{Name: "href", Type: proto.ColumnType_STRING, Description: "The URL of the queue."},
			{Name: "stats", Type: proto.ColumnType_JSON, Description: "Statistics about the queue.", Hydrate: getQueueStats, Transform: transform.FromValue()},
			{Name: "metadata", Type: proto.ColumnType_JSON, Description: "Metadata associated with the queue.", Hydrate: getQueueMetadata, Transform: transform.FromValue()},
		},
	}
}

func listQueues(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)

	// Construct API request URL
	apiUrl := fmt.Sprintf(
		"https://%s.queues.api.rackspacecloud.com/v1/%s/queues",
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
		return nil, fmt.Errorf("failed to retrieve queues: %s", resp.Status)
	}

	// Parse the response
	var result struct {
		Queues []MessageQueue `json:"queues"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Stream each queue to the table
	for _, queue := range result.Queues {
		d.StreamListItem(ctx, queue)
	}

	return nil, nil
}

func getQueueStats(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	queue := h.Item.(MessageQueue)

	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)

	// Construct the API request URL
	apiUrl := fmt.Sprintf(
		"https://%s.queues.api.rackspacecloud.com/v1/%s/queues/%s/stats",
		*rackspaceConfig.Region,
		*rackspaceConfig.TenantID,
		queue.Name,
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
		return nil, fmt.Errorf("failed to retrieve queue stats: %s", resp.Status)
	}

	// Parse the response into the Stats struct
	var result struct {
		Messages MessageStats `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Messages, nil
}

func getQueueMetadata(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	queue := h.Item.(MessageQueue)

	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)

	// Construct the API request URL for metadata
	apiUrl := fmt.Sprintf(
		"https://%s.queues.api.rackspacecloud.com/v1/%s/queues/%s/metadata",
		*rackspaceConfig.Region,
		*rackspaceConfig.TenantID,
		queue.Name,
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
		return nil, fmt.Errorf("failed to retrieve queue metadata: %s", resp.Status)
	}

	// Parse the response into a map
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
