package main

import (
	"flag"
	"github.com/golang/protobuf/proto"
	"stateful-service/pkg/rpcproxy"
	"stateful-service/proto/example1"
	"stateful-service/slog"
	"sync"
	"time"
)

var (
	hostname    string
	method      string
	interval    int
	runningTime int
	epoch       int
	clientNum   int
)

func init() {
	flag.StringVar(&hostname, "hostname", "127.0.0.1:8080", "service hostname")
	flag.StringVar(&method, "method", "", "the name of calling method")
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
	epoch       int
	epochCnt    int
	startTime   time.Time
	rpcClient   rpcproxy.RpcClient
}

func (cli *Client) start() {
	var err error
	cli.rpcClient, err = rpcproxy.NewRpcClient(cli.target)
	if err != nil {
		slog.Errorf("create client for target %s failed, err: %v", cli.target, err)
		return
	}
	cli.epochCnt = 0
	cli.startTime = time.Now()
	for cli.keepRunning() {
		cli.epochCnt++
		addNumReq := &example1.AddNumRequest{
			Num: int32(cli.epochCnt),
		}
		req, _ := proto.Marshal(addNumReq)
		result, err := cli.rpcClient.CallRpc(cli.target, req)
		if err != nil {
			slog.Errorf("call rpc target %v failed, err: %v", cli.target, err)
			continue
		}
		resp := &example1.AddNumReponse{}
		proto.Unmarshal(result, resp)
		slog.Infof("call rpc epoch %v success, result %v", cli.epochCnt, resp.Result)
		<- time.Tick(time.Duration(cli.interval) * time.Second)
	}
}

func (cli *Client) keepRunning() bool {
	if cli.runningTime != -1 {
		return time.Since(cli.startTime) < cli.runningTime
	}
	return cli.epochCnt < cli.epoch
}

func main() {
	flag.Parse()
	if method == "" {
		slog.Error("method name must be specified")
		return
	}

	clients := make([]*Client, clientNum)
	for i := range clients {
		clients[i] = &Client{
			id:          i,
			target:      hostname + "/" + method,
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
		go func() {
			defer wg.Done()
			client.start()
		}()
	}
	wg.Wait()
	costTime := time.Since(startTime)
	slog.Infof("%v clients complete, cost time: %v", costTime)
}

