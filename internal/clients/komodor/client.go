package komodor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL  = "https://api.komodor.com/api/v2/realtime-monitors/config"
	clustersBaseURL = "https://api.komodor.com/api/v2/clusters"
	apiKeyHeader    = "X-API-KEY"
)

// NotFoundError represents a 404 Not Found error from the Komodor API
type NotFoundError struct {
	ID string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("monitor with ID %s not found", e.ID)
}

// Client is a Komodor API client.
type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Komodor API client.
func NewClient(apiKey string) *Client {
	base, _ := url.Parse(defaultBaseURL)
	return &Client{
		baseURL:    base,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// doRequest executes an HTTP request with authentication.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	// Use url.JoinPath to properly append the path to the base URL
	u, err := url.JoinPath(c.baseURL.String(), path)
	if err != nil {
		return nil, err
	}

	// URL construction completed

	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set(apiKeyHeader, c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return c.httpClient.Do(req)
}

// ListMonitors fetches all monitors.
func (c *Client) ListMonitors(ctx context.Context) ([]Monitor, error) {
	resp, err := c.doRequest(ctx, "GET", "", nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			_ = cerr // explicitly ignore
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// Read the response body to debug the issue
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to unmarshal as array first
	var monitors []Monitor
	if err := json.Unmarshal(body, &monitors); err != nil {
		// If that fails, try to unmarshal as object with data field
		var response struct {
			Data []Monitor `json:"data"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal monitors response: %w, body: %s", err, string(body[:min(len(body), 200)]))
		}
		monitors = response.Data
	}

	return monitors, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetMonitor fetches a monitor by ID.
func (c *Client) GetMonitor(ctx context.Context, id string) (*Monitor, error) {
	path := fmt.Sprintf("/%s", id)
	// Making GET request to Komodor API

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			_ = cerr // explicitly ignore
		}
	}()

	// Response received from Komodor API

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{ID: id}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var m Monitor
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}

	// Successfully decoded monitor response
	return &m, nil
}

// CreateMonitor creates a new monitor.
func (c *Client) CreateMonitor(ctx context.Context, monitor *Monitor) (*Monitor, error) {
	resp, err := c.doRequest(ctx, "POST", "", monitor)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			_ = cerr // explicitly ignore
		}
	}()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var m Monitor
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

// UpdateMonitor updates an existing monitor by ID.
func (c *Client) UpdateMonitor(ctx context.Context, id string, monitor *Monitor) (*Monitor, error) {
	resp, err := c.doRequest(ctx, "PATCH", fmt.Sprintf("/%s", id), monitor)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			_ = cerr // explicitly ignore
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var m Monitor
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

// DeleteMonitor deletes a monitor by ID.
func (c *Client) DeleteMonitor(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/%s", id), nil)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			_ = cerr // explicitly ignore
		}
	}()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

// IsNotFound returns true if the error is a 404 Not Found from the Komodor API.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for our custom NotFoundError
	var notFoundErr *NotFoundError
	if errors.As(err, &notFoundErr) {
		return true
	}

	// Check for common 404 patterns in error string
	if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "404 Not Found") {
		return true
	}
	return false
}

// Cluster represents a Komodor cluster
type Cluster struct {
	ID           string            `json:"id,omitempty"`
	Name         string            `json:"name"`
	APIServerURL string            `json:"apiServerUrl,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	CreatedAt    string            `json:"createdAt,omitempty"`
	UpdatedAt    string            `json:"updatedAt,omitempty"`
}

// ClustersResponse represents the response from the clusters API
type ClustersResponse struct {
	Data struct {
		Clusters []Cluster `json:"clusters"`
	} `json:"data"`
}

// ListClusters fetches all clusters from Komodor
func (c *Client) ListClusters(ctx context.Context) ([]Cluster, error) {
	clustersURL, _ := url.Parse(clustersBaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", clustersURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(apiKeyHeader, c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			_ = cerr // explicitly ignore
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var response ClustersResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.Data.Clusters, nil
}

// ValidateCluster checks if a cluster exists using the proper clusters API
func (c *Client) ValidateCluster(ctx context.Context, clusterName string) (bool, error) {
	clusters, err := c.ListClusters(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to list clusters to validate cluster: %w", err)
	}

	// Check if the cluster exists
	for _, cluster := range clusters {
		if cluster.Name == clusterName {
			return true, nil
		}
	}

	return false, nil
}

// Monitor is a struct representing a Komodor Real Time Monitor.
// Fill this out based on the API schema.
type Monitor struct {
	ID           string                   `json:"id,omitempty"`
	CreatedAt    string                   `json:"createdAt,omitempty"`
	UpdatedAt    string                   `json:"updatedAt,omitempty"`
	IsDeleted    bool                     `json:"isDeleted,omitempty"`
	Name         string                   `json:"name"`
	Sensors      []map[string]interface{} `json:"sensors,omitempty"`
	Sinks        map[string]interface{}   `json:"sinks,omitempty"`
	Active       bool                     `json:"active"`
	Type         string                   `json:"type"`
	Variables    map[string]interface{}   `json:"variables,omitempty"`
	SinksOptions map[string][]string      `json:"sinksOptions,omitempty"`
}
 