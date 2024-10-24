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

// Snapshot represents the response format from Rackspace Block Storage v1 API.
type Snapshot struct {
	ID                 string            `json:"id"`
	DisplayName        string            `json:"display_name"`
	VolumeID           string            `json:"volume_id"`
	Status             string            `json:"status"`
	Size               int               `json:"size"`
	CreatedAt          string            `json:"created_at"`
	DisplayDescription string            `json:"display_description"`
	Metadata           map[string]string `json:"metadata"`
}

func tableRackspaceSnapshot() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_snapshot",
		Description: "Rackspace Block Storage Snapshots",
		List: &plugin.ListConfig{
			Hydrate: listSnapshots,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getSnapshot,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_STRING, Description: "The ID of the snapshot"},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "The name of the snapshot"},
			{Name: "volume_id", Type: proto.ColumnType_STRING, Description: "The ID of the volume from which the snapshot was created"},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "The current status of the snapshot (e.g., 'available', 'error')."},
			{Name: "size", Type: proto.ColumnType_INT, Description: "Size of the snapshot in GB"},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Description: "The timestamp when the snapshot was created", Transform: transform.FromGo()},
			{Name: "display_description", Type: proto.ColumnType_STRING, Description: "Description of the snapshot"},
			{Name: "metadata", Type: proto.ColumnType_JSON, Description: "Metadata associated with the snapshot"},
		},
	}
}

// listSnapshots fetches all available snapshots from the Rackspace v1 Block Storage API
func listSnapshots(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)

	// Construct API request URL
	apiURL := fmt.Sprintf(
		"https://%s.blockstorage.api.rackspacecloud.com/v1/%s/snapshots",
		*rackspaceConfig.Region,
		*rackspaceConfig.TenantID,
	)

	// Create HTTP request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set the Authorization header
	req.Header.Add("X-Auth-Token", *rackspaceConfig.TokenID)

	// Perform the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response
	var result struct {
		Snapshots []Snapshot `json:"snapshots"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Stream each snapshot
	for _, snapshot := range result.Snapshots {
		d.StreamListItem(ctx, snapshot)
	}

	return nil, nil
}

// getSnapshot fetches a single snapshot by its ID
func getSnapshot(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// Get the snapshot ID from the query
	snapshotID := d.EqualsQuals["id"].GetStringValue()

	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)

	// Construct the API request URL with the snapshot ID
	apiURL := fmt.Sprintf(
		"https://%s.blockstorage.api.rackspacecloud.com/v1/%s/snapshots/%s",
		*rackspaceConfig.Region,
		*rackspaceConfig.TenantID,
		snapshotID,
	)

	// Create the HTTP request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set the Authorization header
	req.Header.Add("X-Auth-Token", *rackspaceConfig.TokenID)

	// Perform the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response
	var result struct {
		Snapshot Snapshot `json:"snapshot"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Return the snapshot data
	return result.Snapshot, nil
}
