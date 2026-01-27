package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/trimble-oss/tierceron-core/v2/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

func main() {
	// Start a local gRPC server for testing
	fmt.Println("=== Starting gRPC Test Server ===")
	grpcServer, err := StartGRPCServer("50051")
	if err != nil {
		log.Printf("Failed to start gRPC server: %v", err)
	} else {
		fmt.Println("gRPC server started on :50051")
		defer grpcServer.GracefulStop()
		// Give server time to start
		time.Sleep(500 * time.Millisecond)
	}

	// Example 1: Simple REST GET request
	fmt.Println("\n=== Example 1: REST GET ===")
	restEndpoint := api.Endpoint{
		FriendlyName: "GitHub API",
		URL:          "https://api.github.com/zen",
		Type:         api.EndpointTypeREST,
		Timeout:      10 * time.Second,
	}

	params := map[string]interface{}{
		"method": "GET",
		"headers": map[string]string{
			"Accept":     "application/json",
			"User-Agent": "API-Caller-Demo/1.0",
		},
	}

	result, err := restEndpoint.Call(params, nil)
	if err != nil {
		log.Printf("REST GET Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", result["statusCode"])
		fmt.Printf("Body: %v\n", result["body"])
	}

	// Example 2: REST POST request
	fmt.Println("\n=== Example 2: REST POST ===")
	postEndpoint := api.Endpoint{
		FriendlyName: "HTTPBin POST",
		URL:          "https://httpbin.org/post",
		Type:         api.EndpointTypeREST,
		Timeout:      15 * time.Second,
	}

	postParams := map[string]interface{}{
		"method": "POST",
		"headers": map[string]string{
			"Content-Type": "application/json",
		},
		"body": map[string]interface{}{
			"name":    "John Doe",
			"email":   "john@example.com",
			"message": "Hello from API caller!",
		},
	}

	postResult, err := postEndpoint.Call(postParams, nil)
	if err != nil {
		log.Printf("REST POST Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", postResult["statusCode"])
		if body, ok := postResult["body"].(map[string]interface{}); ok {
			if json, ok := body["json"]; ok {
				fmt.Printf("Echoed JSON: %v\n", json)
			}
		}
	}

	// Example 3: Comparison with traditional APICaller approach
	fmt.Println("\n=== Example 3: Traditional APICaller ===")
	endpoint := api.Endpoint{
		FriendlyName: "HTTPBin GET",
		URL:          "https://httpbin.org/get",
		Type:         api.EndpointTypeREST,
		Timeout:      0, // Uses default 30s
	}

	// Traditional approach with APICaller (now uses cached callers)
	caller, err := api.NewAPICaller(endpoint, &api.APICallerConfig{})
	if err != nil {
		log.Fatalf("Failed to create caller: %v", err)
	}
	// Note: No need to defer close - callers are cached globally
	// Context is also managed internally - no need to pass it

	response, err := caller.Call(&api.CallOptions{
		Method: "GET",
		Headers: map[string]string{
			"Accept": "application/json",
		},
	})

	if err != nil {
		log.Printf("APICaller Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		if bodyBytes, ok := response.Body.([]byte); ok {
			fmt.Printf("Body length: %d bytes\n", len(bodyBytes))
		} else {
			fmt.Printf("Body: %v\n", response.Body)
		}
	}

	// Example 4: SOAP API call demonstration
	fmt.Println("\n=== Example 4: SOAP API ===")

	// Try multiple SOAP endpoints until we find one that works
	soapEndpoints := []struct {
		name   string
		url    string
		wsdl   string
		method string
		action string
		params map[string]interface{}
	}{
		{
			name:   "W3Schools Temperature Converter",
			url:    "https://www.w3schools.com/xml/tempconvert.asmx",
			wsdl:   "https://www.w3schools.com/xml/tempconvert.asmx?WSDL",
			method: "CelsiusToFahrenheit",
			action: "https://www.w3schools.com/xml/CelsiusToFahrenheit",
			params: map[string]interface{}{"Celsius": "100"},
		},
		{
			name:   "DNEOnline Calculator",
			url:    "http://www.dneonline.com/calculator.asmx",
			wsdl:   "http://www.dneonline.com/calculator.asmx?WSDL",
			method: "Add",
			action: "http://tempuri.org/Add",
			params: map[string]interface{}{"intA": "25", "intB": "17"},
		},
		{
			name:   "Thomas Bayer Blz Service",
			url:    "http://www.thomas-bayer.com/axis2/services/BLZService",
			wsdl:   "http://www.thomas-bayer.com/axis2/services/BLZService?wsdl",
			method: "getBank",
			action: "",
			params: map[string]interface{}{"blz": "66050000"},
		},
	}

	soapWorked := false
	for _, soapTest := range soapEndpoints {
		fmt.Printf("Trying %s...\n", soapTest.name)

		soapEndpoint := api.Endpoint{
			FriendlyName: soapTest.name,
			URL:          soapTest.url,
			Type:         api.EndpointTypeSOAP,
			Timeout:      10 * time.Second,
			WSDLUrl:      soapTest.wsdl,
		}

		soapParams := map[string]interface{}{
			"method":     soapTest.method,
			"soapAction": soapTest.action,
			"body":       soapTest.params,
		}

		soapResult, err := soapEndpoint.Call(soapParams, nil)

		if err == nil && soapResult != nil {
			if statusCode, ok := soapResult["statusCode"].(int); ok && statusCode == 200 {
				fmt.Printf("✓ SOAP call succeeded!\n")
				fmt.Printf("Status: %d\n", statusCode)

				if body, ok := soapResult["body"].(map[string]interface{}); ok {
					fmt.Printf("Parsed SOAP Response (key-value pairs):\n")
					for key, value := range body {
						fmt.Printf("  %s: %v\n", key, value)
					}
				} else {
					fmt.Printf("Response body: %v\n", soapResult["body"])
				}
				soapWorked = true
				break
			}
		}
		fmt.Printf("  Failed: %v\n", err)
	}

	if !soapWorked {
		fmt.Println("\nNote: All tested public SOAP endpoints failed (common for legacy services).")
		fmt.Println("SOAP features are fully implemented and work with functioning SOAP services:")
		fmt.Println("  ✓ Automatic envelope generation from maps")
		fmt.Println("  ✓ XML response parsing into key-value pairs")
		fmt.Println("  ✓ No manual XML handling required")
	}

	// Example 4: gRPC API call (using local test server)
	fmt.Println("\n=== Example 4: gRPC API ===")
	grpcEndpoint := api.Endpoint{
		FriendlyName: "User Service gRPC",
		URL:          "localhost:50051",
		Type:         api.EndpointTypeGRPC,
		MethodName:   "/user.UserService/GetUser", // Full method path
		Timeout:      15 * time.Second,
	}

	grpcConfig := &api.APICallerConfig{
		InsecureSkipVerify: true, // Use proper TLS in production
	}

	// Simple parameter map - automatically converted to protobuf
	grpcParams := map[string]interface{}{
		"body": map[string]interface{}{
			"user_id": "12345",
		},
	}

	grpcResult, err := grpcEndpoint.Call(grpcParams, grpcConfig)
	if err != nil {
		log.Printf("gRPC Error: %v\n", err)
		fmt.Println("Note: gRPC test server may not have started properly")
	} else {
		fmt.Printf("Status: %d\n", grpcResult["statusCode"])
		// Body is automatically parsed from protobuf to map!
		if body, ok := grpcResult["body"].(map[string]interface{}); ok {
			fmt.Printf("Parsed gRPC Response:\n")
			for key, value := range body {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
		fmt.Println("✓ gRPC call successful!")
	}

	// Example 5: Retry on Timeout
	fmt.Println("\n=== Example 5: Retry on Timeout ===")
	retryEndpoint := api.Endpoint{
		FriendlyName: "HTTPBin Delay (with Retry)",
		URL:          "https://httpbin.org/delay/2", // This endpoint delays 2 seconds
		Type:         api.EndpointTypeREST,
		Timeout:      1 * time.Second, // Short timeout to potentially trigger retry
		MaxRetries:   2,               // Retry up to 2 times on timeout
	}

	retryParams := map[string]interface{}{
		"method": "GET",
		"headers": map[string]string{
			"Accept": "application/json",
		},
	}

	fmt.Println("Making call to endpoint that delays 2s with 1s timeout...")
	fmt.Println("This may timeout and retry with exponential backoff (1s, 2s delays)")
	retryStart := time.Now()
	retryResult, retryErr := retryEndpoint.Call(retryParams, nil)
	retryDuration := time.Since(retryStart)

	if retryErr != nil {
		fmt.Printf("Request failed after retries (took %v): %v\n", retryDuration, retryErr)
		fmt.Println("Note: Retries only occur on timeout errors, not other HTTP errors")
	} else {
		fmt.Printf("Status: %d (took %v)\n", retryResult["statusCode"], retryDuration)
		fmt.Println("Success! The retry logic helped complete the request.")
	}

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("The new Endpoint.Call() method provides a simpler, generic interface")
	fmt.Println("while the APICaller gives you more direct control over the response.")
	fmt.Println("\nKey features demonstrated:")
	fmt.Println("  - Global caller caching for reuse")
	fmt.Println("  - Automatic context management with configurable timeouts")
	fmt.Println("  - Retry logic with exponential backoff for timeout errors")
	fmt.Println("\nSupported endpoint types:")
	fmt.Println("  - REST: JSON APIs with automatic parsing")
	fmt.Println("  - SOAP: Automatic envelope generation and XML parsing")
	fmt.Println("  - gRPC: Dynamic protobuf messages via reflection")
}

// StartGRPCServer starts a simple gRPC server for testing on the given port
func StartGRPCServer(port string) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	// Create the gRPC server
	server := grpc.NewServer()

	// Register a simple test service
	if err := registerTestService(server); err != nil {
		return nil, fmt.Errorf("failed to register service: %w", err)
	}

	// Register reflection service for dynamic discovery
	reflection.Register(server)

	// Start server in goroutine
	go func() {
		log.Printf("gRPC test server listening on :%s with reflection enabled", port)
		if err := server.Serve(lis); err != nil {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	return server, nil
}

// registerTestService registers a simple echo/test service
func registerTestService(server *grpc.Server) error {
	// Create a simple service descriptor programmatically
	// This creates a UserService with a GetUser method
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("user.proto"),
		Package: proto.String("user"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("GetUserRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("user_id"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
			{
				Name: proto.String("GetUserResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("user_id"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:   proto.String("name"),
						Number: proto.Int32(2),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:   proto.String("email"),
						Number: proto.Int32(3),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:   proto.String("message"),
						Number: proto.Int32(4),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: proto.String("UserService"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("GetUser"),
						InputType:  proto.String(".user.GetUserRequest"),
						OutputType: proto.String(".user.GetUserResponse"),
					},
				},
			},
		},
	}

	// Build the file descriptor
	fd, err := protodesc.NewFile(fileDesc, protoregistry.GlobalFiles)
	if err != nil {
		return fmt.Errorf("failed to create file descriptor: %w", err)
	}

	// Register the file descriptor
	err = protoregistry.GlobalFiles.RegisterFile(fd)
	if err != nil {
		// Ignore if already registered (e.g., from multiple runs)
		log.Printf("Note: file descriptor may already be registered: %v", err)
	}

	// Get the service descriptor
	services := fd.Services()
	if services.Len() == 0 {
		return fmt.Errorf("no services found in descriptor")
	}
	serviceDesc := services.Get(0)

	// Get the method descriptor
	methods := serviceDesc.Methods()
	if methods.Len() == 0 {
		return fmt.Errorf("no methods found in service")
	}
	methodDesc := methods.Get(0)

	// Create a handler for the method
	handler := func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
		// Create a dynamic request message
		reqMsg := dynamicpb.NewMessage(methodDesc.Input())
		if err := dec(reqMsg); err != nil {
			return nil, err
		}

		// Extract the user_id from the request
		userID := ""
		reqMsg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			if fd.Name() == "user_id" {
				userID = v.String()
				return false
			}
			return true
		})

		// Create the response message
		respMsg := dynamicpb.NewMessage(methodDesc.Output())
		respMsg.Set(methodDesc.Output().Fields().ByName("user_id"), protoreflect.ValueOfString(userID))
		respMsg.Set(methodDesc.Output().Fields().ByName("name"), protoreflect.ValueOfString("Test User"))
		respMsg.Set(methodDesc.Output().Fields().ByName("email"), protoreflect.ValueOfString("test@example.com"))
		respMsg.Set(methodDesc.Output().Fields().ByName("message"), protoreflect.ValueOfString("Hello from gRPC test server!"))

		return respMsg, nil
	}

	// Register the service and handler
	serviceInfo := &grpc.ServiceDesc{
		ServiceName: string(serviceDesc.FullName()),
		HandlerType: (*interface{})(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: string(methodDesc.Name()),
				Handler:    handler,
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: fileDesc.GetName(),
	}

	server.RegisterService(serviceInfo, struct{}{})

	return nil
}
