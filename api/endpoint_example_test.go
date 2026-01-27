package api_test

import (
	"fmt"
	"time"

	"github.com/trimble-oss/tierceron-core/v2/api"
)

// ExampleEndpoint_Call_rest demonstrates calling a REST endpoint using the generic Call method
func ExampleEndpoint_Call_rest() {
	// Define a REST endpoint
	endpoint := api.Endpoint{
		FriendlyName: "GitHub API",
		URL:          "https://api.github.com/users/octocat",
		Type:         api.EndpointTypeREST,
		Timeout:      10 * time.Second,
	}

	// Create request parameters
	params := map[string]any{
		"method": "GET",
		"headers": map[string]string{
			"Accept":     "application/json",
			"User-Agent": "MyApp/1.0",
		},
	}

	// Make the call - context is managed internally
	result, err := endpoint.Call(params, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Access response data
	statusCode := result["statusCode"].(int)
	fmt.Printf("Status: %d\n", statusCode)

	if body, ok := result["body"]; ok {
		fmt.Printf("Body type: %T\n", body)
	}
}

// ExampleEndpoint_Call_restPost demonstrates a POST request to a REST endpoint
func ExampleEndpoint_Call_restPost() {
	endpoint := api.Endpoint{
		FriendlyName: "User API",
		URL:          "https://api.example.com/users",
		Type:         api.EndpointTypeREST,
		Timeout:      15 * time.Second,
	}

	// Create request with JSON body
	params := map[string]any{
		"method": "POST",
		"headers": map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
		"body": map[string]any{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		},
	}

	result, err := endpoint.Call(params, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Check response
	if errorMsg, ok := result["error"].(string); ok {
		fmt.Printf("API Error: %s\n", errorMsg)
	} else {
		fmt.Printf("Status: %d\n", result["statusCode"])
	}
}

// ExampleEndpoint_Call_soap demonstrates calling a SOAP endpoint
func ExampleEndpoint_Call_soap() {
	endpoint := api.Endpoint{
		FriendlyName: "Weather Service",
		URL:          "http://www.example.com/soap/weatherservice",
		Type:         api.EndpointTypeSOAP,
		Timeout:      20 * time.Second,
		WSDLUrl:      "http://www.example.com/soap/weatherservice?wsdl",
	}

	// Simple parameter map - SOAP envelope generated automatically
	params := map[string]any{
		"method": "GetWeather",
		"body": map[string]any{
			"City":    "New York",
			"Country": "USA",
		},
	}

	result, err := endpoint.Call(params, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Access SOAP response
	fmt.Printf("Status: %d\n", result["statusCode"])
	if body, ok := result["body"].(string); ok {
		fmt.Printf("Response length: %d\n", len(body))
	}
}

// ExampleEndpoint_Call_withTLS demonstrates calling an endpoint with TLS configuration
func ExampleEndpoint_Call_withTLS() {
	endpoint := api.Endpoint{
		FriendlyName: "Secure API",
		URL:          "https://secure-api.example.com/data",
		Type:         api.EndpointTypeREST,
		Timeout:      20 * time.Second,
	}

	// Configure TLS
	config := &api.APICallerConfig{
		TLSCertPath: "/path/to/client-cert.pem",
		TLSKeyPath:  "/path/to/client-key.pem",
		CACertPath:  "/path/to/ca-cert.pem",
	}

	params := map[string]any{
		"method": "GET",
		"headers": map[string]string{
			"Authorization": "Bearer secure-token",
		},
	}

	result, err := endpoint.Call(params, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Status: %d\n", result["statusCode"])
}

// ExampleEndpoint_Call_minimal demonstrates minimal usage with defaults
func ExampleEndpoint_Call_minimal() {
	endpoint := api.Endpoint{
		FriendlyName: "Simple API",
		URL:          "https://httpbin.org/get",
		Type:         api.EndpointTypeREST,
		Timeout:      0, // Use default 30s
	}

	// Minimal call - method defaults to GET, context and config use defaults
	result, err := endpoint.Call(nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Success! Status: %d\n", result["statusCode"])
}

// ExampleEndpoint_Call_errorHandling demonstrates error handling
func ExampleEndpoint_Call_errorHandling() {
	endpoint := api.Endpoint{
		FriendlyName: "Test API",
		URL:          "https://api.example.com/endpoint",
		Type:         api.EndpointTypeREST,
		Timeout:      10 * time.Second,
	}

	params := map[string]any{
		"method": "GET",
	}

	result, err := endpoint.Call(params, nil)
	// Check for errors
	if err != nil {
		fmt.Printf("Call failed: %v\n", err)
	}

	// Also check for response-level errors
	if result != nil {
		if errorMsg, ok := result["error"].(string); ok {
			fmt.Printf("Response error: %s\n", errorMsg)
		}

		// Check status code
		if statusCode, ok := result["statusCode"].(int); ok {
			if statusCode >= 400 {
				fmt.Printf("HTTP error: %d\n", statusCode)
			}
		}
	}
}

// Example showing how to work with the response body
func ExampleEndpoint_Call_responseBody() {
	endpoint := api.Endpoint{
		FriendlyName: "JSON API",
		URL:          "https://jsonplaceholder.typicode.com/users/1",
		Type:         api.EndpointTypeREST,
		Timeout:      10 * time.Second,
	}

	params := map[string]any{
		"method": "GET",
	}

	result, err := endpoint.Call(params, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// The body is automatically parsed as JSON if possible
	if body, ok := result["body"].(map[string]any); ok {
		// Access JSON fields
		if name, ok := body["name"].(string); ok {
			fmt.Printf("User name: %s\n", name)
		}
	} else if bodyStr, ok := result["body"].(string); ok {
		// Non-JSON response returned as string
		fmt.Printf("Response: %s\n", bodyStr)
	}
}
