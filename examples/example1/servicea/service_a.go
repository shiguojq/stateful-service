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

type ServiceA struct {
	Number int32
	mutex  sync.Mutex
}

func (s *ServiceA) AddNum(args manager.MethodArgs) manager.MethodResult {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	req := &example1.AddNumRequest{}
	if err := proto.Unmarshal(args.Message, req); err != nil {
		slog.Errorf("[ServiceA] unmarshal message failed, err: %v", err)
		return manager.MethodResult{}
	}
	slog.Infof("[ServiceA] start handle AddNumReq %+v", req)
	s.Number += req.Num
	storeReq := &example1.StoreNumRequest{
		Num: s.Number,
	}

	data, err := proto.Marshal(storeReq)
	if err != nil {
		slog.Errorf("[ServiceA] marshal message failed, err: %v", err)
		return manager.MethodResult{}
	}

	result := manager.MethodResult{
		Message:    data,
		IsRequest:  true,
		Callback:   "ServiceA.AddNumCallback",
		Target:     "ServiceB.StoreNum",
		TargetHost: "serviceb:8080",
	}
	slog.Infof("[ServiceA] handle AddNumReq %+v finish, result: %+v", req, result)
	return result
}

func (s *ServiceA) AddNumCallback(args manager.MethodArgs) manager.MethodResult {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	req := &example1.StoreNumReponse{}
	if err := proto.Unmarshal(args.Message, req); err != nil {
		slog.Errorf("[ServiceA] unmarshal message failed, err: %v", err)
		return manager.MethodResult{}
	}

	slog.Infof("[ServiceA] start handle AddNumCallbackReq %+v", req)
	s.Number++
	addResp := &example1.AddNumReponse{
		Result: req.Result,
	}

	data, err := proto.Marshal(addResp)
	if err != nil {
		slog.Errorf("[ServiceA] marshal message failed, err: %v", err)
		return manager.MethodResult{}
	}

	result := manager.MethodResult{
		Message:   data,
		IsRequest: false,
	}
	slog.Infof("[ServiceA] handle AddNumCallbackReq %+v finish, result: %+v", req, result)
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
	proxy := rpcproxy.NewProxy("servicea", port)
	slog.Infof("[ServiceA] create new proxy: %v, running port: %v", "ServiceA", port)
	svc := &ServiceA{
		Number: 0,
		mutex:  sync.Mutex{},
	}
	if err := proxy.RegisterService(svc); err != nil {
		slog.Errorf("[ServiceA] register service failed, err: %v", err)
		return
	}
	if err := proxy.RegisterMethod(svc, "AddNum", svc.AddNum, false); err != nil {
		slog.Errorf("[ServiceA] register method %v failed, err: %v", "AddNum", err)
		return
	}
	if err := proxy.RegisterMethod(svc, "AddNumCallback", svc.AddNumCallback, true); err != nil {
		slog.Errorf("[ServiceA] register method %v failed, err: %v", "AddNumCallback", err)
		return
	}
	if err := proxy.RegisterState(svc); err != nil {
		slog.Errorf("[ServiceA] register state failed, err: %v", err)
		return
	}
	if err := proxy.RegisterFields(svc, "Number"); err != nil {
		slog.Errorf("[ServiceA] register field %v failed, err: %v", "Number", err)
		return
	}
	proxy.Start()
}
