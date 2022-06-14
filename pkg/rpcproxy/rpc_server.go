package rpcproxy

import (
	"context"
	"google.golang.org/grpc"
	"reflect"
	"stateful-service/pkg/rpcproxy/manager"
	pb "stateful-service/proto"
	"strings"
)

type RpcProxy interface {
	Start()
	RegisterService(rcvr interface{}) error
	RegisterMethods(rcvr interface{}, methodNames ...string) error
	RegisterState(rcvr interface{}) error
	RegisterFields(rcvr interface{}, fieldNames ...string) error
}

type rpcProxy struct {
	pb.UnimplementedRpcServerServer
	RpcPort        uint
	StateManager   manager.StateManager
	ServiceManager manager.ServiceManager
	MessageManager manager.MessageManager
	ServiceInputCh map[string]chan manager.ReqMsg
	RequestMsgs    map[string]*manager.ReqMsg
}

func NewServer() RpcProxy {
	s := &rpcProxy{}
	s.RpcPort = 0
	s.StateManager = manager.NewStateManager()
	s.ServiceManager = manager.NewServiceManager()
	s.MessageManager = manager.NewMessageManager()
	return s
}

func (s *rpcProxy) Start() {
	grpcServer := grpc.NewServer()
	pb.RegisterRpcServerServer(grpcServer, s)

}

func (s *rpcProxy) RegisterService(rcvr interface{}) error {
	serviceName := reflect.TypeOf(rcvr).Name()
	if _, ok := s.ServiceInputCh[serviceName]; ok {
		return nil
	}
	s.ServiceInputCh[serviceName] = make(chan manager.ReqMsg)
	err := s.ServiceManager.RegisterService(rcvr, s.ServiceInputCh[serviceName])
	return err
}

func (s *rpcProxy) RegisterMethods(rcvr interface{}, methodNames ...string) error {
	serviceName := reflect.TypeOf(rcvr).Name()
	err := s.ServiceManager.RegisterMethods(serviceName, methodNames...)
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

func (s *rpcProxy) GetClient() {

}

func (s *rpcProxy) HandleRequest(_ context.Context, req *pb.RpcMsgRequest) (*pb.RpcMsgResponse, error) {
	dot := strings.LastIndex(req.SvcMeth, ".")
	serviceName := req.SvcMeth[:dot]
	methodName := req.SvcMeth[dot+1:]
	msg := manager.ReqMsg{
		MethodName: methodName,
		Id:         req.Id,
		Args:       req.Msg,
		ReplyCh:    make(chan manager.ReplyMsg),
	}
	s.ServiceInputCh[serviceName] <- msg
	s.RequestMsgs[req.Id] = &msg
	return &pb.RpcMsgResponse{
		Id: req.Id,
		Ok: true,
	}, nil
}

func (s *rpcProxy) HandleResponse(_ context.Context, req *pb.RpcMsgRequest) (*pb.RpcMsgResponse, error) {
	reqMsg := s.MessageManager.GetMessage(req.Id)
	reqMsg.ReplyCh <- manager.ReplyMsg{
		Ok:    true,
		Reply: req.Msg,
	}
	return &pb.RpcMsgResponse{
		Id: req.Id,
		Ok: true,
	}, nil
}

func (s *rpcProxy) HandleCheckpoint() {

}
