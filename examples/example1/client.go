package main

import (
	"context"
	"flag"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"stateful-service/pkg/rpcproxy"
	"stateful-service/proto/example1"
	"stateful-service/proto/pb"
	"stateful-service/slog"
	"sync"
	"time"
)

var (
	hostport    string
	method      string
	service     string
	opset       string
	interval    int
	runningTime int
	epoch       int
	clientNum   int
)

func init() {
	flag.StringVar(&hostport, "hostport", "127.0.0.1:8080", "service hostport")
	flag.StringVar(&service, "service", "", "service name")
	flag.StringVar(&method, "method", "", "the name of calling method")
	flag.StringVar(&opset, "opset", "call", "operation set")
	flag.IntVar(&interval, "interval", 2, "request interval(s)")
	flag.IntVar(&runningTime, "runningTime", -1, "client expect running time(s)")
	flag.IntVar(&epoch, "epoch", 1, "running epoch count")
	flag.IntVar(&clientNum, "clientNum", 1, "the number of client to run")
}

type Client struct {
	id          int
	target      string
	interval    int
	runningTime time.Duration
	svcMeth     string
	epoch       int
	epochCnt    int
	startTime   time.Time
	rpcClient   rpcproxy.RpcClient
}

func (cli *Client) start() {
	slog.Infof("[client %v] start", cli.id)
	var err error
	cli.rpcClient, err = rpcproxy.NewRpcClient(cli.target)
	if err != nil {
		slog.Errorf("[client %v] create client for target %s failed, err: %v", cli.id, cli.target, err)
		return
	}
	cli.epochCnt = 0
	cli.startTime = time.Now()
	for cli.keepRunning() {
		slog.Infof("[client %v] start epoch %v call", cli.id, cli.epochCnt)
		cli.epochCnt++
		addNumReq := &example1.AddNumRequest{
			Num: int32(cli.epochCnt),
		}
		req, _ := proto.Marshal(addNumReq)
		startTime := time.Now()
		result, err := cli.rpcClient.CallRpc(cli.svcMeth, req)
		if err != nil {
			slog.Errorf("[client %v] call rpc target %v failed, err: %v", cli.id, cli.target, err)
			continue
		}
		resp := &example1.AddNumReponse{}
		proto.Unmarshal(result, resp)
		slog.Infof("[client %v] call rpc epoch %v success, result %v, costTime: %v", cli.id, cli.epochCnt, resp.Result, time.Since(startTime))
		<-time.Tick(time.Duration(cli.interval) * time.Second)
	}
	cli.rpcClient.Close()
	slog.Infof("[client %v] complete", cli.id)
}

func (cli *Client) keepRunning() bool {
	if cli.runningTime != time.Duration(-1)*time.Second {
		return time.Since(cli.startTime) < cli.runningTime
	}
	return cli.epochCnt < cli.epoch
}

func callRpc() {
	clients := make([]*Client, clientNum)
	for i := range clients {
		clients[i] = &Client{
			id:          i,
			target:      hostport,
			svcMeth:     service + "." + method,
			epoch:       epoch,
			interval:    interval,
			runningTime: time.Duration(runningTime) * time.Second,
		}
	}
	slog.Infof("init %v clients", len(clients))

	startTime := time.Now()
	wg := &sync.WaitGroup{}
	slog.Infof("start %v clients", len(clients))
	for _, client := range clients {
		wg.Add(1)
		go func(client *Client) {
			defer wg.Done()
			client.start()
		}(client)
	}
	wg.Wait()
	costTime := time.Since(startTime)
	slog.Infof("%v clients complete, cost time: %v", len(clients), costTime)
}

func checkpoint(svcMeth string) {
	conn, err := grpc.Dial(hostport, grpc.WithInsecure())
	if err != nil {
		slog.Errorf("connect to target %v failed, err: %v", hostport, err)
		return
	}
	client := pb.NewRpcProxyClient(conn)
	initReq := &pb.InitCheckpointRequest{
		Source:     "client",
		SourceHost: "",
		Target:     svcMeth,
		TargetHost: hostport,
	}
	startTime := time.Now()
	initResp, err := client.InitCheckpoint(context.Background(), initReq)
	if err != nil {
		slog.Errorf("call initCheckpoint failed, err: %v", err)
	}
	id := initResp.Id

	waitReq := &pb.WaitCheckpointRequest{
		Id: id,
	}
	_, err = client.WaitCheckpoint(context.Background(), waitReq)
	slog.Infof("checkpoint complete, cost time: %v", time.Since(startTime))
}

func main() {
	flag.Parse()
	if service == "" {
		slog.Error("service name must be specified")
		return
	}
	if method == "" {
		slog.Error("method name must be specified")
		return
	}
	if opset == "checkpoint" {
		checkpoint(service + "." + method)
	} else {
		callRpc()
	}

}
