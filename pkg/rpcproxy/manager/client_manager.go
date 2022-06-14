package manager

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	pb "stateful-service/proto"
)

type ClientManager interface {
}

type clientManager struct {
	mutexs map[string]Mutex
}

func NewClientManager() ClientManager {
	m := &clientManager{}
	return m
}

type RpcClient interface {
}

type rpcClient struct {
	client pb.RpcServerClient
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
	cli.client = pb.NewRpcServerClient(conn)
	return cli, nil
}

func (cli *rpcClient) CallRpc(svcMeth string, req []byte) ([]byte, error) {
	request := &pb.RpcMsgRequest{
		Id:      "11", //implement algorithm to generate id
		SvcMeth: svcMeth,
		Msg:     req,
	}

	reqMsg := ReqMsg{
		Id:      request.Id,
		ReplyCh: make(chan ReplyMsg),
	}
	//cli.messageManager.AddMessage(reqMsg)

	resp, err := cli.client.HandleRequest(context.Background(), request)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, errors.New("not found")
	}

	replyMsg := <-reqMsg.ReplyCh
	return replyMsg.Reply, nil
}
