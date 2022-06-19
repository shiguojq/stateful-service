package manager

import "github.com/golang/protobuf/proto"

type MethodResult struct {
	Message   proto.Message
	IsRequest bool
	Target    string
	Callback  string
}
