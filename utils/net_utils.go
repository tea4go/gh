package utils

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sort"
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

func IsAddrInUse(err error) bool {
	opErr, ok := err.(*net.OpError)
	if !ok {
		return false
	}

	// 端口被占用错误的错误码是 "address already in use"
	if opErr.Err.Error() == "address already in use" {
		return true
	}
	if strings.Contains(err.Error(), "Only one usage of each") {
		return true
	}
	return false
}

func IsNetClose(err error) bool {
	if strings.Contains(err.Error(), "closed network connection") {
		return true
	}
	if strings.Contains(err.Error(), "was forcibly closed by the remote host") {
		return true
	}
	if strings.Contains(err.Error(), "broken pipe") {
		return true
	}
	return false
}

// 在日志库里有一个相同代码，需要同步修改。
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
		if opErr.Err.Error() == "address already in use" {
			return "端口已经占用"
		}
		switch t := opErr.Err.(type) {
		case *net.DNSError:
			return "域名解析错误"
		case *os.SyscallError:
			if errno, ok := t.Err.(syscall.Errno); ok {
				switch errno {
				case syscall.ECONNREFUSED:
					return fmt.Sprintf("连接被拒绝")
				case syscall.ETIMEDOUT:
					return fmt.Sprintf("网络连接超时")
				}
			}
		}
	}

	if strings.Contains(err.Error(), "use of closed network connection") {
		return "监听端口已关闭"
	}

	if strings.Contains(err.Error(), "unable to authenticate") {
		return "无法用户密码验证"
	}

	if strings.Contains(err.Error(), "closed network connection") {
		return "使用已关闭网络连接"
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

	if strings.Contains(err.Error(), "broken pipe") {
		return "对端已关闭连接"
	}

	if strings.Contains(err.Error(), "i/o timeout") {
		return "网络连接超时"
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

func GetAllIPAdress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	ip_addrs := make([]string, 0)
	for _, value := range addrs {
		if ipnet, ok := value.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil && !strings.Contains(ipnet.IP.String(), "169") {
				ip_addrs = append(ip_addrs, ipnet.IP.String())
			}
		}
	}

	sort.Strings(ip_addrs)
	return strings.Join(ip_addrs, "; ")
}

func GetAllMacAdress() string {
	addrs, err := net.Interfaces()
	if err != nil {
		return ""
	}
	mac_addrs := make([]string, 0)
	for _, value := range addrs {
		if value.HardwareAddr != nil {
			if !strings.Contains(value.Flags.String(), "running") {
				continue
			}
			if strings.Contains(value.Name, "vEthernet") {
				continue
			}
			ip_addr, _ := value.Addrs()

			ip_addr_text := ""
			for _, ip := range ip_addr {
				if strings.Contains(ip.String(), "::") {
					continue
				}
				if ip_addr_text == "" {
					ip_addr_text = ip.String()
				} else {
					ip_addr_text = ip_addr_text + "," + ip.String()
				}
			}
			if ip_addr_text == "" {
				continue
			}

			mac_addrs = append(mac_addrs, fmt.Sprintf("%s|%s|%s", value.Name, ip_addr_text, value.HardwareAddr.String()))
		}
	}

	sort.Strings(mac_addrs)
	return strings.Join(mac_addrs, "; ")
}
