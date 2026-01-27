package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// GRPCClient implements the Client interface for gRPC APIs
type GRPCClient struct {
	endpoint Endpoint
	conn     *grpc.ClientConn
	config   *APICallerConfig
}

// NewGRPCClient creates a new gRPC client
func NewGRPCClient(endpoint Endpoint, config *APICallerConfig) (*GRPCClient, error) {
	client := &GRPCClient{
		endpoint: endpoint,
		config:   config,
	}

	var opts []grpc.DialOption

	// Configure TLS
	if config.InsecureSkipVerify {
		// Use insecure credentials
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// Create TLS configuration
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}

		// Load CA certificate if provided
		if config.CACertPath != "" {
			caCert, err := os.ReadFile(config.CACertPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate: %w", err)
			}
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}
			tlsConfig.RootCAs = caCertPool
		}

		// Load client certificate if provided
		if config.TLSCertPath != "" && config.TLSKeyPath != "" {
			cert, err := tls.LoadX509KeyPair(config.TLSCertPath, config.TLSKeyPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificate: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	// Set dial timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Establish connection
	conn, err := grpc.DialContext(ctx, endpoint.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client.conn = conn
	return client, nil
}

// Call makes a gRPC call using reflection and dynamic messages
// Accepts map[string]interface{} as body and returns parsed response
func (gc *GRPCClient) Call(options *CallOptions) (*Response, error) {
	if options.Method == "" {
		return nil, fmt.Errorf("method is required for gRPC calls")
	}

	// Create a context with timeout if not already set
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// Add timeout to context if specified
	if options.Timeout != nil {
		if timeout, ok := options.Timeout.(time.Duration); ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
	}

	// Get method descriptor using reflection
	methodDesc, err := gc.getMethodDescriptor(ctx, options.Method)
	if err != nil {
		return nil, fmt.Errorf("failed to get method descriptor: %w", err)
	}

	// Convert map[string]interface{} to dynamic protobuf message
	request := dynamicpb.NewMessage(methodDesc.Input())
	if err := gc.mapToProtoMessage(options.Body, request); err != nil {
		return nil, fmt.Errorf("failed to convert request to protobuf: %w", err)
	}

	// Create dynamic response message
	response := dynamicpb.NewMessage(methodDesc.Output())

	// Invoke the method
	err = gc.conn.Invoke(ctx, options.Method, request, response)
	if err != nil {
		return &Response{
			StatusCode: 500,
			Error:      err,
		}, err
	}

	// Convert response to map[string]interface{}
	responseMap, err := gc.protoMessageToMap(response)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response to map: %w", err)
	}

	// Return in standardized format
	return &Response{
		StatusCode: 200,
		Body:       responseMap,
		Headers:    make(map[string][]string),
	}, nil
}

// getMethodDescriptor retrieves the method descriptor using gRPC reflection
func (gc *GRPCClient) getMethodDescriptor(ctx context.Context, methodName string) (protoreflect.MethodDescriptor, error) {
	// Create reflection client
	reflectClient := grpc_reflection_v1alpha.NewServerReflectionClient(gc.conn)
	stream, err := reflectClient.ServerReflectionInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reflection stream: %w", err)
	}
	defer stream.CloseSend()

	// Extract service and method from full method name (e.g., "/package.Service/Method")
	serviceName, _, err := parseMethodName(methodName)
	if err != nil {
		return nil, err
	}

	// Request file descriptor for service
	req := &grpc_reflection_v1alpha.ServerReflectionRequest{
		MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: serviceName,
		},
	}

	if err := stream.Send(req); err != nil {
		return nil, fmt.Errorf("failed to send reflection request: %w", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("failed to receive reflection response: %w", err)
	}

	// Extract file descriptor
	fdResp, ok := resp.MessageResponse.(*grpc_reflection_v1alpha.ServerReflectionResponse_FileDescriptorResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected reflection response type")
	}

	// Parse file descriptor
	if len(fdResp.FileDescriptorResponse.FileDescriptorProto) == 0 {
		return nil, fmt.Errorf("no file descriptor returned")
	}

	var fdProto descriptorpb.FileDescriptorProto
	if err := proto.Unmarshal(fdResp.FileDescriptorResponse.FileDescriptorProto[0], &fdProto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file descriptor: %w", err)
	}

	fileDesc, err := protodesc.NewFile(&fdProto, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create file descriptor: %w", err)
	}

	// Find the method descriptor
	services := fileDesc.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		if string(service.FullName()) == serviceName {
			methods := service.Methods()
			for j := 0; j < methods.Len(); j++ {
				method := methods.Get(j)
				if "/"+string(service.FullName())+"/"+string(method.Name()) == methodName {
					return method, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("method %s not found", methodName)
}

// parseMethodName extracts service and method from full method path
func parseMethodName(fullMethod string) (string, string, error) {
	if len(fullMethod) == 0 || fullMethod[0] != '/' {
		return "", "", fmt.Errorf("invalid method name format: %s", fullMethod)
	}

	// Remove leading slash
	fullMethod = fullMethod[1:]

	// Split by last slash
	lastSlash := -1
	for i := len(fullMethod) - 1; i >= 0; i-- {
		if fullMethod[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 {
		return "", "", fmt.Errorf("invalid method name format: %s", fullMethod)
	}

	return fullMethod[:lastSlash], fullMethod[lastSlash+1:], nil
}

// mapToProtoMessage converts map[string]interface{} to a protobuf message
func (gc *GRPCClient) mapToProtoMessage(data interface{}, msg protoreflect.Message) error {
	// Convert map to JSON, then use protojson to unmarshal into message
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	// Use protojson to unmarshal into dynamic message
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	if err := unmarshaler.Unmarshal(jsonData, msg.Interface()); err != nil {
		return fmt.Errorf("failed to unmarshal JSON to protobuf: %w", err)
	}

	return nil
}

// protoMessageToMap converts a protobuf message to map[string]interface{}
func (gc *GRPCClient) protoMessageToMap(msg protoreflect.Message) (map[string]interface{}, error) {
	// Use protojson to marshal to JSON
	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}
	jsonData, err := marshaler.Marshal(msg.Interface())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protobuf to JSON: %w", err)
	}

	// Unmarshal JSON to map
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}

	return result, nil
}

// GetConnection returns the underlying gRPC connection for custom service clients
func (gc *GRPCClient) GetConnection() *grpc.ClientConn {
	return gc.conn
}

// Close closes the gRPC connection
func (gc *GRPCClient) Close() error {
	if gc.conn != nil {
		return gc.conn.Close()
	}
	return nil
}
