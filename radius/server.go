package radius

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	logs "github.com/tea4go/gh/log4go"
)

type Handler interface {
	ServeRadius(w ResponseWriter, p *TDataPacket)
}

type HandlerFunc func(w ResponseWriter, p *TDataPacket)

func (h HandlerFunc) ServeRadius(w ResponseWriter, p *TDataPacket) {
	h(w, p)
}

type ResponseWriter interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Write(packet *TDataPacket) error
	AccessAccept(attributes ...*TAttribute) error
	AccessReject(attributes ...*TAttribute) error
	AccessChallenge(attributes ...*TAttribute) error
	AccountingResponse(attributes ...*TAttribute) error
}

// 响应
type responseWriter struct {
	// listener that received the packet
	conn *net.UDPConn
	// where the packet came from
	addr *net.UDPAddr
	// original packet
	packet *TDataPacket
}

func (r *responseWriter) LocalAddr() net.Addr {
	return r.conn.LocalAddr()
}

func (r *responseWriter) RemoteAddr() net.Addr {
	return r.addr
}

func (r *responseWriter) accessRespond(code Code, attributes ...*TAttribute) error {
	packet := TDataPacket{
		Code:          code,
		Identifier:    r.packet.Identifier,
		Authenticator: r.packet.Authenticator,
		Secret:        r.packet.Secret,
		Dictionary:    r.packet.Dictionary,
		AttrItems:     attributes,
	}
	return r.Write(&packet)
}

func (r *responseWriter) AccessAccept(attributes ...*TAttribute) error {
	return r.accessRespond(CodeAccessAccept, attributes...)
}

func (r *responseWriter) AccessReject(attributes ...*TAttribute) error {
	return r.accessRespond(CodeAccessReject, attributes...)
}

func (r *responseWriter) AccessChallenge(attributes ...*TAttribute) error {
	return r.accessRespond(CodeAccessChallenge, attributes...)
}

func (r *responseWriter) AccountingResponse(attributes ...*TAttribute) error {
	return r.accessRespond(CodeAccountingResponse, attributes...)
}

func (r *responseWriter) Write(packet *TDataPacket) error {
	raw, err := packet.Encode()
	if err != nil {
		return err
	}
	if _, err := r.conn.WriteToUDP(raw, r.addr); err != nil {
		return err
	}
	return nil
}

// Server is a server that listens for and handles RADIUS packets.
type Server struct {
	Addr          string
	Port          int
	Network       string
	Secret        []byte
	ClientsMap    map[string]string // Client->Secret mapping
	ClientNets    []net.IPNet
	ClientSecrets [][]byte
	Dictionary    *TDictionary // Dictionary used when decoding incoming packets.
	Handler       Handler      // The packet handler that handles incoming, valid packets.
	listener      *net.UDPConn // Listener
}

func (s *Server) ResetClientNets() error {
	s.ClientNets = nil
	s.ClientSecrets = nil

	if s.ClientsMap != nil {
		for k, v := range s.ClientsMap {

			_, subnet, err := net.ParseCIDR(k)
			if err != nil {
				ip := net.ParseIP(k)
				if ip == nil {
					return errors.New("不合法的CIDR格式或IP地址(" + k + ")")
				} else {
					k = k + "/32"
				}
				_, subnet, err = net.ParseCIDR(k)
				if err != nil {
					return errors.New("不能解析CIDR格式或IP地址(" + k + ")")
				}
			}
			s.ClientNets = append(s.ClientNets, *subnet)
			s.ClientSecrets = append(s.ClientSecrets, []byte(v))
		}
	}

	return nil
}

func (s *Server) GetSecretByIPString(ipaddress string) []byte {

	ip := net.ParseIP(ipaddress)
	if ip == nil {
		return nil
	}

	return s.GetSecretByIP(ip)
}

func (s *Server) GetSecretByIP(ip net.IP) []byte {

	for i, k := range s.ClientNets {
		if k.Contains(ip) {
			return s.ClientSecrets[i]
		}
	}

	return nil
}

func (s *Server) AddClientsMap(m map[string]string) {
	for k, v := range m {
		s.ClientsMap[k] = v
	}
}

func (s *Server) ListenAndServe() error {
	if s.listener != nil {
		return errors.New("Radius Server 已经运行。")
	}

	if s.Handler == nil {
		return errors.New("Radius Server Handler is null.")
	}

	if s.ClientsMap != nil {
		// 双重检查，IP或IPNet范围
		err := s.ResetClientNets()
		if err != nil {
			return err
		}
	}

	logs.Notice("Radius服务器 ...... 正在启动")

	addrStr := fmt.Sprintf("0.0.0.0:%d", s.Port)
	if s.Addr != "" {
		addrStr = fmt.Sprintf("%s:%d", s.Addr, s.Port)
	}

	network := "udp"
	if s.Network != "" {
		network = s.Network
	}
	logs.Notice("==>监听地址： %s", addrStr)
	logs.Notice("==>网络协议： %s", network)
	logs.Notice("==>客户端IP： %v", s.ClientsMap)
	addr, err := net.ResolveUDPAddr(network, addrStr)
	if err != nil {
		return errors.New("监听地址错误，" + err.Error())
	}
	s.listener, err = net.ListenUDP(network, addr)
	if err != nil {
		return errors.New("监听连接错误，" + err.Error())
	}

	type activeKey struct {
		IP         string
		Identifier byte
	}

	var (
		activeLock sync.RWMutex
		active     = map[activeKey]bool{}
	)

	for {
		buff := make([]byte, 4096)
		n, remoteAddr, err := s.listener.ReadFromUDP(buff)
		if err != nil && !err.(*net.OpError).Temporary() {
			logs.Emergency("服务器读取网络数据错误，" + err.Error())
			break
		}

		if n == 0 {
			continue
		}

		buff = buff[:n]
		hexbuff := make([]byte, n*2)
		b := hex.Encode(hexbuff, buff)

		go func(conn *net.UDPConn, buff []byte, remoteAddr *net.UDPAddr) {
			logs.Begin()
			defer logs.End()
			logs.Debug("[%s]==>接收到%d个字符, 内容：%s", remoteAddr, n, hexbuff[:b])

			secret := s.GetSecretByIP(remoteAddr.IP)
			if secret == nil {
				logs.Warning("[%s]==>未授权客户端请求[%s]", remoteAddr, remoteAddr.IP)
				return
			}

			packet, err := ParsePacket(buff, secret, s.Dictionary)
			if err != nil {
				logs.Warning("[%s]==>解析数据包出错，错误：%s", remoteAddr, err.Error())
				return
			}
			logs.Debug(packet.String())
			key := activeKey{
				IP:         remoteAddr.String(),
				Identifier: packet.Identifier,
			}

			activeLock.RLock()
			_, ok := active[key]
			if ok {
				activeLock.RUnlock()
				logs.Warning("[%s]==>接收到的重复请求包(#%d)，当前%d个正在处理。", remoteAddr, packet.Identifier, len(active))
				t := time.NewTicker(5000 * time.Millisecond)
				exit_flag := false
				for !exit_flag {
					select {
					case <-t.C:
						activeLock.RLock()
						_, ok := active[key]
						if ok {
							activeLock.RUnlock()
							logs.Warning("[%s]==>接收到的重复请求包(#%d)，当前%d个正在处理。 ", remoteAddr, packet.Identifier, len(active))
						} else {
							active[key] = true
							activeLock.RUnlock()
							exit_flag = true
							logs.Warning("[%s]==>接收到的重复请求包(#%d)已经处理完成，目前%d个正在处理。 ", remoteAddr, packet.Identifier, len(active))
						}
					}
				}
			} else {
				active[key] = true
				activeLock.RUnlock()
			}

			response := responseWriter{
				conn:   conn,
				addr:   remoteAddr,
				packet: packet,
			}
			s.Handler.ServeRadius(&response, packet)

			activeLock.Lock()
			delete(active, key)
			activeLock.Unlock()

		}(s.listener, buff, remoteAddr)
	}

	logs.Notice("Radius服务器 ...... 正在关闭")
	s.listener = nil
	return nil
}

func (s *Server) Close() error {
	logs.Notice("Radius服务器 ...... 正在停止")
	if s.listener == nil {
		return nil
	}
	return s.listener.Close()
}
