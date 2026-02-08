package radius

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	logs "github.com/tea4go/gh/log4go"
)

// maximum RADIUS packet size
const maxPacketSize = 4095

// Code specifies the kind of RADIUS packet.
type Code byte

// Codes which are defined in RFC 2865.
const (
	CodeAccessRequest      Code = 1
	CodeAccessAccept       Code = 2
	CodeAccessReject       Code = 3
	CodeAccountingRequest  Code = 4
	CodeAccountingResponse Code = 5
	CodeAccessChallenge    Code = 11
	CodeStatusServer       Code = 12
	CodeStatusClient       Code = 13
	CodeReserved           Code = 255
)

var (
	VerdorID     uint32
	VerdorTag    uint8
	VerdorTypeID uint8
)

func SetVendorSpecific(name string) error {
	if name == "hillstone" {
		VerdorID = 28557
		VerdorTag = 0
		VerdorTypeID = 11
		return nil
	}
	return fmt.Errorf("未知的设备！")
}

// Packet defines a RADIUS packet.
type TDataPacket struct {
	Code          Code
	Identifier    byte
	Authenticator [16]byte
	Secret        []byte
	Dictionary    *TDictionary
	AttrItems     []*TAttribute
}

// New returns a new packet with the given code and secret. The identifier and
// authenticator are filled with random data, and the dictionary is set to
// Builtin. nil is returned if not enough random data could be generated.
func NewPacket(code Code, secret []byte) *TDataPacket {
	var buff [17]byte
	if _, err := rand.Read(buff[:]); err != nil {
		return nil
	}

	packet := &TDataPacket{
		Code:       code,
		Identifier: buff[0],
		Secret:     secret,
		Dictionary: Builtin,
	}
	copy(packet.Authenticator[:], buff[1:])
	return packet
}

func (p *TDataPacket) String() string {
	packet_text := ""
	temp_text := fmt.Sprintf("数据包[Code=%d,Identifier=%d,Secret=%s]", p.Code, p.Identifier, p.Secret)
	packet_text = packet_text + temp_text
	for k, v := range p.AttrItems {
		dict := p.Dictionary.IdItems[v.AttrId]

		if dict != nil {
			if dict.Func == AttributeInteger {
				value := p.GetValue(dict.Name)
				temp_text = fmt.Sprintf("\n    [%03d] %s = %d", dict.Id, dict.Name, value.(uint32))
			} else {
				value := p.GetString(dict.Name)
				if dict.Name == "User-Password" {
					if len(value) > 0 {
						temp_text = fmt.Sprintf("\n    [%03d] %s = %c***%s", dict.Id, dict.Name, value[0], value[len(value)-1:])
					} else {
						temp_text = fmt.Sprintf("\n    [%03d] %s = ***", dict.Id, dict.Name)
					}
				} else if dict.Name == "Message-Authenticator" {
					temp_text = fmt.Sprintf("\n    [%03d] %s = 0x%x", dict.Id, dict.Name, value)
				} else {
					temp_text = fmt.Sprintf("\n    [%03d] %s = %s", dict.Id, dict.Name, value)
				}
			}
			packet_text = packet_text + temp_text
		} else {
			temp_text = fmt.Sprintf("\n    Not found %d attribe(%d)", k+1, v.AttrId)
			packet_text = packet_text + temp_text
		}
	}

	return packet_text
}

// Parse parses a RADIUS packet from wire data, using the given shared secret
// and dictionary. nil and an error is returned if there is a problem parsing
// the packet.
//
// Note: this function does not validate the authenticity of a packet.
// Ensuring a packet's authenticity should be done using the IsAuthentic
// method.
func ParsePacket(data, secret []byte, dictionary *TDictionary) (*TDataPacket, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("有效包必须大于20字节。 目前包大小[%d]", len(data))
	}

	packet := &TDataPacket{
		Code:       Code(data[0]),
		Identifier: data[1],
		Secret:     secret,
		Dictionary: dictionary,
	}

	length := binary.BigEndian.Uint16(data[2:4])
	if length < 20 || length > maxPacketSize {
		return nil, fmt.Errorf("无效包长度%d(20-%d)", length, maxPacketSize)
	}

	copy(packet.Authenticator[:], data[4:20])

	attributes := data[20:]
	for len(attributes) > 0 {
		if len(attributes) < 2 {
			return nil, fmt.Errorf("有效属性必须大于2字节。 BufferSize=%d", len(attributes))
		}

		attrLength := attributes[1]
		if attrLength < 1 || attrLength > 253 || len(attributes) < int(attrLength) {
			return nil, fmt.Errorf("无效属性长度%d(2-%d)", attrLength, 253)
		}
		attrType := attributes[0]
		attrValue := attributes[2:attrLength]

		codec := dictionary.GetFunc(attrType)
		if codec == AttributeUnknown {
			logs.Info("未知的属性，请先注册属性编码(%d)。", attrType)
		}

		// if attrType == 26 {
		// 	fmt.Println(DecodeAVPair(attrValue))
		// }

		decoded, err := codec.Decode(packet, attrValue)
		if err != nil {
			if attrType == 2 { //User-Password,如果密码解析不了，则使用默认密码123123
				attr := &TAttribute{
					AttrId:    attrType,
					AttrValue: "123123",
				}
				packet.AttrItems = append(packet.AttrItems, attr)
			} else {
				hexbuff := make([]byte, attrLength*2)
				b := hex.Encode(hexbuff, attrValue)
				return nil, fmt.Errorf("解码属性(%d)失败，数据[%s]，%s", attrType, hexbuff[:b], err.Error())
			}
		} else {
			attr := &TAttribute{
				AttrId:    attrType,
				AttrValue: decoded,
			}
			packet.AttrItems = append(packet.AttrItems, attr)
		}
		attributes = attributes[attrLength:]
	}
	return packet, nil
}

// IsAuthentic returns if the packet is an authenticate response to the given
// request packet. Calling this function is only valid if both:
//   - p.code is one of:
//     CodeAccessAccept
//     CodeAccessReject
//     CodeAccountingRequest
//     CodeAccountingResponse
//     CodeAccessChallenge
//   - p.Authenticator contains the calculated authenticator
func (p *TDataPacket) IsAuthentic(request *TDataPacket) bool {
	switch p.Code {
	case CodeAccessAccept, CodeAccessReject, CodeAccountingRequest, CodeAccessChallenge:
		wire, err := p.Encode()
		if err != nil {
			return false
		}

		hash := md5.New()
		hash.Write(wire[0:4])
		if p.Code == CodeAccountingRequest {
			var nul [16]byte
			hash.Write(nul[:])
		} else {
			hash.Write(request.Authenticator[:])
		}
		hash.Write(wire[20:])
		hash.Write(request.Secret)

		var sum [md5.Size]byte
		//fmt.Println("--->", p.Authenticator[:])
		//fmt.Println("--->", hash.Sum(sum[0:0]))
		return bytes.Equal(hash.Sum(sum[0:0]), p.Authenticator[:])
	}
	return false
}

// ClearAttributes removes all of the packet's attributes.
func (p *TDataPacket) ClearAttr() {
	p.AttrItems = nil
}

// Value returns the value of the first attribute whose dictionary name matches
// the given name. nil is returned if no such attribute exists.
// Value返回根据输入的名字找到字典里值。 如果没有这样的属性，则返回nil。
func (p *TDataPacket) GetValue(name string) interface{} {
	if attr := p.FindAttr(name); attr != nil {
		return attr.AttrValue
	}
	return nil
}

func (p *TDataPacket) FindAttr(name string) *TAttribute {
	for _, attr := range p.AttrItems {
		if attrName, ok := p.Dictionary.GetName(attr.AttrId); ok && attrName == name {
			return attr
		}
	}
	return nil
}

// String returns the string representation of the value of the first attribute
// whose dictionary name matches the given name. The following rules are used
// for converting the attribute value to a string:
//
//   - If no such attribute exists with the given dictionary name, "" is
//     returned
//   - If the attribute's Codec implements AttributeStringer,
//     AttributeStringer.String(value) is returned
//   - If the value implements fmt.Stringer, value.String() is returned
//   - If the value is string, itself is returned
//   - If the value is []byte, string(value) is returned
//   - Otherwise, "" is returned
func (p *TDataPacket) GetString(name string) string {
	attr := p.FindAttr(name)
	if attr == nil {
		return ""
	}
	value := attr.AttrValue

	if codec := p.Dictionary.GetFunc(attr.AttrId); codec != nil {
		if stringer, ok := codec.(IAttributeStringer); ok {
			return stringer.String(value)
		}
	}

	if stringer, ok := value.(interface {
		String() string
	}); ok {
		return stringer.String()
	}
	if inum, ok := value.(uint32); ok {
		return fmt.Sprintf("%d", inum)
	}
	if str, ok := value.(string); ok {
		return str
	}
	if raw, ok := value.([]byte); ok {
		return string(raw)
	}
	return ""
}

// Add adds an attribute whose dictionary name matches the given name.
func (p *TDataPacket) AddAttr(name string, value interface{}) error {
	if p.FindAttr(name) != nil {
		return errors.New("属性已经存在！")
	}
	attr, err := p.Dictionary.NewAttr(name, value)
	if err != nil {
		return err
	}
	p.AttrItems = append(p.AttrItems, attr)
	return nil
}

// Set sets the value of the first attribute whose dictionary name matches the
// given name. If no such attribute exists, a new attribute is added
func (p *TDataPacket) Set(name string, value interface{}) error {
	for _, attr := range p.AttrItems {
		if attrName, ok := p.Dictionary.GetName(attr.AttrId); ok && attrName == name {
			codec := p.Dictionary.GetFunc(attr.AttrId)
			if transformer, ok := codec.(IAttributeTransformer); ok {
				transformed, err := transformer.Transform(value)
				if err != nil {
					return err
				}
				attr.AttrValue = transformed
				return nil
			}
			attr.AttrValue = value
			return nil
		}
	}
	return p.AddAttr(name, value)
}

// PAP returns the User-Name and User-Password attributes of an Access-Request
// packet.
//
// If packet's code is Access-Request, and the packet has a User-Name and
// User-Password attribute, ok is true. Otherwise, it is false.
func (p *TDataPacket) PAP() (username, password string, err error) {
	if p.Code != CodeAccessRequest {
		err = fmt.Errorf("只接收AccessRequest(Code=%d)请求包，当前为(Code=%d)", CodeAccessRequest, p.Code)
		return
	}
	user := p.GetValue("User-Name")
	if user == nil {
		err = errors.New("取属性(User-Name)失败。")
		return
	}
	pass := p.GetValue("User-Password")
	if pass == nil {
		err = errors.New("取属性(User-Password)失败。")
		return
	}
	if userStr, valid := user.(string); valid {
		username = userStr
	} else {
		err = errors.New("转换属性字符串(User-Name)失败。")
		return
	}
	if passStr, valid := pass.(string); valid {
		password = passStr
	} else {
		username = ""
		err = errors.New("转换属性字符串(User-Password)失败。")
		return
	}

	return
}

// Encode encodes the packet to wire format. If there is an error encoding the
// packet, nil and an error is returned.
func (p *TDataPacket) Encode() ([]byte, error) {
	var bufferAttrs bytes.Buffer
	for _, attr := range p.AttrItems {
		if attr == nil {
			continue
		}
		codec := p.Dictionary.GetFunc(attr.AttrId)
		wire, err := codec.Encode(p, attr.AttrValue)
		if err != nil {
			return nil, err
		}
		if len(wire) > 253 {
			return nil, errors.New("radius: encoded attribute is too long")
		}
		bufferAttrs.WriteByte(attr.AttrId)
		bufferAttrs.WriteByte(byte(len(wire) + 2))
		bufferAttrs.Write(wire)
	}

	length := 1 + 1 + 2 + 16 + bufferAttrs.Len()
	if length > maxPacketSize {
		return nil, errors.New("radius: encoded packet is too long")
	}

	var buffer bytes.Buffer
	buffer.Grow(length)
	buffer.WriteByte(byte(p.Code))
	buffer.WriteByte(p.Identifier)
	binary.Write(&buffer, binary.BigEndian, uint16(length))

	switch p.Code {
	case CodeAccessRequest, CodeStatusServer:
		buffer.Write(p.Authenticator[:])
	case CodeAccessAccept, CodeAccessReject, CodeAccountingRequest, CodeAccountingResponse, CodeAccessChallenge:
		hash := md5.New()
		hash.Write(buffer.Bytes())
		if p.Code == CodeAccountingRequest {
			var nul [16]byte
			hash.Write(nul[:])
		} else {
			hash.Write(p.Authenticator[:])
		}
		hash.Write(bufferAttrs.Bytes())
		hash.Write(p.Secret)

		var sum [md5.Size]byte
		buffer.Write(hash.Sum(sum[0:0]))
	default:
		return nil, errors.New("radius: unknown Packet code")
	}

	buffer.ReadFrom(&bufferAttrs)

	return buffer.Bytes(), nil
}
