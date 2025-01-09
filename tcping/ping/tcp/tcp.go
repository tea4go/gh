package tcp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"time"

	"github.com/tea4go/gh/tcping/ping"
)

var _ ping.IPing = (*TPing)(nil)

func NewTCP(host string, port int, op *ping.TOption) *TPing {
	return &TPing{
		host:   host,
		port:   port,
		option: op,
		dialer: &net.Dialer{
			Resolver: op.Resolver,
		},
	}
}

type TPing struct {
	dialer *net.Dialer
	host   string
	port   int
	option *ping.TOption
}

func (p *TPing) SetTarget(t *ping.TTarget) {
	t.IP = p.host
	t.Port = p.port
	t.Timeout = p.option.Timeout
}

func (p *TPing) Ping(ctx context.Context) *ping.TStats {
	timeout := ping.DefaultTimeout
	if p.option.Timeout > 0 {
		timeout = p.option.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var stats ping.TStats

	//#region HttpTrace 追踪 DNS 时间
	var dnsStart time.Time
	ctx = httptrace.WithClientTrace(ctx,
		&httptrace.ClientTrace{
			DNSStart: func(info httptrace.DNSStartInfo) {
				dnsStart = time.Now()
			},
			DNSDone: func(info httptrace.DNSDoneInfo) {
				stats.DNSDuration = time.Since(dnsStart)
			},
		})
	//#endregion

	var (
		conn    net.Conn
		err     error
		tlsConn *tls.Conn
		tlsErr  error
	)

	//#region 连接主机端口
	start := time.Now()
	addr := fmt.Sprintf("%s:%d", p.host, p.port)
	if p.option.IsTls {
		tlsConn, err = tls.DialWithDialer(p.dialer, "tcp", addr, &tls.Config{InsecureSkipVerify: true})
		if err == nil {
			conn = tlsConn.NetConn()
		} else {
			tlsErr = err
			conn, err = p.dialer.DialContext(ctx, "tcp", addr)
		}
	} else {
		conn, err = p.dialer.DialContext(ctx, "tcp", addr)
	}
	stats.Duration = time.Since(start)
	//#endregion

	if err != nil {
		stats.Error = err
		if oe, ok := err.(*net.OpError); ok && oe.Addr != nil {
			stats.Address = oe.Addr.String()
		}
	} else {
		//#region 连接成功处理统计信息
		stats.Connected = true
		stats.Address = conn.RemoteAddr().String()
		if tlsConn != nil && len(tlsConn.ConnectionState().PeerCertificates) > 0 {
			state := tlsConn.ConnectionState()
			stats.Extra = Meta{
				dnsNames:   state.PeerCertificates[0].DNSNames,
				serverName: state.ServerName,
				version:    int(state.Version - tls.VersionTLS10),
				notBefore:  state.PeerCertificates[0].NotBefore,
				notAfter:   state.PeerCertificates[0].NotAfter,
			}
		} else if p.option.IsTls {
			stats.Extra = bytes.NewBufferString("警告：此端口不是SSL/TLS协议，" + ping.FormatError(tlsErr) + "！")
		}
		//#endregion
	}
	return &stats
}

func init() {
	fmt.Println("tcping: 注册 TCP 协议 Ping")
	ping.Register(ping.TCP,
		func(url *url.URL, op *ping.TOption) (ping.IPing, error) {
			port, err := strconv.Atoi(url.Port())
			if err != nil {
				return nil, err
			}
			return NewTCP(url.Hostname(), port, op), nil
		})
}
