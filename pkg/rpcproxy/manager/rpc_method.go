package manager

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"stateful-service/proto/pb"
	"stateful-service/slog"
	"stateful-service/utils"
	"sync/atomic"
)

type RpcMethodFunc func(arg MethodArgs) MethodResult

type RpcMethod struct {
	ServiceName    string
	Name           string
	Fn             RpcMethodFunc
	SourceHost     string
	InputCh        chan ReqMsg
	OutputCh       chan proto.Message
	ClientManager  ClientManager
	SourceSet      utils.StringSet
	SourceBlocked  map[string]bool
	TargetSet      utils.StringSet
	BarrierSet     utils.StringSet
	TargetHosts    map[string]string
	CheckpointCh   chan string
	CheckpointMode CheckpointMode
	BlockCh        chan struct{}
	WaitingCh      chan ReqMsg
	blocking       int32
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
			atomic.StoreInt32(&method.blocking, 1)
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
	case BlockReq:
		method.WaitingCh = make(chan ReqMsg, 1024)
		method.waitCheckpoint()
	case RestoreReq:
	}
}

func (method *RpcMethod) waitCheckpoint() {
	<-method.BlockCh
	atomic.StoreInt32(&method.blocking, 0)
	close(method.WaitingCh)
	for source := range method.SourceBlocked {
		method.SourceBlocked[source] = false
	}
	method.BarrierSet = utils.NewStringSet()
}
