package provider_shadeform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	baseURL           = "https://api.shadeform.ai/v1"
	apiKeyHeader      = "X-API-KEY"
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"

	// Instance routes
	instanceCreateRoute = "/instances/create"
	instanceInfoRoute   = "/instances/%s/info"
	instanceUpdateRoute = "/instances/%s/update"
	instanceDeleteRoute = "/instances/%s/delete"
	instanceTypesRoute  = "/instances/types"

	// Volume routes
	volumeCreateRoute = "/volumes/create"
	volumeInfoRoute   = "/volumes/%s/info"
	volumeDeleteRoute = "/volumes/%s/delete"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("SHADEFORM_API_KEY")
	}

	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) CreateInstance(requestBody map[string]interface{}) (map[string]interface{}, error) {
	return c.makeRequest("POST", instanceCreateRoute, requestBody, true)
}

func (c *Client) GetInstance(instanceID string) (map[string]interface{}, error) {
	return c.makeRequest("GET", fmt.Sprintf(instanceInfoRoute, instanceID), nil, true)
}

func (c *Client) UpdateInstance(instanceID string, requestBody map[string]interface{}) error {
	return c.makeRequestNoResponse("POST", fmt.Sprintf(instanceUpdateRoute, instanceID), requestBody)
}

func (c *Client) DeleteInstance(instanceID string) error {
	return c.makeRequestNoResponse("POST", fmt.Sprintf(instanceDeleteRoute, instanceID), nil)
}

func (c *Client) GetInstanceTypes(params map[string]string) (map[string]interface{}, error) {
	query := ""
	if len(params) > 0 {
		query = "?"
		first := true
		for key, value := range params {
			if !first {
				query += "&"
			}
			query += fmt.Sprintf("%s=%s", key, value)
			first = false
		}
	}

	return c.makeRequest("GET", instanceTypesRoute+query, nil, true)
}

func (c *Client) CreateVolume(requestBody map[string]interface{}) (map[string]interface{}, error) {
	return c.makeRequest("POST", volumeCreateRoute, requestBody, true)
}

func (c *Client) GetVolume(volumeID string) (map[string]interface{}, error) {
	return c.makeRequest("GET", fmt.Sprintf(volumeInfoRoute, volumeID), nil, true)
}

func (c *Client) DeleteVolume(volumeID string) error {
	return c.makeRequestNoResponse("POST", fmt.Sprintf(volumeDeleteRoute, volumeID), nil)
}

func (c *Client) makeRequest(method, path string, body interface{}, expectResponse bool) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(apiKeyHeader, c.apiKey)
	if body != nil {
		req.Header.Set(contentTypeHeader, contentTypeJSON)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// If we don't expect a response (like for delete operations), return early
	if !expectResponse {
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

func (c *Client) makeRequestNoResponse(method, path string, body interface{}) error {
	_, err := c.makeRequest(method, path, body, false)
	return err
}
