package manager

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"reflect"
	pb "stateful-service/proto/pb"
	"stateful-service/slog"
	"stateful-service/utils"
	"sync/atomic"
)

type ServiceManager interface {
	RunService(serviceName string) error
	RegisterService(rcvr interface{}, inputCh chan ReqMsg, fn CheckpointFn) error
	RegisterMethod(serviceName string, methodName string, fn RpcMethodFunc) error
}

type serviceManager struct {
	SourceHost         string
	RegisteredServices map[string]*service
}

func NewServiceManager(sourceHost string) ServiceManager {
	manager := &serviceManager{}
	manager.RegisteredServices = make(map[string]*service)
	manager.SourceHost = sourceHost
	return manager
}

func (m *serviceManager) RegisterService(rcvr interface{}, inputCh chan ReqMsg, fn CheckpointFn) error {
	receiver := reflect.ValueOf(rcvr)
	if receiver.Kind() != reflect.Ptr {
		slog.Errorf("[rpcProxy.manager.serviceManager.RegisterService]: receiver expected kind %v, actual kind %v", reflect.Ptr, receiver.Kind())
		return &ErrReceiverKind{
			kind: receiver.Kind(),
		}
	}
	serviceName := reflect.TypeOf(rcvr).Elem().Name()
	if _, ok := m.RegisteredServices[serviceName]; !ok {
		m.RegisteredServices[serviceName] = &service{
			name:              serviceName,
			typ:               reflect.TypeOf(rcvr),
			rcvr:              receiver.Elem(),
			sourceHost:        m.SourceHost,
			inputCh:           inputCh,
			checkpointFn:      fn,
			checkpointCh:      make(chan string, 1),
			BlockedMethod:     utils.NewStringSet(),
			methodSet:         utils.NewStringSet(),
			registeredMethods: make(map[string]*RpcMethod),
		}
	}
	slog.Infof("[rpcProxy.manager.serviceManager.RegisterService]: register [%v] service successfully", serviceName)
	return nil
}

func (m *serviceManager) RegisterMethod(serviceName string, methodName string, fn RpcMethodFunc) error {
	return m.RegisteredServices[serviceName].registerMethod(methodName, fn)
}

func (m *serviceManager) RunService(serviceName string) error {
	if svc, ok := m.RegisteredServices[serviceName]; !ok {
		return &ErrServiceNotFound{
			serviceName: serviceName,
		}
	} else {
		svc.run()
	}
	return nil
}

type service struct {
	name              string
	typ               reflect.Type
	rcvr              reflect.Value
	sourceHost        string
	inputCh           chan ReqMsg
	running           int32
	checkpointFn      CheckpointFn
	methodSet         utils.StringSet
	checkpointCh      chan string
	BlockedMethod     utils.StringSet
	registeredMethods map[string]*RpcMethod
}

func (svc *service) isRunning() bool {
	return atomic.LoadInt32(&svc.running) == 1
}

func (svc *service) run() {
	if svc.isRunning() {
		return
	}
	defer func() {
		slog.Infof("RpcService %v start running", svc.name)
		atomic.StoreInt32(&svc.running, 1) // maybe not available
	}()
	for _, method := range svc.registeredMethods {
		go method.StartProcess()
	}
	go func() {
		for msg := range svc.inputCh {
			methodName := msg.MethodName
			svc.registeredMethods[methodName].InputCh <- msg
		}
	}()
	go svc.waitCheckpoint()
}

func (svc *service) waitCheckpoint() {
	for methodName := range svc.checkpointCh {
		if svc.BlockedMethod.Has(methodName) {
			continue
		}

		svc.BlockedMethod.Insert(methodName)
		if svc.BlockedMethod.Len() == svc.methodSet.Len() {
			checkpoint := svc.checkpointFn(svc.name)
			slog.Infof("service %s take checkpoint successfully, checkpoint content: %v", svc.name, checkpoint)
			svc.BlockedMethod = utils.NewStringSet()
			for _, method := range svc.registeredMethods {
				method.BlockCh <- struct{}{}
			}
		}
	}
}

func (svc *service) registerMethod(methodName string, fn RpcMethodFunc) error {
	if svc.methodSet.Has(methodName) {
		return nil
	}
	svc.methodSet.Insert(methodName)
	inputCh := make(chan ReqMsg, 1024)
	outputCh := make(chan proto.Message, 1)
	svc.registeredMethods[methodName] = &RpcMethod{
		ServiceName:   svc.name,
		Name:          methodName,
		Fn:            fn,
		SourceHost:    svc.sourceHost,
		InputCh:       inputCh,
		OutputCh:      outputCh,
		CheckpointCh:  svc.checkpointCh,
		BlockCh:       make(chan struct{}),
		ClientManager: NewClientManager(fmt.Sprintf("%v.%v", svc.name, methodName), outputCh),
		SourceSet:     utils.NewStringSet(),
		SourceBlocked: make(map[string]bool),
		TargetSet:     utils.NewStringSet(),
		BarrierSet:    utils.NewStringSet(),
		TargetHosts:   make(map[string]string),
	}
	slog.Infof("[rpcProxy.manager.service.registerMethod]: service [%v] register method [%v] successfully", svc.name, methodName)
	return nil
}

type RpcMethodFunc func(arg MethodArgs) MethodResult

type RpcMethod struct {
	ServiceName   string
	Name          string
	Fn            RpcMethodFunc
	SourceHost    string
	InputCh       chan ReqMsg
	OutputCh      chan proto.Message
	ClientManager ClientManager
	SourceSet     utils.StringSet
	SourceBlocked map[string]bool
	TargetSet     utils.StringSet
	BarrierSet    utils.StringSet
	TargetHosts   map[string]string
	CheckpointCh  chan string
	BlockCh       chan struct{}
	WaitingCh     chan ReqMsg
}

func (method *RpcMethod) StartProcess() {
	slog.Infof("[RpcMethod: %v.%v] start process", method.ServiceName, method.Name)
	for reqMsg := range method.InputCh {
		method.processMsg(reqMsg)
	}
}

func (method *RpcMethod) processMsg(reqMsg ReqMsg) {
	slog.Infof("[RpcMethod: %v.%v] get req message %+v", method.ServiceName, method.Name, reqMsg)
	switch reqMsg.Type {
	case AsyncReq:
		source := reqMsg.Source
		if !method.SourceSet.Has(source) {
			method.SourceSet.Insert(source)
			method.SourceBlocked[source] = false
		}
		if method.SourceBlocked[source] {
			method.WaitingCh <- reqMsg
		}

		args := MethodArgs{
			Message: reqMsg.Args,
		}

		result := method.Fn(args)
		slog.Infof("[RpcMethod: %v.%v] call result: %+v", method.ServiceName, method.Name, result)

		if result.IsRequest {
			req := &pb.AsyncCallRequest{
				Id:     reqMsg.Id,
				Msg:    result.Message,
				Source: fmt.Sprintf("%s.%s", method.ServiceName, method.Name),
				// SourceHost: method.SourceHost,
				SourceHost: method.SourceHost,
			}
			if result.Target != "" {
				req.Target = result.Target
				req.TargetHost = result.TargetHost
				req.Callback = result.Callback
			} else {
				req.Target = reqMsg.Callback
				req.TargetHost = reqMsg.SourceHost
				req.Callback = ""
			}
			method.TargetHosts[req.Target] = req.TargetHost
			if !method.TargetSet.Has(req.Target) {
				method.TargetSet.Insert(req.Target)
			}
			method.OutputCh <- req
		} else {
			replyMsg := ReplyMsg{
				Ok:    true,
				Reply: result.Message,
			}
			reqMsg.ReplyCh <- replyMsg
		}
	case CheckpointReq:
		if method.BarrierSet.Len() == 0 {
			method.WaitingCh = make(chan ReqMsg, 1024)
		}
		source := reqMsg.Source
		if !method.BarrierSet.Has(source) {
			method.BarrierSet.Insert(source)
			method.SourceBlocked[source] = true
		}

		slog.Infof("[RpcMethod: %v.%v] barrier set: %v, source set: %v", method.ServiceName, method.Name, method.BarrierSet, method.SourceSet)

		if method.BarrierSet.Equals(method.SourceSet) {
			for target := range method.TargetSet {
				req := &pb.InitCheckpointRequest{
					Source:     fmt.Sprintf("%s.%s", method.ServiceName, method.Name),
					SourceHost: method.SourceHost,
					Target:     target,
					TargetHost: method.TargetHosts[target],
				}
				method.OutputCh <- req
			}
			method.CheckpointCh <- method.Name
			method.waitCheckpoint()
			replyMsg := ReplyMsg{
				Ok:    true,
				Reply: []byte{},
			}
			reqMsg.ReplyCh <- replyMsg
			for reqMsg := range method.WaitingCh {
				method.processMsg(reqMsg)
			}
		}
	case RestoreReq:
	}
}

func (method *RpcMethod) waitCheckpoint() {
	<-method.BlockCh
	close(method.WaitingCh)
	for source := range method.SourceBlocked {
		method.SourceBlocked[source] = false
	}
	method.BarrierSet = utils.NewStringSet()
}
