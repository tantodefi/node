// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: observer/observer.proto

package types

import (
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"

	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	common "github.com/zeta-chain/node/common"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type ObservationType int32

const (
	ObservationType_EmptyObserverType ObservationType = 0
	ObservationType_InBoundTx         ObservationType = 1
	ObservationType_OutBoundTx        ObservationType = 2
	ObservationType_TSSKeyGen         ObservationType = 3
	ObservationType_TSSKeySign        ObservationType = 4
)

var ObservationType_name = map[int32]string{
	0: "EmptyObserverType",
	1: "InBoundTx",
	2: "OutBoundTx",
	3: "TSSKeyGen",
	4: "TSSKeySign",
}

var ObservationType_value = map[string]int32{
	"EmptyObserverType": 0,
	"InBoundTx":         1,
	"OutBoundTx":        2,
	"TSSKeyGen":         3,
	"TSSKeySign":        4,
}

func (x ObservationType) String() string {
	return proto.EnumName(ObservationType_name, int32(x))
}

func (ObservationType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_3004233a4a5969ce, []int{0}
}

type ObserverMapper struct {
	Index         string        `protobuf:"bytes,1,opt,name=index,proto3" json:"index,omitempty"`
	ObserverChain *common.Chain `protobuf:"bytes,2,opt,name=observer_chain,json=observerChain,proto3" json:"observer_chain,omitempty"`
	ObserverList  []string      `protobuf:"bytes,4,rep,name=observer_list,json=observerList,proto3" json:"observer_list,omitempty"`
}

func (m *ObserverMapper) Reset()         { *m = ObserverMapper{} }
func (m *ObserverMapper) String() string { return proto.CompactTextString(m) }
func (*ObserverMapper) ProtoMessage()    {}
func (*ObserverMapper) Descriptor() ([]byte, []int) {
	return fileDescriptor_3004233a4a5969ce, []int{0}
}
func (m *ObserverMapper) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ObserverMapper) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ObserverMapper.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ObserverMapper) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ObserverMapper.Merge(m, src)
}
func (m *ObserverMapper) XXX_Size() int {
	return m.Size()
}
func (m *ObserverMapper) XXX_DiscardUnknown() {
	xxx_messageInfo_ObserverMapper.DiscardUnknown(m)
}

var xxx_messageInfo_ObserverMapper proto.InternalMessageInfo

func (m *ObserverMapper) GetIndex() string {
	if m != nil {
		return m.Index
	}
	return ""
}

func (m *ObserverMapper) GetObserverChain() *common.Chain {
	if m != nil {
		return m.ObserverChain
	}
	return nil
}

func (m *ObserverMapper) GetObserverList() []string {
	if m != nil {
		return m.ObserverList
	}
	return nil
}

type LastObserverCount struct {
	Count            uint64 `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
	LastChangeHeight int64  `protobuf:"varint,2,opt,name=last_change_height,json=lastChangeHeight,proto3" json:"last_change_height,omitempty"`
}

func (m *LastObserverCount) Reset()         { *m = LastObserverCount{} }
func (m *LastObserverCount) String() string { return proto.CompactTextString(m) }
func (*LastObserverCount) ProtoMessage()    {}
func (*LastObserverCount) Descriptor() ([]byte, []int) {
	return fileDescriptor_3004233a4a5969ce, []int{1}
}
func (m *LastObserverCount) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *LastObserverCount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_LastObserverCount.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *LastObserverCount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LastObserverCount.Merge(m, src)
}
func (m *LastObserverCount) XXX_Size() int {
	return m.Size()
}
func (m *LastObserverCount) XXX_DiscardUnknown() {
	xxx_messageInfo_LastObserverCount.DiscardUnknown(m)
}

var xxx_messageInfo_LastObserverCount proto.InternalMessageInfo

func (m *LastObserverCount) GetCount() uint64 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *LastObserverCount) GetLastChangeHeight() int64 {
	if m != nil {
		return m.LastChangeHeight
	}
	return 0
}

func init() {
	proto.RegisterEnum("zetachain.zetacore.observer.ObservationType", ObservationType_name, ObservationType_value)
	proto.RegisterType((*ObserverMapper)(nil), "zetachain.zetacore.observer.ObserverMapper")
	proto.RegisterType((*LastObserverCount)(nil), "zetachain.zetacore.observer.LastObserverCount")
}

func init() { proto.RegisterFile("observer/observer.proto", fileDescriptor_3004233a4a5969ce) }

var fileDescriptor_3004233a4a5969ce = []byte{
	// 367 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0x51, 0xcf, 0x6a, 0x22, 0x31,
	0x1c, 0x9e, 0xe8, 0xec, 0x82, 0xd9, 0xd5, 0x1d, 0xb3, 0x96, 0x8a, 0x85, 0x20, 0xf6, 0x22, 0xa5,
	0x9d, 0x40, 0xdb, 0x27, 0x50, 0x4a, 0x2b, 0xb5, 0x08, 0xa3, 0x50, 0xe8, 0x45, 0x46, 0x0d, 0x33,
	0x01, 0x4d, 0x86, 0x99, 0x4c, 0x71, 0x7a, 0xeb, 0x1b, 0xf4, 0x21, 0x7a, 0xe8, 0xa3, 0xf4, 0xe8,
	0xb1, 0xc7, 0xa2, 0x2f, 0x52, 0x92, 0x18, 0x4f, 0xf9, 0xfe, 0x24, 0xf9, 0x7d, 0xfc, 0x3e, 0x78,
	0x2c, 0x66, 0x19, 0x4d, 0x9f, 0x69, 0x4a, 0x2c, 0xf0, 0x93, 0x54, 0x48, 0x81, 0x4e, 0x5e, 0xa8,
	0x0c, 0xe7, 0x71, 0xc8, 0xb8, 0xaf, 0x91, 0x48, 0xa9, 0x6f, 0xaf, 0xb4, 0xfe, 0xcf, 0xc5, 0x6a,
	0x25, 0x38, 0x31, 0x87, 0x79, 0xd1, 0x6a, 0x44, 0x22, 0x12, 0x1a, 0x12, 0x85, 0x8c, 0xda, 0x79,
	0x05, 0xb0, 0x36, 0xda, 0xbf, 0x7b, 0x08, 0x93, 0x84, 0xa6, 0xa8, 0x01, 0x7f, 0x31, 0xbe, 0xa0,
	0xeb, 0x26, 0x68, 0x83, 0x6e, 0x25, 0x30, 0x04, 0x5d, 0xc3, 0x9a, 0xfd, 0x7f, 0xaa, 0xe7, 0x36,
	0x4b, 0x6d, 0xd0, 0xfd, 0x73, 0x59, 0xf5, 0xf7, 0x53, 0xfa, 0x4a, 0x0c, 0xaa, 0xf6, 0x92, 0xa6,
	0xe8, 0x14, 0x1e, 0x84, 0xe9, 0x92, 0x65, 0xb2, 0xe9, 0xb6, 0xcb, 0xdd, 0x4a, 0xf0, 0xd7, 0x8a,
	0x43, 0x96, 0xc9, 0xce, 0x23, 0xac, 0x0f, 0xc3, 0x4c, 0xda, 0x18, 0x7d, 0x91, 0x73, 0xa9, 0x52,
	0xcc, 0x15, 0xd0, 0x29, 0xdc, 0xc0, 0x10, 0x74, 0x0e, 0xd1, 0x32, 0xcc, 0xa4, 0x4a, 0xc0, 0x23,
	0x3a, 0x8d, 0x29, 0x8b, 0x62, 0xa9, 0x93, 0x94, 0x03, 0x4f, 0x39, 0x7d, 0x6d, 0xdc, 0x69, 0xfd,
	0x6c, 0x09, 0xff, 0x99, 0x4f, 0x43, 0xc9, 0x04, 0x9f, 0x14, 0x09, 0x45, 0x47, 0xb0, 0x7e, 0xb3,
	0x4a, 0x64, 0x61, 0x87, 0x29, 0xd1, 0x73, 0x50, 0x15, 0x56, 0x06, 0xbc, 0x27, 0x72, 0xbe, 0x98,
	0xac, 0x3d, 0x80, 0x6a, 0x10, 0x8e, 0x72, 0x69, 0x79, 0x49, 0xd9, 0x93, 0xf1, 0xf8, 0x9e, 0x16,
	0xb7, 0x94, 0x7b, 0x65, 0x65, 0x1b, 0x3a, 0x66, 0x11, 0xf7, 0xdc, 0x96, 0xfb, 0xf1, 0x8e, 0x41,
	0x6f, 0xf0, 0xb9, 0xc5, 0x60, 0xb3, 0xc5, 0xe0, 0x7b, 0x8b, 0xc1, 0xdb, 0x0e, 0x3b, 0x9b, 0x1d,
	0x76, 0xbe, 0x76, 0xd8, 0x79, 0x22, 0x11, 0x93, 0x71, 0x3e, 0x53, 0x9b, 0x22, 0xaa, 0xad, 0x0b,
	0xbd, 0x40, 0x62, 0x8b, 0x23, 0xeb, 0x43, 0xbb, 0x44, 0x16, 0x09, 0xcd, 0x66, 0xbf, 0x75, 0x39,
	0x57, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0xc8, 0xad, 0xb9, 0x91, 0xff, 0x01, 0x00, 0x00,
}

func (m *ObserverMapper) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ObserverMapper) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ObserverMapper) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ObserverList) > 0 {
		for iNdEx := len(m.ObserverList) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.ObserverList[iNdEx])
			copy(dAtA[i:], m.ObserverList[iNdEx])
			i = encodeVarintObserver(dAtA, i, uint64(len(m.ObserverList[iNdEx])))
			i--
			dAtA[i] = 0x22
		}
	}
	if m.ObserverChain != nil {
		{
			size, err := m.ObserverChain.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintObserver(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if len(m.Index) > 0 {
		i -= len(m.Index)
		copy(dAtA[i:], m.Index)
		i = encodeVarintObserver(dAtA, i, uint64(len(m.Index)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *LastObserverCount) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *LastObserverCount) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *LastObserverCount) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.LastChangeHeight != 0 {
		i = encodeVarintObserver(dAtA, i, uint64(m.LastChangeHeight))
		i--
		dAtA[i] = 0x10
	}
	if m.Count != 0 {
		i = encodeVarintObserver(dAtA, i, uint64(m.Count))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintObserver(dAtA []byte, offset int, v uint64) int {
	offset -= sovObserver(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ObserverMapper) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Index)
	if l > 0 {
		n += 1 + l + sovObserver(uint64(l))
	}
	if m.ObserverChain != nil {
		l = m.ObserverChain.Size()
		n += 1 + l + sovObserver(uint64(l))
	}
	if len(m.ObserverList) > 0 {
		for _, s := range m.ObserverList {
			l = len(s)
			n += 1 + l + sovObserver(uint64(l))
		}
	}
	return n
}

func (m *LastObserverCount) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Count != 0 {
		n += 1 + sovObserver(uint64(m.Count))
	}
	if m.LastChangeHeight != 0 {
		n += 1 + sovObserver(uint64(m.LastChangeHeight))
	}
	return n
}

func sovObserver(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozObserver(x uint64) (n int) {
	return sovObserver(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ObserverMapper) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowObserver
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ObserverMapper: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ObserverMapper: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowObserver
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthObserver
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthObserver
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Index = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObserverChain", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowObserver
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthObserver
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthObserver
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.ObserverChain == nil {
				m.ObserverChain = &common.Chain{}
			}
			if err := m.ObserverChain.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObserverList", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowObserver
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthObserver
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthObserver
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ObserverList = append(m.ObserverList, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipObserver(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthObserver
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *LastObserverCount) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowObserver
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: LastObserverCount: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LastObserverCount: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Count", wireType)
			}
			m.Count = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowObserver
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Count |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastChangeHeight", wireType)
			}
			m.LastChangeHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowObserver
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LastChangeHeight |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipObserver(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthObserver
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipObserver(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowObserver
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowObserver
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowObserver
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthObserver
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupObserver
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthObserver
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthObserver        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowObserver          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupObserver = fmt.Errorf("proto: unexpected end of group")
)
