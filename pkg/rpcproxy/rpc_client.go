package rpcproxy

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	pb "stateful-service/proto/pb"
	"stateful-service/slog"
	"time"
)

type RpcClient interface {
	CallRpc(target string, req []byte) ([]byte, error)
	Close()
}

type rpcClient struct {
	conn *grpc.ClientConn
}

func NewRpcClient(hostport string) (RpcClient, error) {
	cli := &rpcClient{}
	opts := []grpc.DialOption{
		grpc.WithWriteBufferSize(512 * 1024 * 1024),
		grpc.WithReadBufferSize(512 * 1024 * 1024),
		grpc.WithInsecure(),
	}

	var err error
	cli.conn, err = grpc.Dial(hostport, opts...)
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func (c *rpcClient) CallRpc(target string, req []byte) ([]byte, error) {
	initReq := &pb.InitSyncCallRequest{
		Target: target,
		Msg:    req,
	}

	client := pb.NewRpcProxyClient(c.conn)
	initResp, err := client.InitSyncCall(context.Background(), initReq)
	if err != nil {
		return nil, err
	}
	if !initResp.Ok {
		return nil, errors.New("unkown")
	}

	callReq := &pb.SyncCallRequest{
		Id: initResp.Id,
	}
	slog.Infof("start wait req %v complete", initResp.Id)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	callResp, err := client.SyncCall(ctx, callReq)
	if err != nil {
		return nil, err
	}
	if !callResp.Ok {
		return nil, errors.New("unkown")
	}

	return callResp.Msg, nil
}

func (c *rpcClient) Close() {
	c.conn.Close()
}
