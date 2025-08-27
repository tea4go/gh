package radius

import (
	"encoding/binary"
	"errors"
	"net"
	"time"
	"unicode/utf8"
)

// RFC 2865中定义的基本属性值格式。
var (
	AttributeText    IAttributeCodec // string
	AttributeString  IAttributeCodec // []byte
	AttributeAddress IAttributeCodec // net.IP
	AttributeInteger IAttributeCodec // uint32
	AttributeTime    IAttributeCodec // time.Time
	AttributeUnknown IAttributeCodec // []byte
	AttributeVendor  IAttributeCodec // Vendor-Specific
)

func init() {
	AttributeText = attributeText{}
	AttributeString = attributeString{}
	AttributeAddress = attributeAddress{}
	AttributeInteger = attributeInteger{}
	AttributeTime = attributeTime{}
	AttributeUnknown = attributeUnknown{}
	AttributeVendor = attributeVendor{}
}

type attributeVendor struct{}

func (attributeVendor) Decode(packet *TDataPacket, value []byte) (interface{}, error) {
	vendorID, typeID, re_value, err := DecodeAVPair(value)
	//fmt.Println("AttributeVendor.Decode()", vendorID, typeID, re_value, err)
	if vendorID > 99999 || err != nil {
		return AttributeText.Decode(packet, value)
	}
	VerdorID = vendorID
	VerdorTypeID = typeID
	//return fmt.Sprintf("[%d-%d]%s", vendorID, typeID, re_value), err
	return re_value, err
}

func (attributeVendor) Encode(packet *TDataPacket, value interface{}) ([]byte, error) {
	//fmt.Println("AttributeVendor.Encode()", VerdorID, VerdorTypeID, value.(string))
	if VerdorID > 0 {
		return EncodeAVPair(VerdorID, VerdorTypeID, value.(string)), nil
	} else {
		return AttributeText.Encode(packet, value)
	}
}

func (attributeVendor) GetCodeName() string {
	return "AttributeVendor"
}

type attributeUnknown struct{}

func (attributeUnknown) Decode(packet *TDataPacket, value []byte) (interface{}, error) {
	return nil, errors.New("未注册属性，不能解码！")
}

func (attributeUnknown) Encode(packet *TDataPacket, value interface{}) ([]byte, error) {
	return nil, errors.New("未注册属性，不能编码！")
}

func (attributeUnknown) GetCodeName() string {
	return "AttributeUnknown"
}

type attributeText struct{}

func (attributeText) Decode(packet *TDataPacket, value []byte) (interface{}, error) {
	if !utf8.Valid(value) {
		return nil, errors.New("radius: text attribute is not valid UTF-8")
	}
	return string(value), nil
}

func (attributeText) Encode(packet *TDataPacket, value interface{}) ([]byte, error) {
	str, ok := value.(string)
	if ok {
		return []byte(str), nil
	}
	raw, ok := value.([]byte)
	if ok {
		return raw, nil
	}
	return nil, errors.New("radius: text attribute must be string or []byte")
}

func (attributeText) GetCodeName() string {
	return "AttributeText"
}

type attributeString struct{}

func (attributeString) Decode(packet *TDataPacket, value []byte) (interface{}, error) {
	v := make([]byte, len(value))
	copy(v, value)
	return v, nil
}

func (attributeString) Encode(packet *TDataPacket, value interface{}) ([]byte, error) {
	raw, ok := value.([]byte)
	if ok {
		return raw, nil
	}
	str, ok := value.(string)
	if ok {
		return []byte(str), nil
	}
	return nil, errors.New("radius: string attribute must be []byte or string")
}

func (attributeString) GetCodeName() string {
	return "AttributeString"
}

type attributeAddress struct{}

func (attributeAddress) Decode(packet *TDataPacket, value []byte) (interface{}, error) {
	if len(value) != net.IPv4len {
		return nil, errors.New("radius: address attribute has invalid size")
	}
	v := make([]byte, len(value))
	copy(v, value)
	return net.IP(v), nil
}

func (attributeAddress) Encode(packet *TDataPacket, value interface{}) ([]byte, error) {
	ip, ok := value.(net.IP)
	if !ok {
		return nil, errors.New("radius: address attribute must be net.IP")
	}
	ip = ip.To4()
	if ip == nil {
		return nil, errors.New("radius: address attribute must be an IPv4 net.IP")
	}
	return []byte(ip), nil
}

func (attributeAddress) GetCodeName() string {
	return "AttributeAddress"
}

type attributeInteger struct{}

func (attributeInteger) Decode(packet *TDataPacket, value []byte) (interface{}, error) {
	if len(value) != 4 {
		return nil, errors.New("radius: integer attribute has invalid size")
	}
	//fmt.Printf("attributeInteger.Decode() : %x==>%v\n", value, value)
	//fmt.Println("attributeInteger.Decode() : ", value, binary.BigEndian.Uint32(value))
	return binary.BigEndian.Uint32(value), nil
}

func (attributeInteger) Encode(packet *TDataPacket, value interface{}) ([]byte, error) {
	integer, ok := value.(uint32)
	if !ok {
		return nil, errors.New("radius: integer attribute must be uint32")
	}

	//fmt.Println("attributeInteger.Encode() : ", integer)
	raw := make([]byte, 4)
	binary.BigEndian.PutUint32(raw, integer)
	//fmt.Println("attributeInteger.Encode() : ", raw)
	return raw, nil
}

func (attributeInteger) GetCodeName() string {
	return "AttributeInteger"
}

type attributeTime struct{}

func (attributeTime) Decode(packet *TDataPacket, value []byte) (interface{}, error) {
	if len(value) != 4 {
		return nil, errors.New("radius: time attribute has invalid size")
	}
	return time.Unix(int64(binary.BigEndian.Uint32(value)), 0), nil
}

func (attributeTime) Encode(packet *TDataPacket, value interface{}) ([]byte, error) {
	timestamp, ok := value.(time.Time)
	if !ok {
		return nil, errors.New("radius: time attribute must be time.Time")
	}
	raw := make([]byte, 4)
	binary.BigEndian.PutUint32(raw, uint32(timestamp.Unix()))
	return raw, nil
}

func (attributeTime) GetCodeName() string {
	return "AttributeTime"
}
