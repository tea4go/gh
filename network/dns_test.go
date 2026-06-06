package network

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- GetIPByDomain basic ---

func TestGetIPByDomain(t *testing.T) {
	fmt.Println("=== 测试: GetIPByDomain ===")
	domain := "www.baidu.com"

	name, ips, server, err := GetIPByDomain(domain)
	if err != nil {
		t.Logf("GetIPByDomain failed (network may be unavailable): %v", err)
	} else {
		t.Logf("域名: %s, IPs: %v, DNS服务器: %s", name, ips, server)
		if len(ips) == 0 {
			t.Log("No IPs returned")
		}
	}
}

func TestGetIPByDomainWithCustomDNS(t *testing.T) {
	fmt.Println("=== 测试: GetIPByDomain 自定义 DNS ===")
	domain := "www.baidu.com"

	name, ips, server, err := GetIPByDomain(domain, "114.114.114.114")
	if err != nil {
		t.Logf("GetIPByDomain with custom DNS failed: %v", err)
	} else {
		t.Logf("域名: %s, IPs: %v, DNS服务器: %s", name, ips, server)
	}
}

func TestGetIPByDomainWithCustomDNSPort(t *testing.T) {
	fmt.Println("=== 测试: GetIPByDomain 自定义 DNS 带端口 ===")
	domain := "www.baidu.com"

	name, ips, server, err := GetIPByDomain(domain, "114.114.114.114:53")
	if err != nil {
		t.Logf("GetIPByDomain with custom DNS:port failed: %v", err)
	} else {
		t.Logf("域名: %s, IPs: %v, DNS服务器: %s", name, ips, server)
	}
}

func TestGetIPByDomainInvalidDNS(t *testing.T) {
	fmt.Println("=== 测试: GetIPByDomain 无效 DNS ===")
	domain := "www.baidu.com"

	_, _, _, err := GetIPByDomain(domain, "0.0.0.0:1")
	if err == nil {
		t.Log("GetIPByDomain with invalid DNS did not return error")
	} else {
		t.Logf("GetIPByDomain invalid DNS error (expected): %v", err)
	}
}

// --- DNS test using httptest ---

func TestDNSWithTestServer(t *testing.T) {
	fmt.Println("=== 测试: DNS with test server ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("test server request failed: %v", err)
	}
	resp.Body.Close()
}

// --- Test DNS port appending ---

func TestDNSPortAppend(t *testing.T) {
	fmt.Println("=== 测试: DNS 端口追加 ===")
	_, _, s1, err1 := GetIPByDomain("www.baidu.com", "8.8.8.8")
	if err1 != nil {
		t.Logf("DNS with 8.8.8.8 failed: %v", err1)
	} else {
		if !strings.Contains(s1, ":53") {
			t.Errorf("expected :53 in server, got: %s", s1)
		}
	}

	_, _, s2, err2 := GetIPByDomain("www.baidu.com", "8.8.8.8:5353")
	if err2 != nil {
		t.Logf("DNS with 8.8.8.8:5353 failed: %v", err2)
	} else {
		if !strings.Contains(s2, "5353") {
			t.Errorf("expected 5353 in server, got: %s", s2)
		}
	}
}

// --- Test with invalid domain ---

func TestGetIPByDomainInvalid(t *testing.T) {
	fmt.Println("=== 测试: GetIPByDomain invalid domain ===")
	_, _, _, err := GetIPByDomain("this-domain-does-not-exist-12345.example.invalid", "8.8.8.8")
	if err == nil {
		t.Log("GetIPByDomain with invalid domain did not return error")
	} else {
		t.Logf("GetIPByDomain invalid domain error (expected): %v", err)
	}
}

// --- Test with empty DNS server ---

func TestGetIPByDomainEmptyDNS(t *testing.T) {
	fmt.Println("=== 测试: GetIPByDomain empty DNS ===")
	// Empty string means use system default
	_, _, _, err := GetIPByDomain("www.baidu.com", "")
	if err != nil {
		t.Logf("GetIPByDomain with empty DNS failed: %v", err)
	}
}
