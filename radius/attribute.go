package radius

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

//属性是RADIUS属性，它是RADIUS数据包的一部分。
type TAttribute struct {
	AttrId    byte
	AttrValue interface{}
}

//AttributeCodec定义了如何对属性进行编码和解码数据。
// 注意：不要存储数据; 复制一份。
type IAttributeCodec interface {
	Decode(packet *TDataPacket, wire []byte) (interface{}, error)
	Encode(packet *TDataPacket, value interface{}) ([]byte, error)
	GetCodeName() string
}

//AttributeTransformer定义了属性编解码器的扩展。 它提供了一种将属性值转换为属性允许的值的方法。
type IAttributeTransformer interface {
	Transform(value interface{}) (interface{}, error)
}

//AttributeStringer定义属性编解码器的扩展。 它提供了一个将属性值转换为字符串的方法。
type IAttributeStringer interface {
	String(value interface{}) string
}

// EncodeAVPair encodes AVPair into Vendor-Specific attribute format (string)
func EncodeAVPair(vendorID uint32, typeID uint8, value string) (vsa []byte) {
	return EncodeAVPairByte(vendorID, typeID, []byte(value))
}

// EncodeAVpairTag encodes AVPair into Vendor-Specific attribute format with tag (string)
func EncodeAVpairTag(vendorID uint32, typeID uint8, tag uint8, value string) (vsa []byte) {
	return EncodeAVPairByteTag(vendorID, typeID, tag, []byte(value))
}

// EncodeAVPairByte encodes AVPair into Vendor-Specific attribute format (byte)
func EncodeAVPairByte(vendorID uint32, typeID uint8, value []byte) (vsa []byte) {
	var b bytes.Buffer
	bv := make([]byte, 4)
	binary.BigEndian.PutUint32(bv, vendorID)

	// Vendor-Id(4) + Type-ID(1) + Length(1)
	b.Write(bv)
	b.Write([]byte{byte(typeID), byte(len(value) + 2)})

	// Append attribute value pair
	b.Write(value)

	vsa = b.Bytes()
	return
}

// EncodeAVPairByteTag encodes AVPair into Vendor-Specific attribute format with tag (byte)
func EncodeAVPairByteTag(vendorID uint32, typeID uint8, tag uint8, value []byte) (vsa []byte) {
	var b bytes.Buffer
	bv := make([]byte, 4)
	binary.BigEndian.PutUint32(bv, vendorID)

	// Vendor-Id(4) + Type-ID(1) + Length(1)
	b.Write(bv)
	b.Write([]byte{byte(typeID), byte(len(value) + 3)})

	// Add tag
	b.WriteByte(byte(tag))

	// Append attribute value pair
	b.Write(value)

	vsa = b.Bytes()
	return
}

// DecodeAVPairByte decodes AVP (byte)
func DecodeAVPairByte(vsa []byte) (vendorID uint32, typeID uint8, value []byte, err error) {
	if len(vsa) <= 6 {
		err = fmt.Errorf("Too short VSA: %d bytes", len(vsa))
		return
	}

	vendorID = binary.BigEndian.Uint32([]byte{vsa[0], vsa[1], vsa[2], vsa[3]})
	typeID = uint8(vsa[4])
	value = vsa[6:]
	return
}

// DecodeAVPair decodes AVP (string)
func DecodeAVPair(vsa []byte) (vendorID uint32, typeID uint8, value string, err error) {
	vendorID, typeID, v, err := DecodeAVPairByte(vsa)
	value = string(v)
	return
}
