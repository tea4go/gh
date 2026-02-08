package network

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHttpRequest(t *testing.T) {
	fmt.Println("=== 开始测试: HTTP 请求 (HttpRequest) ===")

	// 启动一个测试服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			fmt.Fprintln(w, `{"code":0, "msg":"success"}`)
		} else if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintln(w, `{"code":0, "msg":"created"}`)
		}
	}))
	defer ts.Close()

	// 测试 GET
	req := HttpGet(ts.URL)
	str, err := req.String()
	if err != nil {
		t.Errorf("GET 请求失败: %v", err)
	}
	t.Logf("GET 响应: %s", str)

	// 测试 POST
	reqPost := HttpPost(ts.URL)
	strPost, err := reqPost.String()
	if err != nil {
		t.Errorf("POST 请求失败: %v", err)
	}
	t.Logf("POST 响应: %s", strPost)
}

func TestSimpleHttpPost(t *testing.T) {
	fmt.Println("=== 开始测试: 简单 HTTP POST (SimpleHttpPost) ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}))
	defer ts.Close()

	code, body, err := SimpleHttpPost(ts.URL, map[string]string{"foo": "bar"}, 5*time.Second, 5*time.Second)
	if err != nil {
		t.Errorf("SimpleHttpPost 失败: %v", err)
	}
	t.Logf("状态码: %d, 响应: %s", code, string(body))
}

func TestSimpleHttpGet(t *testing.T) {
	fmt.Println("=== 开始测试: 简单 HTTP GET (SimpleHttpGet) ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}))
	defer ts.Close()

	code, body, err := SimpleHttpGet(ts.URL, 5*time.Second, 5*time.Second)
	if err != nil {
		t.Errorf("SimpleHttpGet 失败: %v", err)
	}
	t.Logf("状态码: %d, 响应: %s", code, string(body))
}

func TestBasicAuth(t *testing.T) {
	fmt.Println("=== 开始测试: Basic Auth 中间件 (BasicAuth) ===")
	
	SetUserAndPwd("admin", "123456")
	
	handler := BasicAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authenticated"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("admin", "123456")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("鉴权失败，状态码: %d", resp.StatusCode)
	}
}
