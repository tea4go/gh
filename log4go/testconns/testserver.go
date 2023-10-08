package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	fmt.Println("Start Log4go Server ......")
	ln, err := net.Listen("tcp", "0.0.0.0:6002")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			//fmt.Println("网络读取错误，", err)
			return
		}
		if !strings.Contains(message, "{HeartBeat}") {
			fmt.Printf("%v", message)
		}
	}
}
