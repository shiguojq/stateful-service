// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: proto/rpc_proxy.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RpcProxyClient is the client API for RpcProxy service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RpcProxyClient interface {
	AsyncCall(ctx context.Context, in *AsyncCallRequest, opts ...grpc.CallOption) (*AsyncCallResponse, error)
	InitSyncCall(ctx context.Context, in *InitSyncCallRequest, opts ...grpc.CallOption) (*InitSyncCallResponse, error)
	SyncCall(ctx context.Context, in *SyncCallRequest, opts ...grpc.CallOption) (*SyncCallResponse, error)
	InitCheckpoint(ctx context.Context, in *InitCheckpointRequest, opts ...grpc.CallOption) (*InitCheckpointResponse, error)
	WaitCheckpoint(ctx context.Context, in *WaitCheckpointRequest, opts ...grpc.CallOption) (*WaitCheckpointResponse, error)
	SendBarrier(ctx context.Context, in *SendBarrierRequest, opts ...grpc.CallOption) (*SendBarrierResponse, error)
}

type rpcProxyClient struct {
	cc grpc.ClientConnInterface
}

func NewRpcProxyClient(cc grpc.ClientConnInterface) RpcProxyClient {
	return &rpcProxyClient{cc}
}

func (c *rpcProxyClient) AsyncCall(ctx context.Context, in *AsyncCallRequest, opts ...grpc.CallOption) (*AsyncCallResponse, error) {
	out := new(AsyncCallResponse)
	err := c.cc.Invoke(ctx, "/proto.RpcProxy/AsyncCall", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rpcProxyClient) InitSyncCall(ctx context.Context, in *InitSyncCallRequest, opts ...grpc.CallOption) (*InitSyncCallResponse, error) {
	out := new(InitSyncCallResponse)
	err := c.cc.Invoke(ctx, "/proto.RpcProxy/InitSyncCall", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rpcProxyClient) SyncCall(ctx context.Context, in *SyncCallRequest, opts ...grpc.CallOption) (*SyncCallResponse, error) {
	out := new(SyncCallResponse)
	err := c.cc.Invoke(ctx, "/proto.RpcProxy/SyncCall", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rpcProxyClient) InitCheckpoint(ctx context.Context, in *InitCheckpointRequest, opts ...grpc.CallOption) (*InitCheckpointResponse, error) {
	out := new(InitCheckpointResponse)
	err := c.cc.Invoke(ctx, "/proto.RpcProxy/InitCheckpoint", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rpcProxyClient) WaitCheckpoint(ctx context.Context, in *WaitCheckpointRequest, opts ...grpc.CallOption) (*WaitCheckpointResponse, error) {
	out := new(WaitCheckpointResponse)
	err := c.cc.Invoke(ctx, "/proto.RpcProxy/WaitCheckpoint", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rpcProxyClient) SendBarrier(ctx context.Context, in *SendBarrierRequest, opts ...grpc.CallOption) (*SendBarrierResponse, error) {
	out := new(SendBarrierResponse)
	err := c.cc.Invoke(ctx, "/proto.RpcProxy/SendBarrier", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RpcProxyServer is the server API for RpcProxy service.
// All implementations must embed UnimplementedRpcProxyServer
// for forward compatibility
type RpcProxyServer interface {
	AsyncCall(context.Context, *AsyncCallRequest) (*AsyncCallResponse, error)
	InitSyncCall(context.Context, *InitSyncCallRequest) (*InitSyncCallResponse, error)
	SyncCall(context.Context, *SyncCallRequest) (*SyncCallResponse, error)
	InitCheckpoint(context.Context, *InitCheckpointRequest) (*InitCheckpointResponse, error)
	WaitCheckpoint(context.Context, *WaitCheckpointRequest) (*WaitCheckpointResponse, error)
	SendBarrier(context.Context, *SendBarrierRequest) (*SendBarrierResponse, error)
	mustEmbedUnimplementedRpcProxyServer()
}

// UnimplementedRpcProxyServer must be embedded to have forward compatible implementations.
type UnimplementedRpcProxyServer struct {
}

func (UnimplementedRpcProxyServer) AsyncCall(context.Context, *AsyncCallRequest) (*AsyncCallResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AsyncCall not implemented")
}
func (UnimplementedRpcProxyServer) InitSyncCall(context.Context, *InitSyncCallRequest) (*InitSyncCallResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitSyncCall not implemented")
}
func (UnimplementedRpcProxyServer) SyncCall(context.Context, *SyncCallRequest) (*SyncCallResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SyncCall not implemented")
}
func (UnimplementedRpcProxyServer) InitCheckpoint(context.Context, *InitCheckpointRequest) (*InitCheckpointResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitCheckpoint not implemented")
}
func (UnimplementedRpcProxyServer) WaitCheckpoint(context.Context, *WaitCheckpointRequest) (*WaitCheckpointResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WaitCheckpoint not implemented")
}
func (UnimplementedRpcProxyServer) SendBarrier(context.Context, *SendBarrierRequest) (*SendBarrierResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendBarrier not implemented")
}
func (UnimplementedRpcProxyServer) mustEmbedUnimplementedRpcProxyServer() {}

// UnsafeRpcProxyServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RpcProxyServer will
// result in compilation errors.
type UnsafeRpcProxyServer interface {
	mustEmbedUnimplementedRpcProxyServer()
}

func RegisterRpcProxyServer(s grpc.ServiceRegistrar, srv RpcProxyServer) {
	s.RegisterService(&RpcProxy_ServiceDesc, srv)
}

func _RpcProxy_AsyncCall_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AsyncCallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RpcProxyServer).AsyncCall(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RpcProxy/AsyncCall",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RpcProxyServer).AsyncCall(ctx, req.(*AsyncCallRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RpcProxy_InitSyncCall_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InitSyncCallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RpcProxyServer).InitSyncCall(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RpcProxy/InitSyncCall",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RpcProxyServer).InitSyncCall(ctx, req.(*InitSyncCallRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RpcProxy_SyncCall_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SyncCallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RpcProxyServer).SyncCall(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RpcProxy/SyncCall",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RpcProxyServer).SyncCall(ctx, req.(*SyncCallRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RpcProxy_InitCheckpoint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InitCheckpointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RpcProxyServer).InitCheckpoint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RpcProxy/InitCheckpoint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RpcProxyServer).InitCheckpoint(ctx, req.(*InitCheckpointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RpcProxy_WaitCheckpoint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WaitCheckpointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RpcProxyServer).WaitCheckpoint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RpcProxy/WaitCheckpoint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RpcProxyServer).WaitCheckpoint(ctx, req.(*WaitCheckpointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RpcProxy_SendBarrier_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendBarrierRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RpcProxyServer).SendBarrier(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RpcProxy/SendBarrier",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RpcProxyServer).SendBarrier(ctx, req.(*SendBarrierRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RpcProxy_ServiceDesc is the grpc.ServiceDesc for RpcProxy service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RpcProxy_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.RpcProxy",
	HandlerType: (*RpcProxyServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AsyncCall",
			Handler:    _RpcProxy_AsyncCall_Handler,
		},
		{
			MethodName: "InitSyncCall",
			Handler:    _RpcProxy_InitSyncCall_Handler,
		},
		{
			MethodName: "SyncCall",
			Handler:    _RpcProxy_SyncCall_Handler,
		},
		{
			MethodName: "InitCheckpoint",
			Handler:    _RpcProxy_InitCheckpoint_Handler,
		},
		{
			MethodName: "WaitCheckpoint",
			Handler:    _RpcProxy_WaitCheckpoint_Handler,
		},
		{
			MethodName: "SendBarrier",
			Handler:    _RpcProxy_SendBarrier_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/rpc_proxy.proto",
}
