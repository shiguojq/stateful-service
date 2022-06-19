package rpcproxy

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	pb "stateful-service/proto"
)

type RpcClient interface {
	CallRpc(target string, req []byte) ([]byte, error)
}

type rpcClient struct {
	client pb.RpcProxyClient
}

func NewRpcClient(target string) (RpcClient, error) {
	cli := &rpcClient{}
	opts := []grpc.DialOption{
		grpc.WithWriteBufferSize(512 * 1024 * 1024),
		grpc.WithReadBufferSize(512 * 1024 * 1024),
	}

	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}
	cli.client = pb.NewRpcProxyClient(conn)
	return cli, nil
}

func (c *rpcClient) CallRpc(target string, req []byte) ([]byte, error) {
	initReq := &pb.InitSyncCallRequest{
		Target:  target,
		Msg:     req,
	}

	initResp, err := c.client.InitSyncCall(context.Background(), initReq)
	if err != nil {
		return nil, err
	}
	if !initResp.Ok {
		return nil, errors.New("unkown")
	}

	callReq := &pb.SyncCallRequest{
		Id:  initResp.Id,
	}
	callResp, err := c.client.SyncCall(context.Background(), callReq)
	if err != nil {
		return nil, err
	}
	if !callResp.Ok {
		return nil, errors.New("unkown")
	}

	return callResp.Msg, nil
}