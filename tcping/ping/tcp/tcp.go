package tcp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http/httptrace"
	"time"

	"github.com/tea4go/gh/tcping/ping"
)

var _ ping.IPing = (*TPing)(nil)

func New(host string, port int, op *ping.TOption, tls bool) *TPing {
	return &TPing{
		tls:    tls,
		host:   host,
		port:   port,
		option: op,
		dialer: &net.Dialer{
			Resolver: op.Resolver,
		},
	}
}

type TPing struct {
	option *ping.TOption
	host   string
	port   int
	dialer *net.Dialer
	tls    bool
}

func (p *TPing) SetTarget(t *ping.TTarget) {
	t.Protocol = ping.TCP
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
	var dnsStart time.Time
	// trace dns query
	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			stats.DNSDuration = time.Since(dnsStart)
		},
	})

	start := time.Now()
	var (
		conn    net.Conn
		err     error
		tlsConn *tls.Conn
		tlsErr  error
	)
	if p.tls {
		tlsConn, err = tls.DialWithDialer(p.dialer, "tcp", fmt.Sprintf("%s:%d", p.host, p.port), &tls.Config{
			InsecureSkipVerify: true,
		})
		if err == nil {
			conn = tlsConn.NetConn()
		} else {
			tlsErr = err
			conn, err = p.dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", p.host, p.port))
		}
	} else {
		conn, err = p.dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", p.host, p.port))
	}
	stats.Duration = time.Since(start)
	if err != nil {
		stats.Error = err
		if oe, ok := err.(*net.OpError); ok && oe.Addr != nil {
			stats.Address = oe.Addr.String()
		}
	} else {
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
		} else if p.tls {
			stats.Extra = bytes.NewBufferString("警告：此端口不是SSL/TLS协议，" + ping.FormatError(tlsErr) + "！")
		}
	}
	return &stats
}
