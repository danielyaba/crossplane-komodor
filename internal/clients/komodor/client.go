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
	defaultBaseURL = "https://api.komodor.com/api/v2/realtime-monitors/config"
	apiKeyHeader   = "X-API-KEY"
)

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
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	u := c.baseURL.ResolveReference(rel)

	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
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
	var monitors []Monitor
	if err := json.NewDecoder(resp.Body).Decode(&monitors); err != nil {
		return nil, err
	}
	return monitors, nil
}

// GetMonitor fetches a monitor by ID.
func (c *Client) GetMonitor(ctx context.Context, id string) (*Monitor, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/%s", id), nil)
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
	// Check for common 404 patterns in error string
	if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "404 Not Found") {
		return true
	}
	return false
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
