// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.1
// source: mempool.proto

package network_pb

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

type MempoolTxRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start []byte `protobuf:"bytes,1,opt,name=start,proto3" json:"start,omitempty"`
}

func (x *MempoolTxRequest) Reset() {
	*x = MempoolTxRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mempool_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MempoolTxRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MempoolTxRequest) ProtoMessage() {}

func (x *MempoolTxRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mempool_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MempoolTxRequest.ProtoReflect.Descriptor instead.
func (*MempoolTxRequest) Descriptor() ([]byte, []int) {
	return file_mempool_proto_rawDescGZIP(), []int{0}
}

func (x *MempoolTxRequest) GetStart() []byte {
	if x != nil {
		return x.Start
	}
	return nil
}

type MempoolTxResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Txs []*MempoolTx `protobuf:"bytes,1,rep,name=txs,proto3" json:"txs,omitempty"`
}

func (x *MempoolTxResponse) Reset() {
	*x = MempoolTxResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mempool_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MempoolTxResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MempoolTxResponse) ProtoMessage() {}

func (x *MempoolTxResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mempool_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MempoolTxResponse.ProtoReflect.Descriptor instead.
func (*MempoolTxResponse) Descriptor() ([]byte, []int) {
	return file_mempool_proto_rawDescGZIP(), []int{1}
}

func (x *MempoolTxResponse) GetTxs() []*MempoolTx {
	if x != nil {
		return x.Txs
	}
	return nil
}

type MempoolTx struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tx []byte `protobuf:"bytes,1,opt,name=tx,proto3" json:"tx,omitempty"`
}

func (x *MempoolTx) Reset() {
	*x = MempoolTx{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mempool_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MempoolTx) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MempoolTx) ProtoMessage() {}

func (x *MempoolTx) ProtoReflect() protoreflect.Message {
	mi := &file_mempool_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MempoolTx.ProtoReflect.Descriptor instead.
func (*MempoolTx) Descriptor() ([]byte, []int) {
	return file_mempool_proto_rawDescGZIP(), []int{2}
}

func (x *MempoolTx) GetTx() []byte {
	if x != nil {
		return x.Tx
	}
	return nil
}

var File_mempool_proto protoreflect.FileDescriptor

var file_mempool_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6d, 0x65, 0x6d, 0x70, 0x6f, 0x6f, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0a, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x70, 0x62, 0x22, 0x28, 0x0a, 0x10, 0x4d,
	0x65, 0x6d, 0x70, 0x6f, 0x6f, 0x6c, 0x54, 0x78, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05,
	0x73, 0x74, 0x61, 0x72, 0x74, 0x22, 0x3c, 0x0a, 0x11, 0x4d, 0x65, 0x6d, 0x70, 0x6f, 0x6f, 0x6c,
	0x54, 0x78, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x27, 0x0a, 0x03, 0x74, 0x78,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72,
	0x6b, 0x5f, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x6d, 0x70, 0x6f, 0x6f, 0x6c, 0x54, 0x78, 0x52, 0x03,
	0x74, 0x78, 0x73, 0x22, 0x1b, 0x0a, 0x09, 0x4d, 0x65, 0x6d, 0x70, 0x6f, 0x6f, 0x6c, 0x54, 0x78,
	0x12, 0x0e, 0x0a, 0x02, 0x74, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x74, 0x78,
	0x42, 0x36, 0x5a, 0x34, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d,
	0x65, 0x6d, 0x6f, 0x63, 0x61, 0x73, 0x68, 0x2f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2f, 0x72, 0x65,
	0x66, 0x2f, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6e, 0x65,
	0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mempool_proto_rawDescOnce sync.Once
	file_mempool_proto_rawDescData = file_mempool_proto_rawDesc
)

func file_mempool_proto_rawDescGZIP() []byte {
	file_mempool_proto_rawDescOnce.Do(func() {
		file_mempool_proto_rawDescData = protoimpl.X.CompressGZIP(file_mempool_proto_rawDescData)
	})
	return file_mempool_proto_rawDescData
}

var file_mempool_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_mempool_proto_goTypes = []any{
	(*MempoolTxRequest)(nil),  // 0: network_pb.MempoolTxRequest
	(*MempoolTxResponse)(nil), // 1: network_pb.MempoolTxResponse
	(*MempoolTx)(nil),         // 2: network_pb.MempoolTx
}
var file_mempool_proto_depIdxs = []int32{
	2, // 0: network_pb.MempoolTxResponse.txs:type_name -> network_pb.MempoolTx
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_mempool_proto_init() }
func file_mempool_proto_init() {
	if File_mempool_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mempool_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*MempoolTxRequest); i {
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
		file_mempool_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*MempoolTxResponse); i {
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
		file_mempool_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*MempoolTx); i {
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
			RawDescriptor: file_mempool_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mempool_proto_goTypes,
		DependencyIndexes: file_mempool_proto_depIdxs,
		MessageInfos:      file_mempool_proto_msgTypes,
	}.Build()
	File_mempool_proto = out.File
	file_mempool_proto_rawDesc = nil
	file_mempool_proto_goTypes = nil
	file_mempool_proto_depIdxs = nil
}
