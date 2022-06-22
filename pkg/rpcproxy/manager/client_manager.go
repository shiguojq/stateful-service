package manager

import (
	"context"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	pb "stateful-service/proto/pb"
	"stateful-service/slog"
	"strings"
	"time"
)

type ClientManager interface {
}

type clientManager struct {
	Name  string
	Ch    chan proto.Message
	Conns map[string]*grpc.ClientConn
}

func NewClientManager(svcMeth string, ch chan proto.Message) ClientManager {
	mCli := &clientManager{}
	mCli.Name = svcMeth
	mCli.Ch = ch
	mCli.Conns = make(map[string]*grpc.ClientConn)
	go func() {
		slog.Infof("[ClientManager: %v] start process", mCli.Name)
		for req := range mCli.Ch {
			switch req.(type) {
			case *pb.AsyncCallRequest:
				{
					asyncReq := req.(*pb.AsyncCallRequest)
					slog.Infof("[ClientManager: %v] start handle AsyncCallRequest: %+v", mCli.Name, asyncReq)
					dot := strings.LastIndex(asyncReq.Target, ".")
					if dot == -1 {
						continue
					}
					serviceName := asyncReq.Target[:dot]
					conn, err := mCli.GetConn(asyncReq.TargetHost)
					if err != nil {
						continue
					}
					client := pb.NewRpcProxyClient(conn)

					ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
					defer cancel()

					_, err = client.AsyncCall(ctx, asyncReq)
					if err != nil {
						slog.Errorf("call host %v, service %v failed, err: %v", asyncReq.TargetHost, serviceName, err.Error())
					}
					slog.Infof("[ClientManager: %v] source %v asynccall target %v success", mCli.Name, asyncReq.Source, asyncReq.Target)
				}

			case *pb.InitCheckpointRequest:
				{
					checkpointReq := req.(*pb.InitCheckpointRequest)
					dot := strings.LastIndex(checkpointReq.Target, ".")
					if dot == -1 {
						continue
					}
					serviceName := checkpointReq.Target[:dot]
					conn, err := mCli.GetConn(checkpointReq.TargetHost)
					if err != nil {
						continue
					}
					client := pb.NewRpcProxyClient(conn)

					ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
					defer cancel()

					_, err = client.InitCheckpoint(ctx, checkpointReq)
					if err != nil {
						slog.Errorf("call service %v failed, err: %v", serviceName, err.Error())
					}
					slog.Infof("[ClientManager: %v] source %v initCheckpoint target %v success", mCli.Name, checkpointReq.Source, checkpointReq.Target)
				}
			}
		}
	}()
	return mCli
}

func (mCli *clientManager) GetConn(hostname string) (*grpc.ClientConn, error) {
	if _, ok := mCli.Conns[hostname]; !ok {
		opts := []grpc.DialOption{
			grpc.WithWriteBufferSize(512 * 1024 * 1024),
			grpc.WithReadBufferSize(512 * 1024 * 1024),
			grpc.WithInsecure(),
		}

		target := hostname
		conn, err := grpc.Dial(target, opts...)
		if err != nil {
			slog.Errorf("[ClientManager: %v] connect to service %s failed, err: %v", mCli.Name, hostname, err.Error())
			return nil, err
		}
		slog.Infof("[ClientManager: %v] connect to service %s success", mCli.Name, hostname)
		mCli.Conns[hostname] = conn
	}
	return mCli.Conns[hostname], nil
}
