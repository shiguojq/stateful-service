package rpcproxy

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"reflect"
	"stateful-service/config"
	"stateful-service/pkg/rpcproxy/manager"
	pb "stateful-service/proto"
	"stateful-service/slog"
	"stateful-service/utils"
)

type RpcProxy interface {
	Start()
	RegisterService(rcvr interface{}) error
	RegisterMethod(rcvr interface{}, methodName string, fn manager.RpcMethodFunc) error
	RegisterState(rcvr interface{}) error
	RegisterFields(rcvr interface{}, fieldNames ...string) error
}

type rpcProxy struct {
	pb.UnimplementedRpcProxyServer
	RpcPort        int
	ServerName     string
	StateManager   manager.StateManager
	ServiceManager manager.ServiceManager
	MessageManager manager.MessageManager
	ServiceInputCh map[string]chan manager.ReqMsg
	RequestMsgs    map[int64][]manager.ReqMsg
	IdGenClient    pb.IdGeneratorClient
}

func NewProxy(serverName string, port int) RpcProxy {
	s := &rpcProxy{}
	s.RpcPort = port
	s.ServerName = serverName
	s.StateManager = manager.NewStateManager()
	s.ServiceManager = manager.NewServiceManager()
	s.MessageManager = manager.NewMessageManager()
	s.RequestMsgs = make(map[int64][]manager.ReqMsg)
	conn, err := grpc.Dial(config.GetStringEnv(config.EnvIdGeneratorHost), )
	if err != nil {
		slog.Fatal("connect to IdGenerator failed, err: %v", err.Error())
	}
	s.IdGenClient = pb.NewIdGeneratorClient(conn);
	return s
}

func (s *rpcProxy) Start() {
	grpcServer := grpc.NewServer()
	pb.RegisterRpcProxyServer(grpcServer, s)
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", s.RpcPort))
	if err != nil {
		slog.Errorf("[RpcProxy: %v] create listener failed, err: %v", s.ServerName, err)
		return
	}
	if err := grpcServer.Serve(lis); err != nil {
		slog.Errorf("[RpcProxy: %v] start grpc server failed, err: %v", s.ServerName, err)
	}
}

func (s *rpcProxy) RegisterService(rcvr interface{}) error {
	serviceName := reflect.TypeOf(rcvr).Name()
	if _, ok := s.ServiceInputCh[serviceName]; ok {
		return nil
	}
	s.ServiceInputCh[serviceName] = make(chan manager.ReqMsg)
	err := s.ServiceManager.RegisterService(rcvr, s.ServiceInputCh[serviceName], s.StateManager.Checkpoint)
	return err
}

func (s *rpcProxy) RegisterMethod(rcvr interface{}, methodName string, fn manager.RpcMethodFunc) error {
	serviceName := reflect.TypeOf(rcvr).Name()
	err := s.ServiceManager.RegisterMethod(serviceName, methodName, fn)
	return err
}

func (s *rpcProxy) RegisterState(rcvr interface{}) error {
	err := s.StateManager.RegisterState(rcvr)
	return err
}

func (s *rpcProxy) RegisterFields(rcvr interface{}, fieldNames ...string) error {
	stateName := reflect.TypeOf(rcvr).Name()
	err := s.StateManager.RegisterFields(stateName, fieldNames...)
	return err
}

func (s *rpcProxy) GetMutex() {

}

func (s *rpcProxy) AsyncCall(_ context.Context, req *pb.AsyncCallRequest) (*pb.AsyncCallResponse, error) {
	serviceName, methodName := utils.ParseTarget(req.Target)

	msg := manager.ReqMsg{
		MethodName: methodName,
		Id:         req.Id,
		Args:       req.Msg,
		Source:     req.Source,
		Callback:   req.Callback,
		Type:       manager.AsyncReq,
		ReplyCh:    make(chan manager.ReplyMsg),
	}
	s.ServiceInputCh[serviceName] <- msg
	if _, ok := s.RequestMsgs[req.Id]; !ok {
		s.RequestMsgs[req.Id] = make([]manager.ReqMsg, 0)
	}
	s.RequestMsgs[req.Id] = append(s.RequestMsgs[req.Id], msg)
	return &pb.AsyncCallResponse{
		Id: req.Id,
		Ok: true,
	}, nil
}

func (s *rpcProxy) InitSyncCall(_ context.Context, req *pb.InitSyncCallRequest) (*pb.InitSyncCallResponse, error) {
	serviceName, _ := utils.ParseTarget(req.Target)

	resp := &pb.InitSyncCallResponse{}
	if serviceName != s.ServerName {
		resp.Ok = false
		return resp, nil
	}
	idReq := &pb.GenerateIdRequest{
		SvcName: serviceName,
	}
	idResp, err := s.IdGenClient.GenerateId(context.Background(), idReq)
	if err != nil || !idResp.Ok {
		if err == nil {
			err = errors.New("UNKOWN")
		}
		slog.Errorf("micro service %s call GenerateId failed, err: %v", serviceName, err.Error())
		resp.Ok = false
		return resp, err
	}

	id := idResp.Id
	callReq := &pb.AsyncCallRequest{
		Id:       id,
		Target:   req.Target,
		Msg:      req.Msg,
		Callback: "",
		Source:   "client",
	}
	callResp, err := s.AsyncCall(context.Background(), callReq)
	if err != nil || !callResp.Ok {
		if err == nil {
			err = errors.New("UNKOWN")
		}
		slog.Errorf("micro service %s init sync call failed, err: %v", serviceName, err.Error())
		resp.Ok = false
		return resp, err
	}

	resp.Id = idResp.Id
	resp.Ok = true
	return resp, nil
}

func (s *rpcProxy) SyncCall(_ context.Context, req *pb.SyncCallRequest) (*pb.SyncCallResponse, error) {
	resp := &pb.SyncCallResponse{}
	if _, ok := s.RequestMsgs[req.Id]; !ok {
		slog.Errorf("micro service %s can not message of request %v", s.ServerName, req.Id)
		resp.Ok = false
		return resp, nil
	}

	msg := s.RequestMsgs[req.Id][len(s.RequestMsgs) - 1]
	replyMsg := <- msg.ReplyCh
	if !replyMsg.Ok {
		slog.Errorf("micro service %s handle sync request %v failed", s.ServerName, req.Id)
		resp.Ok = false
		return resp, nil
	}

	resp.Msg = replyMsg.Reply
	resp.Ok = true
	return resp, nil
}

func (s *rpcProxy) InitCheckpoint(_ context.Context, req *pb.InitCheckpointRequest) (*pb.InitCheckpointResponse, error) {
	serviceName, methodName := utils.ParseTarget(req.Target)

	resp := &pb.InitCheckpointResponse{}
	if serviceName != s.ServerName {
		resp.Ok = false
		return resp, nil
	}

	idReq := &pb.GenerateIdRequest{
		SvcName: serviceName,
	}
	idResp, err := s.IdGenClient.GenerateId(context.Background(), idReq)
	if err != nil || !idResp.Ok {
		if err == nil {
			err = errors.New("UNKOWN")
		}
		slog.Errorf("micro service %s call GenerateId failed, err: %v", serviceName, err.Error())
		resp.Ok = false
		return resp, err
	}
	id := resp.Id

	msg := manager.ReqMsg{
		MethodName: methodName,
		Id:         id,
		Type:       manager.CheckpointReq,
		Source:     req.Source,
		ReplyCh:    make(chan manager.ReplyMsg),
	}

	s.ServiceInputCh[serviceName] <- msg
	if _, ok := s.RequestMsgs[id]; !ok {
		s.RequestMsgs[id] = make([]manager.ReqMsg, 0)
	}
	s.RequestMsgs[id] = append(s.RequestMsgs[id], msg)
	resp.Id = id
	resp.Ok = true
	return resp, nil
}

func (s *rpcProxy) WaitCheckpoint(_ context.Context, req *pb.WaitCheckpointRequest) (*pb.WaitCheckpointResponse, error) {
	resp := &pb.WaitCheckpointResponse{}
	if _, ok := s.RequestMsgs[req.Id]; !ok {
		slog.Errorf("micro service %s can not message of request %v", s.ServerName, req.Id)
		resp.Ok = false
		return resp, nil
	}
	msg := s.RequestMsgs[req.Id][len(s.RequestMsgs) - 1]
	<- msg.ReplyCh
	resp.Ok = true
	return resp, nil
}

func (s *rpcProxy) SendCheckpoint(_ context.Context, req *pb.SendBarrierRequest) (*pb.SendBarrierResponse, error) {
	slog.Infof("rpc proxy %s start SendCheckpoint, req: %v", s.ServerName, req)
	serviceName, methodName := utils.ParseTarget(req.Target)

	resp := &pb.SendBarrierResponse{}
	if serviceName != s.ServerName {
		resp.Ok = false
		return resp, nil
	}

	msg := manager.ReqMsg{
		MethodName: methodName,
		Id:         req.Id,
		Type:       manager.CheckpointReq,
		Source:     req.Source,
		ReplyCh:    make(chan manager.ReplyMsg),
	}

	s.ServiceInputCh[serviceName] <- msg

	resp.Ok = true
	return resp, nil
}

