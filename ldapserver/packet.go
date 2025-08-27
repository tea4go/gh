package ldapserver

import (
	"bufio"
	"fmt"

	ldap "github.com/openstandia/goldap/message"
	logs "github.com/tea4go/gh/log4go"
)

type messagePacket struct {
	bytes []byte
}

func readMessagePacket(br *bufio.Reader) (*messagePacket, error) {
	var err error
	var bytes *[]byte
	bytes, err = readLdapMessageBytes(br)

	if err == nil {
		messagePacket := &messagePacket{bytes: *bytes}
		return messagePacket, err
	}
	return &messagePacket{}, err

}

func (msg *messagePacket) readMessage() (m ldap.LDAPMessage, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("收到无效数据包(Hex=%x,%#v)", msg.bytes, r)
		}
	}()

	return decodeMessage(msg.bytes)
}

func decodeMessage(bytes []byte) (ret ldap.LDAPMessage, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("解析数据包失败，%v", e)
		}
	}()
	zero := 0
	ret, err = ldap.ReadLDAPMessage(ldap.NewBytes(zero, bytes))
	return
}

// BELLOW SHOULD BE IN ROOX PACKAGE

func readLdapMessageBytes(br *bufio.Reader) (ret *[]byte, err error) {
	var bytes []byte
	var tagAndLength ldap.TagAndLength
	tagAndLength, err = readTagAndLength(br, &bytes)
	if err != nil {
		return
	}
	readBytes(br, &bytes, tagAndLength.Length)
	return &bytes, err
}

// readTagAndLength parses an ASN.1 tag and length pair from a live connection
// into a byte slice. It returns the parsed data and the new offset. SET and
// SET OF (tag 17) are mapped to SEQUENCE and SEQUENCE OF (tag 16) since we
// don't distinguish between ordered and unordered objects in this code.
func readTagAndLength(conn *bufio.Reader, bytes *[]byte) (ret ldap.TagAndLength, err error) {
	// offset = initOffset
	//b := bytes[offset]
	//offset++
	var b byte
	b, err = readBytes(conn, bytes, 1)
	if err != nil {
		return
	}
	ret.Class = int(b >> 6)
	ret.IsCompound = b&0x20 == 0x20
	ret.Tag = int(b & 0x1f)

	//	// If the bottom five bits are set, then the tag number is actually base 128
	//	// encoded afterwards
	//	if ret.tag == 0x1f {
	//		ret.tag, err = parseBase128Int(conn, bytes)
	//		if err != nil {
	//			return
	//		}
	//	}
	// We are expecting the LDAP sequence tag 0x30 as first byte
	// TO-DO: tonyQQ
	// if b != 0x30 {
	// 	err = fmt.Errorf("packet error: expecting 0x30 as first byte, got %#x instead", b)
	// 	return
	// }

	b, err = readBytes(conn, bytes, 1)
	if err != nil {
		return
	}
	if b&0x80 == 0 {
		// The length is encoded in the bottom 7 bits.
		ret.Length = int(b & 0x7f)
	} else {
		// Bottom 7 bits give the number of length bytes to follow.
		numBytes := int(b & 0x7f)
		if numBytes == 0 {
			err = ldap.SyntaxError{"indefinite length found (not DER)"}
			return
		}
		ret.Length = 0
		for i := 0; i < numBytes; i++ {

			b, err = readBytes(conn, bytes, 1)
			if err != nil {
				return
			}
			if ret.Length >= 1<<23 {
				// We can't shift ret.length up without
				// overflowing.
				err = ldap.StructuralError{"溢出错误(length too large)"}
				return
			}
			ret.Length <<= 8
			ret.Length |= int(b)
			// Compat some lib which use go-ldap or someone else,
			// they encode int may have leading zeros when it's greater then 127
			// if ret.Length == 0 {
			// 	// DER requires that lengths be minimal.
			// 	err = ldap.StructuralError{"superfluous leading zeros in length"}
			// 	return
			// }
		}
	}

	return
}

// 从连接中读取“长度”字节
// 将读取的字节附加到“bytes”
// 返回最后读取的字节
func readBytes(conn *bufio.Reader, bytes *[]byte, length int) (b byte, err error) {
	newbytes := make([]byte, length)
	n, err := conn.Read(newbytes)
	if err != nil {
		return
	} else if n != length {
		if n > 0 {
			logs.Debug("从数据中读取长度(%d)，与数据包长度(%d)不一致。", n, length)
		}
	}
	*bytes = append(*bytes, newbytes...)
	b = (*bytes)[len(*bytes)-1]
	return
}
