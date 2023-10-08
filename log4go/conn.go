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
	"io"
	"net"
	"sync"
	"time"
	//ffmt "gopkg.in/ffmt.v1"
)

// connWriter implements LoggerInterface.
// it writes messages in keep-live tcp connection.
type connWriter struct {
	mu           sync.Mutex
	lgconn       net.Conn
	lgwi         *logWriter     //网络连接读写接口
	lgwc         io.WriteCloser //网络连接关闭接口(网络连接如果需要关闭，只能通过这个，因为上面接口没有关闭功能）
	conn_timeout time.Duration
	rw_timeout   time.Duration
	Reconnect    bool   `json:"reconnect"`
	Net          string `json:"net"`
	Addr         string `json:"addr"`
	Level        int    `json:"level"`
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

func (c *connWriter) connect() error {
	FDebug("Connect() : 连接日志服务器(%s://%s)", c.Net, c.Addr)

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lgwc != nil {
		c.lgwc.Close()
		c.lgwc = nil
	}

	conn, err := net.DialTimeout(c.Net, c.Addr, c.conn_timeout)
	if err != nil {
		return err
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
	}

	c.lgwc = conn
	c.lgwi = newLogWriter(conn)
	c.lgconn = conn
	return nil
}

// Init init connection writer with json config.
// json config only need key "level".
func (c *connWriter) Init(jsonConfig string) error {
	FDebug("InitLogger(conn,%s) : %s", GetLevelName(c.Level), jsonConfig)
	err := json.Unmarshal([]byte(jsonConfig), c)
	if err != nil {
		return err
	}

	c.connect()
	go func() {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()

		for range ticker.C {
			if c.lgwi == nil {
				c.connect()
			} else {
				_, err := c.lgwi.writeln("{HeartBeat}\n")
				if err != nil {
					c.connect()
				}
			}
		}

	}()
	return nil
}

// WriteMsg write message in connection.
// if connection is down, try to re-connect.
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
	fmt.Printf("%s", msg)
	//c.lgconn.SetDeadline(time.Now().Add(c.rw_timeout))
	//fmt.Println(c.lgconn.Write([]byte(h)))
	//c.lgwc.SetDeadline(time.Now().Add(c.rw_timeout))
	//fmt.Println(c.lgwi.writeln(h))
	if c.lgwi != nil {
		c.mu.Lock()
		defer c.mu.Unlock()
		_, err := c.lgconn.Write([]byte(msg))
		if err != nil {
			c.Destroy()
		}
	}
	return nil
}

// Flush implementing method. empty.
func (c *connWriter) Flush() {

}

// Destroy destroy connection writer and close tcp listener.
func (c *connWriter) Destroy() {
	if c.lgwc != nil {
		c.lgwc.Close()
		c.lgwc = nil
	}
}

func (w *connWriter) SetLevel(l int) {
	w.Level = l
}

func (w *connWriter) GetLevel() int {
	return w.Level
}

func init() {
	Register(AdapterConn, NewConn)
}
