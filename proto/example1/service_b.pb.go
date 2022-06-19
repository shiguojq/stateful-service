// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        (unknown)
// source: proto/example1/service_b.proto

package example1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StoreNumRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Num int32 `protobuf:"varint,1,opt,name=num,proto3" json:"num,omitempty"`
}

func (x *StoreNumRequest) Reset() {
	*x = StoreNumRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_example1_service_b_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StoreNumRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StoreNumRequest) ProtoMessage() {}

func (x *StoreNumRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_example1_service_b_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StoreNumRequest.ProtoReflect.Descriptor instead.
func (*StoreNumRequest) Descriptor() ([]byte, []int) {
	return file_proto_example1_service_b_proto_rawDescGZIP(), []int{0}
}

func (x *StoreNumRequest) GetNum() int32 {
	if x != nil {
		return x.Num
	}
	return 0
}

type StoreNumReponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result int32 `protobuf:"varint,1,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *StoreNumReponse) Reset() {
	*x = StoreNumReponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_example1_service_b_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StoreNumReponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StoreNumReponse) ProtoMessage() {}

func (x *StoreNumReponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_example1_service_b_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StoreNumReponse.ProtoReflect.Descriptor instead.
func (*StoreNumReponse) Descriptor() ([]byte, []int) {
	return file_proto_example1_service_b_proto_rawDescGZIP(), []int{1}
}

func (x *StoreNumReponse) GetResult() int32 {
	if x != nil {
		return x.Result
	}
	return 0
}

var File_proto_example1_service_b_proto protoreflect.FileDescriptor

var file_proto_example1_service_b_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x31,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x08, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x31, 0x22, 0x23, 0x0a, 0x0f, 0x53, 0x74,
	0x6f, 0x72, 0x65, 0x4e, 0x75, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a,
	0x03, 0x6e, 0x75, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x6e, 0x75, 0x6d, 0x22,
	0x29, 0x0a, 0x0f, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x4e, 0x75, 0x6d, 0x52, 0x65, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x42, 0x12, 0x5a, 0x10, 0x2e, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x31, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_example1_service_b_proto_rawDescOnce sync.Once
	file_proto_example1_service_b_proto_rawDescData = file_proto_example1_service_b_proto_rawDesc
)

func file_proto_example1_service_b_proto_rawDescGZIP() []byte {
	file_proto_example1_service_b_proto_rawDescOnce.Do(func() {
		file_proto_example1_service_b_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_example1_service_b_proto_rawDescData)
	})
	return file_proto_example1_service_b_proto_rawDescData
}

var file_proto_example1_service_b_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_example1_service_b_proto_goTypes = []interface{}{
	(*StoreNumRequest)(nil), // 0: example1.StoreNumRequest
	(*StoreNumReponse)(nil), // 1: example1.StoreNumReponse
}
var file_proto_example1_service_b_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_example1_service_b_proto_init() }
func file_proto_example1_service_b_proto_init() {
	if File_proto_example1_service_b_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_example1_service_b_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StoreNumRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_example1_service_b_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StoreNumReponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_example1_service_b_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_example1_service_b_proto_goTypes,
		DependencyIndexes: file_proto_example1_service_b_proto_depIdxs,
		MessageInfos:      file_proto_example1_service_b_proto_msgTypes,
	}.Build()
	File_proto_example1_service_b_proto = out.File
	file_proto_example1_service_b_proto_rawDesc = nil
	file_proto_example1_service_b_proto_goTypes = nil
	file_proto_example1_service_b_proto_depIdxs = nil
}
