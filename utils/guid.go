package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

const (
	ObjectGUIDAttribute = "objectGUID"
)

// GUID represents a GUID/UUID. It has the same structure as
// golang.org/x/sys/windows.GUID so that it can be used with
// functions expecting that type.
// Reference: https://github.com/Microsoft/go-winio/blob/v0.4.14/pkg/guid/guid.go
type TGUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// FromString parses a string containing a GUID and returns the GUID. The only
// format currently supported is the `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
// format.
func GUIDFromString(s string) (TGUID, error) {
	if len(s) != 36 {
		return TGUID{}, fmt.Errorf("非法的 GUID %q", s)
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return TGUID{}, fmt.Errorf("无效的  GUID %q", s)
	}

	var g TGUID

	data1, err := strconv.ParseUint(s[0:8], 16, 32)
	if err != nil {
		return TGUID{}, fmt.Errorf("无效的 GUID %q", s)
	}
	g.Data1 = uint32(data1)

	data2, err := strconv.ParseUint(s[9:13], 16, 16)
	if err != nil {
		return TGUID{}, fmt.Errorf("无效的 GUID %q", s)
	}
	g.Data2 = uint16(data2)

	data3, err := strconv.ParseUint(s[14:18], 16, 16)
	if err != nil {
		return TGUID{}, fmt.Errorf("无效的 GUID %q", s)
	}
	g.Data3 = uint16(data3)

	for i, x := range []int{19, 21, 24, 26, 28, 30, 32, 34} {
		v, err := strconv.ParseUint(s[x:x+2], 16, 8)
		if err != nil {
			return TGUID{}, fmt.Errorf("无效的 GUID %q", s)
		}
		g.Data4[i] = uint8(v)
	}

	return g, nil
}

// FromWindowsArray constructs a GUID from a Windows encoding array of bytes.
func GUIDFromBytes(bs []byte) TGUID {
	var b [16]byte
	copy(b[:], bs)
	return fromArray(b, binary.LittleEndian)
}

// FromWindowsArray constructs a GUID from a Windows encoding array of bytes.
func GUIDFromWindowsArray(b [16]byte) TGUID {
	return fromArray(b, binary.LittleEndian)
}

func fromArray(b [16]byte, order binary.ByteOrder) TGUID {
	var g TGUID
	g.Data1 = order.Uint32(b[0:4])
	g.Data2 = order.Uint16(b[4:6])
	g.Data3 = order.Uint16(b[6:8])
	copy(g.Data4[:], b[8:16])
	return g
}

func (g TGUID) toArray(order binary.ByteOrder) [16]byte {
	b := [16]byte{}
	order.PutUint32(b[0:4], g.Data1)
	order.PutUint16(b[4:6], g.Data2)
	order.PutUint16(b[6:8], g.Data3)
	copy(b[8:16], g.Data4[:])
	return b
}

func (g TGUID) String() string {
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		g.Data1,
		g.Data2,
		g.Data3,
		g.Data4[:2],
		g.Data4[2:])
}

// OctetString parses a GUID to octet string in Windows
// encoding. Format supported for AD(ObjectGUID Attribute) search
// is `\xx\xx\xx\xx\xx\xx\xx\xx\xx\xx\xx\xx\xx\xx\xx\xx`
func (g TGUID) OctetString() string {
	byteArray := g.ToWindowsArray()
	var buffer bytes.Buffer
	for _, b := range byteArray {
		buffer.WriteString("\\" + fmt.Sprintf("%02x", b))
	}
	return buffer.String()
}

// ToWindowsArray returns an array of 16 bytes representing the GUID in Windows
// encoding.
func (g TGUID) ToWindowsArray() [16]byte {
	return g.toArray(binary.LittleEndian)
}
