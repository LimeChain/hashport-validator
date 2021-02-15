// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.13.0
// source: crypto_transfer_message.proto

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

type CryptoTransferMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransactionId string `protobuf:"bytes,1,opt,name=transactionId,proto3" json:"transactionId,omitempty"`
	EthAddress    string `protobuf:"bytes,2,opt,name=ethAddress,proto3" json:"ethAddress,omitempty"`
	Amount        string `protobuf:"bytes,3,opt,name=amount,proto3" json:"amount,omitempty"`
	Fee           string `protobuf:"bytes,4,opt,name=fee,proto3" json:"fee,omitempty"`
	GasPrice      string `protobuf:"bytes,5,opt,name=gasPrice,proto3" json:"gasPrice,omitempty"`
}

func (x *CryptoTransferMessage) Reset() {
	*x = CryptoTransferMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_crypto_transfer_message_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CryptoTransferMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CryptoTransferMessage) ProtoMessage() {}

func (x *CryptoTransferMessage) ProtoReflect() protoreflect.Message {
	mi := &file_crypto_transfer_message_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CryptoTransferMessage.ProtoReflect.Descriptor instead.
func (*CryptoTransferMessage) Descriptor() ([]byte, []int) {
	return file_crypto_transfer_message_proto_rawDescGZIP(), []int{0}
}

func (x *CryptoTransferMessage) GetTransactionId() string {
	if x != nil {
		return x.TransactionId
	}
	return ""
}

func (x *CryptoTransferMessage) GetEthAddress() string {
	if x != nil {
		return x.EthAddress
	}
	return ""
}

func (x *CryptoTransferMessage) GetAmount() string {
	if x != nil {
		return x.Amount
	}
	return ""
}

func (x *CryptoTransferMessage) GetFee() string {
	if x != nil {
		return x.Fee
	}
	return ""
}

func (x *CryptoTransferMessage) GetGasPrice() string {
	if x != nil {
		return x.GasPrice
	}
	return ""
}

var File_crypto_transfer_message_proto protoreflect.FileDescriptor

var file_crypto_transfer_message_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65,
	0x72, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa3, 0x01, 0x0a, 0x15, 0x43, 0x72, 0x79, 0x70, 0x74,
	0x6f, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x24, 0x0a, 0x0d, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x74, 0x68, 0x41, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x65, 0x74, 0x68, 0x41,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x10,
	0x0a, 0x03, 0x66, 0x65, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x66, 0x65, 0x65,
	0x12, 0x1a, 0x0a, 0x08, 0x67, 0x61, 0x73, 0x50, 0x72, 0x69, 0x63, 0x65, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x67, 0x61, 0x73, 0x50, 0x72, 0x69, 0x63, 0x65, 0x42, 0x38, 0x5a, 0x36,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x6d, 0x65, 0x63,
	0x68, 0x61, 0x69, 0x6e, 0x2f, 0x68, 0x65, 0x64, 0x65, 0x72, 0x61, 0x2d, 0x65, 0x74, 0x68, 0x2d,
	0x62, 0x72, 0x69, 0x64, 0x67, 0x65, 0x2d, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_crypto_transfer_message_proto_rawDescOnce sync.Once
	file_crypto_transfer_message_proto_rawDescData = file_crypto_transfer_message_proto_rawDesc
)

func file_crypto_transfer_message_proto_rawDescGZIP() []byte {
	file_crypto_transfer_message_proto_rawDescOnce.Do(func() {
		file_crypto_transfer_message_proto_rawDescData = protoimpl.X.CompressGZIP(file_crypto_transfer_message_proto_rawDescData)
	})
	return file_crypto_transfer_message_proto_rawDescData
}

var file_crypto_transfer_message_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_crypto_transfer_message_proto_goTypes = []interface{}{
	(*CryptoTransferMessage)(nil), // 0: proto.CryptoTransferMessage
}
var file_crypto_transfer_message_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_crypto_transfer_message_proto_init() }
func file_crypto_transfer_message_proto_init() {
	if File_crypto_transfer_message_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_crypto_transfer_message_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CryptoTransferMessage); i {
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
			RawDescriptor: file_crypto_transfer_message_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_crypto_transfer_message_proto_goTypes,
		DependencyIndexes: file_crypto_transfer_message_proto_depIdxs,
		MessageInfos:      file_crypto_transfer_message_proto_msgTypes,
	}.Build()
	File_crypto_transfer_message_proto = out.File
	file_crypto_transfer_message_proto_rawDesc = nil
	file_crypto_transfer_message_proto_goTypes = nil
	file_crypto_transfer_message_proto_depIdxs = nil
}
