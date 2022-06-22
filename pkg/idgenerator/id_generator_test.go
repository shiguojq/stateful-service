package idgenerator

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "stateful-service/proto/pb"
	"stateful-service/slog"
	"testing"
	"time"
)

func callIdGenerator(target string) error {
	conn, err := grpc.Dial(fmt.Sprintf(target), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	cli := pb.NewIdGeneratorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	idSet := make(map[int64]struct{})
	for i := 0; i < 1024; i++ {
		r, err := cli.GenerateId(ctx, &pb.GenerateIdRequest{SvcName: "test"})
		if err != nil {
			return err
		}
		if _, ok := idSet[r.Id]; ok {
			panic(fmt.Sprintf("id %v has been generated before", r.Id))
		}
		idSet[r.Id] = struct{}{}
		fmt.Println(r.Id)
	}
	fmt.Println(len(idSet))
	return nil
}

func TestIdGenerator(t *testing.T) {
	gen := NewIdGenerator()
	for i := 0; i < 10; i++ {
		req := &pb.GenerateIdRequest{
			SvcName: "aaa",
		}
		resp, _ := gen.GenerateId(context.Background(), req)
		fmt.Printf("%d\n", resp.Id)
	}
}

func TestIdGeneratorServer(t *testing.T) {
	runningPort := 8080
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", runningPort))
	if err != nil {
		slog.Fatalf("failed to listen: %v", err)
	}
	generator := NewIdGenerator()
	gServer := grpc.NewServer()
	pb.RegisterIdGeneratorServer(gServer, generator)
	slog.Infof("id generator listening at %v", lis.Addr())
	go func() {
		if err := gServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	if err := callIdGenerator(fmt.Sprintf("127.0.0.1:%v", runningPort)); err != nil {
		t.Fatalf("call id generator failed, err: %v", err)
	}
}

func TestIdGeneratorClient(t *testing.T) {
	if err := callIdGenerator("10.102.113.6:8080"); err != nil {
		t.Fatalf("call id generator failed, err: %v", err)
	}
}
