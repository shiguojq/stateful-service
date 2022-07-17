package manager

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"reflect"
	"stateful-service/slog"
	"stateful-service/utils"
	"sync"
	"sync/atomic"
)

type ServiceManager interface {
	RunService(serviceName string) error
	RegisterService(rcvr interface{}, inputCh chan ReqMsg, fn CheckpointFn, mode CheckpointMode) error
	RegisterMethod(serviceName string, methodName string, fn RpcMethodFunc, callback bool) error
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

func (m *serviceManager) RegisterService(rcvr interface{}, inputCh chan ReqMsg, fn CheckpointFn, mode CheckpointMode) error {
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
			checkpointMode:    mode,
			blockedMethod:     utils.NewStringSet(),
			methodSet:         utils.NewStringSet(),
			callMethod:        utils.NewStringSet(),
			callbackMethod:    utils.NewStringSet(),
			registeredMethods: make(map[string]*RpcMethod),
		}
	}
	slog.Infof("[rpcProxy.manager.serviceManager.RegisterService]: register [%v] service successfully", serviceName)
	return nil
}

func (m *serviceManager) RegisterMethod(serviceName string, methodName string, fn RpcMethodFunc, callback bool) error {
	return m.RegisteredServices[serviceName].registerMethod(methodName, fn, callback)
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
	// meta info
	name       string
	typ        reflect.Type
	rcvr       reflect.Value
	inputCh    chan ReqMsg
	sourceHost string
	// running state
	running     int32
	blocking    int32
	logging     int32
	recordCount int32
	// checkpoint state
	checkpointFn      CheckpointFn
	checkpointCh      chan string
	checkpointMode    CheckpointMode
	checkpointReplyCh chan ReplyMsg
	// method set
	methodSet         utils.StringSet //full method set
	blockedMethod     utils.StringSet //blocked method set
	callMethod        utils.StringSet //direct call method set {full set} - {call set} = {callback set}
	callbackMethod    utils.StringSet //callback method set
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
			slog.Infof("[RpcService %v] msg %+v start wait", svc.name, msg)
			svc.waitBlock()
			if msg.Type == CheckpointReq && svc.checkpointReplyCh == nil {
				svc.checkpointReplyCh = msg.ReplyCh
			}
			if atomic.LoadInt32(&svc.logging) == 1 {
				atomic.AddInt32(&svc.recordCount, 1)
			}
			slog.Infof("[RpcService %v] get msg %+v", svc.name, msg)
			methodName := msg.MethodName
			svc.registeredMethods[methodName].InputCh <- msg
		}
	}()
	switch svc.checkpointMode {
	case MicroServiceCheckpoint:
		{
			go svc.waitMicroServiceCheckpoint()
		}
	case ChandyLamportCheckpoint:
		{
			go svc.waitLamportCheckpoint()
		}
	case FlinkCheckpoint:
		{
			go svc.waitFlinkCheckpoint()
		}
	}

}

func (svc *service) waitBlock() {
	switch svc.checkpointMode {
	case MicroServiceCheckpoint:
		return
	case FlinkCheckpoint:
		for atomic.LoadInt32(&svc.blocking) == 1 {
		}
	}
}

func (svc *service) waitMicroServiceCheckpoint() {
	for methodName := range svc.checkpointCh {
		if svc.blockedMethod.Has(methodName) {
			continue
		}
		svc.blockedMethod.Insert(methodName)
		if svc.blockedMethod.Len() == svc.methodSet.Len() {
			checkpoint := svc.checkpointFn(svc.name)
			slog.Infof("service %s take checkpoint successfully, checkpoint content: %v", svc.name, checkpoint)
			svc.blockedMethod = utils.NewStringSet()
			for _, method := range svc.registeredMethods {
				method.BlockCh <- struct{}{}
			}
		}
	}
}

func (svc *service) waitLamportCheckpoint() {

}

func (svc *service) waitFlinkCheckpoint() {
	for methodName := range svc.checkpointCh {
		if svc.blockedMethod.Has(methodName) {
			continue
		}
		svc.blockedMethod.Insert(methodName)
		if atomic.LoadInt32(&svc.blocking) == 0 &&
			atomic.LoadInt32(&svc.logging) == 0 &&
			svc.callMethod.Less(svc.blockedMethod) {
			atomic.StoreInt32(&svc.blocking, 1)
			wg := sync.WaitGroup{}
			for callbackMethod := range svc.callbackMethod {
				if !svc.blockedMethod.Has(callbackMethod) {
					wg.Add(1)
					go func() {
						defer wg.Done()
						svc.registeredMethods[callbackMethod].InputCh <- ReqMsg{
							Type: BlockReq,
						}
					}()
				}
			}
			wg.Wait()
			checkpoint := svc.checkpointFn(svc.name)
			atomic.StoreInt32(&svc.logging, 1)
			atomic.StoreInt32(&svc.recordCount, 0)
			slog.Infof("[RpcService %v] take checkpoint successfully, checkpoint content: %v", svc.name, checkpoint)
			for _, method := range svc.registeredMethods {
				if !svc.blockedMethod.Has(method.Name) {
					slog.Infof("[RpcService %v] start method %v", svc.name, method.Name)
					method.BlockCh <- struct{}{}
					slog.Infof("[RpcService %v] start method %v success", svc.name, method.Name)
				}
			}
			atomic.StoreInt32(&svc.blocking, 0)
		}
		if atomic.LoadInt32(&svc.logging) == 1 && svc.blockedMethod.Equals(svc.methodSet) {
			svc.blockedMethod = utils.NewStringSet()
			count := atomic.LoadInt32(&svc.recordCount)
			atomic.StoreInt32(&svc.recordCount, 0)
			atomic.StoreInt32(&svc.logging, 0)
			svc.checkpointReplyCh <- ReplyMsg{
				Ok:    true,
				Reply: nil,
			}
			svc.checkpointReplyCh = nil
			slog.Infof("[RpcService %v] record request message %v", svc.name, count)
		}
	}
}

func (svc *service) registerMethod(methodName string, fn RpcMethodFunc, callback bool) error {
	if svc.methodSet.Has(methodName) {
		return nil
	}
	svc.methodSet.Insert(methodName)
	if svc.checkpointMode == FlinkCheckpoint {
		if callback {
			svc.callbackMethod.Insert(methodName)
		} else {
			svc.callMethod.Insert(methodName)
		}
	}
	inputCh := make(chan ReqMsg, 1024)
	outputCh := make(chan proto.Message, 1)
	svc.registeredMethods[methodName] = &RpcMethod{
		ServiceName:    svc.name,
		Name:           methodName,
		Fn:             fn,
		SourceHost:     svc.sourceHost,
		InputCh:        inputCh,
		OutputCh:       outputCh,
		CheckpointCh:   svc.checkpointCh,
		CheckpointMode: svc.checkpointMode,
		BlockCh:        make(chan struct{}),
		ClientManager:  NewClientManager(fmt.Sprintf("%v.%v", svc.name, methodName), outputCh),
		SourceSet:      utils.NewStringSet(),
		SourceBlocked:  make(map[string]bool),
		TargetSet:      utils.NewStringSet(),
		BarrierSet:     utils.NewStringSet(),
		TargetHosts:    make(map[string]string),
	}
	slog.Infof("[rpcProxy.manager.service.registerMethod]: service [%v] register method [%v] successfully", svc.name, methodName)
	return nil
}
