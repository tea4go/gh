package network

import (
	"fmt"
	"testing"
)

func TestGetMacByIP(t *testing.T) {
	fmt.Println("=== 开始测试: 通过IP获取MAC地址 (GetMacByIP) ===")
	// 测试本机回环地址（通常不会有 ARP 记录，预期失败或特定行为）
	// 注意：GetMacByIP 依赖 arp 命令或 ARP 协议，需要真实网络环境。
	// 这里仅测试函数调用不 panic，并打印结果。
	ip := "127.0.0.1"
	mac, err := GetMacByIP(ip)
	if err != nil {
		t.Logf("获取 %s 的 MAC 失败 (符合预期，如果是回环地址): %v", ip, err)
	} else {
		t.Logf("获取 %s 的 MAC 成功: %s", ip, mac)
	}
}

func TestGetPublicIP(t *testing.T) {
	fmt.Println("=== 开始测试: 获取公网IP (GetPublicIP) ===")
	// 依赖外部服务，可能失败或超时。
	ip := GetPublicIP()
	if ip == "" {
		t.Log("获取公网IP超时或失败 (可能是网络原因)")
	} else {
		t.Logf("公网IP: %s", ip)
	}
}

func TestGetPublicIPDetail(t *testing.T) {
	fmt.Println("=== 开始测试: 获取公网IP详情 (GetPublicIPDetail) ===")
	// 依赖淘宝IP库，经常不稳定。
	detail, err := GetPublicIPDetail()
	if err != nil {
		t.Logf("获取公网IP详情失败 (可能是接口不可用): %v", err)
	} else {
		t.Logf("公网IP详情: %+v", detail)
	}
}

func TestGetIPByURL(t *testing.T) {
	fmt.Println("=== 开始测试: 从URL获取IP (GetIPByURL) ===")
	// 使用一个返回IP的测试服务
	url := "http://myip.ipip.net"
	ip, err := GetIPByURL(url)
	if err != nil {
		t.Logf("从 %s 获取IP失败: %v", url, err)
	} else {
		t.Logf("从 %s 获取到的IP: %s", url, ip)
	}
}
