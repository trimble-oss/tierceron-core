// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: statsdk/statsdk.proto

package statsdk

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	StatService_GetStats_FullMethodName = "/statsdk.StatService/GetStats"
	StatService_SetStats_FullMethodName = "/statsdk.StatService/SetStats"
)

// StatServiceClient is the client API for StatService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StatServiceClient interface {
	GetStats(ctx context.Context, in *GetStatRequest, opts ...grpc.CallOption) (*GetStatResponse, error)
	SetStats(ctx context.Context, in *SetStatRequest, opts ...grpc.CallOption) (*SetStatResponse, error)
}

type statServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewStatServiceClient(cc grpc.ClientConnInterface) StatServiceClient {
	return &statServiceClient{cc}
}

func (c *statServiceClient) GetStats(ctx context.Context, in *GetStatRequest, opts ...grpc.CallOption) (*GetStatResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetStatResponse)
	err := c.cc.Invoke(ctx, StatService_GetStats_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *statServiceClient) SetStats(ctx context.Context, in *SetStatRequest, opts ...grpc.CallOption) (*SetStatResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SetStatResponse)
	err := c.cc.Invoke(ctx, StatService_SetStats_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StatServiceServer is the server API for StatService service.
// All implementations must embed UnimplementedStatServiceServer
// for forward compatibility.
type StatServiceServer interface {
	GetStats(context.Context, *GetStatRequest) (*GetStatResponse, error)
	SetStats(context.Context, *SetStatRequest) (*SetStatResponse, error)
	mustEmbedUnimplementedStatServiceServer()
}

// UnimplementedStatServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedStatServiceServer struct{}

func (UnimplementedStatServiceServer) GetStats(context.Context, *GetStatRequest) (*GetStatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStats not implemented")
}
func (UnimplementedStatServiceServer) SetStats(context.Context, *SetStatRequest) (*SetStatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetStats not implemented")
}
func (UnimplementedStatServiceServer) mustEmbedUnimplementedStatServiceServer() {}
func (UnimplementedStatServiceServer) testEmbeddedByValue()                     {}

// UnsafeStatServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StatServiceServer will
// result in compilation errors.
type UnsafeStatServiceServer interface {
	mustEmbedUnimplementedStatServiceServer()
}

func RegisterStatServiceServer(s grpc.ServiceRegistrar, srv StatServiceServer) {
	// If the following call pancis, it indicates UnimplementedStatServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&StatService_ServiceDesc, srv)
}

func _StatService_GetStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StatServiceServer).GetStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: StatService_GetStats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StatServiceServer).GetStats(ctx, req.(*GetStatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StatService_SetStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetStatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StatServiceServer).SetStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: StatService_SetStats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StatServiceServer).SetStats(ctx, req.(*SetStatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// StatService_ServiceDesc is the grpc.ServiceDesc for StatService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var StatService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "statsdk.StatService",
	HandlerType: (*StatServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetStats",
			Handler:    _StatService_GetStats_Handler,
		},
		{
			MethodName: "SetStats",
			Handler:    _StatService_SetStats_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "statsdk/statsdk.proto",
}
