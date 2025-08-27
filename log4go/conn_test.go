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
	"bufio"
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

// ConnTCPListener takes a TCP listener and accepts n TCP connections
// Returns connections using connChan
func connTCPListener(t *testing.T, n int, ln net.Listener, connChan chan<- net.Conn) {
	FDebug("Listener ....")
	// Listen and accept n incoming connections
	for i := 0; i < n; i++ {
		conn, err := ln.Accept()
		if err != nil {
			t.Log("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		FDebug("Add Client ....")
		// Send accepted connection to channel
		connChan <- conn
	}
	ln.Close()
	close(connChan)
}

// need to rewrite this test, it's not stable
func TestReconnect(t *testing.T) {
	//t.SkipNow()
	// Setup connection listener
	newConns := make(chan net.Conn)
	connNum := 2
	ln, err := net.Listen("tcp", "0.0.0.0:6002")
	if err != nil {
		t.Log("Error listening:", err.Error())
		os.Exit(1)
	}
	go connTCPListener(t, connNum, ln, newConns)
	time.Sleep(1 * time.Second)

	// Setup logger
	log := NewLogger()
	log.SetLogger(AdapterConn, fmt.Sprintf(`{"net":"tcp","level":%d,"addr":"127.0.0.1:6002"}`, LevelDebug))
	log.Info("测试日志A（Info）")

	// Refuse first connection
	first := <-newConns
	message, err := bufio.NewReader(first).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("接收到日志：%s", string(message))
	first.Close()

	time.Sleep(1 * time.Second)

	// Send another log after conn closed
	FDebug("========================================")
	log.Info("测试日志B（Info）")
	message, err = bufio.NewReader(first).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("接收到日志：%s", string(message))

	// Check if there was a second connection attempt
	select {
	case second := <-newConns:
		message, err := bufio.NewReader(second).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("接收到日志：%s", string(message))
		second.Close()
	default:
		t.Error("没有重连")
	}
}
