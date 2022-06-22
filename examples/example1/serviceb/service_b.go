package main

import (
	"flag"
	"github.com/golang/protobuf/proto"
	"stateful-service/pkg/rpcproxy"
	"stateful-service/pkg/rpcproxy/manager"
	"stateful-service/proto/example1"
	"stateful-service/slog"
	"sync"
)

type ServiceB struct {
	Number int32
	mutex  sync.Mutex
}

func (s *ServiceB) StoreNum(args manager.MethodArgs) manager.MethodResult {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	req := &example1.StoreNumRequest{}
	if err := proto.Unmarshal(args.Message, req); err != nil {
		slog.Errorf("[ServiceB] unmarshal message failed, err: %v", err)
		return manager.MethodResult{}
	}

	slog.Infof("[ServiceB] start handle StoreNumReq %+v", req)
	s.Number = req.Num
	storeResp := &example1.StoreNumReponse{
		Result: s.Number,
	}
	data, err := proto.Marshal(storeResp)
	if err != nil {
		slog.Errorf("[ServiceB] marshal message failed, err: %v", err)
		return manager.MethodResult{}
	}

	result := manager.MethodResult{
		Message:    data,
		IsRequest:  true,
		Callback:   "",
		Target:     "ServiceB.StoreNumCallback",
		TargetHost: "serviceb:8080",
	}
	slog.Infof("[ServiceB] handle StoreNumReq %+v finish, result: %v", req, result)
	return result
}

func (s *ServiceB) StoreNumCallback(args manager.MethodArgs) manager.MethodResult {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	req := &example1.StoreNumReponse{}
	if err := proto.Unmarshal(args.Message, req); err != nil {
		slog.Errorf("[ServiceB] unmarshal message failed, err: %v", err)
		return manager.MethodResult{}
	}

	slog.Infof("[ServiceB] start handle StoreNumCallbackReq %+v", req)
	s.Number++
	storeResp := &example1.StoreNumReponse{
		Result: s.Number,
	}
	data, err := proto.Marshal(storeResp)
	if err != nil {
		slog.Errorf("[ServiceB] marshal message failed, err: %v", err)
		return manager.MethodResult{}
	}

	result := manager.MethodResult{
		Message:   data,
		IsRequest: true,
	}
	slog.Infof("[ServiceB] handle StoreNumCallbackReq %+v finish, result: %v", req, result)
	return result
}

var (
	port int
)

func init() {
	flag.IntVar(&port, "port", 8080, "server listen port")
}

func main() {
	flag.Parse()
	proxy := rpcproxy.NewProxy("ServiceB", port)
	slog.Infof("[ServiceB] create new proxy: %v, running port: %v", "ServiceB", port)
	svc := &ServiceB{
		Number: 0,
		mutex:  sync.Mutex{},
	}
	if err := proxy.RegisterService(svc); err != nil {
		slog.Errorf("[ServiceB] register service failed, err: %v", err)
		return
	}
	if err := proxy.RegisterMethod(svc, "StoreNum", svc.StoreNum); err != nil {
		slog.Errorf("[ServiceB] register method %v failed, err: %v", "AddNum", err)
		return
	}
	if err := proxy.RegisterMethod(svc, "StoreNumCallback", svc.StoreNumCallback); err != nil {
		slog.Errorf("[ServiceB] register method %v failed, err: %v", "AddNumCallback", err)
		return
	}
	if err := proxy.RegisterState(svc); err != nil {
		slog.Errorf("[ServiceB] register state failed, err: %v", err)
		return
	}
	if err := proxy.RegisterFields(svc, "Number"); err != nil {
		slog.Errorf("[ServiceB] register field %v failed, err: %v", "Number", err)
		return
	}
	proxy.Start()
}
