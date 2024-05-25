package ldapserver

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	proxyproto "github.com/pires/go-proxyproto"
	logs "github.com/tea4go/gh/log4go"
)

// Server is an LDAP server.
type Server struct {
	Listener      net.Listener
	proxyListener *proxyproto.Listener
	ReadTimeout   time.Duration  // optional read timeout
	WriteTimeout  time.Duration  // optional write timeout
	wg            sync.WaitGroup // group of goroutines (1 by client)
	chDone        chan bool      // Channel Done, value => shutdown
	client_nets   []net.IPNet
	client_ips    sync.Map
	// OnNewConnection, if non-nil, is called on new connections.
	// If it returns non-nil, the connection is closed.
	onNewConnection func(c net.Conn) error

	// Handler handles ldap message received from client
	// it SHOULD "implement" RequestHandler interface
	Handler Handler
}

// NewServer return a LDAP Server
func NewServer() *Server {
	return &Server{
		chDone:      make(chan bool),
		client_nets: make([]net.IPNet, 0),
	}
}

// Handle registers the handler for the server.
// If a handler already exists for pattern, Handle panics
func (s *Server) Handle(h Handler) {
	if s.Handler != nil {
		logs.Emergency("注册请求处理程序错误：多次注册")
		os.Exit(1)
	}
	s.Handler = h
}

// ListenAndServe listens on the TCP network address s.Addr and then
// calls Serve to handle requests on incoming connections.  If
// s.Addr is blank, ":389" is used.
func (s *Server) ListenAndServe(addr string, ch chan error, options ...func(*Server)) {
	if addr == "" {
		addr = "0.0.0.0:389"
	}
	var e error
	s.Listener, e = net.Listen("tcp", addr)

	if e != nil {
		ch <- fmt.Errorf("开始监听失败，%s", e)
		return
	}

	if ch != nil {
		close(ch)
	}

	logs.Debug("开启服务监听 ...... %s", addr)

	for _, option := range options {
		option(s)
	}
	s.proxyListener = &proxyproto.Listener{Listener: s.Listener}
	s.serve()
}

// ListenAndServeTLS doing the same as ListenAndServe,
// but uses tls.Listen instead of net.Listen. If
// s.Addr is blank, ":686" is used.
func (s *Server) ListenAndServeTLS(addr string, certFile string, keyFile string, ch chan error, options ...func(*Server)) {
	if addr == "" {
		addr = "0.0.0.0:686"
	}

	cert, e := tls.LoadX509KeyPair(certFile, keyFile)
	if e != nil {
		ch <- fmt.Errorf("创建证书链时失败，%s", e.Error())
		return
	}

	tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionSSL30, MaxVersion: tls.VersionTLS12}
	s.Listener, e = tls.Listen("tcp", addr, &tlsConfig)
	if e != nil {
		ch <- fmt.Errorf("开始监听失败，%s", e)
		return
	}

	if ch != nil {
		close(ch)
	}

	logs.Debug("开启服务监听（TLS） ...... %s", addr)

	for _, option := range options {
		option(s)
	}
	s.proxyListener = &proxyproto.Listener{Listener: s.Listener}
	s.serve()
}

// Handle requests messages on the listener
func (s *Server) serve() {
	defer s.Listener.Close()

	if s.Handler == nil {
		logs.Emergency("error handling request messages: no request handler defined")
		os.Exit(1)
	}

	i := 0

	for {
		select {
		case <-s.chDone:
			logs.Debug("stopping server")
			s.Listener.Close()
			return
		default:
		}

		rw, err := s.proxyListener.Accept()

		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			logs.Error("创建本地连接失败，%s", err.Error())
		}

		//判断是否白名单
		re_addr := rw.RemoteAddr()
		if s.CheckClient(re_addr.String()) == false {
			time.Sleep(1000)
			rw.Close()
			logs.Warning("创建本地连接失败，客户端(%s)未授权访问。", re_addr)
			continue
		}

		if s.ReadTimeout != 0 {
			rw.SetReadDeadline(time.Now().Add(s.ReadTimeout))
		}
		if s.WriteTimeout != 0 {
			rw.SetWriteDeadline(time.Now().Add(s.WriteTimeout))
		}

		cli := s.newClient(rw)

		i = i + 1
		cli.Numero = i
		logs.Debug("创建本地连接成功[%d] %s", cli.Numero, re_addr)
		s.wg.Add(1)
		go cli.serve()
	}
}

func (s *Server) GetClientNets() []net.IPNet {
	return s.client_nets
}

func (s *Server) GetClientIPs() *sync.Map {
	return &s.client_ips
}

func (s *Server) IsClient(ip_text string) bool {
	for _, k := range s.client_nets {
		if k.String() == ip_text {
			return true
		}
	}
	return false
}

func (s *Server) GetClients() string {
	t := ""
	for _, k := range s.client_nets {
		if t == "" {
			t = k.String()
		} else {
			t = t + ";" + k.String()
		}
	}
	return t
}

func (s *Server) SetClients(ip_text string) error {
	for _, client := range strings.Split(ip_text, ";") {
		_, subnet, err := net.ParseCIDR(client)
		if err != nil {
			ip := net.ParseIP(client)
			if ip == nil {
				return fmt.Errorf("不合法的CIDR格式或IP地址(" + client + ")")
			} else {
				client = client + "/32"
			}
			_, subnet, err = net.ParseCIDR(client)
			if err != nil {
				return fmt.Errorf("不能解析CIDR格式或IP地址(" + client + ")")
			}
		}

		if s.IsClient(subnet.String()) == false {
			s.client_nets = append(s.client_nets, *subnet)
		}
	}
	return nil
}

func (s *Server) CheckClient(ip_text string) bool {
	ips := strings.Split(ip_text, ":")
	if len(ips) == 2 {
		ip_text = ips[0]
	}
	t, ok := s.client_ips.Load(ip_text)
	if ok {
		value := t.(int)
		s.client_ips.Store(ip_text, value+1)
	} else {
		s.client_ips.Store(ip_text, 1)
	}

	ip := net.ParseIP(ip_text)
	if ip == nil {
		return false
	}

	for _, k := range s.client_nets {
		if k.Contains(ip) {
			return true
		}
	}
	return false
}

// Return a new session with the connection
// client has a writer and reader buffer
func (s *Server) newClient(rwc net.Conn) (c *client) {
	c = &client{
		srv: s,
		rwc: rwc,
		br:  bufio.NewReader(rwc),
		bw:  bufio.NewWriter(rwc),
	}
	return c
}

// Termination of the LDAP session is initiated by the server sending a
// Notice of Disconnection.  In this case, each
// protocol peer gracefully terminates the LDAP session by ceasing
// exchanges at the LDAP message layer, tearing down any SASL layer,
// tearing down any TLS layer, and closing the transport connection.
// A protocol peer may determine that the continuation of any
// communication would be pernicious, and in this case, it may abruptly
// terminate the session by ceasing communication and closing the
// transport connection.
// In either case, when the LDAP session is terminated.
func (s *Server) Stop() {
	close(s.chDone)
	logs.Debug("优雅地关闭客户端连接")
	s.wg.Wait()
	logs.Debug("所有的客户端连接已关闭")
}
