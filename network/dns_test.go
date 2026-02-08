package network

import (
	"fmt"
	"testing"
)

func TestGetIPByDomain(t *testing.T) {
	fmt.Println("=== 开始测试: 域名解析 (GetIPByDomain) ===")
	domain := "www.baidu.com"
	
	// 测试默认 DNS
	name, ips, server, err := GetIPByDomain(domain)
	if err != nil {
		t.Errorf("解析域名 %s 失败: %v", domain, err)
	} else {
		t.Logf("域名: %s, IPs: %v, DNS服务器: %s", name, ips, server)
		if len(ips) == 0 {
			t.Error("解析成功但未返回IP")
		}
	}

	// 测试指定 DNS (114.114.114.114)
	fmt.Println("=== 测试指定 DNS ===")
	name, ips, server, err = GetIPByDomain(domain, "114.114.114.114")
	if err != nil {
		t.Logf("指定DNS解析失败 (可能是网络限制): %v", err)
	} else {
		t.Logf("域名: %s, IPs: %v, DNS服务器: %s", name, ips, server)
	}
}
