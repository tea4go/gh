package http2tcp

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/gh/utils"
)

var ConnectTimeout time.Duration

// Client
// =================================================

type TClient struct {
	server string
	token  string
}

func NewClient(server string, token string) *TClient {
	if !strings.Contains(server, "://") {
		server = "http://" + server
	}
	return &TClient{
		server: server,
		token:  token,
	}
}

type IStdReadWriter struct {
	io.Reader
	io.Writer
}

func (c *TClient) Proxy(name, to string) error {
	if to == "" {
		return fmt.Errorf("建立本地代理连接失败，请传入本地代理地址！")
	}

	conn, err := net.DialTimeout(`tcp`, to, ConnectTimeout*time.Second)
	if err != nil {
		return fmt.Errorf("建立本地代理连接(%s)失败，%s", to, err.Error())
	}
	defer conn.Close()

	err = c.connectServer(conn, "reverse_proxy", name)
	return err
}

func (c *TClient) ProxyDaemon(name, to string) error {
	if to == "" {
		return fmt.Errorf("建立本地代理连接失败，请传入本地代理地址！")
	}
	ticker := time.NewTicker(1 * time.Second)
	for {
		logs.Notice("")
		logs.Notice("建立本地代理连接(%s) ...... %s", name, to)
		err := c.Proxy(name, to)
		if err != nil {
			logs.Error(err.Error())
		} else {
			logs.Notice("断开本地代理连接(%s) ...... %s", name, to)
		}
		<-ticker.C
	}
}

func (c *TClient) Server(listen string, ctype, to string) error {
	if listen == "" {
		return fmt.Errorf("本地监听失败，请传入本地监听端口！")
	}

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		if utils.IsAddrInUse(err) {
			return fmt.Errorf("本地监听失败，端口已经被占用。")
		} else {
			return fmt.Errorf("本地监听失败，%s", err.Error())
		}
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			return fmt.Errorf("接收客户端请求失败，%s", err.Error())
		}

		go func(conn net.Conn) {
			readdr := conn.RemoteAddr()
			logs.Notice("")
			logs.Notice("新客户端接入 ...... %s", readdr)
			defer func() {
				logs.Notice("新客户端断开 ...... %s", readdr)
				conn.Close()
			}()
			if err := c.connectServer(conn, ctype, to); err != nil {
				logs.Error("[%18s]%s", readdr, err.Error())
			}
		}(conn)
	}
}

// 通过服务器端连接：1、目标机器（例如：127.0.0.1:51234）2、代理连接(例如：home)
func (c *TClient) connectServer(clientConn net.Conn, ctype, value string) error {
	readdr := clientConn.RemoteAddr()
	u, err := url.Parse(c.server)
	if err != nil {
		return fmt.Errorf("解析服务器地址失败(%s)，%s", c.server, err.Error())
	}
	host := u.Hostname()
	port := u.Port()
	if port == `` {
		switch u.Scheme {
		case `http`:
			port = "80"
		case `https`:
			port = `443`
		default:
			return fmt.Errorf(`解析服务器地址失败，未知的协议(%s)`, u.Scheme)
		}
	}
	serverAddr := net.JoinHostPort(host, port)
BeginConnectHTTPServer:
	logs.Debug("[%18s]连接服务器地址(%s:%s) ......", readdr, host, port)
	var serverConn net.Conn
	if u.Scheme == `http` {
		serverConn, err = net.DialTimeout(`tcp`, serverAddr, ConnectTimeout*time.Second)
		if err != nil {
			return fmt.Errorf(`连接服务器(%s)失败，%s`, serverAddr, err.Error())
		}
	} else if u.Scheme == `https` {
		d := net.Dialer{Timeout: ConnectTimeout * time.Second}
		serverConn, err = tls.DialWithDialer(&d, `tcp`, serverAddr, nil)
		if err != nil {
			return fmt.Errorf(`连接服务器(%s)失败，%s`, serverAddr, err.Error())
		}
	}
	if serverConn == nil {
		return fmt.Errorf(`连接服务器(%s)失败，未知错误`, serverAddr)
	}
	defer serverConn.Close()

	v := u.Query()
	if ctype == "server" { //功能与server一样，目的通过服务器连接目标地址，server是通过服务器再连接
		v.Set(`addr`, value)
	} else if ctype == "reverse_proxy" { //通过“反向代理”
		v.Set(`reverse_proxy`, value)
	} else if ctype == "proxy" { //功能与server一样，目的通过服务器连接目标地址，proxy是连上“反向代理”连接
		v.Set(`proxy`, value)
	}
	u.RawQuery = v.Encode()
	url, _ := url.QueryUnescape(u.String())

	logs.Debug("[%18s]通过Http的协议升级(Upgrade)，调用GET请求获取TCP连接(%s)", readdr, url)
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("通过Http的协议升级(Upgrade)，调用GET请求获取TCP连接(%s)失败，%s", url, err.Error())
	}
	req.Header.Add(`Connection`, `upgrade`)
	req.Header.Add(`Upgrade`, httpHeaderUpgrade)
	req.Header.Add(`Authorization`, fmt.Sprintf(`%s %s`, authHeaderType, c.token))
	if err := req.Write(serverConn); err != nil {
		return fmt.Errorf("通过Http的协议升级(Upgrade)，调用WriteHeader失败，%s", err.Error())
	}
	bior := bufio.NewReader(serverConn)
	resp, err := http.ReadResponse(bior, req)
	if err != nil {
		return fmt.Errorf("通过Http的协议升级(Upgrade)，调用ReadResponse失败，%s", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == 502 || resp.StatusCode == 401 || resp.StatusCode == 500 || resp.StatusCode == 404 {
		//服务器没启动，忽略此错误
		serverConn.Close()
		time.Sleep(1 * time.Second)
		goto BeginConnectHTTPServer
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("服务端不支持协议升级(Upgrade)，返回码：%s\n%s", resp.Status, string(b))
	}
	logs.Debug("[%18s]服务端支持协议升级(%s)", readdr, resp.Status)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	logs.Info("[%18s]启动处理线程 ...... [Http服务器 <--> 本地]", readdr)
	go func() {
		logs.Debug("[%18s]===>处理线程 ...... [Http服务器 --> 本地]", readdr)
		var n int64
		var err error
		defer func() {
			logs.Debug("[%18s]===>关闭线程 ...... [Http服务器 --> 本地]", readdr)
			clientConn.Close()
			wg.Done()
		}()
		if n = int64(bior.Buffered()); n > 0 {
			logs.Debug("bior.Buffered() %d", n)
			nc, err := io.CopyN(clientConn, bior, n)
			if err != nil || nc != n {
				logs.Error("[%18s][Http服务器-->本地]复制未处理的缓冲数据到本地失败，%s (%d!=%d)", readdr, err.Error(), nc, n)
				return
			}
		}
		n, err = io.Copy(clientConn, serverConn)
		if err != nil {
			logs.Warning("[%18s][Http服务器-->本地]服务器连接错误，%s (%d)", readdr, utils.GetNetError(err), n)
		}
	}()

	go func() {
		logs.Debug("[%18s]===>处理线程 ...... [本地 --> Http服务器]", readdr)
		var n int64
		var err error
		defer func() {
			logs.Debug("[%18s]===>关闭线程 ...... [本地 --> Http服务器]", readdr)
			serverConn.Close()
			wg.Done()
		}()
		n, err = io.Copy(serverConn, clientConn)
		if err != nil {
			logs.Warning("[%18s][本地-->Http服务器]本地连接错误，%s (%d)", readdr, utils.GetNetError(err), n)
		}
	}()
	wg.Wait()
	time.Sleep(20 * time.Microsecond)
	logs.Info("[%18s]关闭处理线程 ...... [Http服务器 <--> 本地]", readdr)
	return nil
}

// Server
// =================================================
type TServer struct {
	token string
}

var m_conn sync.Map
var m_chan sync.Map

const (
	authHeaderType    = `HTTP2TCP`
	httpHeaderUpgrade = `upgrade`
)

func NewServer(token string) *TServer {
	return &TServer{
		token: token,
	}
}

func (s *TServer) auth(r *http.Request) bool {
	a := strings.Fields(r.Header.Get("Authorization"))
	if len(a) == 2 && a[0] == authHeaderType && a[1] == s.token {
		return true
	}
	return false
}

func (s *TServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logs.Notice("")
	url, _ := url.QueryUnescape(r.URL.String())
	logs.Info("[%18s]Http服务器接到请求(%s)", r.RemoteAddr, url)

	var err error
	var code int
	var msg string
	var atip string
	var avalue string

	code = http.StatusOK
	defer func() {
		if msg != "" {
			if code == http.StatusOK {
				logs.Notice("[%18s]%s", r.RemoteAddr, msg)
				http.Error(w, msg, code)
			} else {
				logs.Error("[%18s]%s", r.RemoteAddr, msg)
				http.Error(w, msg, code)
			}
		}
	}()

	if !s.auth(r) {
		code = http.StatusUnauthorized
		time.Sleep(5 * time.Second)
		msg = "鉴权不成功！"
		return
	}

	if upgrade := r.Header.Get(`Upgrade`); upgrade != httpHeaderUpgrade {
		code = http.StatusBadRequest
		//time.Sleep(1 * time.Second)
		msg = fmt.Sprintf(`请求格式Upgrade错误(%s)`, upgrade)
		return
	}

	var remote net.Conn
	var remote_chd chan int
	query := r.URL.Query().Get("query")                         //查询连接
	addr := r.URL.Query().Get("addr")                           //建立网络连接
	reverse_proxy := r.URL.Query().Get("reverse_proxy")         //建立反向代理连接
	del_reverse_proxy := r.URL.Query().Get("del_reverse_proxy") //关闭反向代理连接
	proxy := r.URL.Query().Get("proxy")                         //建立与反向代理桥接
	if addr != "" {
		atip = "网络连接"
		avalue = addr
		logs.Debug("[%18s]建立网络连接(%s)", r.RemoteAddr, addr)

		remote, err = net.DialTimeout(`tcp`, addr, ConnectTimeout*time.Second)
		if err != nil {
			code = http.StatusBadRequest
			msg = fmt.Sprintf("建立网络连接(%s)失败，%s", addr, err.Error())
			return
		}
	} else if reverse_proxy != "" {
		atip = "反向代理连接"
		avalue = reverse_proxy
		logs.Debug("[%18s]建立反向代理连接(%s) ......", r.RemoteAddr, reverse_proxy)

		//首先关闭老的连接
		var chd chan int
		v, ok := m_chan.Load(reverse_proxy)
		m_conn.Delete(reverse_proxy)
		m_chan.Delete(reverse_proxy)

		if ok {
			logs.Debug("[%18s]关闭反向代理连接(%s) ...... OK", r.RemoteAddr, reverse_proxy)
			chd = v.(chan int)
			<-chd
		}
		time.Sleep(20 * time.Microsecond)

		ch := make(chan int)
		w.Header().Add(`Content-Length`, `0`)
		w.WriteHeader(http.StatusSwitchingProtocols)

		//hijack方法让我们可以从响应(Response)中拿到这个 TCP 连接。
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			code = http.StatusInternalServerError
			msg = fmt.Sprintf("建立反向代理连接(%s)失败，从响应(Response)中拿到TCP连接错误，%s", reverse_proxy, err.Error())
			return
		}
		m_conn.Store(reverse_proxy, conn)
		m_chan.Store(reverse_proxy, ch)

		logs.Notice("[%18s]建立反向代理连接(%s)成功，等待连接 ...... ", r.RemoteAddr, reverse_proxy)
		ch <- 1 //阻塞等待

		time.Sleep(20 * time.Microsecond)
		logs.Notice("[%18s]释放反向代理连接(%s)成功", r.RemoteAddr, reverse_proxy)
		m_conn.Delete(reverse_proxy)
		m_chan.Delete(reverse_proxy)
		defer conn.Close()
		defer close(ch)
		msg = ""
		return
	} else if del_reverse_proxy != "" {
		atip = "删除代理连接"
		avalue = del_reverse_proxy
		logs.Debug("[%18s]关闭反向代理连接(%s) ......", r.RemoteAddr, del_reverse_proxy)

		var chd chan int
		v, ok := m_chan.Load(del_reverse_proxy)
		m_conn.Delete(del_reverse_proxy)
		m_chan.Delete(del_reverse_proxy)

		if ok {
			msg = fmt.Sprintf("关闭反向代理连接(%s) ...... OK", del_reverse_proxy)
			chd = v.(chan int)
			<-chd
		} else {
			code = http.StatusBadRequest
			msg = fmt.Sprintf("关闭反向代理桥接(%s)失败，找不到反向代理连接。", del_reverse_proxy)
		}
		return
	} else if proxy != "" {
		atip = "反向代理桥接"
		avalue = proxy
		logs.Debug("[%18s]请求反向代理桥接(%s) ......", r.RemoteAddr, proxy)

		v, ok := m_conn.Load(proxy)
		if ok {
			remote = v.(net.Conn)
			logs.Notice("[%18s]请求反向代理桥接(%s) ...... OK", r.RemoteAddr, proxy)
		} else {
			code = http.StatusBadRequest
			msg = fmt.Sprintf("请求反向代理桥接(%s)失败，找不到反向代理连接。", proxy)
			return
		}
		v, ok = m_chan.Load(proxy)
		if ok {
			remote_chd = v.(chan int)
		}
	} else if query != "" {
		atip = "查询连接"
		logs.Debug("[%18s]请求查询连接 ......", r.RemoteAddr)

		conns := make([]string, 0)
		m_chan.Range(func(key, value interface{}) bool {
			conns = append(conns, value.(string))
			return true
		})
		code = http.StatusOK
		msg = utils.GetJson(conns)
		return
	} else {
		code = http.StatusBadRequest
		msg = `请求未知动作！`
		return
	}
	defer func() {
		if addr != "" {
			logs.Debug("[%18s]断开目标地址连接(%s)]", r.RemoteAddr, addr)
			remote.Close()
		}
	}()

	w.Header().Add(`Content-Length`, `0`)
	w.WriteHeader(http.StatusSwitchingProtocols)

	//hijack方法让我们可以从响应(Response)中拿到这个 TCP 连接。
	conn, bio, err := w.(http.Hijacker).Hijack()
	if err != nil {
		code = http.StatusInternalServerError
		msg = fmt.Sprintf("从响应(Response)中拿到TCP连接失败，%s", err.Error())
		return
	}
	defer conn.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		var tip string
		if proxy != "" {
			tip = fmt.Sprintf("Http服务器 --> 代理连接(%s)", proxy)
		} else {
			tip = fmt.Sprintf("Http服务器 --> 目标机器(%s)", addr)
		}
		logs.Debug("[%18s]启动处理线程 ...... [%s]", r.RemoteAddr, tip)
		var n int64
		var err error
		defer func() {
			logs.Debug("[%18s]关闭处理线程 ...... [%s]", r.RemoteAddr, tip)
			if remote_chd != nil {
				<-remote_chd
			}
			wg.Done()
		}()

		// 返回的 bufio.Reader 可能包含来自客户端的未处理缓冲数据。
		// 将它们复制到 remote 以便我们可以直接使用 未处理缓冲数据
		if n := int64(bio.Reader.Buffered()); n > 0 {
			nc, err := io.CopyN(remote, bio, n)
			if err != nil || nc != n {
				logs.Error("[%18s]处理缓冲数据失败(%s)，%s (%d!=%d)", r.RemoteAddr, tip, utils.GetNetError(err), nc, n)
				return
			}
		}

		// conn --> remote
		n, err = io.Copy(remote, conn)
		if err != nil {
			logs.Warning("[%18s]服务器连接错误(%s)，%s (%d)", r.RemoteAddr, tip, utils.GetNetError(err), n) //conn报错
		}
	}()

	go func() {
		var tip string
		if proxy != "" {
			tip = fmt.Sprintf("代理连接(%s) --> Http服务器", proxy)
		} else {
			tip = fmt.Sprintf("目标机器(%s) --> Http服务器", addr)
		}
		logs.Debug("[%18s]启动处理线程 ...... [%s]", r.RemoteAddr, tip)
		var n int64
		var err error
		defer func() {
			logs.Debug("[%18s]关闭处理线程 ...... [%s]", r.RemoteAddr, tip)
			conn.Close()
			wg.Done()
		}()

		// 刷新任何未写入的数据.
		if err := bio.Writer.Flush(); err != nil {
			logs.Error(`[%18s]刷新缓存数据失败(%s)，%s`, r.RemoteAddr, tip, utils.GetNetError(err))
			return
		}

		// remote --> conn
		n, err = io.Copy(conn, remote) //remote为目标机器/代理连接
		if err != nil {
			logs.Warning("[%18s]本地连接错误(%s)，%s (%d)", r.RemoteAddr, tip, utils.GetNetError(err), n)
		}
	}()

	wg.Wait()
	logs.Notice("[%18s]断开%s(%s)", r.RemoteAddr, atip, avalue)
}
