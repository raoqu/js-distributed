package net

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPResponse represents the response from an HTTP request
type HTTPResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       string
	Error      error
}

// HTTPClient is a wrapper around http.Client with additional functionality
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTP client with the specified timeout
func NewHTTPClient(timeoutSeconds int) *HTTPClient {
	timeout := time.Duration(timeoutSeconds) * time.Second
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// DefaultHTTPClient returns a new HTTP client with a default timeout of 30 seconds
func DefaultHTTPClient() *HTTPClient {
	return NewHTTPClient(30)
}

// Get performs a synchronous HTTP GET request
func (c *HTTPClient) Get(urlStr string, params map[string]string, headers map[string]string) HTTPResponse {
	// Add query parameters to URL if provided
	if len(params) > 0 {
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			return HTTPResponse{Error: fmt.Errorf("invalid URL: %v", err)}
		}
		
		q := parsedURL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		parsedURL.RawQuery = q.Encode()
		urlStr = parsedURL.String()
	}

	// Create request
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return HTTPResponse{Error: fmt.Errorf("error creating request: %v", err)}
	}

	// Add headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return c.doRequest(req)
}

// Post performs a synchronous HTTP POST request
func (c *HTTPClient) Post(urlStr string, body interface{}, headers map[string]string) HTTPResponse {
	var bodyReader io.Reader
	var contentType string

	// Process body based on type
	switch b := body.(type) {
	case string:
		bodyReader = strings.NewReader(b)
		contentType = "text/plain"
	case []byte:
		bodyReader = bytes.NewReader(b)
		contentType = "application/octet-stream"
	case nil:
		bodyReader = nil
	default:
		// Try to marshal as JSON
		jsonData, err := json.Marshal(body)
		if err != nil {
			return HTTPResponse{Error: fmt.Errorf("error marshaling JSON body: %v", err)}
		}
		bodyReader = bytes.NewReader(jsonData)
		contentType = "application/json"
	}

	// Create request
	req, err := http.NewRequest(http.MethodPost, urlStr, bodyReader)
	if err != nil {
		return HTTPResponse{Error: fmt.Errorf("error creating request: %v", err)}
	}

	// Set default content type if not overridden in headers
	if contentType != "" && headers["Content-Type"] == "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Add headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return c.doRequest(req)
}

// doRequest executes the HTTP request and returns a response
func (c *HTTPClient) doRequest(req *http.Request) HTTPResponse {
	resp, err := c.client.Do(req)
	if err != nil {
		return HTTPResponse{Error: fmt.Errorf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return HTTPResponse{Error: fmt.Errorf("error reading response body: %v", err)}
	}

	return HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       string(body),
		Error:      nil,
	}
}
