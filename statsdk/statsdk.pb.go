// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.21.12
// source: statsdk/statsdk.proto

package statsdk

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

type GetStatRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Token string `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *GetStatRequest) Reset() {
	*x = GetStatRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_statsdk_statsdk_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStatRequest) ProtoMessage() {}

func (x *GetStatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_statsdk_statsdk_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStatRequest.ProtoReflect.Descriptor instead.
func (*GetStatRequest) Descriptor() ([]byte, []int) {
	return file_statsdk_statsdk_proto_rawDescGZIP(), []int{0}
}

func (x *GetStatRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *GetStatRequest) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

type SetStatRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Token string `protobuf:"bytes,3,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *SetStatRequest) Reset() {
	*x = SetStatRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_statsdk_statsdk_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetStatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetStatRequest) ProtoMessage() {}

func (x *SetStatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_statsdk_statsdk_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetStatRequest.ProtoReflect.Descriptor instead.
func (*SetStatRequest) Descriptor() ([]byte, []int) {
	return file_statsdk_statsdk_proto_rawDescGZIP(), []int{1}
}

func (x *SetStatRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *SetStatRequest) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *SetStatRequest) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

type UpdateStatRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key      string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value    string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Token    string `protobuf:"bytes,3,opt,name=token,proto3" json:"token,omitempty"`
	Datatype string `protobuf:"bytes,4,opt,name=datatype,proto3" json:"datatype,omitempty"`
}

func (x *UpdateStatRequest) Reset() {
	*x = UpdateStatRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_statsdk_statsdk_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateStatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateStatRequest) ProtoMessage() {}

func (x *UpdateStatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_statsdk_statsdk_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateStatRequest.ProtoReflect.Descriptor instead.
func (*UpdateStatRequest) Descriptor() ([]byte, []int) {
	return file_statsdk_statsdk_proto_rawDescGZIP(), []int{2}
}

func (x *UpdateStatRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *UpdateStatRequest) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *UpdateStatRequest) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *UpdateStatRequest) GetDatatype() string {
	if x != nil {
		return x.Datatype
	}
	return ""
}

type GetStatResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Results string `protobuf:"bytes,1,opt,name=results,proto3" json:"results,omitempty"`
}

func (x *GetStatResponse) Reset() {
	*x = GetStatResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_statsdk_statsdk_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStatResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStatResponse) ProtoMessage() {}

func (x *GetStatResponse) ProtoReflect() protoreflect.Message {
	mi := &file_statsdk_statsdk_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStatResponse.ProtoReflect.Descriptor instead.
func (*GetStatResponse) Descriptor() ([]byte, []int) {
	return file_statsdk_statsdk_proto_rawDescGZIP(), []int{3}
}

func (x *GetStatResponse) GetResults() string {
	if x != nil {
		return x.Results
	}
	return ""
}

type UpdateStatResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *UpdateStatResponse) Reset() {
	*x = UpdateStatResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_statsdk_statsdk_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateStatResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateStatResponse) ProtoMessage() {}

func (x *UpdateStatResponse) ProtoReflect() protoreflect.Message {
	mi := &file_statsdk_statsdk_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateStatResponse.ProtoReflect.Descriptor instead.
func (*UpdateStatResponse) Descriptor() ([]byte, []int) {
	return file_statsdk_statsdk_proto_rawDescGZIP(), []int{4}
}

func (x *UpdateStatResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

var File_statsdk_statsdk_proto protoreflect.FileDescriptor

var file_statsdk_statsdk_proto_rawDesc = []byte{
	0x0a, 0x15, 0x73, 0x74, 0x61, 0x74, 0x73, 0x64, 0x6b, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x73, 0x64,
	0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x73, 0x74, 0x61, 0x74, 0x73, 0x64, 0x6b,
	0x22, 0x38, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x4e, 0x0a, 0x0e, 0x53, 0x65,
	0x74, 0x53, 0x74, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x6d, 0x0a, 0x11, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x1a, 0x0a,
	0x08, 0x64, 0x61, 0x74, 0x61, 0x74, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x64, 0x61, 0x74, 0x61, 0x74, 0x79, 0x70, 0x65, 0x22, 0x2b, 0x0a, 0x0f, 0x47, 0x65, 0x74,
	0x53, 0x74, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x72,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x2e, 0x0a, 0x12, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x53, 0x74, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x32, 0xa4, 0x02, 0x0a, 0x0b, 0x53, 0x74, 0x61, 0x74, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3d, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61,
	0x74, 0x73, 0x12, 0x17, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x64, 0x6b, 0x2e, 0x47, 0x65, 0x74,
	0x53, 0x74, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x73, 0x74,
	0x61, 0x74, 0x73, 0x64, 0x6b, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x40, 0x0a, 0x08, 0x53, 0x65, 0x74, 0x53, 0x74, 0x61, 0x74,
	0x73, 0x12, 0x17, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x64, 0x6b, 0x2e, 0x53, 0x65, 0x74, 0x53,
	0x74, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x73, 0x74, 0x61,
	0x74, 0x73, 0x64, 0x6b, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x49, 0x0a, 0x0e, 0x49, 0x6e, 0x63, 0x72, 0x65,
	0x6d, 0x65, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x73, 0x12, 0x1a, 0x2e, 0x73, 0x74, 0x61, 0x74,
	0x73, 0x64, 0x6b, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x64, 0x6b, 0x2e,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x49, 0x0a, 0x0e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4d, 0x61, 0x78, 0x53,
	0x74, 0x61, 0x74, 0x73, 0x12, 0x1a, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x64, 0x6b, 0x2e, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x1b, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x64, 0x6b, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x53, 0x74, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x0b, 0x5a,
	0x09, 0x2e, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x73, 0x64, 0x6b, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_statsdk_statsdk_proto_rawDescOnce sync.Once
	file_statsdk_statsdk_proto_rawDescData = file_statsdk_statsdk_proto_rawDesc
)

func file_statsdk_statsdk_proto_rawDescGZIP() []byte {
	file_statsdk_statsdk_proto_rawDescOnce.Do(func() {
		file_statsdk_statsdk_proto_rawDescData = protoimpl.X.CompressGZIP(file_statsdk_statsdk_proto_rawDescData)
	})
	return file_statsdk_statsdk_proto_rawDescData
}

var file_statsdk_statsdk_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_statsdk_statsdk_proto_goTypes = []any{
	(*GetStatRequest)(nil),     // 0: statsdk.GetStatRequest
	(*SetStatRequest)(nil),     // 1: statsdk.SetStatRequest
	(*UpdateStatRequest)(nil),  // 2: statsdk.UpdateStatRequest
	(*GetStatResponse)(nil),    // 3: statsdk.GetStatResponse
	(*UpdateStatResponse)(nil), // 4: statsdk.UpdateStatResponse
}
var file_statsdk_statsdk_proto_depIdxs = []int32{
	0, // 0: statsdk.StatService.GetStats:input_type -> statsdk.GetStatRequest
	1, // 1: statsdk.StatService.SetStats:input_type -> statsdk.SetStatRequest
	2, // 2: statsdk.StatService.IncrementStats:input_type -> statsdk.UpdateStatRequest
	2, // 3: statsdk.StatService.UpdateMaxStats:input_type -> statsdk.UpdateStatRequest
	3, // 4: statsdk.StatService.GetStats:output_type -> statsdk.GetStatResponse
	4, // 5: statsdk.StatService.SetStats:output_type -> statsdk.UpdateStatResponse
	4, // 6: statsdk.StatService.IncrementStats:output_type -> statsdk.UpdateStatResponse
	4, // 7: statsdk.StatService.UpdateMaxStats:output_type -> statsdk.UpdateStatResponse
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_statsdk_statsdk_proto_init() }
func file_statsdk_statsdk_proto_init() {
	if File_statsdk_statsdk_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_statsdk_statsdk_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*GetStatRequest); i {
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
		file_statsdk_statsdk_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*SetStatRequest); i {
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
		file_statsdk_statsdk_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*UpdateStatRequest); i {
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
		file_statsdk_statsdk_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*GetStatResponse); i {
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
		file_statsdk_statsdk_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*UpdateStatResponse); i {
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
			RawDescriptor: file_statsdk_statsdk_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_statsdk_statsdk_proto_goTypes,
		DependencyIndexes: file_statsdk_statsdk_proto_depIdxs,
		MessageInfos:      file_statsdk_statsdk_proto_msgTypes,
	}.Build()
	File_statsdk_statsdk_proto = out.File
	file_statsdk_statsdk_proto_rawDesc = nil
	file_statsdk_statsdk_proto_goTypes = nil
	file_statsdk_statsdk_proto_depIdxs = nil
}
