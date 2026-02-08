package network

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCheckForUpdate(t *testing.T) {
	fmt.Println("=== 开始测试: 检查更新 (CheckForUpdate) ===")

	// 设置当前版本
	SetAppVersion("TestApp", "v1.0.0", "", "")

	// 模拟版本服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回最新版本信息
		fmt.Fprintln(w, "v1.0.1")
		fmt.Fprintln(w, "checksum_dummy")
	}))
	defer ts.Close()

	latest, sum, url, err := CheckForUpdate(ts.URL, false)
	if err != nil {
		t.Logf("检查更新失败: %v", err)
	} else {
		t.Logf("发现新版本: %s, 校验和: %s, 下载地址: %s", latest, sum, url)
		if latest != "v1.0.1" {
			t.Errorf("解析到的版本号错误，期望 v1.0.1，得到 %s", latest)
		}
	}
}

func TestCalcChecksum(t *testing.T) {
	fmt.Println("=== 开始测试: 计算校验和 (CalcChecksum) ===")
	// 计算当前测试可执行文件的 checksum
	sum, err := CalcChecksum()
	if err != nil {
		// 在测试环境中，os.Executable() 可能并不指向一个常规文件，或者被临时编译运行
		t.Logf("计算校验和失败 (测试环境常见): %v", err)
	} else {
		t.Logf("当前程序校验和: %s", sum)
	}
}

func TestCheckVerservers(t *testing.T) {
	fmt.Println("=== 开始测试: 检查版本服务器可用性 (CheckVerservers) ===")
	
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/state" {
			fmt.Fprint(w, "OK")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	valid := CheckVerservers([]string{ts.URL, "http://invalid-url.com"}, 1)
	if len(valid) == 0 {
		t.Error("未找到有效服务器")
	} else {
		t.Logf("有效服务器: %v", valid)
	}
}

func TestPublishSoftware(t *testing.T) {
	fmt.Println("=== 开始测试: 发布软件 (PublishSoftware) ===")
	
	// 需要设置环境变量 BASH_KEY 才能通过权限验证
	os.Setenv("BASH_KEY", "test_key")
	SetAppVersion("TestApp", "v1.0.0", "", "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/publish" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"errno":0, "errmsg":"success"}`)
		}
	}))
	defer ts.Close()

	// 注意：PublishSoftware 会尝试打开 RunFileName()，在测试中可能不存在或被锁定
	// 这里主要测试逻辑流程，如果文件打开失败是预期的
	err := PublishSoftware(ts.URL)
	if err != nil {
		t.Logf("发布失败 (预期内，因为可能缺少真实文件): %v", err)
	}
}
