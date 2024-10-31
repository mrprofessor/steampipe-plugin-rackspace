package rackspace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// DNSDomain represents a Rackspace DNS domain
type DNSDomain struct {
	ID           string      `json:"id"`
	AccountID    string      `json:"accountId"`
	Name         string      `json:"name"`
	TTL          int         `json:"ttl"`
	EmailAddress string      `json:"emailAddress"`
	Updated      time.Time   `json:"updated"`
	Created      time.Time   `json:"created"`
	RecordsList  []DNSRecord `json:"recordsList"` // Slice to hold DNS records
}

// DNSRecord represents a single DNS record in a domain
type DNSRecord struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Data    string    `json:"data"`
	TTL     int       `json:"ttl"`
	Updated time.Time `json:"updated"`
	Created time.Time `json:"created"`
	Comment string    `json:"comment"`
}

func tableRackspaceDNSDomain() *plugin.Table {
	return &plugin.Table{
		Name:        "rackspace_dns_domain",
		Description: "Retrieve details of Rackspace DNS domains.",
		List: &plugin.ListConfig{
			Hydrate: listDNSDomains,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_STRING, Description: "The unique identifier of the DNS domain."},
			{Name: "account_id", Type: proto.ColumnType_STRING, Description: "The account ID associated with the DNS domain."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the DNS domain."},
			{Name: "ttl", Type: proto.ColumnType_INT, Description: "Time-to-live for the domain."},
			{Name: "email_address", Type: proto.ColumnType_STRING, Description: "The contact email address for the DNS domain."},
			{Name: "updated", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the DNS domain was last updated.", Transform: transform.FromField("Updated")},
			{Name: "created", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when the DNS domain was created.", Transform: transform.FromField("Created")},
			{Name: "records_list", Type: proto.ColumnType_JSON, Description: "List of DNS records for the domain.", Hydrate: getDNSRecords, Transform: transform.FromValue()},
		},
	}
}

func listDNSDomains(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)

	// Construct API request URL
	apiUrl := fmt.Sprintf(
		"https://dns.api.rackspacecloud.com/v1.0/%s/domains",
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
		return nil, fmt.Errorf("failed to retrieve DNS domains: %s", resp.Status)
	}

	// Parse the response
	var result struct {
		Domains []DNSDomain `json:"domains"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Stream each domain to the table
	for _, domain := range result.Domains {
		d.StreamListItem(ctx, domain)
	}

	return nil, nil
}

func getDNSRecords(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	domain := h.Item.(DNSDomain)

	// Get connection config
	rackspaceConfig := GetConfig(d.Connection)

	// Construct the API request URL for DNS records
	apiUrl := fmt.Sprintf(
		"https://dns.api.rackspacecloud.com/v1.0/%s/domains/%s/records",
		*rackspaceConfig.TenantID,
		domain.ID,
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
		return nil, fmt.Errorf("failed to retrieve DNS records for domain %s: %s", domain.Name, resp.Status)
	}

	// Parse the response into the records list
	var result struct {
		Records []DNSRecord `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Set the records list for the domain in Hydrate data
	return result.Records, nil
}
