package idgenerator

import (
	"context"
	"stateful-service/config"
	pb "stateful-service/proto"
	"stateful-service/utils"
	"time"
)

const (
	TIMESTAMP_BITS       = int64(39)
	TIMESTAMP_MASK       = -1 ^ (-1 << TIMESTAMP_BITS)
	POD_IP_BITS          = int64(16)
	POD_IP_MASK          = -1 ^ (-1 << POD_IP_BITS)
	POD_IP_RIGHT_SHIFT   = int64(8)
	POD_IP_LEFT_SHIFT    = SEQUENCE_BITS
	SEQUENCE_BITS        = int64(8)
	SEQUENCE_MASK        = -1 ^ (-1 << SEQUENCE_BITS)
	TIMESTAMP_LEFT_SHIFT = SEQUENCE_BITS + POD_IP_BITS
)

type IdGenerator struct {
	pb.UnimplementedIdGeneratorServer
	startTime     int64
	podIp         int32
	sequence      int64
	lastTimeMills int64
	svcNames      utils.StringSet
}

func NewIdGenerator() *IdGenerator {
	_, ip := config.GetIp()
	generator := &IdGenerator{
		podIp:         ip,
		sequence:      0,
		startTime:     time.Now().UnixMilli(),
		lastTimeMills: time.Now().UnixMilli(),
		svcNames:      utils.NewStringSet(),
	}
	return generator
}

func (gen *IdGenerator) GenerateId(_ context.Context, req *pb.GenerateIdRequest) (*pb.GenerateIdResponse, error) {
	timestamp := gen.waitNextMillis()

	if timestamp == gen.lastTimeMills {
		gen.sequence = (gen.sequence + 1) & SEQUENCE_MASK
		if gen.sequence == 0 {
			timestamp = gen.waitNextMillis()
		}
	} else {
		gen.sequence = 0
	}

	gen.lastTimeMills = timestamp
	id := ((timestamp & TIMESTAMP_MASK) << TIMESTAMP_LEFT_SHIFT) | (((int64(gen.podIp) >> POD_IP_RIGHT_SHIFT) & POD_IP_MASK) << POD_IP_LEFT_SHIFT) | gen.sequence
	resp := &pb.GenerateIdResponse{
		Id: id,
	}
	if _, ok := gen.svcNames[req.SvcName]; !ok {
		gen.svcNames.Insert(req.SvcName)
		//Todo(jiayk) report new direct call to controller
	}
	return resp, nil
}

func (gen *IdGenerator) waitNextMillis() int64 {
	timestamp := time.Now().UnixMilli()
	for timestamp < gen.lastTimeMills {
		timestamp = time.Now().UnixMilli()
	}
	return timestamp
}
