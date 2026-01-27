package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/trimble-oss/tierceron-core/v2/api"
)

// ExampleNewAPICaller demonstrates basic usage of the API caller
func ExampleNewAPICaller() {
	// Define a REST endpoint
	endpoint := api.Endpoint{
		FriendlyName: "Example API",
		URL:          "https://api.example.com/v1/users",
		Type:         api.EndpointTypeREST,
		Timeout:      30 * time.Second,
	}

	// Create configuration
	config := &api.APICallerConfig{
		InsecureSkipVerify: false,
	}

	// Create API caller (uses cached instance)
	caller, err := api.NewAPICaller(endpoint, config)
	if err != nil {
		panic(err)
	}
	// Note: No need to close - callers are cached globally

	// Make a GET request
	response, err := caller.Call(&api.CallOptions{
		Context: context.Background(),
		Method:  "GET",
		Headers: map[string]string{
			"Accept": "application/json",
		},
	})
	if err != nil {
		panic(err)
	}

	_ = response // Use response
}

// ExampleNewAPICaller_rest demonstrates REST API usage
func ExampleNewAPICaller_rest() {
	endpoint := api.Endpoint{
		FriendlyName: "User API",
		URL:          "https://api.example.com/users",
		Type:         api.EndpointTypeREST,
		Timeout:      15 * time.Second,
	}

	caller, _ := api.NewAPICaller(endpoint, &api.APICallerConfig{})
	// Note: No need to close - callers are cached globally

	// POST request with JSON body
	requestBody := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := caller.Call(&api.CallOptions{
		Context: ctx,
		Method:  "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
		Body: requestBody,
	})
	if err != nil {
		panic(err)
	}

	_ = response // Use response
}

// ExampleNewAPICaller_grpc demonstrates gRPC usage
func ExampleNewAPICaller_grpc() {
	endpoint := api.Endpoint{
		FriendlyName: "User Service",
		URL:          "localhost:50051",
		Type:         api.EndpointTypeGRPC,
		Timeout:      -1, // No timeout for long-lived gRPC streams
	}

	config := &api.APICallerConfig{
		InsecureSkipVerify: true, // For development only
	}

	_, _ = api.NewAPICaller(endpoint, config)
	// Note: No need to close - callers are cached globally

	// Access the gRPC connection for use with generated clients
	// Note: In production, you would cast to *api.GRPCClient and use GetConnection()
	// conn := caller.(*api.GRPCClient).GetConnection()
	// client := yourpb.NewYourServiceClient(conn)
}

// ExampleNewAPICaller_soap demonstrates SOAP API usage
func ExampleNewAPICaller_soap() {
	endpoint := api.Endpoint{
		FriendlyName: "Weather Service",
		URL:          "http://www.example.com/soap/weatherservice",
		Type:         api.EndpointTypeSOAP,
		Timeout:      20 * time.Second,
		WSDLUrl:      "http://www.example.com/soap/weatherservice?wsdl",
	}

	caller, _ := api.NewAPICaller(endpoint, &api.APICallerConfig{})
	// Note: No need to close - callers are cached globally

	// Simple parameter map - SOAP envelope generated automatically
	params := map[string]interface{}{
		"City":    "New York",
		"Country": "USA",
	}

	response, err := caller.Call(&api.CallOptions{
		Context: context.Background(),
		Method:  "GetWeather",
		Body:    params, // SOAP client auto-generates envelope
	})
	if err != nil {
		panic(err)
	}

	_ = response // Use response
}

// ExampleNewAPICaller_tls demonstrates TLS configuration
func ExampleNewAPICaller_tls() {
	endpoint := api.Endpoint{
		FriendlyName: "Secure API",
		URL:          "https://secure-api.example.com/v1/data",
		Type:         api.EndpointTypeREST,
		Timeout:      25 * time.Second,
	}

	config := &api.APICallerConfig{
		TLSCertPath: "/path/to/client-cert.pem",
		TLSKeyPath:  "/path/to/client-key.pem",
		CACertPath:  "/path/to/ca-cert.pem",
	}

	caller, _ := api.NewAPICaller(endpoint, config)
	// Note: No need to close - callers are cached globally

	response, err := caller.Call(&api.CallOptions{
		Context: context.Background(),
		Method:  "GET",
	})
	if err != nil {
		panic(err)
	}

	_ = response // Use response
}

// TestEndpointTypes verifies endpoint type constants
func TestEndpointTypes(t *testing.T) {
	tests := []struct {
		name         string
		endpointType api.EndpointType
		expected     string
	}{
		{"REST", api.EndpointTypeREST, "rest"},
		{"gRPC", api.EndpointTypeGRPC, "grpc"},
		{"SOAP", api.EndpointTypeSOAP, "soap"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.endpointType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.endpointType)
			}
		})
	}
}

// TestEndpointValidation tests endpoint validation
func TestEndpointValidation(t *testing.T) {
	tests := []struct {
		name      string
		endpoint  api.Endpoint
		shouldErr bool
	}{
		{
			name: "Valid endpoint",
			endpoint: api.Endpoint{
				FriendlyName: "Test API",
				URL:          "https://api.example.com",
				Type:         api.EndpointTypeREST,
				Timeout:      30 * time.Second,
			},
			shouldErr: false,
		},
		{
			name: "Missing friendly name",
			endpoint: api.Endpoint{
				URL:  "https://api.example.com",
				Type: api.EndpointTypeREST,
			},
			shouldErr: true,
		},
		{
			name: "Missing URL",
			endpoint: api.Endpoint{
				FriendlyName: "Test API",
				Type:         api.EndpointTypeREST,
			},
			shouldErr: true,
		},
		{
			name: "Invalid endpoint type",
			endpoint: api.Endpoint{
				FriendlyName: "Test API",
				URL:          "https://api.example.com",
				Type:         "invalid",
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := api.NewAPICaller(tt.endpoint, &api.APICallerConfig{})
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
