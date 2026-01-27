package api_test

import (
	"fmt"
	"time"

	"github.com/trimble-oss/tierceron-core/v2/api"
)

// ExampleEndpoint_timeout_default demonstrates using the default 30s timeout
func ExampleEndpoint_timeout_default() {
	endpoint := api.Endpoint{
		FriendlyName: "Default Timeout API",
		URL:          "https://httpbin.org/delay/1",
		Type:         api.EndpointTypeREST,
		// Timeout not set = uses default 30 seconds
	}

	result, err := endpoint.Call(map[string]any{
		"method": "GET",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Success with default timeout! Status: %d\n", result["statusCode"])
}

// ExampleEndpoint_timeout_custom demonstrates using a custom timeout
func ExampleEndpoint_timeout_custom() {
	endpoint := api.Endpoint{
		FriendlyName: "Fast API",
		URL:          "https://httpbin.org/delay/1",
		Type:         api.EndpointTypeREST,
		Timeout:      5 * time.Second, // Custom 5 second timeout
	}

	result, err := endpoint.Call(map[string]any{
		"method": "GET",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Success with 5s timeout! Status: %d\n", result["statusCode"])
}

// ExampleEndpoint_timeout_none demonstrates using no timeout
func ExampleEndpoint_timeout_none() {
	endpoint := api.Endpoint{
		FriendlyName: "Long Running API",
		URL:          "https://httpbin.org/delay/2",
		Type:         api.EndpointTypeREST,
		Timeout:      -1, // -1 = no timeout, wait indefinitely
	}

	result, err := endpoint.Call(map[string]any{
		"method": "GET",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Success with no timeout! Status: %d\n", result["statusCode"])
}

// ExampleEndpoint_timeout_short demonstrates what happens with a timeout that's too short
func ExampleEndpoint_timeout_short() {
	endpoint := api.Endpoint{
		FriendlyName: "Will Timeout API",
		URL:          "https://httpbin.org/delay/5", // Takes 5 seconds
		Type:         api.EndpointTypeREST,
		Timeout:      1 * time.Second, // Only wait 1 second
	}

	result, err := endpoint.Call(map[string]any{
		"method": "GET",
	})
	if err != nil {
		fmt.Printf("Expected timeout error: %v\n", err)
		return
	}

	// This won't execute because the call will timeout
	fmt.Printf("Status: %d\n", result["statusCode"])
}
