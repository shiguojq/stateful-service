package idgenerator

import (
	"context"
	"fmt"
	pb "stateful-service/proto"
	"testing"
)

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
