package manager

import (
	"context"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"stateful-service/config"
	pb "stateful-service/proto"
	"stateful-service/slog"
	"strings"
)

type ClientManager interface {
}

type clientManager struct {
	Ch      chan proto.Message
	Clients map[string]pb.RpcProxyClient
}

func NewClientManager(ch chan proto.Message) ClientManager {
	mCli := &clientManager{}
	mCli.Ch = ch
	go func() {
		for req := range ch {
			switch req.(type) {

			case *pb.AsyncCallRequest:
				{
					asyncReq := req.(*pb.AsyncCallRequest)
					dot := strings.LastIndex(asyncReq.Target, ".")
					if dot == -1 {
						continue
					}
					serviceName := asyncReq.Target[:dot]
					client, err := mCli.GetClient(serviceName)
					if err != nil {
						continue
					}
					_, err = client.AsyncCall(context.Background(), asyncReq)
					if err != nil {
						slog.Errorf("call service %v failed, err: %v", serviceName, err.Error())
					}
				}

			case *pb.InitCheckpointRequest:
				{
					checkpointReq := req.(*pb.InitCheckpointRequest)
					dot := strings.LastIndex(checkpointReq.Target, ".")
					if dot == -1 {
						continue
					}
					serviceName := checkpointReq.Target[:dot]
					client, err := mCli.GetClient(serviceName)
					if err != nil {
						continue
					}
					_, err = client.InitCheckpoint(context.Background(), checkpointReq)
					if err != nil {
						slog.Errorf("call service %v failed, err: %v", serviceName, err.Error())
					}
				}
			}
		}
	}()
	return mCli
}

func (mCli *clientManager) GetClient(serviceName string) (pb.RpcProxyClient, error) {
	if _, ok := mCli.Clients[serviceName]; !ok {
		opts := []grpc.DialOption{
			grpc.WithWriteBufferSize(512 * 1024 * 1024),
			grpc.WithReadBufferSize(512 * 1024 * 1024),
		}

		target := serviceName + ":" + config.GetStringEnv(config.EnvRunningPort)
		conn, err := grpc.Dial(target, opts...)
		if err != nil {
			slog.Errorf("connect to service %s failed, err: %v", serviceName, err.Error())
			return nil, err
		}
		mCli.Clients[serviceName] = pb.NewRpcProxyClient(conn)
	}
	return mCli.Clients[serviceName], nil
}
