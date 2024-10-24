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

// Volume represents the response format from Rackspace Block Storage v1 API.
type Volume struct {
	ID                 string            `json:"id"`
	DisplayName        string            `json:"display_name"`
	Status             string            `json:"status"`
	Size               int               `json:"size"`
	CreatedAt          string            `json:"created_at"`
	AvailabilityZone   string            `json:"availability_zone"`
	Bootable           string            `json:"bootable"`
	Encrypted          bool              `json:"encrypted"`
	VolumeType         string            `json:"volume_type"`
	SnapshotID         string            `json:"snapshot_id"`
	SourceVolID        string            `json:"source_volid"`
	DisplayDescription string            `json:"display_description"`
	MultiAttach        string            `json:"multiattach"` // NULLABLE field
	Metadata           map[string]string `json:"metadata"`
	Attachments        []Attachment      `json:"attachments"`
}

// Attachment represents the attached devices data for a volume.
type Attachment struct {
	ServerID     string `json:"server_id"`
	AttachmentID string `json:"attachment_id"`
	HostName     string `json:"host_name"`
	VolumeID     string `json:"volume_id"`
	Device       string `json:"device"`
	ID           string `json:"id"`
}

func tableRackspaceVolume() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_volume",
		Description: "Rackspace Block Storage Volumes",
		List: &plugin.ListConfig{
			Hydrate: listVolumes,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getVolume,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_STRING, Description: "The ID of the volume"},
			{Name: "display_name", Type: proto.ColumnType_STRING, Description: "The name of the volume"},
			{Name: "status", Type: proto.ColumnType_STRING, Description: "The current status of the volume (e.g., 'available', 'in-use')."},
			{Name: "size", Type: proto.ColumnType_INT, Description: "Size of the volume in GB"},
			{Name: "created_at", Type: proto.ColumnType_TIMESTAMP, Description: "The timestamp when the volume was created", Transform: transform.FromGo()},
			{Name: "availability_zone", Type: proto.ColumnType_STRING, Description: "The availability zone where the volume is located"},
			{Name: "bootable", Type: proto.ColumnType_STRING, Description: "Whether the volume is bootable (true/false)"},
			{Name: "encrypted", Type: proto.ColumnType_BOOL, Description: "Whether the volume is encrypted (true/false)"},
			{Name: "volume_type", Type: proto.ColumnType_STRING, Description: "The type of the volume (e.g., SSD, HDD)"},
			{Name: "snapshot_id", Type: proto.ColumnType_STRING, Description: "The ID of the snapshot from which the volume was created, if any"},
			{Name: "source_volid", Type: proto.ColumnType_STRING, Description: "The ID of the source volume, if any"},
			{Name: "display_description", Type: proto.ColumnType_STRING, Description: "The description of the volume"},
			{Name: "multiattach", Type: proto.ColumnType_STRING, Description: "Whether the volume supports multiple attachments (true/false)", Transform: transform.FromField("MultiAttach")},
			{Name: "metadata", Type: proto.ColumnType_JSON, Description: "Metadata associated with the volume"},
			{Name: "attachments", Type: proto.ColumnType_JSON, Description: "The attached devices information for the volume"},
		},
	}
}

// listVolumes fetches all available volumes from the Rackspace v1 Block Storage API
func listVolumes(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)

	// Construct API request URL
	apiURL := fmt.Sprintf(
		"https://%s.blockstorage.api.rackspacecloud.com/v1/%s/volumes",
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
		Volumes []Volume `json:"volumes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Stream each volume
	for _, volume := range result.Volumes {
		d.StreamListItem(ctx, volume)
	}

	return nil, nil
}

// getVolume fetches a single volume by its ID
func getVolume(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// Get the volume ID from the query
	volumeID := d.EqualsQuals["id"].GetStringValue()

	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)

	// Construct the API request URL with the volume ID
	apiURL := fmt.Sprintf(
		"https://%s.blockstorage.api.rackspacecloud.com/v1/%s/volumes/%s",
		*rackspaceConfig.Region,
		*rackspaceConfig.TenantID,
		volumeID,
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
		Volume Volume `json:"volume"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Return the volume data
	return result.Volume, nil
}
