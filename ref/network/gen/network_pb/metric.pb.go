// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.1
// source: metric.proto

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

type MetricRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      []byte   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Parents [][]byte `protobuf:"bytes,2,rep,name=parents,proto3" json:"parents,omitempty"`
}

func (x *MetricRequest) Reset() {
	*x = MetricRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_metric_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricRequest) ProtoMessage() {}

func (x *MetricRequest) ProtoReflect() protoreflect.Message {
	mi := &file_metric_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricRequest.ProtoReflect.Descriptor instead.
func (*MetricRequest) Descriptor() ([]byte, []int) {
	return file_metric_proto_rawDescGZIP(), []int{0}
}

func (x *MetricRequest) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *MetricRequest) GetParents() [][]byte {
	if x != nil {
		return x.Parents
	}
	return nil
}

type MetricResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Infos []*MetricInfo `protobuf:"bytes,1,rep,name=infos,proto3" json:"infos,omitempty"`
}

func (x *MetricResponse) Reset() {
	*x = MetricResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_metric_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricResponse) ProtoMessage() {}

func (x *MetricResponse) ProtoReflect() protoreflect.Message {
	mi := &file_metric_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricResponse.ProtoReflect.Descriptor instead.
func (*MetricResponse) Descriptor() ([]byte, []int) {
	return file_metric_proto_rawDescGZIP(), []int{1}
}

func (x *MetricResponse) GetInfos() []*MetricInfo {
	if x != nil {
		return x.Infos
	}
	return nil
}

type MetricInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Parent   []byte `protobuf:"bytes,2,opt,name=parent,proto3" json:"parent,omitempty"`
	Action   string `protobuf:"bytes,3,opt,name=action,proto3" json:"action,omitempty"`
	Order    int32  `protobuf:"varint,4,opt,name=order,proto3" json:"order,omitempty"`
	Count    int32  `protobuf:"varint,5,opt,name=count,proto3" json:"count,omitempty"`
	Start    int64  `protobuf:"varint,6,opt,name=start,proto3" json:"start,omitempty"`
	Duration int64  `protobuf:"varint,7,opt,name=duration,proto3" json:"duration,omitempty"`
}

func (x *MetricInfo) Reset() {
	*x = MetricInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_metric_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricInfo) ProtoMessage() {}

func (x *MetricInfo) ProtoReflect() protoreflect.Message {
	mi := &file_metric_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricInfo.ProtoReflect.Descriptor instead.
func (*MetricInfo) Descriptor() ([]byte, []int) {
	return file_metric_proto_rawDescGZIP(), []int{2}
}

func (x *MetricInfo) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *MetricInfo) GetParent() []byte {
	if x != nil {
		return x.Parent
	}
	return nil
}

func (x *MetricInfo) GetAction() string {
	if x != nil {
		return x.Action
	}
	return ""
}

func (x *MetricInfo) GetOrder() int32 {
	if x != nil {
		return x.Order
	}
	return 0
}

func (x *MetricInfo) GetCount() int32 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *MetricInfo) GetStart() int64 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *MetricInfo) GetDuration() int64 {
	if x != nil {
		return x.Duration
	}
	return 0
}

type MetricTimeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start int64 `protobuf:"varint,1,opt,name=start,proto3" json:"start,omitempty"`
}

func (x *MetricTimeRequest) Reset() {
	*x = MetricTimeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_metric_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricTimeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricTimeRequest) ProtoMessage() {}

func (x *MetricTimeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_metric_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricTimeRequest.ProtoReflect.Descriptor instead.
func (*MetricTimeRequest) Descriptor() ([]byte, []int) {
	return file_metric_proto_rawDescGZIP(), []int{3}
}

func (x *MetricTimeRequest) GetStart() int64 {
	if x != nil {
		return x.Start
	}
	return 0
}

type MetricTimeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metrics []*MetricTime `protobuf:"bytes,1,rep,name=metrics,proto3" json:"metrics,omitempty"`
}

func (x *MetricTimeResponse) Reset() {
	*x = MetricTimeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_metric_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricTimeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricTimeResponse) ProtoMessage() {}

func (x *MetricTimeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_metric_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricTimeResponse.ProtoReflect.Descriptor instead.
func (*MetricTimeResponse) Descriptor() ([]byte, []int) {
	return file_metric_proto_rawDescGZIP(), []int{4}
}

func (x *MetricTimeResponse) GetMetrics() []*MetricTime {
	if x != nil {
		return x.Metrics
	}
	return nil
}

type MetricTime struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id   []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Time int64  `protobuf:"varint,2,opt,name=time,proto3" json:"time,omitempty"`
}

func (x *MetricTime) Reset() {
	*x = MetricTime{}
	if protoimpl.UnsafeEnabled {
		mi := &file_metric_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricTime) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricTime) ProtoMessage() {}

func (x *MetricTime) ProtoReflect() protoreflect.Message {
	mi := &file_metric_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricTime.ProtoReflect.Descriptor instead.
func (*MetricTime) Descriptor() ([]byte, []int) {
	return file_metric_proto_rawDescGZIP(), []int{5}
}

func (x *MetricTime) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *MetricTime) GetTime() int64 {
	if x != nil {
		return x.Time
	}
	return 0
}

var File_metric_proto protoreflect.FileDescriptor

var file_metric_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a,
	0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x70, 0x62, 0x22, 0x39, 0x0a, 0x0d, 0x4d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x70,
	0x61, 0x72, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61,
	0x72, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x3e, 0x0a, 0x0e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2c, 0x0a, 0x05, 0x69, 0x6e, 0x66, 0x6f, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x5f, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x05,
	0x69, 0x6e, 0x66, 0x6f, 0x73, 0x22, 0xaa, 0x01, 0x0a, 0x0a, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63,
	0x49, 0x6e, 0x66, 0x6f, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x05, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x22, 0x29, 0x0a, 0x11, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x54, 0x69, 0x6d, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x22, 0x46, 0x0a,
	0x12, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x54, 0x69, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x30, 0x0a, 0x07, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x70,
	0x62, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x54, 0x69, 0x6d, 0x65, 0x52, 0x07, 0x6d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x22, 0x30, 0x0a, 0x0a, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x54,
	0x69, 0x6d, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x42, 0x36, 0x5a, 0x34, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x65, 0x6d, 0x6f, 0x63, 0x61, 0x73, 0x68, 0x2f, 0x69,
	0x6e, 0x64, 0x65, 0x78, 0x2f, 0x72, 0x65, 0x66, 0x2f, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_metric_proto_rawDescOnce sync.Once
	file_metric_proto_rawDescData = file_metric_proto_rawDesc
)

func file_metric_proto_rawDescGZIP() []byte {
	file_metric_proto_rawDescOnce.Do(func() {
		file_metric_proto_rawDescData = protoimpl.X.CompressGZIP(file_metric_proto_rawDescData)
	})
	return file_metric_proto_rawDescData
}

var file_metric_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_metric_proto_goTypes = []any{
	(*MetricRequest)(nil),      // 0: network_pb.MetricRequest
	(*MetricResponse)(nil),     // 1: network_pb.MetricResponse
	(*MetricInfo)(nil),         // 2: network_pb.MetricInfo
	(*MetricTimeRequest)(nil),  // 3: network_pb.MetricTimeRequest
	(*MetricTimeResponse)(nil), // 4: network_pb.MetricTimeResponse
	(*MetricTime)(nil),         // 5: network_pb.MetricTime
}
var file_metric_proto_depIdxs = []int32{
	2, // 0: network_pb.MetricResponse.infos:type_name -> network_pb.MetricInfo
	5, // 1: network_pb.MetricTimeResponse.metrics:type_name -> network_pb.MetricTime
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_metric_proto_init() }
func file_metric_proto_init() {
	if File_metric_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_metric_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*MetricRequest); i {
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
		file_metric_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*MetricResponse); i {
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
		file_metric_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*MetricInfo); i {
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
		file_metric_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*MetricTimeRequest); i {
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
		file_metric_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*MetricTimeResponse); i {
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
		file_metric_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*MetricTime); i {
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
			RawDescriptor: file_metric_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_metric_proto_goTypes,
		DependencyIndexes: file_metric_proto_depIdxs,
		MessageInfos:      file_metric_proto_msgTypes,
	}.Build()
	File_metric_proto = out.File
	file_metric_proto_rawDesc = nil
	file_metric_proto_goTypes = nil
	file_metric_proto_depIdxs = nil
}
