package network

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestAutoLogin(t *testing.T) {
	fmt.Println("=== 开始测试: AutoLogin 自动登录 ===")

	// 从环境变量读取测试数据
	loginURL := os.Getenv("GO_TEST_LOGIN_URL")
	testUser := os.Getenv("GO_TEST_USER")
	testPass := os.Getenv("GO_TEST_PASS")
	testAppKey := os.Getenv("GO_TEST_APPKEY")

	if testUser == "" || testPass == "" || testAppKey == "" || loginURL == "" {
		t.Skip("跳过测试: 请设置环境变量 GO_TEST_USER, GO_TEST_PASS, GO_TEST_APPKEY, GO_TEST_LOGIN_URL")
	}
	loginURL = loginURL + "/v2/auth/check_login"

	// 测试用例1: 登录成功
	t.Run("登录成功", func(t *testing.T) {
		code, _, err := AutoLogin(loginURL, testAppKey, testUser, testPass)
		if err != nil {
			t.Fatalf("AutoLogin 失败: %v", err)
		}
		t.Logf("登录成功, code: %s", code)
	})

	// 测试用例2: 登录失败 (errno != 0)
	t.Run("登录失败_错误响应", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"errno":  1001,
				"errmsg": "用户名或密码错误",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		}))
		defer ts.Close()

		code, _, err := AutoLogin(ts.URL, "wrong_key", "wrong_user", "wrong_pwd")
		if err == nil {
			t.Fatal("期望返回错误，但 err 为 nil")
		}
		if code != "" {
			t.Errorf("期望 code 为空，实际: '%s'", code)
		}
		t.Logf("登录失败(符合预期): %v", err)
	})

	// 测试用例3: 服务器不可达
	t.Run("连接失败", func(t *testing.T) {
		code, _, err := AutoLogin("http://127.0.0.1:1/notexist", "key", "user", "pwd")
		if err == nil {
			t.Fatal("期望返回错误，但 err 为 nil")
		}
		if code != "" {
			t.Errorf("期望 code 为空，实际: '%s'", code)
		}
		t.Logf("连接失败(符合预期): %v", err)
	})

	// 测试用例4: 服务器返回500
	t.Run("服务器内部错误", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Internal Server Error")
		}))
		defer ts.Close()

		code, resp, err := AutoLogin(ts.URL, "key", "user", "pwd")
		if err == nil {
			t.Fatal("期望返回错误，但 err 为 nil")
		}
		if code != "" {
			t.Errorf("期望 code 为空，实际: '%s'", code)
		}
		t.Logf("500错误(符合预期): code=%s, resp=%v, err=%v", code, resp, err)
	})
}
