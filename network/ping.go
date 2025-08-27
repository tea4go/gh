package network

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	icmpv4EchoRequest = 8
	icmpv4EchoReply   = 0
	icmpv6EchoRequest = 128
	icmpv6EchoReply   = 129
)

type TPingResult struct {
	Domain       string
	IPAddr       string
	DelayShort   float64
	DelayLong    float64
	DelayAverage float64
	Lost         int
}

func (Self *TPingResult) String() string {
	return fmt.Sprintf("%s ---> %.2f ms", Self.IPAddr, Self.DelayAverage)
}

type icmpMessage struct {
	Type     int
	Code     int
	Checksum int
	Ttl      int
	Body     icmpMessageBody
}

func (p *icmpMessage) String() string {
	return fmt.Sprintf("Type=%d,Code=%d,Check=%d,TTL=%d,Body={%s}", p.Type, p.Code, p.Checksum, p.Ttl, p.Body)
}

type icmpMessageBody interface {
	Len() int
	Marshal() ([]byte, error)
	GetId() int
	GetSeq() int
}

func (m *icmpMessage) Marshal() ([]byte, error) {
	b := []byte{byte(m.Type), byte(m.Code), 0, 0}
	if m.Body != nil && m.Body.Len() != 0 {
		mb, err := m.Body.Marshal()
		if err != nil {
			return nil, err
		}
		b = append(b, mb...)
	}
	switch m.Type {
	case icmpv6EchoRequest,
		icmpv6EchoReply:
		return b, nil
	}
	csumcv := len(b) - 1
	s := uint32(0)
	for i := 0; i < csumcv; i += 2 {
		s += uint32(b[i+1])<<8 | uint32(b[i])
	}
	if csumcv&1 == 0 {
		s += uint32(b[csumcv])
	}
	s = s>>16 + s&0xffff
	s = s + s>>16

	b[2] ^= byte(^s & 0xff)
	b[3] ^= byte(^s >> 8)
	return b, nil
}

func parseICMPMessage(b []byte) (*icmpMessage, error) {
	if len(b) >= 20 {
		hdrlen := int(b[0]&0x0f) << 2
		b = b[hdrlen:]
	}

	msglen := len(b)
	if msglen < 4 {
		return nil, errors.New("数据包格式错误，长度不足4个字节。")
	}
	m := &icmpMessage{
		Type:     int(b[0]),
		Code:     int(b[1]),
		Checksum: int(b[2])<<8 | int(b[3]),
		Ttl:      int(b[8]),
	}

	switch m.Type {
	case icmpv4EchoRequest,
		icmpv4EchoReply,
		icmpv6EchoRequest,
		icmpv6EchoReply:
		m.Body = parseICMPEchoBody(b[4:])
	}

	return m, nil
}

type icmpEcho struct {
	ID   int
	SEQ  int
	Data []byte
}

func (p *icmpEcho) String() string {
	return fmt.Sprintf("Id=%d,Seq=%d,Data=%s", p.ID, p.SEQ, p.Data[:16])
}

func (p *icmpEcho) Len() int {
	if p == nil {
		return 0
	}
	return 4 + len(p.Data)
}

func (p *icmpEcho) GetId() int {
	return p.ID
}

func (p *icmpEcho) GetSeq() int {
	return p.SEQ
}

func (p *icmpEcho) Marshal() ([]byte, error) {
	b := make([]byte, 4+len(p.Data))
	b[0], b[1] = byte(p.ID>>8), byte(p.ID&0xff)
	b[2], b[3] = byte(p.SEQ>>8), byte(p.SEQ&0xff)
	copy(b[4:], p.Data)
	return b, nil
}

func parseICMPEchoBody(b []byte) *icmpEcho {
	bodylen := len(b)
	p := &icmpEcho{
		ID:  int(b[0])<<8 | int(b[1]),
		SEQ: int(b[2])<<8 | int(b[3]),
	}

	if bodylen > 4 {
		p.Data = make([]byte, bodylen-4)
		copy(p.Data, b[4:])
	}
	return p
}

func connect_icmp(host string, timeout time.Duration, size int, index int64) (int, error) {
	var err error
	var conn net.Conn
	var wb, rb []byte
	var wo, ro *icmpMessage
	var ttl int

	//生成数据包
	wo = &icmpMessage{
		Type: icmpv4EchoRequest,
		Code: 0,
		Body: &icmpEcho{
			ID:   os.Getpid() & 0xffff,
			SEQ:  int(index & 0xffff),
			Data: bytes.Repeat([]byte("."), size),
		},
	}
	if wb, err = wo.Marshal(); err != nil {
		return -1, errors.New("创建ICMP数据包出错，原因：" + err.Error())
	}

	if conn, err = net.DialTimeout("ip4:icmp", host, timeout); err != nil {
		return -2, errors.New("连接服务器(" + host + ")失败，原因：" + err.Error())
	}
	defer conn.Close()

	//设置超时时间。
	conn.SetDeadline(time.Now().Add(timeout))

	//发送数据包
	if _, err = conn.Write(wb); err != nil {
		return -3, errors.New("发送ICMP数据包失败，原因：" + err.Error())
	}

	//接收数据包
	rb = make([]byte, 20+len(wb))
	for {
		if _, err = conn.Read(rb); err != nil {
			return -4, errors.New("接收ICMP数据包失败，原因：" + err.Error())
		}

		if ro, err = parseICMPMessage(rb); err != nil {
			return -5, errors.New("解析ICMP数据包失败，原因：" + err.Error())
		}

		if ro.Body.GetId() != wo.Body.GetId() || ro.Body.GetSeq() != wo.Body.GetSeq() {
			continue
		}

		switch ro.Type {
		case icmpv4EchoRequest,
			icmpv6EchoRequest:
			continue
		}

		ttl = ro.Ttl
		break
	}

	return ttl, nil
}

//输入：次数,发送包大小,网络超时
//返回：是否OK，延时，是否稳定（ping超时），错误
func Pinger(host_name string, count int, size int, timeout time.Duration) (bool, *TPingResult, error) {
	var ip_str string
	var err, out_err error
	var result TPingResult

	if size < 16 {
		size = 16
	}
	if count < 1 {
		count = 1
	}
	if timeout < 0 {
		timeout = 1
	}

	//是否能直接得出IP，如果传入是IP地址，则直接能得IP地址
	ip_flag := net.ParseIP(host_name)
	if ip_flag == nil {
		conn, err := net.DialTimeout("ip4:icmp", host_name, 3*time.Second)
		if err != nil {
			return false, &result, err
		}
		defer conn.Close()
		ip_str = conn.RemoteAddr().String()
		tt, err := net.LookupCNAME(host_name)
		if err != nil {
			return false, &result, err
		}
		host_name = tt
	} else {
		ip_str = ip_flag.String()
	}

	var sendN int64 = 0  //发送次数
	var recvN int64 = 0  //接收次数
	var lostN int = 0    //超时次数
	var shortT int64 = 0 //最短返回时间
	var longT int64 = -1 //最长返回时间
	var sumT int64 = 0   //总计返时时间

	for count > 0 {
		sendN++
		count--

		starttime := time.Now()
		_, err = connect_icmp(ip_str, timeout, size, sendN)
		if err == nil {
			endduration := time.Since(starttime).Nanoseconds()
			sumT += endduration

			recvN++
			if shortT > endduration {
				shortT = endduration
			}
			if longT < endduration {
				longT = endduration
			}
		} else {
			out_err = err
			lostN++
		}
		time.Sleep(1e9)
	}
	result.Domain = host_name
	result.IPAddr = ip_str
	result.DelayShort = float64(shortT) / 1000000.0
	result.DelayLong = float64(longT) / 1000000.0
	result.DelayAverage = float64(sumT) / 1000000.0 / float64(recvN)
	result.Lost = lostN

	return recvN > 0, &result, out_err
}
