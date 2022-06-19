package manager

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"reflect"
	pb "stateful-service/proto"
	"stateful-service/slog"
	"stateful-service/utils"
	"sync/atomic"
)

type ServiceManager interface {
	RegisterService(rcvr interface{}, inputCh chan ReqMsg, fn func(svcName string) map[string]interface{}) error
	RegisterMethod(serviceName string, methodName string, fn RpcMethodFunc) error
}

type serviceManager struct {
	RegisteredServices map[string]*service
}

func NewServiceManager() ServiceManager {
	manager := &serviceManager{}
	manager.RegisteredServices = make(map[string]*service)
	return manager
}

func (m *serviceManager) RegisterService(rcvr interface{}, inputCh chan ReqMsg, fn func(svcName string) map[string]interface{}) error {
	receiver := reflect.ValueOf(rcvr)
	if receiver.Kind() != reflect.Ptr {
		slog.Errorf("[rpcProxy.manager.serviceManager.RegisterService]: receiver expected kind %v, actual kind %v", reflect.Ptr, receiver.Kind())
		return &ErrReceiverKind{
			kind: receiver.Kind(),
		}
	}
	serviceName := reflect.TypeOf(rcvr).Name()
	if _, ok := m.RegisteredServices[serviceName]; !ok {
		m.RegisteredServices[serviceName] = &service{
			name:              serviceName,
			typ:               reflect.TypeOf(rcvr),
			rcvr:              receiver,
			inputCh:           inputCh,
			checkpointFn:      fn,
			checkpointCh:      make(chan string, 1),
			receivedMethod:    utils.NewStringSet(),
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
	inputCh           chan ReqMsg
	running           int32
	checkpointFn      func(svcName string) map[string]interface{}
	methodSet         utils.StringSet
	checkpointCh      chan string
	receivedMethod    utils.StringSet
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
		atomic.StoreInt32(&svc.running, 1) // maybe not available
	}()
	for _, method := range svc.registeredMethods {
		go method.StartProcess(svc.name)
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
		if svc.receivedMethod.Has(methodName) {
			continue
		}
		svc.receivedMethod.Insert(methodName)
		if svc.receivedMethod.Len() == svc.methodSet.Len() {
			checkpoint := svc.checkpointFn(svc.name)
			slog.Infof("service %s take checkpoint successfully, checkpoint content: %v", svc.name, checkpoint)
			svc.receivedMethod = utils.NewStringSet()
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
		Name:          methodName,
		Fn:            fn,
		InputCh:       inputCh,
		OutputCh:      outputCh,
		CheckpointCh:  svc.checkpointCh,
		ClientManager: NewClientManager(outputCh),
		SourceSet:     utils.NewStringSet(),
		TargetSet:     utils.NewStringSet(),
		BarrierSet:    utils.NewStringSet(),
	}
	slog.Infof("[rpcProxy.manager.service.registerMethod]: service [%v] register method [%v] successfully", svc.name, methodName)
	return nil
}

type RpcMethodFunc func(arg MethodArgs) MethodResult

type RpcMethod struct {
	Name          string
	Fn            RpcMethodFunc
	InputCh       chan ReqMsg
	OutputCh      chan proto.Message
	ClientManager ClientManager
	SourceSet     utils.StringSet
	TargetSet     utils.StringSet
	BarrierSet    utils.StringSet
	CheckpointCh  chan string
}

func (method *RpcMethod) StartProcess(serviceName string) {
	for reqMsg := range method.InputCh {
		switch reqMsg.Type {
		case AsyncReq:
			source := reqMsg.Source
			if !method.SourceSet.Has(source) {
				method.SourceSet.Insert(source)
			}
			args := MethodArgs{}

			if err := utils.Unmarshal(reqMsg.Args, args.Message); err != nil {
				continue
			}

			result := method.Fn(args)

			data, err := utils.Marshal(result.Message)
			if err != nil {
				slog.Errorf("marshal message %v failed, err: %v", result.Message, err)
				continue
			}
			if result.IsRequest {
				req := &pb.AsyncCallRequest{
					Id:     reqMsg.Id,
					Msg:    data,
					Source: fmt.Sprintf("%s.%s", serviceName, method.Name),
				}
				if result.Target != "" {
					req.Target = result.Target
					req.Callback = result.Callback
				} else {
					req.Target = reqMsg.Callback
					req.Callback = ""
				}

				if !method.TargetSet.Has(req.Target) {
					method.TargetSet.Insert(req.Target)
				}

				method.OutputCh <- req
			} else {
				replyMsg := ReplyMsg{
					Ok:    true,
					Reply: data,
				}
				reqMsg.ReplyCh <- replyMsg
			}
		case CheckpointReq:
			source := reqMsg.Source
			if !method.BarrierSet.Has(source) {
				method.BarrierSet.Insert(source)
			}

			if method.BarrierSet.Equals(method.SourceSet) {
				for target := range method.TargetSet {
					req := &pb.InitCheckpointRequest{
						Source: fmt.Sprintf("%s.%s", serviceName, method.Name),
						Target: target,
					}
					method.OutputCh <- req
				}
				method.CheckpointCh <- method.Name
			}
		case RestoreReq:
		}

	}
}
