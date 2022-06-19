package utils

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"strings"
)

func Marshal(v interface{}) ([]byte, error) {
	vv, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to marshal, message is %T, want proto.Message", v)
	}
	return proto.Marshal(vv)
}

func Unmarshal(data []byte, v interface{}) error {
	vv, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("failed to unmarshal, message is %T, want proto.Message", v)
	}
	return proto.Unmarshal(data, vv)
}

func ParseTarget(target string) (string, string) {
	dot := strings.LastIndex(target, ".")
	serviceName := target[:dot]
	methodName := target[dot+1:]
	return serviceName, methodName
}
