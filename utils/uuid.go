package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

// 保证每秒/每个主机/进程有16,777,216（24位）唯一ID
// - 4字节秒数，
// - 3字节机器标识符，
// - 2字节进程ID，
// - 3字节计数器，以随机值开头。
type TUUID [rawLen]byte

const (
	encodedLen = 20                                 // string encoded len
	decodedLen = 15                                 // len after base32 decoding with the padded data
	rawLen     = 12                                 // binary raw len
	encoding   = "0123456789abcdefghijklmnopqrstuv" //使用小写字母存储base32编码的自定义版本。
)

var KeyID = "rfoMzV4D8O9owOET33vJ"
var ErrInvalidID = errors.New("无效的ID")
var objectIDCounter = randInt()
var machineID = ReadMachineID(3)
var pid = os.Getpid()
var dec [256]byte

func init() {
	for i := 0; i < len(dec); i++ {
		dec[i] = 0xFF
	}
	for i := 0; i < len(encoding); i++ {
		dec[encoding[i]] = byte(i)
	}
}

// 获得主机ID
func ReadMachineID(size int) []byte {
	id := make([]byte, size)
	if hostname, err := os.Hostname(); err == nil {
		hw := md5.New()
		hw.Write([]byte(hostname))
		copy(id, hw.Sum(nil))
	} else {
		if _, randErr := rand.Reader.Read(id); randErr != nil {
			panic(fmt.Errorf("无法获取主机名也无法生成随机数，错误：%v；%v", err, randErr))
		}
	}
	return id
}

// 生成随机数
func randInt() uint32 {
	b := make([]byte, 3)
	if _, err := rand.Reader.Read(b); err != nil {
		panic(fmt.Errorf("无法生成随机数，错误：%s", err))
	}
	return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
}

func NewUUID() *TUUID {
	var id TUUID

	binary.BigEndian.PutUint32(id[:], uint32(time.Now().Unix()))

	id[4] = machineID[0]
	id[5] = machineID[1]
	id[6] = machineID[2]

	id[7] = byte(pid >> 8)
	id[8] = byte(pid)

	i := atomic.AddUint32(&objectIDCounter, 1)

	id[9] = byte(i >> 16)
	id[10] = byte(i >> 8)
	id[11] = byte(i)

	return &id
}

func UUIDFromString(id string) (*TUUID, error) {
	i := &TUUID{}
	err := i.UnmarshalText([]byte(id))
	return i, err
}

func (id TUUID) String() string {
	text, _ := id.MarshalText()
	return string(text)
}

func (id TUUID) MarshalText() ([]byte, error) {
	text := make([]byte, encodedLen)
	text[0] = encoding[id[0]>>3]
	text[1] = encoding[(id[1]>>6)&0x1F|(id[0]<<2)&0x1F]
	text[2] = encoding[(id[1]>>1)&0x1F]
	text[3] = encoding[(id[2]>>4)&0x1F|(id[1]<<4)&0x1F]
	text[4] = encoding[id[3]>>7|(id[2]<<1)&0x1F]
	text[5] = encoding[(id[3]>>2)&0x1F]
	text[6] = encoding[id[4]>>5|(id[3]<<3)&0x1F]
	text[7] = encoding[id[4]&0x1F]
	text[8] = encoding[id[5]>>3]
	text[9] = encoding[(id[6]>>6)&0x1F|(id[5]<<2)&0x1F]
	text[10] = encoding[(id[6]>>1)&0x1F]
	text[11] = encoding[(id[7]>>4)&0x1F|(id[6]<<4)&0x1F]
	text[12] = encoding[id[8]>>7|(id[7]<<1)&0x1F]
	text[13] = encoding[(id[8]>>2)&0x1F]
	text[14] = encoding[(id[9]>>5)|(id[8]<<3)&0x1F]
	text[15] = encoding[id[9]&0x1F]
	text[16] = encoding[id[10]>>3]
	text[17] = encoding[(id[11]>>6)&0x1F|(id[10]<<2)&0x1F]
	text[18] = encoding[(id[11]>>1)&0x1F]
	text[19] = encoding[(id[11]<<4)&0x1F]
	return text, nil
}

func (id *TUUID) UnmarshalText(text []byte) error {
	if len(text) != encodedLen {
		return ErrInvalidID
	}
	for _, c := range text {
		if dec[c] == 0xFF {
			return ErrInvalidID
		}
	}

	id[0] = dec[text[0]]<<3 | dec[text[1]]>>2
	id[1] = dec[text[1]]<<6 | dec[text[2]]<<1 | dec[text[3]]>>4
	id[2] = dec[text[3]]<<4 | dec[text[4]]>>1
	id[3] = dec[text[4]]<<7 | dec[text[5]]<<2 | dec[text[6]]>>3
	id[4] = dec[text[6]]<<5 | dec[text[7]]
	id[5] = dec[text[8]]<<3 | dec[text[9]]>>2
	id[6] = dec[text[9]]<<6 | dec[text[10]]<<1 | dec[text[11]]>>4
	id[7] = dec[text[11]]<<4 | dec[text[12]]>>1
	id[8] = dec[text[12]]<<7 | dec[text[13]]<<2 | dec[text[14]]>>3
	id[9] = dec[text[14]]<<5 | dec[text[15]]
	id[10] = dec[text[16]]<<3 | dec[text[17]]>>2
	id[11] = dec[text[17]]<<6 | dec[text[18]]<<1 | dec[text[19]]>>4

	return nil
}

func (id TUUID) GetTime() time.Time {
	secs := int64(binary.BigEndian.Uint32(id[0:4]))
	return time.Unix(secs, 0)
}

func (id TUUID) GetMachine() []byte {
	return id[4:7]
}

func (id TUUID) GetPID() uint16 {
	return binary.BigEndian.Uint16(id[7:9])
}
