// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.13.0
// source: topic_eth_signature_message.proto

package proto

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type TopicEthSignatureMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransactionId string `protobuf:"bytes,1,opt,name=transactionId,proto3" json:"transactionId,omitempty"`
	EthAddress    string `protobuf:"bytes,2,opt,name=ethAddress,proto3" json:"ethAddress,omitempty"`
	Amount        string `protobuf:"bytes,3,opt,name=amount,proto3" json:"amount,omitempty"`
	Fee           string `protobuf:"bytes,4,opt,name=fee,proto3" json:"fee,omitempty"`
	Signature     string `protobuf:"bytes,5,opt,name=signature,proto3" json:"signature,omitempty"`
	Asset         string `protobuf:"bytes,6,opt,name=asset,proto3" json:"asset,omitempty"`
}

func (x *TopicEthSignatureMessage) Reset() {
	*x = TopicEthSignatureMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_topic_eth_signature_message_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TopicEthSignatureMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TopicEthSignatureMessage) ProtoMessage() {}

func (x *TopicEthSignatureMessage) ProtoReflect() protoreflect.Message {
	mi := &file_topic_eth_signature_message_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TopicEthSignatureMessage.ProtoReflect.Descriptor instead.
func (*TopicEthSignatureMessage) Descriptor() ([]byte, []int) {
	return file_topic_eth_signature_message_proto_rawDescGZIP(), []int{0}
}

func (x *TopicEthSignatureMessage) GetTransactionId() string {
	if x != nil {
		return x.TransactionId
	}
	return ""
}

func (x *TopicEthSignatureMessage) GetEthAddress() string {
	if x != nil {
		return x.EthAddress
	}
	return ""
}

func (x *TopicEthSignatureMessage) GetAmount() string {
	if x != nil {
		return x.Amount
	}
	return ""
}

func (x *TopicEthSignatureMessage) GetFee() string {
	if x != nil {
		return x.Fee
	}
	return ""
}

func (x *TopicEthSignatureMessage) GetSignature() string {
	if x != nil {
		return x.Signature
	}
	return ""
}

func (x *TopicEthSignatureMessage) GetAsset() string {
	if x != nil {
		return x.Asset
	}
	return ""
}

var File_topic_eth_signature_message_proto protoreflect.FileDescriptor

var file_topic_eth_signature_message_proto_rawDesc = []byte{
	0x0a, 0x21, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x5f, 0x65, 0x74, 0x68, 0x5f, 0x73, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbe, 0x01, 0x0a, 0x18, 0x54,
	0x6f, 0x70, 0x69, 0x63, 0x45, 0x74, 0x68, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x1e, 0x0a,
	0x0a, 0x65, 0x74, 0x68, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0a, 0x65, 0x74, 0x68, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x16, 0x0a,
	0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61,
	0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x66, 0x65, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x66, 0x65, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x73, 0x73, 0x65, 0x74, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x61, 0x73, 0x73, 0x65, 0x74, 0x42, 0x38, 0x5a, 0x36, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x6d, 0x65, 0x63, 0x68,
	0x61, 0x69, 0x6e, 0x2f, 0x68, 0x65, 0x64, 0x65, 0x72, 0x61, 0x2d, 0x65, 0x74, 0x68, 0x2d, 0x62,
	0x72, 0x69, 0x64, 0x67, 0x65, 0x2d, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_topic_eth_signature_message_proto_rawDescOnce sync.Once
	file_topic_eth_signature_message_proto_rawDescData = file_topic_eth_signature_message_proto_rawDesc
)

func file_topic_eth_signature_message_proto_rawDescGZIP() []byte {
	file_topic_eth_signature_message_proto_rawDescOnce.Do(func() {
		file_topic_eth_signature_message_proto_rawDescData = protoimpl.X.CompressGZIP(file_topic_eth_signature_message_proto_rawDescData)
	})
	return file_topic_eth_signature_message_proto_rawDescData
}

var file_topic_eth_signature_message_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_topic_eth_signature_message_proto_goTypes = []interface{}{
	(*TopicEthSignatureMessage)(nil), // 0: proto.TopicEthSignatureMessage
}
var file_topic_eth_signature_message_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_topic_eth_signature_message_proto_init() }
func file_topic_eth_signature_message_proto_init() {
	if File_topic_eth_signature_message_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_topic_eth_signature_message_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TopicEthSignatureMessage); i {
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
			RawDescriptor: file_topic_eth_signature_message_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_topic_eth_signature_message_proto_goTypes,
		DependencyIndexes: file_topic_eth_signature_message_proto_depIdxs,
		MessageInfos:      file_topic_eth_signature_message_proto_msgTypes,
	}.Build()
	File_topic_eth_signature_message_proto = out.File
	file_topic_eth_signature_message_proto_rawDesc = nil
	file_topic_eth_signature_message_proto_goTypes = nil
	file_topic_eth_signature_message_proto_depIdxs = nil
}
