package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

// Client is an HTTP client for the Zenodo API.
type Client struct {
	baseURL     string
	token       string
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

// NewClient creates a new API client with rate limiting.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: NewRateLimiter(),
	}
}

// Get performs a GET request and decodes the JSON response into result.
func (c *Client) Get(path string, query url.Values, result interface{}) error {
	return c.do(http.MethodGet, path, query, nil, result)
}

// Post performs a POST request with a JSON body and decodes the response.
func (c *Client) Post(path string, body interface{}, result interface{}) error {
	return c.do(http.MethodPost, path, nil, body, result)
}

// Put performs a PUT request with a JSON body and decodes the response.
func (c *Client) Put(path string, body interface{}, result interface{}) error {
	return c.do(http.MethodPut, path, nil, body, result)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string, result interface{}) error {
	return c.do(http.MethodDelete, path, nil, nil, result)
}

func (c *Client) do(method, path string, query url.Values, body interface{}, result interface{}) error {
	reqURL := c.baseURL + path
	if query != nil {
		reqURL += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	// Auth header.
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Content headers.
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	slog.Debug("API request", "method", method, "url", reqURL)

	// Rate limit before sending.
	if c.rateLimiter != nil {
		c.rateLimiter.Wait(path)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limiter from response headers.
	if c.rateLimiter != nil {
		c.rateLimiter.UpdateFromHeaders(resp, path)
	}

	slog.Debug("API response", "status", resp.StatusCode)

	// Read response body.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	// Check for error status codes.
	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, respBody)
	}

	// 204 No Content — nothing to decode.
	if resp.StatusCode == http.StatusNoContent || len(respBody) == 0 {
		return nil
	}

	// Decode response.
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

func parseAPIError(status int, body []byte) error {
	apiErr := &model.APIError{Status: status}

	// Try to parse structured error from Zenodo.
	if err := json.Unmarshal(body, apiErr); err != nil {
		// Couldn't parse — use raw body as message.
		apiErr.Message = string(body)
		if apiErr.Message == "" {
			apiErr.Message = http.StatusText(status)
		}
	}

	// Ensure status is set (API might not include it in body).
	apiErr.Status = status

	if hint := apiErr.Hint(); hint != "" {
		slog.Warn(hint)
	}

	return apiErr
}

// BaseURL returns the client's base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// GetRaw performs a GET request with a custom Accept header and returns raw bytes.
// Useful for non-JSON formats like BibTeX or DataCite XML.
func (c *Client) GetRaw(path string, accept string) ([]byte, error) {
	reqURL := c.baseURL + path

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", accept)

	slog.Debug("API request (raw)", "url", reqURL, "accept", accept)

	if c.rateLimiter != nil {
		c.rateLimiter.Wait(path)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if c.rateLimiter != nil {
		c.rateLimiter.UpdateFromHeaders(resp, path)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, body)
	}

	return body, nil
}
