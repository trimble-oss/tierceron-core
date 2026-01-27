package api

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// TestGRPCEndpointWithLocalServer tests gRPC functionality with a local test server
func TestGRPCEndpointWithLocalServer(t *testing.T) {
	// Start a test gRPC server
	server, port, err := startTestGRPCServer()
	if err != nil {
		t.Fatalf("Failed to start test gRPC server: %v", err)
	}
	defer server.GracefulStop()

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	// Create gRPC endpoint
	endpoint := Endpoint{
		FriendlyName: "Test User Service",
		URL:          fmt.Sprintf("localhost:%s", port),
		Type:         EndpointTypeGRPC,
		MethodName:   "/user.UserService/GetUser",
		Timeout:      10 * time.Second,
	}

	config := &APICallerConfig{
		InsecureSkipVerify: true,
	}

	// Test with user_id parameter
	params := map[string]interface{}{
		"body": map[string]interface{}{
			"user_id": "test123",
		},
	}

	result, err := endpoint.Call(params, config)
	if err != nil {
		t.Fatalf("gRPC call failed: %v", err)
	}

	// Verify status code
	statusCode, ok := result["statusCode"].(int)
	if !ok {
		t.Fatalf("Status code not found or wrong type")
	}
	if statusCode != 200 {
		t.Errorf("Expected status code 200, got %d", statusCode)
	}

	// Verify body
	body, ok := result["body"].(map[string]interface{})
	if !ok {
		t.Fatalf("Body not found or not a map")
	}

	// Check response fields
	if userID, ok := body["user_id"].(string); !ok || userID != "test123" {
		t.Errorf("Expected user_id 'test123', got %v", body["user_id"])
	}

	if name, ok := body["name"].(string); !ok || name != "Test User" {
		t.Errorf("Expected name 'Test User', got %v", body["name"])
	}

	if email, ok := body["email"].(string); !ok || email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %v", body["email"])
	}

	if message, ok := body["message"].(string); !ok || message == "" {
		t.Errorf("Expected non-empty message, got %v", body["message"])
	}

	t.Logf("✓ gRPC test passed with response: %+v", body)
}

// startTestGRPCServer starts a simple gRPC test server and returns the server and port
func startTestGRPCServer() (*grpc.Server, string, error) {
	// Find an available port
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, "", fmt.Errorf("failed to listen: %w", err)
	}

	port := fmt.Sprintf("%d", lis.Addr().(*net.TCPAddr).Port)

	server := grpc.NewServer()

	// Register test service
	if err := registerTestUserService(server); err != nil {
		return nil, "", fmt.Errorf("failed to register service: %w", err)
	}

	// Register reflection
	reflection.Register(server)

	// Start server
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Printf("Test gRPC server stopped: %v", err)
		}
	}()

	return server, port, nil
}

// registerTestUserService registers a test UserService
func registerTestUserService(server *grpc.Server) error {
	// Create service descriptor
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

	// Build file descriptor
	fd, err := protodesc.NewFile(fileDesc, protoregistry.GlobalFiles)
	if err != nil {
		return fmt.Errorf("failed to create file descriptor: %w", err)
	}

	// Try to register (ignore if already registered from demo)
	_ = protoregistry.GlobalFiles.RegisterFile(fd)

	// Get service and method descriptors
	services := fd.Services()
	if services.Len() == 0 {
		return fmt.Errorf("no services found")
	}
	serviceDesc := services.Get(0)

	methods := serviceDesc.Methods()
	if methods.Len() == 0 {
		return fmt.Errorf("no methods found")
	}
	methodDesc := methods.Get(0)

	// Create handler
	handler := func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
		reqMsg := dynamicpb.NewMessage(methodDesc.Input())
		if err := dec(reqMsg); err != nil {
			return nil, err
		}

		// Extract user_id
		userID := ""
		reqMsg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			if fd.Name() == "user_id" {
				userID = v.String()
				return false
			}
			return true
		})

		// Create response
		respMsg := dynamicpb.NewMessage(methodDesc.Output())
		respMsg.Set(methodDesc.Output().Fields().ByName("user_id"), protoreflect.ValueOfString(userID))
		respMsg.Set(methodDesc.Output().Fields().ByName("name"), protoreflect.ValueOfString("Test User"))
		respMsg.Set(methodDesc.Output().Fields().ByName("email"), protoreflect.ValueOfString("test@example.com"))
		respMsg.Set(methodDesc.Output().Fields().ByName("message"), protoreflect.ValueOfString("gRPC test response"))

		return respMsg, nil
	}

	// Register service
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
