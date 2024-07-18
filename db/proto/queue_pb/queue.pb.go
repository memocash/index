// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.1
// source: queue.proto

package queue_pb

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

type Messages struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Messages []*Message `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
}

func (x *Messages) Reset() {
	*x = Messages{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Messages) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Messages) ProtoMessage() {}

func (x *Messages) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Messages.ProtoReflect.Descriptor instead.
func (*Messages) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{0}
}

func (x *Messages) GetMessages() []*Message {
	if x != nil {
		return x.Messages
	}
	return nil
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid       []byte `protobuf:"bytes,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Topic     string `protobuf:"bytes,2,opt,name=topic,proto3" json:"topic,omitempty"`
	Message   []byte `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	Timestamp int64  `protobuf:"varint,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{1}
}

func (x *Message) GetUid() []byte {
	if x != nil {
		return x.Uid
	}
	return nil
}

func (x *Message) GetTopic() string {
	if x != nil {
		return x.Topic
	}
	return ""
}

func (x *Message) GetMessage() []byte {
	if x != nil {
		return x.Message
	}
	return nil
}

func (x *Message) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type ErrorReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error string `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *ErrorReply) Reset() {
	*x = ErrorReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ErrorReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ErrorReply) ProtoMessage() {}

func (x *ErrorReply) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ErrorReply.ProtoReflect.Descriptor instead.
func (*ErrorReply) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{2}
}

func (x *ErrorReply) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type MessageUids struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Topic string   `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
	Uids  [][]byte `protobuf:"bytes,2,rep,name=uids,proto3" json:"uids,omitempty"`
}

func (x *MessageUids) Reset() {
	*x = MessageUids{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageUids) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageUids) ProtoMessage() {}

func (x *MessageUids) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageUids.ProtoReflect.Descriptor instead.
func (*MessageUids) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{3}
}

func (x *MessageUids) GetTopic() string {
	if x != nil {
		return x.Topic
	}
	return ""
}

func (x *MessageUids) GetUids() [][]byte {
	if x != nil {
		return x.Uids
	}
	return nil
}

type RequestSingle struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Topic string `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
	Uid   []byte `protobuf:"bytes,2,opt,name=uid,proto3" json:"uid,omitempty"`
}

func (x *RequestSingle) Reset() {
	*x = RequestSingle{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestSingle) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestSingle) ProtoMessage() {}

func (x *RequestSingle) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestSingle.ProtoReflect.Descriptor instead.
func (*RequestSingle) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{4}
}

func (x *RequestSingle) GetTopic() string {
	if x != nil {
		return x.Topic
	}
	return ""
}

func (x *RequestSingle) GetUid() []byte {
	if x != nil {
		return x.Uid
	}
	return nil
}

type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Topic    string   `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
	Start    []byte   `protobuf:"bytes,2,opt,name=start,proto3" json:"start,omitempty"`
	Max      uint32   `protobuf:"varint,3,opt,name=max,proto3" json:"max,omitempty"`
	Prefixes [][]byte `protobuf:"bytes,4,rep,name=prefixes,proto3" json:"prefixes,omitempty"`
	Uids     [][]byte `protobuf:"bytes,5,rep,name=uids,proto3" json:"uids,omitempty"`
	Wait     bool     `protobuf:"varint,6,opt,name=wait,proto3" json:"wait,omitempty"`
	Newest   bool     `protobuf:"varint,7,opt,name=newest,proto3" json:"newest,omitempty"`
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{5}
}

func (x *Request) GetTopic() string {
	if x != nil {
		return x.Topic
	}
	return ""
}

func (x *Request) GetStart() []byte {
	if x != nil {
		return x.Start
	}
	return nil
}

func (x *Request) GetMax() uint32 {
	if x != nil {
		return x.Max
	}
	return 0
}

func (x *Request) GetPrefixes() [][]byte {
	if x != nil {
		return x.Prefixes
	}
	return nil
}

func (x *Request) GetUids() [][]byte {
	if x != nil {
		return x.Uids
	}
	return nil
}

func (x *Request) GetWait() bool {
	if x != nil {
		return x.Wait
	}
	return false
}

func (x *Request) GetNewest() bool {
	if x != nil {
		return x.Newest
	}
	return false
}

type RequestStream struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Topic    string   `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
	Prefixes [][]byte `protobuf:"bytes,2,rep,name=prefixes,proto3" json:"prefixes,omitempty"`
}

func (x *RequestStream) Reset() {
	*x = RequestStream{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestStream) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestStream) ProtoMessage() {}

func (x *RequestStream) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestStream.ProtoReflect.Descriptor instead.
func (*RequestStream) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{6}
}

func (x *RequestStream) GetTopic() string {
	if x != nil {
		return x.Topic
	}
	return ""
}

func (x *RequestStream) GetPrefixes() [][]byte {
	if x != nil {
		return x.Prefixes
	}
	return nil
}

type EmptyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *EmptyRequest) Reset() {
	*x = EmptyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EmptyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmptyRequest) ProtoMessage() {}

func (x *EmptyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EmptyRequest.ProtoReflect.Descriptor instead.
func (*EmptyRequest) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{7}
}

type Topic struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name  string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Count uint64 `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *Topic) Reset() {
	*x = Topic{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Topic) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Topic) ProtoMessage() {}

func (x *Topic) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Topic.ProtoReflect.Descriptor instead.
func (*Topic) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{8}
}

func (x *Topic) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Topic) GetCount() uint64 {
	if x != nil {
		return x.Count
	}
	return 0
}

type TopicListReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Topics []*Topic `protobuf:"bytes,1,rep,name=topics,proto3" json:"topics,omitempty"`
}

func (x *TopicListReply) Reset() {
	*x = TopicListReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TopicListReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TopicListReply) ProtoMessage() {}

func (x *TopicListReply) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TopicListReply.ProtoReflect.Descriptor instead.
func (*TopicListReply) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{9}
}

func (x *TopicListReply) GetTopics() []*Topic {
	if x != nil {
		return x.Topics
	}
	return nil
}

type CountRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Topic  string `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
	Prefix []byte `protobuf:"bytes,2,opt,name=prefix,proto3" json:"prefix,omitempty"`
}

func (x *CountRequest) Reset() {
	*x = CountRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CountRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CountRequest) ProtoMessage() {}

func (x *CountRequest) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CountRequest.ProtoReflect.Descriptor instead.
func (*CountRequest) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{10}
}

func (x *CountRequest) GetTopic() string {
	if x != nil {
		return x.Topic
	}
	return ""
}

func (x *CountRequest) GetPrefix() []byte {
	if x != nil {
		return x.Prefix
	}
	return nil
}

type TopicCount struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Count uint64 `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *TopicCount) Reset() {
	*x = TopicCount{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TopicCount) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TopicCount) ProtoMessage() {}

func (x *TopicCount) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TopicCount.ProtoReflect.Descriptor instead.
func (*TopicCount) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{11}
}

func (x *TopicCount) GetCount() uint64 {
	if x != nil {
		return x.Count
	}
	return 0
}

type RequestPrefixes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Topic    string           `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
	Prefixes []*RequestPrefix `protobuf:"bytes,2,rep,name=prefixes,proto3" json:"prefixes,omitempty"`
	Max      uint32           `protobuf:"varint,3,opt,name=max,proto3" json:"max,omitempty"`
	Newest   bool             `protobuf:"varint,4,opt,name=newest,proto3" json:"newest,omitempty"`
}

func (x *RequestPrefixes) Reset() {
	*x = RequestPrefixes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestPrefixes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestPrefixes) ProtoMessage() {}

func (x *RequestPrefixes) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestPrefixes.ProtoReflect.Descriptor instead.
func (*RequestPrefixes) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{12}
}

func (x *RequestPrefixes) GetTopic() string {
	if x != nil {
		return x.Topic
	}
	return ""
}

func (x *RequestPrefixes) GetPrefixes() []*RequestPrefix {
	if x != nil {
		return x.Prefixes
	}
	return nil
}

func (x *RequestPrefixes) GetMax() uint32 {
	if x != nil {
		return x.Max
	}
	return 0
}

func (x *RequestPrefixes) GetNewest() bool {
	if x != nil {
		return x.Newest
	}
	return false
}

type RequestPrefix struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix []byte `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
	Start  []byte `protobuf:"bytes,2,opt,name=start,proto3" json:"start,omitempty"`
	Max    uint32 `protobuf:"varint,3,opt,name=max,proto3" json:"max,omitempty"`
}

func (x *RequestPrefix) Reset() {
	*x = RequestPrefix{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestPrefix) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestPrefix) ProtoMessage() {}

func (x *RequestPrefix) ProtoReflect() protoreflect.Message {
	mi := &file_queue_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestPrefix.ProtoReflect.Descriptor instead.
func (*RequestPrefix) Descriptor() ([]byte, []int) {
	return file_queue_proto_rawDescGZIP(), []int{13}
}

func (x *RequestPrefix) GetPrefix() []byte {
	if x != nil {
		return x.Prefix
	}
	return nil
}

func (x *RequestPrefix) GetStart() []byte {
	if x != nil {
		return x.Start
	}
	return nil
}

func (x *RequestPrefix) GetMax() uint32 {
	if x != nil {
		return x.Max
	}
	return 0
}

var File_queue_proto protoreflect.FileDescriptor

var file_queue_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x71, 0x75, 0x65, 0x75, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x71,
	0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x22, 0x39, 0x0a, 0x08, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x73, 0x12, 0x2d, 0x0a, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62,
	0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x73, 0x22, 0x69, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12,
	0x14, 0x0a, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x74, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12,
	0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x22, 0x0a,
	0x0a, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x22, 0x37, 0x0a, 0x0b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x55, 0x69, 0x64, 0x73,
	0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x69, 0x64, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0c, 0x52, 0x04, 0x75, 0x69, 0x64, 0x73, 0x22, 0x37, 0x0a, 0x0d, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x53, 0x69, 0x6e, 0x67, 0x6c, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74,
	0x6f, 0x70, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x70, 0x69,
	0x63, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03,
	0x75, 0x69, 0x64, 0x22, 0xa3, 0x01, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x14, 0x0a, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x74, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6d,
	0x61, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x6d, 0x61, 0x78, 0x12, 0x1a, 0x0a,
	0x08, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x65, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0c, 0x52,
	0x08, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x65, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x69, 0x64,
	0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x04, 0x75, 0x69, 0x64, 0x73, 0x12, 0x12, 0x0a,
	0x04, 0x77, 0x61, 0x69, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x77, 0x61, 0x69,
	0x74, 0x12, 0x16, 0x0a, 0x06, 0x6e, 0x65, 0x77, 0x65, 0x73, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x06, 0x6e, 0x65, 0x77, 0x65, 0x73, 0x74, 0x22, 0x41, 0x0a, 0x0d, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f,
	0x70, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63,
	0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0c, 0x52, 0x08, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x65, 0x73, 0x22, 0x0e, 0x0a, 0x0c,
	0x45, 0x6d, 0x70, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x31, 0x0a, 0x05,
	0x54, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22,
	0x39, 0x0a, 0x0e, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x12, 0x27, 0x0a, 0x06, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x0f, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x2e, 0x54, 0x6f, 0x70,
	0x69, 0x63, 0x52, 0x06, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x73, 0x22, 0x3c, 0x0a, 0x0c, 0x43, 0x6f,
	0x75, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f,
	0x70, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63,
	0x12, 0x16, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x22, 0x22, 0x0a, 0x0a, 0x54, 0x6f, 0x70, 0x69,
	0x63, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x86, 0x01, 0x0a,
	0x0f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x65, 0x73,
	0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x33, 0x0a, 0x08, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78,
	0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65,
	0x5f, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x50, 0x72, 0x65, 0x66, 0x69,
	0x78, 0x52, 0x08, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x65, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x6d,
	0x61, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x6d, 0x61, 0x78, 0x12, 0x16, 0x0a,
	0x06, 0x6e, 0x65, 0x77, 0x65, 0x73, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x6e,
	0x65, 0x77, 0x65, 0x73, 0x74, 0x22, 0x4f, 0x0a, 0x0d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x14,
	0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x73,
	0x74, 0x61, 0x72, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x03, 0x6d, 0x61, 0x78, 0x32, 0x86, 0x04, 0x0a, 0x05, 0x51, 0x75, 0x65, 0x75, 0x65,
	0x12, 0x3a, 0x0a, 0x0c, 0x53, 0x61, 0x76, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73,
	0x12, 0x12, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x73, 0x1a, 0x14, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x2e,
	0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x3f, 0x0a, 0x0e,
	0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x15,
	0x2e, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x55, 0x69, 0x64, 0x73, 0x1a, 0x14, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62,
	0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x3a, 0x0a,
	0x0a, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x17, 0x2e, 0x71, 0x75,
	0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x53, 0x69,
	0x6e, 0x67, 0x6c, 0x65, 0x1a, 0x11, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x2e,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x00, 0x12, 0x36, 0x0a, 0x0b, 0x47, 0x65, 0x74,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x11, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65,
	0x5f, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x71, 0x75,
	0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x22,
	0x00, 0x12, 0x40, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x42, 0x79, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78,
	0x65, 0x73, 0x12, 0x19, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x2e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x65, 0x73, 0x1a, 0x12, 0x2e,
	0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x73, 0x22, 0x00, 0x12, 0x43, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x17, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65,
	0x5f, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x1a, 0x11, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x42, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x54,
	0x6f, 0x70, 0x69, 0x63, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x16, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65,
	0x5f, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x18, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x2e, 0x54, 0x6f, 0x70, 0x69,
	0x63, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x41, 0x0a, 0x0f,
	0x47, 0x65, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12,
	0x16, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x2e, 0x43, 0x6f, 0x75, 0x6e, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f,
	0x70, 0x62, 0x2e, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x00, 0x42,
	0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x65,
	0x6d, 0x6f, 0x63, 0x61, 0x73, 0x68, 0x2f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2f, 0x64, 0x62, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x70, 0x62, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_queue_proto_rawDescOnce sync.Once
	file_queue_proto_rawDescData = file_queue_proto_rawDesc
)

func file_queue_proto_rawDescGZIP() []byte {
	file_queue_proto_rawDescOnce.Do(func() {
		file_queue_proto_rawDescData = protoimpl.X.CompressGZIP(file_queue_proto_rawDescData)
	})
	return file_queue_proto_rawDescData
}

var file_queue_proto_msgTypes = make([]protoimpl.MessageInfo, 14)
var file_queue_proto_goTypes = []any{
	(*Messages)(nil),        // 0: queue_pb.Messages
	(*Message)(nil),         // 1: queue_pb.Message
	(*ErrorReply)(nil),      // 2: queue_pb.ErrorReply
	(*MessageUids)(nil),     // 3: queue_pb.MessageUids
	(*RequestSingle)(nil),   // 4: queue_pb.RequestSingle
	(*Request)(nil),         // 5: queue_pb.Request
	(*RequestStream)(nil),   // 6: queue_pb.RequestStream
	(*EmptyRequest)(nil),    // 7: queue_pb.EmptyRequest
	(*Topic)(nil),           // 8: queue_pb.Topic
	(*TopicListReply)(nil),  // 9: queue_pb.TopicListReply
	(*CountRequest)(nil),    // 10: queue_pb.CountRequest
	(*TopicCount)(nil),      // 11: queue_pb.TopicCount
	(*RequestPrefixes)(nil), // 12: queue_pb.RequestPrefixes
	(*RequestPrefix)(nil),   // 13: queue_pb.RequestPrefix
}
var file_queue_proto_depIdxs = []int32{
	1,  // 0: queue_pb.Messages.messages:type_name -> queue_pb.Message
	8,  // 1: queue_pb.TopicListReply.topics:type_name -> queue_pb.Topic
	13, // 2: queue_pb.RequestPrefixes.prefixes:type_name -> queue_pb.RequestPrefix
	0,  // 3: queue_pb.Queue.SaveMessages:input_type -> queue_pb.Messages
	3,  // 4: queue_pb.Queue.DeleteMessages:input_type -> queue_pb.MessageUids
	4,  // 5: queue_pb.Queue.GetMessage:input_type -> queue_pb.RequestSingle
	5,  // 6: queue_pb.Queue.GetMessages:input_type -> queue_pb.Request
	12, // 7: queue_pb.Queue.GetByPrefixes:input_type -> queue_pb.RequestPrefixes
	6,  // 8: queue_pb.Queue.GetStreamMessages:input_type -> queue_pb.RequestStream
	7,  // 9: queue_pb.Queue.GetTopicList:input_type -> queue_pb.EmptyRequest
	10, // 10: queue_pb.Queue.GetMessageCount:input_type -> queue_pb.CountRequest
	2,  // 11: queue_pb.Queue.SaveMessages:output_type -> queue_pb.ErrorReply
	2,  // 12: queue_pb.Queue.DeleteMessages:output_type -> queue_pb.ErrorReply
	1,  // 13: queue_pb.Queue.GetMessage:output_type -> queue_pb.Message
	0,  // 14: queue_pb.Queue.GetMessages:output_type -> queue_pb.Messages
	0,  // 15: queue_pb.Queue.GetByPrefixes:output_type -> queue_pb.Messages
	1,  // 16: queue_pb.Queue.GetStreamMessages:output_type -> queue_pb.Message
	9,  // 17: queue_pb.Queue.GetTopicList:output_type -> queue_pb.TopicListReply
	11, // 18: queue_pb.Queue.GetMessageCount:output_type -> queue_pb.TopicCount
	11, // [11:19] is the sub-list for method output_type
	3,  // [3:11] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_queue_proto_init() }
func file_queue_proto_init() {
	if File_queue_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_queue_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Messages); i {
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
		file_queue_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*Message); i {
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
		file_queue_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*ErrorReply); i {
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
		file_queue_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*MessageUids); i {
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
		file_queue_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*RequestSingle); i {
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
		file_queue_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*Request); i {
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
		file_queue_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*RequestStream); i {
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
		file_queue_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*EmptyRequest); i {
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
		file_queue_proto_msgTypes[8].Exporter = func(v any, i int) any {
			switch v := v.(*Topic); i {
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
		file_queue_proto_msgTypes[9].Exporter = func(v any, i int) any {
			switch v := v.(*TopicListReply); i {
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
		file_queue_proto_msgTypes[10].Exporter = func(v any, i int) any {
			switch v := v.(*CountRequest); i {
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
		file_queue_proto_msgTypes[11].Exporter = func(v any, i int) any {
			switch v := v.(*TopicCount); i {
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
		file_queue_proto_msgTypes[12].Exporter = func(v any, i int) any {
			switch v := v.(*RequestPrefixes); i {
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
		file_queue_proto_msgTypes[13].Exporter = func(v any, i int) any {
			switch v := v.(*RequestPrefix); i {
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
			RawDescriptor: file_queue_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   14,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_queue_proto_goTypes,
		DependencyIndexes: file_queue_proto_depIdxs,
		MessageInfos:      file_queue_proto_msgTypes,
	}.Build()
	File_queue_proto = out.File
	file_queue_proto_rawDesc = nil
	file_queue_proto_goTypes = nil
	file_queue_proto_depIdxs = nil
}
