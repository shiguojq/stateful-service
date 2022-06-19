package main

import (
	"flag"
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
	req := args.Message.(*example1.AddNumRequest)
	slog.Infof("[ServiceA] start handle AddNumReq %+v", req)
	s.Number += req.Num
	storeReq := &example1.StoreNumRequest{
		Num: s.Number,
	}
	result := manager.MethodResult{
		Message:   storeReq,
		IsRequest: true,
		Callback:  "AddNumCallback",
		Target:    "",
	}
	slog.Infof("[ServiceA] handle AddNumReq %+v finish, result: %v", req, result)
	return result
}

func (s *ServiceA) AddNumCallback(args manager.MethodArgs) manager.MethodResult {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	req := args.Message.(*example1.StoreNumReponse)
	slog.Infof("[ServiceA] start handle AddNumCallbackReq %+v", req)
	s.Number++
	addResp := &example1.AddNumReponse{
		Result: req.Result,
	}
	result := manager.MethodResult{
		Message:   addResp,
		IsRequest: false,
	}
	slog.Infof("[ServiceA] handle AddNumCallbackReq %+v finish, result: %v", req, result)
	return result
}

var (
	port int
)

func init() {
	flag.IntVar(&port, "port", 80, "server listen port")
}

func main() {
	flag.Parse()
	proxy := rpcproxy.NewProxy("ServiceA", port)
	slog.Infof("[ServiceA] create new proxy: %v, running port: %v", "ServiceA", port)
	svc := &ServiceA{
		Number:  0,
		mutex:   sync.Mutex{},
	}
	if err := proxy.RegisterService(svc); err != nil {
		slog.Errorf("[ServiceA] register service failed, err: %v", err)
		return
	}
	if err := proxy.RegisterMethod(svc, "AddNum", svc.AddNum); err != nil {
		slog.Errorf("[ServiceA] register method %v failed, err: %v", "AddNum", err)
		return
	}
	if err := proxy.RegisterMethod(svc, "AddNumCallback", svc.AddNumCallback); err != nil {
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