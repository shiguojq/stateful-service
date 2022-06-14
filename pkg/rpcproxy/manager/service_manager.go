package manager

import (
	"reflect"
	"stateful-service/slog"
	"stateful-service/utils"
	"sync/atomic"
)

type ServiceManager interface {
	RegisterService(rcvr interface{}, inputCh chan ReqMsg) error
	RegisterMethod(serviceName string, methodName string) error
	RegisterMethods(serviceName string, methodNames ...string) error
}

type serviceManager struct {
	RegisteredServices map[string]*service
}

func NewServiceManager() ServiceManager {
	manager := &serviceManager{}
	manager.RegisteredServices = make(map[string]*service)
	return manager
}

func (m *serviceManager) RegisterService(rcvr interface{}, inputCh chan ReqMsg) error {
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
			registeredMethods: make(map[string]reflect.Method),
		}
	}
	slog.Infof("[rpcProxy.manager.serviceManager.RegisterService]: register [%v] service successfully", serviceName)
	return nil
}

func (m *serviceManager) RegisterMethod(serviceName string, methodName string) error {
	return m.RegisteredServices[serviceName].registerMethod(methodName)
}

func (m *serviceManager) RegisterMethods(serviceName string, methodNames ...string) error {
	return m.RegisteredServices[serviceName].registerMethods(methodNames...)
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
	registeredMethods map[string]reflect.Method
	methodCh          map[string]chan ReqMsg
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
	for methodName, ch := range svc.methodCh {
		svc.Process(methodName, svc.registeredMethods[methodName], ch)
	}
	go func() {
		for msg := range svc.inputCh {
			methodName := msg.MethodName
			svc.methodCh[methodName] <- msg

		}
	}()
}

func (svc *service) registerMethod(methodName string) error {
	method, ok := svc.typ.MethodByName(methodName)
	if !ok {
		return &ErrMethodNotFound{
			clsName:    svc.name,
			methodName: methodName,
		}
	}
	if _, ok := svc.registeredMethods[methodName]; ok {
		return nil
	}
	svc.registeredMethods[methodName] = method
	svc.methodCh[methodName] = make(chan ReqMsg, 1024) // add channel buffer to avoid overhead
	slog.Infof("[rpcProxy.manager.service.registerMethod]: service [%v] register method [%v] successfully", svc.name, methodName)
	return nil
}

func (svc *service) registerMethods(methodNames ...string) error {
	for _, methodName := range methodNames {
		err := svc.registerMethod(methodName)
		if err != nil {
			return err
		}
	}
	slog.Infof("[rpcProxy.manager.service.registerMethod]: service [%v] register method %v successfully", svc.name, methodNames)
	return nil
}

func (svc *service) Process(methodName string, method reflect.Method, inputCh chan ReqMsg) {
	for reqMsg := range inputCh {
		replyMsg := ReplyMsg{}

		reqType := method.Type.In(1).Elem()
		req := reflect.New(reqType)
		if err := utils.Unmarshal(reqMsg.Args, req); err != nil {
			replyMsg.Ok = false
			reqMsg.ReplyCh <- replyMsg
			continue
		}

		replyType := method.Type.In(2).Elem()
		reply := reflect.New(replyType)

		function := method.Func
		function.Call([]reflect.Value{svc.rcvr, req.Elem(), reply})

		data, err := utils.Marshal(reply)
		if err != nil {
			replyMsg.Ok = false
			reqMsg.ReplyCh <- replyMsg
			continue
		}
		replyMsg.Ok = true
		replyMsg.Reply = data
		reqMsg.ReplyCh <- replyMsg
	}
}
