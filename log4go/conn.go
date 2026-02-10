// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logs

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/net/proxy"
	//ffmt "gopkg.in/ffmt.v1"
)

// connWriter implements LoggerInterface.
// it writes messages in keep-live tcp connection.
type connWriter struct {
	mu           sync.Mutex
	lgconn       net.Conn
	conn_timeout time.Duration
	rw_timeout   time.Duration
	Net          string `json:"net"`
	Addr         string `json:"addr"`
	Level        int    `json:"level"`
	Name         string `json:"name"`
	ColorFlag    bool   `json:"color"` //this filed is useful only when system's terminal supports color
}

// NewConn create new ConnWrite returning as LoggerInterface.
func NewConn() ILogger {
	conn := new(connWriter)
	conn.Net = "tcp"
	conn.Level = LevelNotice
	conn.ColorFlag = true
	conn.conn_timeout = 5 * time.Second
	conn.rw_timeout = 3 * time.Second
	return conn
}

func (c *connWriter) connect(tos ...time.Duration) error {
	tos = append(tos, c.conn_timeout)
	to := tos[0]
	proxyAddr := GetParamString("log_http_proxy", "", "")
	FDebug("Connect() : 连接日志服务器(%s://%s) %s", c.Net, c.Addr, proxyAddr)

	c.Destroy()

	var conn net.Conn
	var err error
	if proxyAddr != "" {
		dialer := &net.Dialer{
			Timeout:   to,
			KeepAlive: 30 * time.Second,
		}
		dialer_proxy, err := proxy.SOCKS5("tcp", proxyAddr, nil, dialer)
		if err != nil {
			FDebug("Connect() : 设置代理服务器失败(%s)，%s", proxyAddr, GetNetError(err))
			conn, err = net.DialTimeout(c.Net, c.Addr, to)
		} else {
			conn, err = dialer_proxy.Dial(c.Net, c.Addr)
		}
	} else {
		conn, err = net.DialTimeout(c.Net, c.Addr, to)
	}
	if err != nil {
		FDebug("Connect() : 连接日志服务器(%s://%s) ...... %s", c.Net, c.Addr, GetNetError(err))
		return err
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
	}

	//设置日志名称
	if c.Name != "" {
		msg := fmt.Sprintf("{LogName}%s{LogName}\n", c.Name)
		_, err = conn.Write([]byte(msg))
		if err != nil {
			conn.Close()
			return err
		}
	}

	c.mu.Lock()
	c.lgconn = conn
	c.mu.Unlock()

	//FDebug("Connect() : 连接日志服务器(%s://%s) ...... OK", c.Net, c.Addr)
	return nil
}

func (c *connWriter) connect_tcp_proxy(tos ...time.Duration) error {
	tos = append(tos, c.conn_timeout)
	to := tos[0]
	FDebug("Connect() : 连接日志服务器(%s://%s)", c.Net, c.Addr)

	c.Destroy()

	var conn net.Conn
	var err error
	// 设置代理服务器的地址和端口
	proxyAddr := GetParamString("log_socks5_proxy", "", "192.168.3.164:32129")
	if proxyAddr != "" {
		dialer := &net.Dialer{
			Timeout:   to,
			KeepAlive: 30 * time.Second,
		}
		dialer_proxy, err := proxy.SOCKS5("tcp", proxyAddr, nil, dialer)
		if err != nil {
			FDebug("Connect() : 设置代理服务器失败(%s)，%s", proxyAddr, GetNetError(err))
		}
		conn, err = dialer_proxy.Dial(c.Net, c.Addr)
	} else {
		conn, err = net.DialTimeout(c.Net, c.Addr, to)
	}
	if err != nil {
		FDebug("Connect() : 连接日志服务器(%s://%s) ...... %s", c.Net, c.Addr, GetNetError(err))
		return err
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
	}

	c.mu.Lock()
	c.lgconn = conn
	c.mu.Unlock()

	//FDebug("Connect() : 连接日志服务器(%s://%s) ...... OK", c.Net, c.Addr)
	return nil
}

// Init init connection writer with json config.
// json config only need key "level".
func (c *connWriter) Init(jsonConfig string) error {
	if len(jsonConfig) == 0 {
		FDebug("InitLogger(%s,conn,color=%v) : %s", GetLevelName(c.Level), c.ColorFlag, jsonConfig)
		return nil
	}
	err := json.Unmarshal([]byte(jsonConfig), c)
	if err != nil {
		FDebug("InitLogger(%s,conn,color=%v) : %s", GetLevelName(c.Level), c.ColorFlag, jsonConfig)
		return err
	}
	FDebug("InitLogger(%s,conn,color=%v) : %s", GetLevelName(c.Level), c.ColorFlag, jsonConfig)

	//第一次连接
	c.connect(1 * time.Second)

	go func() {

		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()

		for range ticker.C {
			if c.lgconn == nil {
				c.connect()
			} else {
				err := c.writeMsgByConn("{HeartBeat}\n")
				if err != nil {
					FDebug("WriteLogger() : %s", GetNetError(err))
					c.connect()
				}
			}
		}

	}()

	return nil
}

func (c *connWriter) writeMsgByConn(msg string) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lgconn != nil {
		_, err = c.lgconn.Write([]byte(msg))
	}
	return err
}

// WriteMsg 写入消息
func (c *connWriter) WriteMsg(fileName string, fileLine int, callLevel int, callFunc string, logLevel int, when time.Time, msg string) error {
	if logLevel > c.Level {
		return nil
	}

	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[0 : len(msg)-1]
	}
	msg = msg + "\n"
	if logLevel != LevelPrint {
		head := fmt.Sprintf("(%s:%d)", fileName, fileLine)
		msg = fmt.Sprintf("%s %-25s %s> %s", when.Format("15.04.05"), head, levelPrefix[logLevel], msg)
	}
	if c.ColorFlag {
		msg = colors[logLevel](msg)
	}

	if c.lgconn != nil {
		err := c.writeMsgByConn(msg)
		if err != nil {
			c.Destroy()
		}
		//time.Sleep(100 * time.Microsecond)
	}
	return nil
}

// Flush implementing method. empty.
func (c *connWriter) Flush() {

}

// Destroy destroy connection writer and close tcp listener.
func (c *connWriter) Destroy() {
	if c.lgconn != nil {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.lgconn.Close()
		c.lgconn = nil
	}
}

// SetLevel 设置日志级别
func (w *connWriter) SetLevel(l int) {
	w.Level = l
}

// GetLevel 获取日志级别
func (w *connWriter) GetLevel() int {
	return w.Level
}

func init() {
	Register(AdapterConn, NewConn)
}
