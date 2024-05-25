package radius

import (
	"net"
	"strconv"
	"time"

	logs "github.com/tea4go/gh/log4go"
)

// Client is a RADIUS client that can send and receive packets to and from a
// RADIUS server.
type Client struct {
	// Network on which to make the connection. Defaults to "udp".
	Net string

	// Local address to use for outgoing connections (can be nil).
	LocalAddr net.Addr

	// Timeouts for various operations. Default values for each field is 10
	// seconds.
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// TODO: add ability to resend packets every X seconds

// Exchange sends the packet to the given server address and waits for a
// response. nil and an error is returned upon failure.
func (c *Client) SendPacket(packet *TDataPacket, addr string) (*TDataPacket, error) {
	connNet := c.Net
	if connNet == "" {
		connNet = "udp"
	}

	const defaultTimeout = 10 * time.Second
	dialTimeout := c.DialTimeout
	if dialTimeout == 0 {
		dialTimeout = defaultTimeout
	}

	dialer := net.Dialer{
		Timeout:   dialTimeout,
		LocalAddr: c.LocalAddr,
	}
	logs.Debug("连接RadiusServer(%s-%s),参数： Timeout = %.f, LocalAddr = %v", connNet, addr, dialTimeout.Seconds(), c.LocalAddr)
	conn, err := dialer.Dial(connNet, addr)
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(conn.LocalAddr().String())
	if err == nil {
		packet.AddAttr("NAS-IP-Address", net.ParseIP(host))
		nasPort, _ := strconv.Atoi(port)
		packet.AddAttr("NAS-Port", uint32(nasPort))
	}

	logs.Debug("发送%s", packet.String())
	packet_data, err := packet.Encode()

	if err != nil {
		return nil, err
	}

	writeTimeout := c.WriteTimeout
	if writeTimeout == 0 {
		writeTimeout = defaultTimeout
	}
	conn.SetWriteDeadline(time.Now().Add(writeTimeout))

	if _, err := conn.Write(packet_data); err != nil {
		conn.Close()
		return nil, err
	}

	logs.Debug("准备读数据 ......")
	var incoming [maxPacketSize]byte

	readTimeout := c.ReadTimeout
	if readTimeout == 0 {
		readTimeout = defaultTimeout
	}
	conn.SetReadDeadline(time.Now().Add(readTimeout))

	for {
		n, err := conn.Read(incoming[:])
		if err != nil {
			conn.Close()
			return nil, err
		}
		received, err := ParsePacket(incoming[:n], packet.Secret, packet.Dictionary)
		if err != nil {
			conn.Close()
			return received, err
		}

		logs.Debug("接收%s", received.String())
		conn.Close()
		return received, nil
	}
}
