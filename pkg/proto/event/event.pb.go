// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: event/event.proto

package event

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

type Action int32

const (
	Action_Add    Action = 0
	Action_Remove Action = 1
)

// Enum value maps for Action.
var (
	Action_name = map[int32]string{
		0: "Add",
		1: "Remove",
	}
	Action_value = map[string]int32{
		"Add":    0,
		"Remove": 1,
	}
)

func (x Action) Enum() *Action {
	p := new(Action)
	*p = x
	return p
}

func (x Action) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Action) Descriptor() protoreflect.EnumDescriptor {
	return file_event_event_proto_enumTypes[0].Descriptor()
}

func (Action) Type() protoreflect.EnumType {
	return &file_event_event_proto_enumTypes[0]
}

func (x Action) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Action.Descriptor instead.
func (Action) EnumDescriptor() ([]byte, []int) {
	return file_event_event_proto_rawDescGZIP(), []int{0}
}

type FileEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserID   int64  `protobuf:"varint,1,opt,name=userID,proto3" json:"userID,omitempty"`
	UserName string `protobuf:"bytes,2,opt,name=userName,proto3" json:"userName,omitempty"`
	FileName string `protobuf:"bytes,3,opt,name=fileName,proto3" json:"fileName,omitempty"`
	Time     int64  `protobuf:"varint,4,opt,name=time,proto3" json:"time,omitempty"`
	Size     int64  `protobuf:"varint,5,opt,name=size,proto3" json:"size,omitempty"`
	Action   Action `protobuf:"varint,6,opt,name=action,proto3,enum=event.Action" json:"action,omitempty"`
}

func (x *FileEvent) Reset() {
	*x = FileEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_event_event_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileEvent) ProtoMessage() {}

func (x *FileEvent) ProtoReflect() protoreflect.Message {
	mi := &file_event_event_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileEvent.ProtoReflect.Descriptor instead.
func (*FileEvent) Descriptor() ([]byte, []int) {
	return file_event_event_proto_rawDescGZIP(), []int{0}
}

func (x *FileEvent) GetUserID() int64 {
	if x != nil {
		return x.UserID
	}
	return 0
}

func (x *FileEvent) GetUserName() string {
	if x != nil {
		return x.UserName
	}
	return ""
}

func (x *FileEvent) GetFileName() string {
	if x != nil {
		return x.FileName
	}
	return ""
}

func (x *FileEvent) GetTime() int64 {
	if x != nil {
		return x.Time
	}
	return 0
}

func (x *FileEvent) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *FileEvent) GetAction() Action {
	if x != nil {
		return x.Action
	}
	return Action_Add
}

var File_event_event_proto protoreflect.FileDescriptor

var file_event_event_proto_rawDesc = []byte{
	0x0a, 0x11, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0xaa, 0x01, 0x0a, 0x09, 0x46,
	0x69, 0x6c, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x75, 0x73, 0x65, 0x72,
	0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x44,
	0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08,
	0x66, 0x69, 0x6c, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x66, 0x69, 0x6c, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x73, 0x69, 0x7a, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65,
	0x12, 0x25, 0x0a, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x0d, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2a, 0x1d, 0x0a, 0x06, 0x41, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x07, 0x0a, 0x03, 0x41, 0x64, 0x64, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x52, 0x65,
	0x6d, 0x6f, 0x76, 0x65, 0x10, 0x01, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4d, 0x69, 0x6b, 0x68, 0x61, 0x6c, 0x65, 0x76, 0x69, 0x63, 0x68,
	0x2f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x68, 0x61, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x6b, 0x67,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_event_event_proto_rawDescOnce sync.Once
	file_event_event_proto_rawDescData = file_event_event_proto_rawDesc
)

func file_event_event_proto_rawDescGZIP() []byte {
	file_event_event_proto_rawDescOnce.Do(func() {
		file_event_event_proto_rawDescData = protoimpl.X.CompressGZIP(file_event_event_proto_rawDescData)
	})
	return file_event_event_proto_rawDescData
}

var file_event_event_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_event_event_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_event_event_proto_goTypes = []interface{}{
	(Action)(0),       // 0: event.Action
	(*FileEvent)(nil), // 1: event.FileEvent
}
var file_event_event_proto_depIdxs = []int32{
	0, // 0: event.FileEvent.action:type_name -> event.Action
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_event_event_proto_init() }
func file_event_event_proto_init() {
	if File_event_event_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_event_event_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileEvent); i {
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
			RawDescriptor: file_event_event_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_event_event_proto_goTypes,
		DependencyIndexes: file_event_event_proto_depIdxs,
		EnumInfos:         file_event_event_proto_enumTypes,
		MessageInfos:      file_event_event_proto_msgTypes,
	}.Build()
	File_event_event_proto = out.File
	file_event_event_proto_rawDesc = nil
	file_event_event_proto_goTypes = nil
	file_event_event_proto_depIdxs = nil
}
