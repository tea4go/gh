package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	port := flag.Int("port", 9514, "设置日志服务器监听端口。")

	flag.Usage = func() {
		fmt.Printf("使用方法： %s [参数 ...]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	fmt.Printf("Start Log4go Server ...... (0.0.0.0:%d)\n", *port)
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
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
		message := make([]byte, 1024)
		_, err := conn.Read(message)
		if err != nil {
			return
		}
		if !strings.Contains(string(message), "{HeartBeat}") {
			fmt.Printf("%s", message)
		}
	}
}
