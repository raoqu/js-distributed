package script

import (
	"encoding/json"
	"fmt"
	"main/util/net"

	"github.com/dop251/goja"
)

// RequestOptions holds all the options for a fetch request
// It encapsulates all parameters that can be passed to the net.fetch API
type RequestOptions struct {
	// Method specifies the HTTP method (GET, POST, etc.)
	Method string
	// Params contains URL query parameters for GET requests
	Params map[string]string
	// Headers contains HTTP request headers
	Headers map[string]string
	// Body contains the request body for POST requests
	Body interface{}
	// Timeout specifies the request timeout in seconds
	Timeout int
}

// Net_fetch implements a synchronous HTTP request function for scripts
// This function provides a JavaScript-friendly API for making HTTP requests
// from scripts. It supports both GET and POST methods with various options.
//
// Usage in JS:
//
//	net.fetch(url, {
//	  method: "GET" | "POST", // default is GET if not specified
//	  params: {key: value},   // URL parameters for GET requests
//	  headers: {key: value},  // HTTP headers
//	  body: object | string,  // Request body for POST requests
//	  timeout: number         // Timeout in seconds (default: 30)
//	})
//
// Returns an object with the following properties:
//
//	{
//	  status: number,         // HTTP status code
//	  headers: object,        // Response headers
//	  data: string,           // Response body as string
//	  error: string,          // Error message if request failed, null otherwise
//	  json: object            // Parsed JSON response (if content-type is application/json)
//	}
func Net_fetch(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	if len(call.Arguments) < 1 {
		return nil, fmt.Errorf("net.fetch requires at least a URL argument")
	}

	// Get URL
	urlStr := call.Arguments[0].String()

	// Parse options
	opts := parseRequestOptions(rt, call)

	// Execute the request
	response := executeRequest(urlStr, opts)

	// Process the response
	result := processResponse(response)

	return rt.ToValue(result), nil
}

// parseRequestOptions extracts and processes options from JavaScript arguments
// It converts the JavaScript options object into a Go RequestOptions struct
func parseRequestOptions(rt *goja.Runtime, call goja.FunctionCall) RequestOptions {
	// Default values
	opts := RequestOptions{
		Method:  "GET",
		Timeout: 30,
	}

	// Process options if provided
	if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) && !goja.IsNull(call.Arguments[1]) {
		options := call.Arguments[1].ToObject(rt)
		if options != nil {
			// Extract method
			opts.Method = extractStringOption(rt, options, "method", opts.Method)

			// Extract params
			opts.Params = extractMapOption(rt, options, "params")

			// Extract headers
			opts.Headers = extractMapOption(rt, options, "headers")

			// Extract body
			opts.Body = extractBodyOption(rt, options)

			// Extract timeout
			opts.Timeout = extractTimeoutOption(rt, options, opts.Timeout)
		}
	}

	return opts
}

// extractStringOption extracts a string option with a default value
// It safely handles undefined and null values in JavaScript
func extractStringOption(rt *goja.Runtime, options *goja.Object, key string, defaultValue string) string {
	val := options.Get(key)
	if val != nil && !goja.IsUndefined(val) && !goja.IsNull(val) {
		return val.String()
	}
	return defaultValue
}

// extractMapOption extracts a map[string]string from a JavaScript object
// It converts a JavaScript object with string keys and values to a Go map
func extractMapOption(rt *goja.Runtime, options *goja.Object, key string) map[string]string {
	val := options.Get(key)
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return nil
	}

	result := make(map[string]string)
	obj := val.ToObject(rt)
	if obj != nil {
		for _, k := range obj.Keys() {
			itemVal := obj.Get(k)
			if itemVal != nil {
				result[k] = itemVal.String()
			}
		}
	}

	return result
}

// extractBodyOption extracts the body option
// It handles both string and object body types from JavaScript
func extractBodyOption(rt *goja.Runtime, options *goja.Object) interface{} {
	val := options.Get("body")
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return nil
	}

	if val.ExportType().Kind().String() == "string" {
		return val.String()
	}
	return val.Export()
}

// extractTimeoutOption extracts the timeout option with validation
// It ensures the timeout value is positive, otherwise returns the default value
func extractTimeoutOption(rt *goja.Runtime, options *goja.Object, defaultValue int) int {
	val := options.Get("timeout")
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return defaultValue
	}

	timeout := int(val.ToInteger())
	if timeout <= 0 {
		return defaultValue // Default if invalid
	}
	return timeout
}

// executeRequest performs the actual HTTP request based on the method
// It creates an HTTP client with the specified timeout and executes the request
func executeRequest(urlStr string, opts RequestOptions) net.HTTPResponse {
	// Create HTTP client with specified timeout
	client := net.NewHTTPClient(opts.Timeout)

	// Execute request based on method
	switch opts.Method {
	case "POST":
		return client.Post(urlStr, opts.Body, opts.Headers)
	default: // Default to GET
		return client.Get(urlStr, opts.Params, opts.Headers)
	}
}

// processResponse converts an HTTP response to a JavaScript-friendly format
// It creates a map that will be converted to a JavaScript object with status, headers, data, and error
func processResponse(response net.HTTPResponse) map[string]interface{} {
	// Create result object
	result := make(map[string]interface{})
	result["status"] = response.StatusCode
	result["headers"] = response.Headers
	result["data"] = response.Body

	// Handle error if any
	if response.Error != nil {
		result["error"] = response.Error.Error()
	} else {
		result["error"] = nil
	}

	// Try to parse JSON response if content-type is application/json
	processJsonResponse(response, result)

	return result
}

// processJsonResponse attempts to parse JSON responses and add them to the result
// If the response has a JSON content type, it will try to parse the body as JSON
// and add it to the result map as the 'json' property
func processJsonResponse(response net.HTTPResponse, result map[string]interface{}) {
	contentType := ""
	if ct, ok := response.Headers["Content-Type"]; ok && len(ct) > 0 {
		contentType = ct[0]
	}

	if contentType == "application/json" || contentType == "application/json; charset=utf-8" {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(response.Body), &jsonData); err == nil {
			result["json"] = jsonData
		}
	}
}
