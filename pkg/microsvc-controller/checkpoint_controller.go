package microsvccontroller

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	v1 "stateful-service/api/v1"
	"stateful-service/config"
	pb "stateful-service/proto"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type CheckpointController interface {
	Start()
}

type checkpointController struct {
	checkpointInterval int
	dead               int32
	reconclier         *MicroServiceReconciler
}

func NewCheckpointController(reconclier *MicroServiceReconciler) CheckpointController {
	cc := &checkpointController{}
	cc.checkpointInterval = config.GetIntEnv(config.EnvCheckpointInterval)
	cc.reconclier = reconclier
	return cc
}

func (cc *checkpointController) Start() {
	clog := cc.reconclier.Log.WithName("checkpointController")
	for atomic.LoadInt32(&cc.dead) != 1 {
		clog.Info("try to start checkpoint")
		go cc.Checkpoint()
		<-time.Tick(time.Duration(cc.checkpointInterval) * time.Second)
	}
}

func (cc *checkpointController) Checkpoint() {
	clog := cc.reconclier.Log.WithName("start checkpoint")
	microServiceList := &v1.MicroServiceList{}
	if err := cc.reconclier.List(context.Background(), microServiceList); err != nil {
		clog.Error(err, "get microServiceList failed")
		return
	}
	wg := &sync.WaitGroup{}
	clog.Info("init checkpoint for start microservies", "count", len(microServiceList.Items), "microservice", microServiceList.Items)
	succCh := make(chan string, len(microServiceList.Items))
	failCh := make(chan string, len(microServiceList.Items))
	for _, microService := range microServiceList.Items {
		if microService.Spec.StartPoint {
			wg.Add(1)
			go func() {
				clog := cc.reconclier.Log.WithName("CheckpointStarter")
				defer func() {
					wg.Done()
				}()
				target := fmt.Sprintf("%s:%v", microService.Name, microService.Spec.Port)
				initReq := &pb.InitCheckpointRequest{
					Target: target,
					Source: config.ControllerSourceString,
				}
				clog.Info("start init connect", "target", target)
				conn, err := grpc.Dial(target)
				if err != nil {
					clog.Error(err, "init connect success", "target", target)
					failCh <- microService.Name
					return
				}
				clog.Info("init connect success", "target", target)

				client := pb.NewRpcProxyClient(conn)
				clog.Info("start init checkpoint", "target", target)
				initResp, err := client.InitCheckpoint(context.Background(), initReq)
				if err != nil {
					clog.Error(err, "init checkpoint failed", "target", target)
					failCh <- microService.Name
					return
				}
				clog.Info("init checkpoint success", "target", target, "checkpoint id", initResp.Id)

				waitReq := &pb.WaitCheckpointRequest{
					Id: initResp.Id,
				}
				clog.Info("start wait checkpoint", "target", target)
				waitResp, err := client.WaitCheckpoint(context.Background(), waitReq)
				if err != nil || !waitResp.Ok {
					clog.Error(err, "wait checkpoint failed", target, "target")
					failCh <- microService.Name
					return
				}
				succCh <- microService.Name
			}()
		}
	}
	wg.Done()

	close(succCh)
	close(failCh)
	result := &strings.Builder{}
	result.WriteString("checkpoint result: \n" +
		"success list: \n")
	succCnt := 0
	for microName := range succCh {
		succCnt++
		result.WriteString(microName)
		result.WriteByte('\n')
	}
	result.WriteString(fmt.Sprintf("success total: %v\n", succCnt))
	result.WriteString("fail list: \n")
	failCnt := 0
	for microName := range failCh {
		failCnt++
		result.WriteString(microName)
		result.WriteByte('\n')
	}
	result.WriteString(fmt.Sprintf("fail total: %v\n", failCnt))
	clog.Info(result.String())
}

func (cc *checkpointController) Kill() {
	atomic.StoreInt32(&cc.dead, 1)
}
