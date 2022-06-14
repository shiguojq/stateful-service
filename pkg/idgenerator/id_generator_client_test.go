package idgenerator

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"stateful-service/config"
	pb "stateful-service/proto"
	"testing"
	"time"
)

func TestIdGeneratorClient(t *testing.T) {
	runningPort := config.GetIntEnv(config.EnvRunningPort)
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%v", runningPort), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	cli := pb.NewIdGeneratorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	idSet := make(map[int64]struct{})
	for i := 0; i < 1024; i++ {
		r, err := cli.GenerateId(ctx, &pb.GenerateIdRequest{SvcName: "test"})
		if err != nil {
			t.Fatalf("could not greet: %v", err)
		}
		if _, ok := idSet[r.Id]; ok {
			panic(fmt.Sprintf("id %v has been generated before", r.Id))
		}
		idSet[r.Id] = struct{}{}
	}
	fmt.Println(len(idSet))
}
