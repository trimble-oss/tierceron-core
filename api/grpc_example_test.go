package api_test

import (
	"fmt"
	"log"
	"time"

	"github.com/trimble-oss/tierceron-core/v2/api"
)

// Example demonstrates calling a gRPC endpoint using the generic map-based interface
// This example assumes you have a gRPC server with reflection enabled
func Example_grpcEndpoint() {
	// Define a gRPC endpoint
	endpoint := api.Endpoint{
		FriendlyName: "User Service",
		URL:          "localhost:50051",
		Type:         api.EndpointTypeGRPC,
		MethodName:   "/user.UserService/GetUser", // Full method path
		Timeout:      15 * time.Second,
	}

	// Configure the API caller with insecure connection (for testing)
	config := &api.APICallerConfig{
		InsecureSkipVerify: true, // Use TLS certificates in production
	}

	// Get or create API caller
	caller, err := api.NewAPICaller(endpoint, config)
	if err != nil {
		log.Fatalf("Failed to create API caller: %v", err)
	}

	// Call the gRPC endpoint with map parameters
	// The map will be automatically converted to the appropriate protobuf message
	params := map[string]any{
		"body": map[string]any{
			"user_id": "12345",
		},
	}

	response, err := endpoint.Call(params, config)
	if err != nil {
		log.Fatalf("gRPC call failed: %v", err)
	}

	// Access the response
	statusCode := response["statusCode"].(int)
	body := response["body"].(map[string]any)

	fmt.Printf("Status: %d\n", statusCode)
	fmt.Printf("User Name: %s\n", body["name"])
	fmt.Printf("User Email: %s\n", body["email"])

	// Close the caller when done
	caller.Close()

	// Output format would be:
	// Status: 200
	// User Name: John Doe
	// User Email: john@example.com
}

// Example demonstrates calling a gRPC streaming method
func Example_grpcStreamingEndpoint() {
	endpoint := api.Endpoint{
		FriendlyName: "Chat Service",
		URL:          "localhost:50052",
		Type:         api.EndpointTypeGRPC,
		MethodName:   "/chat.ChatService/SendMessage",
		Timeout:      30 * time.Second,
	}

	config := &api.APICallerConfig{
		InsecureSkipVerify: true,
	}

	// Send a message via gRPC
	params := map[string]any{
		"body": map[string]any{
			"message": "Hello from gRPC!",
			"user_id": "user123",
			"room_id": "room456",
		},
	}

	response, err := endpoint.Call(params, config)
	if err != nil {
		log.Fatalf("gRPC call failed: %v", err)
	}

	// Process response
	if errorMsg, ok := response["error"].(string); ok && errorMsg != "" {
		fmt.Printf("Error: %s\n", errorMsg)
		return
	}

	body := response["body"].(map[string]any)
	fmt.Printf("Message sent successfully: %v\n", body["success"])
	fmt.Printf("Message ID: %s\n", body["message_id"])
}

// Example demonstrates using TLS with gRPC
func Example_grpcWithTLS() {
	endpoint := api.Endpoint{
		FriendlyName: "Secure Service",
		URL:          "secure.example.com:443",
		Type:         api.EndpointTypeGRPC,
		MethodName:   "/api.SecureService/GetData",
		Timeout:      20 * time.Second,
	}

	// Configure with TLS certificates
	config := &api.APICallerConfig{
		CACertPath:  "/path/to/ca.crt",
		TLSCertPath: "/path/to/client.crt",
		TLSKeyPath:  "/path/to/client.key",
	}

	params := map[string]any{
		"body": map[string]any{
			"query": "SELECT * FROM users",
			"limit": 100,
		},
	}

	response, err := endpoint.Call(params, config)
	if err != nil {
		log.Fatalf("gRPC call failed: %v", err)
	}

	body := response["body"].(map[string]any)
	results := body["results"].([]any)
	fmt.Printf("Retrieved %d results\n", len(results))
}

// Example demonstrates timeout handling with gRPC
func Example_grpcWithTimeout() {
	// Short timeout for demonstration
	endpoint := api.Endpoint{
		FriendlyName: "Slow Service",
		URL:          "localhost:50053",
		Type:         api.EndpointTypeGRPC,
		MethodName:   "/slow.SlowService/ProcessData",
		Timeout:      2 * time.Second, // 2 second timeout
	}

	config := &api.APICallerConfig{
		InsecureSkipVerify: true,
	}

	params := map[string]any{
		"body": map[string]any{
			"data": "large dataset",
		},
	}

	response, err := endpoint.Call(params, config)
	if err != nil {
		// Handle timeout error
		fmt.Printf("Call timed out: %v\n", err)
		return
	}

	fmt.Printf("Response: %v\n", response["body"])
}

// Example demonstrates complex nested data structures with gRPC
func Example_grpcComplexData() {
	endpoint := api.Endpoint{
		FriendlyName: "Order Service",
		URL:          "localhost:50054",
		Type:         api.EndpointTypeGRPC,
		MethodName:   "/order.OrderService/CreateOrder",
		Timeout:      15 * time.Second,
	}

	config := &api.APICallerConfig{
		InsecureSkipVerify: true,
	}

	// Complex nested structure - automatically converted to protobuf
	params := map[string]any{
		"body": map[string]any{
			"customer_id": "cust_12345",
			"items": []any{
				map[string]any{
					"product_id": "prod_001",
					"quantity":   2,
					"price":      29.99,
				},
				map[string]any{
					"product_id": "prod_002",
					"quantity":   1,
					"price":      49.99,
				},
			},
			"shipping_address": map[string]any{
				"street":  "123 Main St",
				"city":    "Springfield",
				"state":   "IL",
				"zip":     "62701",
				"country": "USA",
			},
			"payment_method": map[string]any{
				"type":       "credit_card",
				"last_four":  "1234",
				"expiration": "12/25",
			},
		},
	}

	response, err := endpoint.Call(params, config)
	if err != nil {
		log.Fatalf("Failed to create order: %v", err)
	}

	body := response["body"].(map[string]any)
	fmt.Printf("Order created: %s\n", body["order_id"])
	fmt.Printf("Total amount: $%.2f\n", body["total_amount"])
	fmt.Printf("Status: %s\n", body["status"])
}
