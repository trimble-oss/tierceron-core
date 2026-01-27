# API Caller Package

A generic API caller package for making calls to REST, gRPC, and SOAP endpoints in Go.

## Features

- **Multi-Protocol Support**: Support for REST, gRPC, and SOAP API endpoints
- **TLS/SSL Support**: Configurable TLS with support for custom certificates
- **Flexible Configuration**: Easy endpoint configuration with friendly names
- **Caller Caching**: Global caching of API callers for reuse and efficiency
- **Automatic Context Management**: Built-in context creation with configurable timeouts
- **Configurable Timeouts**: Per-endpoint timeout configuration (default 30s, or set to -1 for no timeout)
- **Automatic Retries**: Exponential backoff retry logic for timeout errors
- **Type-Safe**: Strongly typed endpoint types and responses

## Installation

```go
import "github.com/trimble-oss/tierceron-core/v2/api"
```

## Usage

### Quick Start - Generic Endpoint Call

The simplest way to make API calls is using the `Endpoint.Call()` method which accepts generic parameters:

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/trimble-oss/tierceron-core/v2/api"
)

func main() {
    // Define an endpoint
    endpoint := api.Endpoint{
        FriendlyName: "My REST API",
        URL:          "https://api.example.com/v1/users",
        Type:         api.EndpointTypeREST,
    }
    
    // Make a call with generic parameters
    // Context is managed internally with a 30s timeout
    params := map[string]interface{}{
        "method": "GET",
        "headers": map[string]string{
            "Accept": "application/json",
        },
    }
    
    result, err := endpoint.Call(params, nil)
    if err != nil {
        log.Fatalf("API call failed: %v", err)
    }
    
    // Access response as a map
    fmt.Printf("Status: %d\n", result["statusCode"])
    fmt.Printf("Body: %v\n", result["body"])
}
```

### Advanced - Using APICaller Directly

For more control, you can use the `APICaller` directly:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/trimble-oss/tierceron-core/v2/api"
)

func main() {
    // Define an endpoint
    endpoint := api.Endpoint{
        FriendlyName: "My REST API",
        URL:          "https://api.example.com/v1/resource",
        Type:         api.EndpointTypeREST,
    }
    
    // Create configuration
    config := &api.APICallerConfig{
        InsecureSkipVerify: false, // Set to true only for testing
    }
    
    // Create API caller (cached globally - no need to close)
    caller, err := api.NewAPICaller(endpoint, config)
    if err != nil {
        log.Fatalf("Failed to create API caller: %v", err)
    }
    
    // Make a call - context is managed internally
    response, err := caller.Call(&api.CallOptions{
        Method:  "GET",
        Headers: map[string]string{
            "Accept": "application/json",
        },
    })
    
    if err != nil {
        log.Fatalf("API call failed: %v", err)
    }
    
    fmt.Printf("Status: %d\n", response.StatusCode)
    fmt.Printf("Body: %s\n", string(response.Body))
}
```

### REST API Example

#### Using Generic Call Method

```go
// Create a REST endpoint
endpoint := api.Endpoint{
    FriendlyName: "User API",
    URL:          "https://api.example.com/users",
    Type:         api.EndpointTypeREST,
}

// POST request with JSON body using generic parameters
params := map[string]interface{}{
    "method": "POST",
    "headers": map[string]string{
        "Content-Type":  "application/json",
        "Authorization": "Bearer token123",
    },
    "body": map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
    },
}

config := &api.APICallerConfig{
    TLSCertPath: "/path/to/cert.pem",
    TLSKeyPath:  "/path/to/key.pem",
    CACertPath:  "/path/to/ca.pem",
}

result, err := endpoint.Call(params, config)
if err != nil {
    log.Fatal(err)
}

// Access response data
statusCode := result["statusCode"].(int)
body := result["body"] // Automatically parsed as JSON
```

#### Using APICaller Directly

```go
endpoint := api.Endpoint{
    FriendlyName: "User API",
    URL:          "https://api.example.com/users",
    Type:         api.EndpointTypeREST,
}

config := &api.APICallerConfig{
    TLSCertPath: "/path/to/cert.pem",
    TLSKeyPath:  "/path/to/key.pem",
    CACertPath:  "/path/to/ca.pem",
}

caller, _ := api.NewAPICaller(endpoint, config)
defer caller.Close()

// POST request with JSON body
requestBody := map[string]interface{}{
    "name":  "John Doe",
    "email": "john@example.com",
}

response, err := caller.Call(&api.CallOptions{
    Context: context.Background(),
    Method:  "POST",
    Headers: map[string]string{
        "Content-Type": "application/json",
        "Authorization": "Bearer token123",
    },
    Body: requestBody,
})
```

### gRPC Example

```go
// Create a gRPC endpoint
endpoint := api.Endpoint{
    FriendlyName: "User Service",
    URL:          "localhost:50051",
    Type:         api.EndpointTypeGRPC,
}

config := &api.APICallerConfig{
    TLSCertPath: "/path/to/cert.pem",
    TLSKeyPath:  "/path/to/key.pem",
    CACertPath:  "/path/to/ca.pem",
}

caller, _ := api.NewAPICaller(endpoint, config)
defer caller.Close()

// For gRPC, you typically need to access the underlying connection
// and use service-specific generated clients
if grpcClient, ok := caller.(*api.GRPCClient); ok {
    conn := grpcClient.GetConnection()
    // Use conn with your generated protobuf client
    // client := yourpb.NewYourServiceClient(conn)
}
```

### SOAP Example

SOAP endpoints automatically generate SOAP envelopes from simple parameter maps. No need to manually create XML!

#### Simple SOAP Call with Auto-Generated Envelope

```go
// Create a SOAP endpoint with WSDL URL
endpoint := api.Endpoint{
    FriendlyName: "Temperature Converter",
    URL:          "https://www.w3schools.com/xml/tempconvert.asmx",
    Type:         api.EndpointTypeSOAP,
    Timeout:      15 * time.Second,
    WSDLUrl:      "https://www.w3schools.com/xml/tempconvert.asmx?WSDL",
}

// Just provide the parameters - SOAP envelope is generated automatically!
params := map[string]interface{}{
    "method": "CelsiusToFahrenheit",
    "body": map[string]interface{}{
        "Celsius": "100",
    },
}

result, err := endpoint.Call(params, nil)
if err != nil {
    log.Fatal(err)
}

// Response body is automatically parsed into key-value pairs!
if body, ok := result["body"].(map[string]interface{}); ok {
    fmt.Printf("Temperature: %s\n", body["CelsiusToFahrenheitResult"])
    // Access other fields by name
    for key, value := range body {
        fmt.Printf("%s: %v\n", key, value)
    }
}
fmt.Printf("WSDL: %s\n", endpoint.WSDLUrl)
```

#### Another SOAP Example

```go
endpoint := api.Endpoint{
    FriendlyName: "Weather Service",
    URL:          "http://www.example.com/soap/service",
    Type:         api.EndpointTypeSOAP,
    Timeout:      20 * time.Second,
    WSDLUrl:      "http://www.example.com/soap/service?wsdl",
}

// The SOAP client automatically creates the envelope
params := map[string]interface{}{
    "method": "GetWeather",
    "body": map[string]interface{}{
        "City":    "New York",
        "Country": "USA",
    },
}

result, err := endpoint.Call(params, nil)
if err != nil {
    log.Fatal(err)
}

// Response is parsed into key-value pairs
if body, ok := result["body"].(map[string]interface{}); ok {
    fmt.Printf("City: %s\n", body["City"])
    fmt.Printf("Temperature: %s\n", body["Temperature"])
    fmt.Printf("Condition: %s\n", body["Condition"])
}
statusCode := result["statusCode"].(int)
fmt.Printf("Status: %d\n", statusCode)
```

The SOAP client automatically:
- Creates the SOAP envelope structure
- Extracts namespace from WSDL URL
- Marshals parameters into XML elements
- Adds proper SOAP headers
- **Parses the SOAP response into key-value pairs**
- No manual XML creation or parsing needed!

#### Using APICaller Directly

```go
endpoint := api.Endpoint{
    FriendlyName: "Weather Service",
    URL:          "http://www.example.com/soap/service",
    Type:         api.EndpointTypeSOAP,
}

config := &api.APICallerConfig{
    InsecureSkipVerify: false,
}

caller, _ := api.NewAPICaller(endpoint, config)
defer caller.Close()

// SOAP request with custom envelope
type WeatherRequest struct {
    XMLName xml.Name `xml:"GetWeather"`
    City    string   `xml:"city"`
}

soapEnvelope := api.CreateSOAPEnvelope(
    WeatherRequest{City: "New York"},
    nil, // No custom header
)

response, err := caller.Call(&api.CallOptions{
    Context: context.Background(),
    Method:  "GetWeather", // SOAPAction
    Body:    soapEnvelope,
})
```

## Configuration Options

### Generic Call Parameters (for Endpoint.Call)

When using `Endpoint.Call()`, pass parameters as a `map[string]interface{}`:

**Common parameters (all endpoint types):**
- `"method"` (string): HTTP method for REST (GET, POST, etc.), operation name for SOAP/gRPC
- `"body"` (interface{}): Request body - can be map, struct, []byte, or string
- `"headers"` (map[string]string): HTTP headers (REST/SOAP)

**SOAP-specific parameters:**
- `"soapAction"` (string): SOAP action header value (automatically wrapped in quotes)

**Example:**
```go
params := map[string]interface{}{
    "method": "POST",
    "headers": map[string]string{
        "Content-Type": "application/json",
        "Authorization": "Bearer token",
    },
    "body": map[string]interface{}{
        "key": "value",
    },
}

result, err := endpoint.Call(ctx, params, config)
```

**Response map keys:**
- `"statusCode"` (int): HTTP status code (REST/SOAP only)
- `"body"` (interface{}): Response body, automatically parsed as JSON if possible, otherwise string
- `"headers"` (map[string][]string): Response headers (REST/SOAP only)
- `"error"` (string): Error message if the call failed

### APICallerConfig

- `InsecureSkipVerify` (bool): Skip TLS certificate verification (use with caution, only for testing)
- `TLSCertPath` (string): Path to client TLS certificate file
- `TLSKeyPath` (string): Path to client TLS key file
- `CACertPath` (string): Path to CA certificate file for server verification

### CallOptions

- `Context` (context.Context): Request context for timeout/cancellation
- `Method` (string): HTTP method for REST (GET, POST, etc.), operation name for SOAP, method path for gRPC
- `Headers` (map[string]string): Request headers (REST and SOAP)
- `Body` (interface{}): Request body - can be []byte, string, io.Reader, or any struct (will be marshaled to JSON/XML)
- `Timeout` (interface{}): Optional timeout duration

## Endpoint Types

- `api.EndpointTypeREST`: For RESTful HTTP APIs
- `api.EndpointTypeGRPC`: For gRPC services
- `api.EndpointTypeSOAP`: For SOAP web services

## Response Structure

```go
type Response struct {
    StatusCode int                 // HTTP status code (REST/SOAP)
    Body       []byte              // Response body
    Headers    map[string][]string // Response headers (REST/SOAP)
    Error      error               // Error if the call failed
}
```

## Error Handling

The package returns errors in the following cases:
- Invalid endpoint configuration
- Network/connection failures
- HTTP 4xx/5xx errors (REST/SOAP)
- SOAP faults
- TLS/certificate errors

Always check both the returned error and the `Response.Error` field:

```go
response, err := caller.Call(options)
if err != nil {
    // Handle error
}
if response.Error != nil {
    // Handle response-specific error (e.g., HTTP 4xx/5xx)
}
```

## Thread Safety

Each `APICaller` instance maintains its own client connection and is safe for concurrent use. However, for best performance with high concurrency, consider creating a pool of callers or using connection pooling at the HTTP transport level.

## Best Practices

1. **No need to close callers**: Callers are cached globally - use `ClearCallerCache()` only when needed
2. **Configure timeouts per endpoint**: Set `Timeout` on `Endpoint` struct for per-call control
3. **TLS verification**: Only set `InsecureSkipVerify: true` for testing/development
4. **Cache management**: Use `RemoveCallerFromCache()` to remove specific cached callers when needed

## Timeout Configuration

You can configure timeouts on a per-endpoint basis using the `Timeout` field:

```go
// Default timeout (30 seconds)
endpoint := api.Endpoint{
    FriendlyName: "Default Timeout API",
    URL:          "https://api.example.com",
    Type:         api.EndpointTypeREST,
    // Timeout is 0 or unset = uses 30 second default
}

// Custom timeout (5 seconds)
endpoint := api.Endpoint{
    FriendlyName: "Fast API",
    URL:          "https://api.example.com/fast",
    Type:         api.EndpointTypeREST,
    Timeout:      5 * time.Second,
}

// No timeout (wait indefinitely)
endpoint := api.Endpoint{
    FriendlyName: "Long Running API",
    URL:          "https://api.example.com/batch",
    Type:         api.EndpointTypeREST,
    Timeout:      -1, // -1 means no timeout
}
```

## Retry on Timeout

The package includes automatic retry logic for timeout errors with exponential backoff:

```go
// Configure retries for timeout scenarios
endpoint := api.Endpoint{
    FriendlyName: "Unreliable API",
    URL:          "https://api.example.com",
    Type:         api.EndpointTypeREST,
    Timeout:      2 * time.Second, // Request timeout
    MaxRetries:   3,                // Retry up to 3 times on timeout
}

// When a request times out:
// - Attempt 1: Initial request (fails with timeout)
// - Wait 1 second (exponential backoff: 2^0 = 1s)
// - Attempt 2: Retry (fails with timeout)
// - Wait 2 seconds (exponential backoff: 2^1 = 2s)
// - Attempt 3: Retry (fails with timeout)  
// - Wait 4 seconds (exponential backoff: 2^2 = 4s)
// - Attempt 4: Final retry (succeeds or fails)

result, err := endpoint.Call(params, config)
```

**Key Features:**
- Only retries on `context.DeadlineExceeded` errors (timeout errors)
- Non-timeout errors (4xx, 5xx, network errors) are not retried
- Exponential backoff: 1s, 2s, 4s, 8s between retries
- `MaxRetries: 0` (default) means no retries
- Each retry recreates the context with the configured timeout
- Total time = (MaxRetries + 1) × Timeout + sum of backoff delays

```go
// No retries (default behavior)
endpoint := api.Endpoint{
    FriendlyName: "Quick Fail API",
    URL:          "https://api.example.com",
    Type:         api.EndpointTypeREST,
    Timeout:      5 * time.Second,
    MaxRetries:   0, // or omit - no retries by default
}

// Aggressive retries for flaky endpoints
endpoint := api.Endpoint{
    FriendlyName: "Flaky API",
    URL:          "https://api.example.com",
    Type:         api.EndpointTypeREST,
    Timeout:      3 * time.Second,
    MaxRetries:   5, // Will retry up to 5 times on timeout
}
```

## Caller Caching

All API callers are automatically cached globally based on endpoint configuration:

```go
// Create first caller - added to cache
caller1, _ := api.NewAPICaller(endpoint, config)

// Create second caller with same endpoint/config - returns cached instance
caller2, _ := api.NewAPICaller(endpoint, config)
// caller1 == caller2 (same instance)

// Clear entire cache
api.ClearCallerCache()

// Remove specific caller from cache
api.RemoveCallerFromCache(endpoint, config)
```
4. **Error handling**: Check both the error return value and `Response.Error`
5. **Connection reuse**: Reuse `APICaller` instances when possible for connection pooling

### gRPC Example

gRPC endpoints use reflection to dynamically discover services and methods, allowing you to make calls using simple map-based parameters without generated code!

#### Prerequisites

Your gRPC server must have **reflection enabled** for dynamic discovery to work:

```go
// In your gRPC server code:
import "google.golang.org/grpc/reflection"

func main() {
    s := grpc.NewServer()
    // Register your services...
    
    // Enable reflection
    reflection.Register(s)
    
    s.Serve(lis)
}
```

#### Simple gRPC Call with Dynamic Messages

```go
// Create a gRPC endpoint
endpoint := api.Endpoint{
    FriendlyName: "User Service",
    URL:          "localhost:50051",
    Type:         api.EndpointTypeGRPC,
    MethodName:   "/user.UserService/GetUser", // Full method path: /package.Service/Method
    Timeout:      15 * time.Second,
}

config := &api.APICallerConfig{
    InsecureSkipVerify: true, // Use proper TLS in production
}

// Just provide the parameters - they're automatically converted to protobuf!
params := map[string]interface{}{
    "body": map[string]interface{}{
        "user_id": "12345",
    },
}

result, err := endpoint.Call(params, config)
if err != nil {
    log.Fatal(err)
}

// Response is automatically parsed into a map!
statusCode := result["statusCode"].(int)
body := result["body"].(map[string]interface{})

fmt.Printf("Status: %d\n", statusCode)
fmt.Printf("User Name: %s\n", body["name"])
fmt.Printf("User Email: %s\n", body["email"])
```

#### gRPC with Complex Nested Data

The package automatically converts complex nested structures to protobuf messages:

```go
endpoint := api.Endpoint{
    FriendlyName: "Order Service",
    URL:          "localhost:50054",
    Type:         api.EndpointTypeGRPC,
    MethodName:   "/order.OrderService/CreateOrder",
    Timeout:      15 * time.Second,
}

// Complex nested structure - automatically converted to protobuf
params := map[string]interface{}{
    "body": map[string]interface{}{
        "customer_id": "cust_12345",
        "items": []interface{}{
            map[string]interface{}{
                "product_id": "prod_001",
                "quantity":   2,
                "price":      29.99,
            },
            map[string]interface{}{
                "product_id": "prod_002",
                "quantity":   1,
                "price":      49.99,
            },
        },
        "shipping_address": map[string]interface{}{
            "street":  "123 Main St",
            "city":    "Springfield",
            "state":   "IL",
            "zip":     "62701",
            "country": "USA",
        },
        "payment_method": map[string]interface{}{
            "type":       "credit_card",
            "last_four":  "1234",
            "expiration": "12/25",
        },
    },
}

result, err := endpoint.Call(params, nil)
if err != nil {
    log.Fatal(err)
}

body := result["body"].(map[string]interface{})
fmt.Printf("Order created: %s\n", body["order_id"])
fmt.Printf("Total amount: $%.2f\n", body["total_amount"])
```

#### gRPC with TLS

```go
endpoint := api.Endpoint{
    FriendlyName: "Secure Service",
    URL:          "secure.example.com:443",
    Type:         api.EndpointTypeGRPC,
    MethodName:   "/api.SecureService/GetData",
    Timeout:      20 * time.Second,
}

config := &api.APICallerConfig{
    CACertPath:  "/path/to/ca.crt",
    TLSCertPath: "/path/to/client.crt",
    TLSKeyPath:  "/path/to/client.key",
}

params := map[string]interface{}{
    "body": map[string]interface{}{
        "query": "SELECT * FROM users",
        "limit": 100,
    },
}

result, err := endpoint.Call(params, config)
// Process result...
```

#### How gRPC Reflection Works

1. **Service Discovery**: Uses gRPC reflection to discover service and method definitions from your server
2. **Dynamic Messages**: Converts request parameters from `map[string]interface{}` to dynamic protobuf messages
3. **Automatic Invocation**: Invokes the gRPC method with the dynamic message
4. **Response Parsing**: Automatically converts the protobuf response back to `map[string]interface{}`

**Note:** The `MethodName` field must contain the full method path in the format `/package.ServiceName/MethodName`. You can find this in your `.proto` file or by using tools like `grpcurl`.

The gRPC client automatically:
- Discovers service methods using gRPC reflection
- Converts map parameters to dynamic protobuf messages
- Invokes methods without generated code
- Parses protobuf responses into key-value pairs
- Handles complex nested data structures

