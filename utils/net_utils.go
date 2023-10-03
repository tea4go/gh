package utils

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
)

func GetStatusCode(resp *http.Response) int {
	if resp == nil {
		return 0
	} else if resp.Request != nil && resp.Request.Response != nil {
		return resp.Request.Response.StatusCode
	} else {
		return resp.StatusCode
	}
}

func GetNetError(err error) string {
	if err == io.EOF {
		return "网络主动断开"
	}

	netErr, ok := err.(net.Error)
	if ok {
		if netErr.Timeout() {
			return "网络连接超时"
		}
		if netErr.Temporary() {
			return "网络临时错误"
		}
	}

	opErr, ok := netErr.(*net.OpError)
	if ok {

		switch t := opErr.Err.(type) {
		case *net.DNSError:
			return "域名解析错误"
		case *os.SyscallError:
			if errno, ok := t.Err.(syscall.Errno); ok {
				switch errno {
				case syscall.ECONNREFUSED:
					return fmt.Sprintf("连接被服务器拒绝")
				case syscall.ETIMEDOUT:
					return fmt.Sprintf("网络连接超时")
				}
			}
		}
	}

	if strings.Contains(err.Error(), "closed network connection") {
		return "使用已关闭的网络连接"
	}

	if strings.Contains(err.Error(), "connection refused") {
		return "连接被拒绝"
	}

	if strings.Contains(err.Error(), "server gave HTTP response to HTTPS client") {
		return "服务器需要https访问"
	}

	if strings.Contains(err.Error(), "x509: certificate is not valid") {
		return "无效的网站证书"
	}

	if strings.Contains(err.Error(), "x509: certificate is valid") {
		return "网站证书不匹配"
	}

	if strings.Contains(err.Error(), "no such host") {
		return "网站域名不存在"
	}

	if strings.Contains(err.Error(), "actively refused it") {
		return "无法建立连接"
	}

	if strings.Contains(err.Error(), "was forcibly closed by the remote host") {
		return "远程主机强制关闭了现有连接"
	}
	return err.Error()
}

func GetIPAdress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	ip_addr := ""
	for _, value := range addrs {
		if ipnet, ok := value.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip_addr = ipnet.IP.String()
				if strings.Contains(ip_addr, "192.168.3") || strings.Contains(ip_addr, "10.45") || strings.Contains(ip_addr, "192.168.50") || strings.Contains(ip_addr, "192.168.50") {
					return ip_addr
				}
			}
		}
	}
	return ip_addr
}
