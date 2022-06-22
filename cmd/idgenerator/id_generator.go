package main

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"stateful-service/config"
	"stateful-service/pkg/idgenerator"
	pb "stateful-service/proto/pb"
	"stateful-service/slog"
)

func main() {
	runningPort := config.GetIntEnv(config.EnvRunningPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", runningPort))
	if err != nil {
		slog.Fatalf("failed to listen: %v", err)
	}
	generator := idgenerator.NewIdGenerator()
	gServer := grpc.NewServer()
	pb.RegisterIdGeneratorServer(gServer, generator)
	slog.Infof("id generator listening at %v", lis.Addr())
	if err := gServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
