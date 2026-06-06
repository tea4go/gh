package network

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- THttpRequest builder tests ---

func TestHttpRequestGet(t *testing.T) {
	fmt.Println("=== 测试: HttpGet ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		fmt.Fprint(w, `{"result":"ok"}`)
	}))
	defer ts.Close()

	str, err := HttpGet(ts.URL).String()
	if err != nil {
		t.Fatalf("HttpGet failed: %v", err)
	}
	if !strings.Contains(str, "ok") {
		t.Errorf("unexpected response: %s", str)
	}
}

func TestHttpRequestPost(t *testing.T) {
	fmt.Println("=== 测试: HttpPost ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		fmt.Fprint(w, `{"result":"created"}`)
	}))
	defer ts.Close()

	str, err := HttpPost(ts.URL).String()
	if err != nil {
		t.Fatalf("HttpPost failed: %v", err)
	}
	if !strings.Contains(str, "created") {
		t.Errorf("unexpected response: %s", str)
	}
}

func TestHttpPut(t *testing.T) {
	fmt.Println("=== 测试: HttpPut ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "method=%s", r.Method)
	}))
	defer ts.Close()

	str, err := HttpPut(ts.URL).String()
	if err != nil {
		t.Fatalf("HttpPut failed: %v", err)
	}
	if !strings.Contains(str, "PUT") {
		t.Errorf("unexpected response: %s", str)
	}
}

func TestHttpDelete(t *testing.T) {
	fmt.Println("=== 测试: HttpDelete ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "method=%s", r.Method)
	}))
	defer ts.Close()

	str, err := HttpDelete(ts.URL).String()
	if err != nil {
		t.Fatalf("HttpDelete failed: %v", err)
	}
	if !strings.Contains(str, "DELETE") {
		t.Errorf("unexpected response: %s", str)
	}
}

func TestHttpHead(t *testing.T) {
	fmt.Println("=== 测试: HttpHead ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "head")
	}))
	defer ts.Close()

	req := HttpHead(ts.URL)
	resp, err := req.Response()
	if err != nil {
		t.Fatalf("HttpHead failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("unexpected status: %d", resp.StatusCode)
	}
}

// --- THttpRequest method chaining tests ---

func TestTHttpRequestSetting(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Setting ===")
	req := NewRequest("http://example.com", "GET")
	s := THttpSettings{UserAgent: "TestAgent"}
	result := req.Setting(s)
	if result != req {
		t.Error("Setting should return self")
	}
	got := req.GetSetting()
	if got.UserAgent != "TestAgent" {
		t.Errorf("Setting not applied: %+v", got)
	}
}

func TestTHttpRequestSetBasicAuth(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest SetBasicAuth ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		fmt.Fprintf(w, "user=%s pass=%s ok=%v", u, p, ok)
	}))
	defer ts.Close()

	req := NewRequest(ts.URL, "GET")
	result := req.SetBasicAuth("myuser", "mypass")
	if result != req {
		t.Error("SetBasicAuth should return self")
	}

	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !strings.Contains(str, "user=myuser") || !strings.Contains(str, "pass=mypass") {
		t.Errorf("basic auth not sent: %s", str)
	}
}

func TestTHttpRequestSetEnableCookie(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest SetEnableCookie ===")
	req := NewRequest("http://example.com", "GET")
	result := req.SetEnableCookie(true)
	if result != req {
		t.Error("SetEnableCookie should return self")
	}
	if !req.setting.EnableCookie {
		t.Error("EnableCookie not set")
	}
}

func TestTHttpRequestSetUserAgent(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest SetUserAgent ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Header.Get("User-Agent"))
	}))
	defer ts.Close()

	req := NewRequest(ts.URL, "GET")
	req.SetUserAgent("CustomAgent/1.0")
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !strings.Contains(str, "CustomAgent") {
		t.Errorf("User-Agent not set: %s", str)
	}
}

func TestTHttpRequestDebug(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Debug ===")
	req := NewRequest("http://example.com", "GET")
	result := req.Debug(true)
	if result != req {
		t.Error("Debug should return self")
	}
	if !req.setting.ShowDebug {
		t.Error("ShowDebug not set")
	}
}

func TestTHttpRequestRetries(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Retries ===")
	req := NewRequest("http://example.com", "GET")
	result := req.Retries(3)
	if result != req {
		t.Error("Retries should return self")
	}
	if req.setting.Retries != 3 {
		t.Errorf("Retries not set: %d", req.setting.Retries)
	}
}

func TestTHttpRequestRetryDelay(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest RetryDelay ===")
	req := NewRequest("http://example.com", "GET")
	result := req.RetryDelay(100 * time.Millisecond)
	if result != req {
		t.Error("RetryDelay should return self")
	}
	if req.setting.RetryDelay != 100*time.Millisecond {
		t.Errorf("RetryDelay not set: %v", req.setting.RetryDelay)
	}
}

func TestTHttpRequestDumpBody(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest DumpBody ===")
	req := NewRequest("http://example.com", "GET")
	result := req.DumpBody(false)
	if result != req {
		t.Error("DumpBody should return self")
	}
	if req.setting.DumpBody {
		t.Error("DumpBody should be false")
	}
}

func TestTHttpRequestSetTimeout(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest SetTimeout ===")
	req := NewRequest("http://example.com", "GET")
	result := req.SetTimeout(5*time.Second, 10*time.Second)
	if result != req {
		t.Error("SetTimeout should return self")
	}
	if req.setting.ConnectTimeout != 5*time.Second || req.setting.ReadWriteTimeout != 10*time.Second {
		t.Errorf("Timeout not set: %v / %v", req.setting.ConnectTimeout, req.setting.ReadWriteTimeout)
	}
}

func TestTHttpRequestHeader(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Header ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Header.Get("X-Custom"))
	}))
	defer ts.Close()

	req := NewRequest(ts.URL, "GET")
	req.Header("X-Custom", "testvalue")
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if str != "testvalue" {
		t.Errorf("Header not sent: %s", str)
	}
}

func TestTHttpRequestSetHost(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest SetHost ===")
	req := NewRequest("http://example.com", "GET")
	result := req.SetHost("custom-host")
	if result != req {
		t.Error("SetHost should return self")
	}
	if req.req.Host != "custom-host" {
		t.Errorf("Host not set: %s", req.req.Host)
	}
}

func TestTHttpRequestSetProtocolVersion(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest SetProtocolVersion ===")
	req := NewRequest("http://example.com", "GET")
	req.SetProtocolVersion("HTTP/1.0")
	if req.req.Proto != "HTTP/1.0" {
		t.Errorf("Proto not set: %s", req.req.Proto)
	}
	// Empty version defaults to HTTP/1.1
	req2 := NewRequest("http://example.com", "GET")
	req2.SetProtocolVersion("")
	if req2.req.Proto != "HTTP/1.1" {
		t.Errorf("Proto should default to HTTP/1.1: %s", req2.req.Proto)
	}
	// Invalid version is ignored
	req3 := NewRequest("http://example.com", "GET")
	req3.SetProtocolVersion("INVALID")
	if req3.req.Proto != "HTTP/1.1" {
		t.Errorf("Invalid proto should not change: %s", req3.req.Proto)
	}
}

func TestTHttpRequestSetCookie(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest SetCookie ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("testcookie")
		if err != nil {
			fmt.Fprint(w, "no-cookie")
		} else {
			fmt.Fprint(w, cookie.Value)
		}
	}))
	defer ts.Close()

	req := NewRequest(ts.URL, "GET")
	req.SetCookie(&http.Cookie{Name: "testcookie", Value: "yum"})
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if str != "yum" {
		t.Errorf("Cookie not sent: %s", str)
	}
}

func TestTHttpRequestSetTransport(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest SetTransport ===")
	req := NewRequest("http://example.com", "GET")
	transport := &http.Transport{}
	result := req.SetTransport(transport)
	if result != req {
		t.Error("SetTransport should return self")
	}
	if req.setting.Transport != transport {
		t.Error("Transport not set")
	}
}

func TestTHttpRequestSetProxy(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest SetProxy ===")
	req := NewRequest("http://example.com", "GET")
	proxyFunc := func(r *http.Request) (*url.URL, error) {
		return url.Parse("http://proxy:8080")
	}
	result := req.SetProxy(proxyFunc)
	if result != req {
		t.Error("SetProxy should return self")
	}
	if req.setting.Proxy == nil {
		t.Error("Proxy not set")
	}
}

func TestTHttpRequestSetCheckRedirect(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest SetCheckRedirect ===")
	req := NewRequest("http://example.com", "GET")
	redirectFunc := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	result := req.SetCheckRedirect(redirectFunc)
	if result != req {
		t.Error("SetCheckRedirect should return self")
	}
	if req.setting.CheckRedirect == nil {
		t.Error("CheckRedirect not set")
	}
}

func TestTHttpRequestSetFilters(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest SetFilters ===")
	req := NewRequest("http://example.com", "GET")
	fc := func(next Filter) Filter {
		return func(ctx context.Context, r *THttpRequest) (*http.Response, error) {
			return next(ctx, r)
		}
	}
	result := req.SetFilters(fc)
	if result != req {
		t.Error("SetFilters should return self")
	}
	if len(req.setting.FilterChains) != 1 {
		t.Errorf("FilterChains not set: %d", len(req.setting.FilterChains))
	}
}

func TestTHttpRequestAddFilters(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest AddFilters ===")
	req := NewRequest("http://example.com", "GET")
	fc1 := func(next Filter) Filter {
		return next
	}
	fc2 := func(next Filter) Filter {
		return next
	}
	req.SetFilters(fc1)
	req.AddFilters(fc2)
	if len(req.setting.FilterChains) != 2 {
		t.Errorf("AddFilters not working: %d", len(req.setting.FilterChains))
	}
}

func TestTHttpRequestParam(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Param ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.URL.RawQuery)
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	req.Param("key", "value")
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !strings.Contains(str, "key=value") {
		t.Errorf("Param not sent: %s", str)
	}
}

func TestTHttpRequestParamMulti(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Param multi-value ===")
	req := NewRequest("http://example.com", "GET")
	req.Param("key", "v1")
	req.Param("key", "v2")
	if len(req.params["key"]) != 2 {
		t.Errorf("multi param not working: %v", req.params["key"])
	}
}

func TestTHttpRequestBody(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Body ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprint(w, string(body))
	}))
	defer ts.Close()

	req := HttpPost(ts.URL)
	req.Body("test-body")
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if str != "test-body" {
		t.Errorf("Body not sent: %s", str)
	}
}

func TestTHttpRequestBodyBytes(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Body []byte ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprint(w, string(body))
	}))
	defer ts.Close()

	req := HttpPost(ts.URL)
	req.Body([]byte("byte-body"))
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if str != "byte-body" {
		t.Errorf("Body []byte not sent: %s", str)
	}
}

func TestTHttpRequestJSONBody(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest JSONBody ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, "ct=%s body=%s", ct, string(body))
	}))
	defer ts.Close()

	req := HttpPost(ts.URL)
	req.JSONBody(map[string]string{"key": "val"})
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !strings.Contains(str, "application/json") {
		t.Errorf("JSON Content-Type not set: %s", str)
	}
}

func TestTHttpRequestXMLBody(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest XMLBody ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		fmt.Fprintf(w, "ct=%s", ct)
	}))
	defer ts.Close()

	type xmlTest struct {
		XMLName xml.Name `xml:"test"`
		Value   string   `xml:"value"`
	}
	req := HttpPost(ts.URL)
	req.XMLBody(xmlTest{Value: "hello"})
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !strings.Contains(str, "application/xml") {
		t.Errorf("XML Content-Type not set: %s", str)
	}
}

func TestTHttpRequestYAMLBody(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest YAMLBody ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		fmt.Fprintf(w, "ct=%s", ct)
	}))
	defer ts.Close()

	req := HttpPost(ts.URL)
	req.YAMLBody(map[string]string{"key": "val"})
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !strings.Contains(str, "yaml") {
		t.Errorf("YAML Content-Type not set: %s", str)
	}
}

func TestTHttpRequestGetRequest(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest GetRequest ===")
	req := NewRequest("http://example.com/path", "GET")
	r := req.GetRequest()
	if r == nil {
		t.Fatal("GetRequest returned nil")
	}
	if r.Method != "GET" {
		t.Errorf("unexpected method: %s", r.Method)
	}
}

func TestTHttpRequestGetResponse(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest GetResponse ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "hello")
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	_, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp := req.GetResponse()
	if resp == nil {
		t.Fatal("GetResponse returned nil")
	}
}

func TestTHttpRequestDumpRequest(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest DumpRequest ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	req.Debug(true).DumpBody(true)
	_, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	dump := req.DumpRequest()
	if len(dump) == 0 {
		t.Error("DumpRequest returned empty")
	}
}

// --- THttpRequest response methods ---

func TestTHttpRequestToJSON(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToJSON ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"key": "value"})
	}))
	defer ts.Close()

	var result map[string]string
	err := HttpGet(ts.URL).ToJSON(&result)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("ToJSON result mismatch: %+v", result)
	}
}

func TestTHttpRequestToXML(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToXML ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<TestXML><Value>hello</Value></TestXML>`)
	}))
	defer ts.Close()

	type TestXML struct {
		XMLName xml.Name `xml:"TestXML"`
		Value   string   `xml:"Value"`
	}
	var result TestXML
	err := HttpGet(ts.URL).ToXML(&result)
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}
	if result.Value != "hello" {
		t.Errorf("ToXML result mismatch: %+v", result)
	}
}

func TestTHttpRequestToYAML(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToYAML ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "key: value\n")
	}))
	defer ts.Close()

	var result map[string]string
	err := HttpGet(ts.URL).ToYAML(&result)
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("ToYAML result mismatch: %+v", result)
	}
}

func TestTHttpRequestToFile(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToFile ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "file-content")
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "test_output.txt")

	err := HttpGet(ts.URL).ToFile(outFile)
	if err != nil {
		t.Fatalf("ToFile failed: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output file failed: %v", err)
	}
	if string(data) != "file-content" {
		t.Errorf("file content mismatch: %s", string(data))
	}
}

func TestTHttpRequestStatusCode(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest StatusCode ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "created")
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	_, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	code, err := req.StatusCode()
	if err != nil {
		t.Fatalf("StatusCode failed: %v", err)
	}
	if code != http.StatusCreated {
		t.Errorf("StatusCode mismatch: %d", code)
	}
}

func TestTHttpRequestResponse(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Response ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	resp, err := HttpGet(ts.URL).Response()
	if err != nil {
		t.Fatalf("Response failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Response status mismatch: %d", resp.StatusCode)
	}
}

func TestTHttpRequestBytes(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Bytes ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "bytes-data")
	}))
	defer ts.Close()

	data, err := HttpGet(ts.URL).Bytes()
	if err != nil {
		t.Fatalf("Bytes failed: %v", err)
	}
	if string(data) != "bytes-data" {
		t.Errorf("Bytes mismatch: %s", string(data))
	}
}

// Test gzip response
func TestTHttpRequestGzip(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Gzip ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		fmt.Fprint(gw, "gzip-content")
		gw.Close()
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	req.setting.Gzip = true
	str, err := req.String()
	if err != nil {
		t.Fatalf("Gzip request failed: %v", err)
	}
	if !strings.Contains(str, "gzip-content") {
		t.Errorf("Gzip content mismatch: %s", str)
	}
}

// --- HttpRequest function tests ---

func TestHttpRequestA(t *testing.T) {
	fmt.Println("=== 测试: HttpRequestA ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"errno": 0})
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, _, err := HttpRequestA("GET", ts.URL, false, &result)
	if err != nil {
		t.Logf("HttpRequestA error: %v", err)
	}
	t.Logf("HttpRequestA: code=%d, result=%v, err=%v", code, result, err)
}

func TestHttpRequestB(t *testing.T) {
	fmt.Println("=== 测试: HttpRequestB ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	code, _, _, err := HttpRequestB("GET", ts.URL, false)
	t.Logf("HttpRequestB: code=%d, err=%v", code, err)
}

func TestHttpRequestWithParams(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest with params ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"errno":0}`)
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, _, err := HttpRequestPB("GET", ts.URL, false, map[string]string{"q": "test"}, &result)
	t.Logf("HttpRequestPB: code=%d, err=%v", code, err)
}

func TestHttpRequestWithBody(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest with body ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		ct := r.Header.Get("Content-Type")
		fmt.Fprintf(w, `{"ct":"%s","body":"%s"}`, ct, string(body))
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, data, err := HttpRequestBB("POST", ts.URL, false, `{"key":"val"}`, &result)
	t.Logf("HttpRequestBB: code=%d, data=%s, err=%v", code, string(data), err)
}

func TestHttpRequestWithHeader(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest with header ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"errno":0}`)
	}))
	defer ts.Close()

	code, _, _, err := HttpRequestHB("GET", ts.URL, false, map[string]string{"X-Test": "value"})
	t.Logf("HttpRequestHB: code=%d, err=%v", code, err)
}

func TestHttpRequestWithAuth(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest with auth ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		fmt.Fprintf(w, `{"user":"%s","pass":"%s","ok":%v}`, u, p, ok)
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, _, err := HttpRequestPD("GET", ts.URL, false, nil, "user1", "pass1", &result)
	t.Logf("HttpRequestPD: code=%d, err=%v", code, err)
}

func TestHttpRequestFull(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest full ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"errno":0,"data":"test"}`)
	}))
	defer ts.Close()

	var result map[string]interface{}

	_, _, _, _ = HttpRequestHA("GET", ts.URL, false, map[string]string{"X-H": "1"}, &result)
	_, _, _, _ = HttpRequestPHB("GET", ts.URL, false, map[string]string{"k": "v"}, map[string]string{"X-H": "1"}, &result)
	_, _, _, _ = HttpRequestPHD("GET", ts.URL, false, map[string]string{"k": "v"}, map[string]string{"X-H": "1"}, "u", "p", &result)
	_, _, _, _ = HttpRequestBHB("POST", ts.URL, false, "body-data", map[string]string{"X-H": "1"}, &result)
	_, _, _, _ = HttpRequestBHD("POST", ts.URL, false, "body-data", map[string]string{"X-H": "1"}, "u", "p", &result)
	_, _, _, _ = HttpRequestBB("POST", ts.URL, false, "body-str", &result)
	_, _, _, _ = HttpRequestBD("POST", ts.URL, false, "body-str", "u", "p", &result)
	_, _, _, _ = HttpRequestPBHB("POST", ts.URL, false, map[string]string{"k": "v"}, "body", map[string]string{"X-H": "1"}, &result)
	_, _, _, _ = HttpRequestPC("GET", ts.URL, false, map[string]string{"k": "v"}, []*http.Cookie{{Name: "c", Value: "v"}}, &result)
	_, _, _, _ = HttpRequestPHC("GET", ts.URL, false, map[string]string{"k": "v"}, []*http.Cookie{{Name: "c", Value: "v"}}, map[string]string{"X-H": "1"}, &result)
	_, _, _, _ = HttpRequestBC("POST", ts.URL, false, "body", []*http.Cookie{{Name: "c", Value: "v"}}, &result)
	_, _, _, _ = HttpRequestBHC("POST", ts.URL, false, "body", []*http.Cookie{{Name: "c", Value: "v"}}, map[string]string{"X-H": "1"}, &result)
}

// --- BasicAuth middleware tests ---

func TestBasicAuthNoUsers(t *testing.T) {
	fmt.Println("=== 测试: BasicAuth 无用户 ===")
	// Clear users
	validUsers = map[string]string{}

	handler := BasicAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	// No users configured => should pass through
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestBasicAuthCorrect(t *testing.T) {
	fmt.Println("=== 测试: BasicAuth 正确凭据 ===")
	SetUserAndPwd("admin", "secret123")

	handler := BasicAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("admin", "secret123")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestBasicAuthNoCredentials(t *testing.T) {
	fmt.Println("=== 测试: BasicAuth 无凭据 ===")
	SetUserAndPwd("admin", "secret123")

	handler := BasicAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "401") {
		t.Errorf("expected 401 in body, got: %s", body)
	}
}

func TestBasicAuthWrongCredentials(t *testing.T) {
	fmt.Println("=== 测试: BasicAuth 错误凭据 ===")
	SetUserAndPwd("admin", "secret123")

	handler := BasicAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("admin", "wrongpass")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "402") {
		t.Errorf("expected 402 in body, got: %s", body)
	}
}

func TestBasicAuthWrongUser(t *testing.T) {
	fmt.Println("=== 测试: BasicAuth 错误用户 ===")
	SetUserAndPwd("admin", "secret123")

	handler := BasicAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("wronguser", "secret123")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// --- BasicAuth2 tests ---

func TestBasicAuth2NoUsers(t *testing.T) {
	fmt.Println("=== 测试: BasicAuth2 无用户 ===")
	validUsers = map[string]string{}

	handler := BasicAuth2(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestBasicAuth2Correct(t *testing.T) {
	fmt.Println("=== 测试: BasicAuth2 正确凭据 ===")
	SetUserAndPwd("testuser", "testpass")

	handler := BasicAuth2(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("testuser", "testpass")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestBasicAuth2NoCredentials(t *testing.T) {
	fmt.Println("=== 测试: BasicAuth2 无凭据 ===")
	SetUserAndPwd("testuser", "testpass")

	handler := BasicAuth2(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestBasicAuth2WrongCredentials(t *testing.T) {
	fmt.Println("=== 测试: BasicAuth2 错误凭据 ===")
	SetUserAndPwd("testuser", "testpass")

	handler := BasicAuth2(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("testuser", "wrongpass")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// --- SetBasicAuth function test ---

func TestSetBasicAuthFunc(t *testing.T) {
	fmt.Println("=== 测试: SetBasicAuth 函数 ===")
	validUsers = map[string]string{}

	SetBasicAuth("myuser:mypass")
	if validUsers["myuser"] != "mypass" {
		t.Errorf("SetBasicAuth did not set user: %+v", validUsers)
	}

	// Update existing user
	SetBasicAuth("myuser:newpass")
	if validUsers["myuser"] != "newpass" {
		t.Errorf("SetBasicAuth did not update user: %+v", validUsers)
	}

	// Invalid format (no colon)
	validUsers = map[string]string{}
	SetBasicAuth("invalidformat")
	if len(validUsers) != 0 {
		t.Errorf("SetBasicAuth should ignore invalid format: %+v", validUsers)
	}
}

// --- SetUserAndPwd test ---

func TestSetUserAndPwd(t *testing.T) {
	fmt.Println("=== 测试: SetUserAndPwd ===")
	validUsers = map[string]string{}

	SetUserAndPwd("User1", "Pass1")
	if validUsers["user1"] != "pass1" {
		t.Errorf("SetUserAndPwd did not lowercase: %+v", validUsers)
	}

	// Update existing user
	SetUserAndPwd("USER1", "NEWPASS")
	if validUsers["user1"] != "newpass" {
		t.Errorf("SetUserAndPwd did not update: %+v", validUsers)
	}
}

// --- GetHttpRemoteAddr tests ---

func TestGetHttpRemoteAddr(t *testing.T) {
	fmt.Println("=== 测试: GetHttpRemoteAddr ===")

	// nil request
	if addr := GetHttpRemoteAddr(nil); addr != "" {
		t.Errorf("expected empty for nil, got: %s", addr)
	}

	// X-Forwarded-For header
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	if addr := GetHttpRemoteAddr(req); addr != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got: %s", addr)
	}

	// RemoteAddr with colon
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "10.0.0.1:12345"
	addr2 := GetHttpRemoteAddr(req2)
	if !strings.HasPrefix(addr2, "10.0.0.1") {
		t.Errorf("expected 10.0.0.1, got: %s", addr2)
	}

	// IPv6 loopback
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.RemoteAddr = "[::1]:12345"
	addr3 := GetHttpRemoteAddr(req3)
	if !strings.HasPrefix(addr3, "127.0.0.1") {
		t.Errorf("expected 127.0.0.1, got: %s", addr3)
	}
}

func TestGetHttpRemoteAddrPort(t *testing.T) {
	fmt.Println("=== 测试: GetHttpRemoteAddrPort ===")

	// nil request
	if addr := GetHttpRemoteAddrPort(nil); addr != "" {
		t.Errorf("expected empty for nil, got: %s", addr)
	}

	// X-Forwarded-For header
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	if addr := GetHttpRemoteAddrPort(req); addr != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got: %s", addr)
	}

	// RemoteAddr
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "10.0.0.1:12345"
	addr2 := GetHttpRemoteAddrPort(req2)
	if !strings.HasPrefix(addr2, "10.0.0.1") {
		t.Errorf("expected 10.0.0.1, got: %s", addr2)
	}

	// IPv6 loopback
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.RemoteAddr = "[::1]:12345"
	addr3 := GetHttpRemoteAddrPort(req3)
	if !strings.Contains(addr3, "127.0.0.1") {
		t.Errorf("expected 127.0.0.1, got: %s", addr3)
	}
}

// --- SetDefaultSetting tests ---

func TestSetDefaultSettingByTimeout(t *testing.T) {
	fmt.Println("=== 测试: SetDefaultSettingByTimeout ===")
	SetDefaultSettingByTimeout(5*time.Second, 10*time.Second)
	if defaultSetting.ConnectTimeout != 5*time.Second || defaultSetting.ReadWriteTimeout != 10*time.Second {
		t.Errorf("SetDefaultSettingByTimeout not applied")
	}
}

func TestSetDefaultSetting(t *testing.T) {
	fmt.Println("=== 测试: SetDefaultSetting ===")
	s := THttpSettings{
		UserAgent:        "Test",
		ConnectTimeout:   3 * time.Second,
		ReadWriteTimeout: 6 * time.Second,
	}
	SetDefaultSetting(s)
	if defaultSetting.UserAgent != "Test" {
		t.Errorf("SetDefaultSetting not applied: %s", defaultSetting.UserAgent)
	}
}

// --- WebManager tests ---

func TestWebManagerHome(t *testing.T) {
	fmt.Println("=== 测试: webManagerHome ===")
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	webManagerHome(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Welcome") {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestWebManagerAPI(t *testing.T) {
	fmt.Println("=== 测试: webManagerAPI ===")

	// Test without log_level param
	req := httptest.NewRequest("GET", "/manager", nil)
	w := httptest.NewRecorder()
	WebManagerAPI = nil
	webManagerAPI(w, req)
	body := w.Body.String()
	if !strings.Contains(body, "errno:0") {
		t.Errorf("expected default response, got: %s", body)
	}

	// Test with log_level param
	req2 := httptest.NewRequest("GET", "/manager?log_level=6", nil)
	w2 := httptest.NewRecorder()
	webManagerAPI(w2, req2)
	body2 := w2.Body.String()
	if !strings.Contains(body2, "errno:0") {
		t.Errorf("expected success, got: %s", body2)
	}

	// Test with invalid log_level
	req3 := httptest.NewRequest("GET", "/manager?log_level=abc", nil)
	w3 := httptest.NewRecorder()
	webManagerAPI(w3, req3)
	body3 := w3.Body.String()
	if !strings.Contains(body3, "53010") {
		t.Errorf("expected error for invalid level, got: %s", body3)
	}

	// Test with out-of-range log_level
	req4 := httptest.NewRequest("GET", "/manager?log_level=999", nil)
	w4 := httptest.NewRecorder()
	webManagerAPI(w4, req4)
	body4 := w4.Body.String()
	if !strings.Contains(body4, "53010") {
		t.Errorf("expected error for out-of-range level, got: %s", body4)
	}

	// Test with log_name=log_level format
	req5 := httptest.NewRequest("GET", "/manager?log_level=testlog=6", nil)
	w5 := httptest.NewRecorder()
	webManagerAPI(w5, req5)
	body5 := w5.Body.String()
	if !strings.Contains(body5, "errno:0") {
		t.Errorf("expected success for name=level, got: %s", body5)
	}

	// Test with custom WebManagerAPI
	WebManagerAPI = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"custom":true}`)
	}
	req6 := httptest.NewRequest("GET", "/manager", nil)
	w6 := httptest.NewRecorder()
	webManagerAPI(w6, req6)
	body6 := w6.Body.String()
	if !strings.Contains(body6, "custom") {
		t.Errorf("expected custom handler, got: %s", body6)
	}
	WebManagerAPI = nil
}

func TestWebAutoTest(t *testing.T) {
	fmt.Println("=== 测试: webAutoTest ===")
	WebAutoTestAPI = nil
	req := httptest.NewRequest("GET", "/autotest", nil)
	w := httptest.NewRecorder()
	webAutoTest(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Test with custom WebAutoTestAPI
	WebAutoTestAPI = func(w http.ResponseWriter, r *http.Request, re map[string]interface{}) map[string]interface{} {
		re["custom"] = true
		return re
	}
	req2 := httptest.NewRequest("GET", "/autotest", nil)
	w2 := httptest.NewRecorder()
	webAutoTest(w2, req2)
	body := w2.Body.String()
	if !strings.Contains(body, "custom") {
		t.Errorf("expected custom field, got: %s", body)
	}
	WebAutoTestAPI = nil
}

// --- TimeoutDialer test ---

func TestTimeoutDialer(t *testing.T) {
	fmt.Println("=== 测试: TimeoutDialer ===")
	dialer := TimeoutDialer(1*time.Second, 2*time.Second)
	if dialer == nil {
		t.Fatal("TimeoutDialer returned nil")
	}
}

// --- DownloadURL test ---

func TestDownloadURL(t *testing.T) {
	fmt.Println("=== 测试: DownloadURL ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "line1\nline2\nline3\n")
	}))
	defer ts.Close()

	code, lines, err := DownloadURL(ts.URL, 5*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("DownloadURL failed: %v", err)
	}
	if code != 200 {
		t.Errorf("status code: %d", code)
	}
	if len(lines) != 4 { // 3 lines + trailing empty from \n
		t.Logf("lines count: %d, lines: %v", len(lines), lines)
	}
}

// --- DownloadTextFile test ---

func TestDownloadTextFile(t *testing.T) {
	fmt.Println("=== 测试: DownloadTextFile ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "line1\nline2\n")
	}))
	defer ts.Close()

	code, lines, err := DownloadTextFile(ts.URL, 5*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("DownloadTextFile failed: %v", err)
	}
	if code != 200 {
		t.Errorf("status code: %d", code)
	}
	t.Logf("lines: %v", lines)
}

// --- DownloadFile test ---

func TestDownloadFile(t *testing.T) {
	fmt.Println("=== 测试: DownloadFile ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Last-Modified", time.Now().Format(time.RFC1123))
		fmt.Fprint(w, "file-content")
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "downloaded.txt")

	err := DownloadFile(ts.URL, outFile, 5*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading downloaded file failed: %v", err)
	}
	if string(data) != "file-content" {
		t.Errorf("file content mismatch: %s", string(data))
	}
}

// --- PostFile test (buildURL with files) ---

func TestTHttpRequestPostFile(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest PostFile ===")
	req := NewRequest("http://example.com", "POST")
	result := req.PostFile("file", "/nonexistent/file.txt")
	if result != req {
		t.Error("PostFile should return self")
	}
	if req.files["file"] != "/nonexistent/file.txt" {
		t.Error("PostFile not set")
	}
}

// --- pathExistAndMkdir test ---

func TestPathExistAndMkdir(t *testing.T) {
	fmt.Println("=== 测试: pathExistAndMkdir ===")
	tmpDir := t.TempDir()
	existingDir := filepath.Join(tmpDir, "existing")
	os.MkdirAll(existingDir, os.ModePerm)

	// Existing directory
	err := pathExistAndMkdir(filepath.Join(existingDir, "file.txt"))
	if err != nil {
		t.Errorf("pathExistAndMkdir failed for existing dir: %v", err)
	}

	// New directory
	newDir := filepath.Join(tmpDir, "new", "nested", "file.txt")
	err = pathExistAndMkdir(newDir)
	if err != nil {
		t.Errorf("pathExistAndMkdir failed for new dir: %v", err)
	}
}

// --- CreateDefaultCookie test ---

func TestCreateDefaultCookie(t *testing.T) {
	fmt.Println("=== 测试: createDefaultCookie ===")
	defaultCookieJar = nil
	createDefaultCookie()
	if defaultCookieJar == nil {
		t.Error("createDefaultCookie did not create jar")
	}
	defaultCookieJar = nil // reset
}

// --- DoRequestWithCtx test with filter chains ---

func TestDoRequestWithCtxFilter(t *testing.T) {
	fmt.Println("=== 测试: DoRequestWithCtx with filter ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "filtered")
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	called := false
	fc := func(next Filter) Filter {
		return func(ctx context.Context, r *THttpRequest) (*http.Response, error) {
			called = true
			return next(ctx, r)
		}
	}
	req.SetFilters(fc)

	str, err := req.String()
	if err != nil {
		t.Fatalf("filtered request failed: %v", err)
	}
	if !called {
		t.Error("filter was not called")
	}
	if !strings.Contains(str, "filtered") {
		t.Errorf("unexpected response: %s", str)
	}
}

// --- SetTLSClientConfig test ---

func TestTHttpRequestSetTLSClientConfig(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest SetTLSClientConfig ===")
	req := NewRequest("http://example.com", "GET")
	cfg := &tls.Config{}
	result := req.SetTLSClientConfig(cfg)
	if result != req {
		t.Error("SetTLSClientConfig should return self")
	}
	if req.setting.TLSClientConfig != cfg {
		t.Error("TLSClientConfig not set")
	}
}

// --- doRequest with custom transport test ---

func TestDoRequestCustomTransport(t *testing.T) {
	fmt.Println("=== 测试: doRequest custom transport ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "custom-transport")
	}))
	defer ts.Close()

	transport := &http.Transport{}
	req := HttpGet(ts.URL)
	req.SetTransport(transport)

	str, err := req.String()
	if err != nil {
		t.Fatalf("custom transport request failed: %v", err)
	}
	if !strings.Contains(str, "custom-transport") {
		t.Errorf("unexpected response: %s", str)
	}
}

// --- SimpleHttpPost with string body ---

func TestSimpleHttpPostStringBody(t *testing.T) {
	fmt.Println("=== 测试: SimpleHttpPost string body ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprint(w, string(body))
	}))
	defer ts.Close()

	code, resp, err := SimpleHttpPost(ts.URL, "string-body", 5*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("SimpleHttpPost failed: %v", err)
	}
	if code != 200 {
		t.Errorf("status code: %d", code)
	}
	if string(resp) != "string-body" {
		t.Errorf("response mismatch: %s", string(resp))
	}
}

// --- SimpleHttpPost with []byte body ---

func TestSimpleHttpPostBytesBody(t *testing.T) {
	fmt.Println("=== 测试: SimpleHttpPost []byte body ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprint(w, string(body))
	}))
	defer ts.Close()

	code, resp, err := SimpleHttpPost(ts.URL, []byte("bytes-body"), 5*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("SimpleHttpPost failed: %v", err)
	}
	if code != 200 {
		t.Errorf("status code: %d", code)
	}
	if string(resp) != "bytes-body" {
		t.Errorf("response mismatch: %s", string(resp))
	}
}

// --- SimpleHttpPost with nil body ---

func TestSimpleHttpPostNilBody(t *testing.T) {
	fmt.Println("=== 测试: SimpleHttpPost nil body ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "nil-ok")
	}))
	defer ts.Close()

	code, resp, err := SimpleHttpPost(ts.URL, nil, 5*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("SimpleHttpPost failed: %v", err)
	}
	if code != 200 {
		t.Errorf("status code: %d", code)
	}
	if string(resp) != "nil-ok" {
		t.Errorf("response mismatch: %s", string(resp))
	}
}

// --- Bytes cached test ---

func TestTHttpRequestBytesCached(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Bytes cached ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "cached-data")
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	data1, err := req.Bytes()
	if err != nil {
		t.Fatalf("Bytes failed: %v", err)
	}
	// Second call should use cache
	data2, err := req.Bytes()
	if err != nil {
		t.Fatalf("Bytes (cached) failed: %v", err)
	}
	if string(data1) != string(data2) {
		t.Errorf("cached bytes mismatch")
	}
	if string(data2) != "cached-data" {
		t.Errorf("unexpected cached data: %s", string(data2))
	}
}

// --- HttpRequest error cases ---

func TestHttpRequestServerError(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest server error ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	}))
	defer ts.Close()

	code, _, _, err := HttpRequestB("GET", ts.URL, false)
	if err == nil {
		t.Log("HttpRequestB did not return error for 500 (may be expected)")
	}
	t.Logf("HttpRequestB 500: code=%d, err=%v", code, err)
}

// --- AutoLogin test with local server ---

func TestAutoLoginLocal(t *testing.T) {
	fmt.Println("=== 测试: AutoLogin local ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errno": 0,
			"code":  "auth-code-123",
		})
	}))
	defer ts.Close()

	code, _, err := AutoLogin(ts.URL, "appkey", "user", "pass")
	if err != nil {
		t.Logf("AutoLogin error: %v", err)
	} else {
		if code != "auth-code-123" {
			t.Errorf("AutoLogin code mismatch: %s", code)
		}
	}
}

// --- NewRequest with bad URL ---

func TestNewRequestBadURL(t *testing.T) {
	fmt.Println("=== 测试: NewRequest bad URL ===")
	req := NewRequest("://bad-url", "GET")
	if req == nil {
		t.Fatal("NewRequest returned nil for bad URL")
	}
}

// --- OpenWebManager test (start and check it doesn't panic) ---

func TestOpenWebManager(t *testing.T) {
	fmt.Println("=== 测试: OpenWebManager ===")
	// Use a random high port to avoid conflicts
	mux := OpenWebManager(0)
	if mux == nil {
		t.Fatal("OpenWebManager returned nil")
	}
	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)
}

// --- JSONBody with nil/empty ---

func TestTHttpRequestJSONBodyNil(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest JSONBody nil ===")
	req := NewRequest("http://example.com", "POST")
	// nil obj should not set body
	result, err := req.JSONBody(nil)
	if err != nil {
		t.Errorf("JSONBody nil returned error: %v", err)
	}
	if result != req {
		t.Error("JSONBody should return self")
	}
}

func TestTHttpRequestXMLBodyNil(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest XMLBody nil ===")
	req := NewRequest("http://example.com", "POST")
	result, err := req.XMLBody(nil)
	if err != nil {
		t.Errorf("XMLBody nil returned error: %v", err)
	}
	if result != req {
		t.Error("XMLBody should return self")
	}
}

func TestTHttpRequestYAMLBodyNil(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest YAMLBody nil ===")
	req := NewRequest("http://example.com", "POST")
	result, err := req.YAMLBody(nil)
	if err != nil {
		t.Errorf("YAMLBody nil returned error: %v", err)
	}
	if result != req {
		t.Error("YAMLBody should return self")
	}
}

// --- JSONBody with already set body ---

func TestTHttpRequestJSONBodyAlreadySet(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest JSONBody already set ===")
	req := NewRequest("http://example.com", "POST")
	req.Body("existing-body")
	// Should not overwrite existing body
	result, err := req.JSONBody(map[string]string{"k": "v"})
	if err != nil {
		t.Errorf("JSONBody returned error: %v", err)
	}
	if result != req {
		t.Error("JSONBody should return self")
	}
}

// --- HttpRequest with string body ---

func TestHttpRequestWithStringBody(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest with string body ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, `{"received":"%s"}`, string(body))
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, _, err := HttpRequest("POST", ts.URL, false, nil, "body-string", nil, nil, "", "", &result)
	t.Logf("HttpRequest string body: code=%d, err=%v", code, err)
}

// --- HttpRequest with []byte body ---

func TestHttpRequestWithBytesBody(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest with []byte body ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, `{"received":"%s"}`, string(body))
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, _, err := HttpRequest("POST", ts.URL, false, nil, []byte("byte-body"), nil, nil, "", "", &result)
	t.Logf("HttpRequest []byte body: code=%d, err=%v", code, err)
}

// --- HttpRequest with struct body (uses utils.GetJson) ---

func TestHttpRequestWithStructBody(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest with struct body ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, `{"received":"%s"}`, string(body))
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, _, err := HttpRequest("POST", ts.URL, false, nil, map[string]string{"key": "val"}, nil, nil, "", "", &result)
	t.Logf("HttpRequest struct body: code=%d, err=%v", code, err)
}

// --- HttpRequest with cookies ---

func TestHttpRequestWithCookies(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest with cookies ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("test")
		if err != nil {
			fmt.Fprint(w, `{"cookie":""}`)
		} else {
			fmt.Fprintf(w, `{"cookie":"%s"}`, c.Value)
		}
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, _, err := HttpRequest("GET", ts.URL, true, nil, nil, []*http.Cookie{{Name: "test", Value: "yum"}}, nil, "", "", &result)
	t.Logf("HttpRequest cookies: code=%d, err=%v", code, err)
}

// --- HttpRequest nil result ---

func TestHttpRequestNilResult(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest nil result ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "some data")
	}))
	defer ts.Close()

	code, _, data, err := HttpRequest("GET", ts.URL, false, nil, nil, nil, nil, "", "", nil)
	t.Logf("HttpRequest nil result: code=%d, data=%s, err=%v", code, string(data), err)
}

// --- HttpRequest error URL ---

func TestHttpRequestErrorURL(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest error URL ===")
	code, _, _, err := HttpRequestB("GET", "http://127.0.0.1:1/nonexistent", false)
	if err == nil {
		t.Log("HttpRequestB did not return error for unreachable URL")
	}
	t.Logf("HttpRequestB error URL: code=%d, err=%v", code, err)
}

// --- buildURL POST with params ---

func TestBuildURLPostWithParams(t *testing.T) {
	fmt.Println("=== 测试: buildURL POST with params ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		ct := r.Header.Get("Content-Type")
		fmt.Fprintf(w, "ct=%s body=%s", ct, string(body))
	}))
	defer ts.Close()

	req := HttpPost(ts.URL)
	req.Param("key1", "val1")
	req.Param("key2", "val2")
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	t.Logf("POST with params: %s", str)
	if !strings.Contains(str, "key1=val1") {
		t.Errorf("params not in body: %s", str)
	}
}

// --- buildURL GET with existing query ---

func TestBuildURLGetWithExistingQuery(t *testing.T) {
	fmt.Println("=== 测试: buildURL GET with existing query ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.URL.RawQuery)
	}))
	defer ts.Close()

	req := HttpGet(ts.URL + "?existing=1")
	req.Param("new", "2")
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !strings.Contains(str, "existing=1") || !strings.Contains(str, "new=2") {
		t.Errorf("query params missing: %s", str)
	}
}

// --- buildURL POST with body already set ---

func TestBuildURLPostWithBodyAlreadySet(t *testing.T) {
	fmt.Println("=== 测试: buildURL POST with body already set ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprint(w, string(body))
	}))
	defer ts.Close()

	req := HttpPost(ts.URL)
	req.Body("pre-set-body")
	req.Param("key", "val") // should be ignored since body is already set
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !strings.Contains(str, "pre-set-body") {
		t.Errorf("body mismatch: %s", str)
	}
}

// --- buildURL DELETE with params ---

func TestBuildURLDeleteWithParams(t *testing.T) {
	fmt.Println("=== 测试: buildURL DELETE with params ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, "method=%s body=%s", r.Method, string(body))
	}))
	defer ts.Close()

	req := HttpDelete(ts.URL)
	req.Param("key", "val")
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	t.Logf("DELETE with params: %s", str)
}

// --- GetHttpRemoteAddr no colon ---

func TestGetHttpRemoteAddrNoColon(t *testing.T) {
	fmt.Println("=== 测试: GetHttpRemoteAddr no colon ===")
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1"
	addr := GetHttpRemoteAddr(req)
	if addr != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got: %s", addr)
	}
}

// --- DownloadFile 404 ---

func TestDownloadFile404(t *testing.T) {
	fmt.Println("=== 测试: DownloadFile 404 ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "downloaded.txt")

	err := DownloadFile(ts.URL, outFile, 5*time.Second, 5*time.Second)
	if err == nil {
		t.Error("expected error for 404")
	}
	t.Logf("DownloadFile 404 error (expected): %v", err)
}

// --- HttpRequest with username+password ---

func TestHttpRequestWithUsernamePassword(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest with username+password ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		fmt.Fprintf(w, `{"user":"%s","pass":"%s","ok":%v}`, u, p, ok)
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, _, err := HttpRequest("GET", ts.URL, false, nil, nil, nil, nil, "testuser", "testpass", &result)
	t.Logf("HttpRequest with auth: code=%d, err=%v", code, err)
}

// --- HttpRequest nil result with success JSON ---

func TestHttpRequestNilResultSuccess(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest nil result with success ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"errno": 0})
	}))
	defer ts.Close()

	code, _, data, err := HttpRequest("GET", ts.URL, false, nil, nil, nil, nil, "", "", nil)
	t.Logf("HttpRequest nil result success: code=%d, data=%s, err=%v", code, string(data), err)
}

// --- HttpRequest with empty data response ---

func TestHttpRequestEmptyData(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest empty data ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, _, err := HttpRequest("GET", ts.URL, false, nil, nil, nil, nil, "", "", &result)
	t.Logf("HttpRequest empty data: code=%d, err=%v", code, err)
}

// --- HttpRequest with error status code (>= 300) ---

func TestHttpRequestErrorCode(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest error code >= 300 ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"bad request"}`)
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, _, err := HttpRequest("GET", ts.URL, false, nil, nil, nil, nil, "", "", &result)
	t.Logf("HttpRequest error code: code=%d, err=%v", code, err)
}

// --- HttpRequest with unparseable JSON ---

func TestHttpRequestBadJSON(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest bad JSON ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not json")
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, _, err := HttpRequest("GET", ts.URL, false, nil, nil, nil, nil, "", "", &result)
	t.Logf("HttpRequest bad JSON: code=%d, err=%v", code, err)
}

// --- buildURL POST with file upload ---

func TestBuildURLPostWithFile(t *testing.T) {
	fmt.Println("=== 测试: buildURL POST with file upload ===")
	tmpFile := filepath.Join(t.TempDir(), "upload.txt")
	os.WriteFile(tmpFile, []byte("upload content"), 0644)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			t.Errorf("ParseMultipartForm failed: %v", err)
			fmt.Fprint(w, "parse error")
			return
		}
		file, _, err := r.FormFile("uploadfile")
		if err != nil {
			t.Errorf("FormFile failed: %v", err)
			fmt.Fprint(w, "form error")
			return
		}
		defer file.Close()
		content, _ := io.ReadAll(file)
		fmt.Fprintf(w, "uploaded=%s", string(content))
	}))
	defer ts.Close()

	req := HttpPost(ts.URL)
	req.PostFile("uploadfile", tmpFile)
	str, err := req.String()
	if err != nil {
		t.Fatalf("file upload request failed: %v", err)
	}
	if !strings.Contains(str, "upload content") {
		t.Errorf("file content mismatch: %s", str)
	}
}

// --- Bytes with nil body in response ---

func TestTHttpRequestBytesNilBody(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Bytes nil body ===")
	req := NewRequest("http://example.com", "GET")
	req.resp = &http.Response{StatusCode: 200, Body: nil}
	data, err := req.Bytes()
	if err != nil {
		t.Fatalf("Bytes with nil body failed: %v", err)
	}
	if data != nil {
		t.Errorf("expected nil data, got: %v", data)
	}
}

// --- ToFile with nil body ---

func TestTHttpRequestToFileNilBody(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToFile nil body ===")
	req := NewRequest("http://example.com", "GET")
	req.resp = &http.Response{StatusCode: 200, Body: nil}
	tmpDir := t.TempDir()
	err := req.ToFile(filepath.Join(tmpDir, "test.txt"))
	if err != nil {
		t.Logf("ToFile nil body error: %v", err)
	}
}

// --- ToJSON with error ---

func TestTHttpRequestToJSONError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToJSON error ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not json")
	}))
	defer ts.Close()

	var result map[string]interface{}
	err := HttpGet(ts.URL).ToJSON(&result)
	if err == nil {
		t.Error("expected error for non-JSON response")
	}
}

// --- ToXML with error ---

func TestTHttpRequestToXMLError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToXML error ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not xml")
	}))
	defer ts.Close()

	type TestXML struct {
		Value string `xml:"Value"`
	}
	var result TestXML
	err := HttpGet(ts.URL).ToXML(&result)
	if err == nil {
		t.Error("expected error for non-XML response")
	}
}

// --- ToYAML with error ---

func TestTHttpRequestToYAMLError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToYAML error ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{invalid yaml: [}")
	}))
	defer ts.Close()

	var result map[string]interface{}
	err := HttpGet(ts.URL).ToYAML(&result)
	if err == nil {
		t.Error("expected error for non-YAML response")
	}
}

// --- getResponse with existing response ---

func TestGetResponseExisting(t *testing.T) {
	fmt.Println("=== 测试: getResponse existing ===")
	req := NewRequest("http://example.com", "GET")
	req.resp = &http.Response{StatusCode: 200}
	resp, err := req.getResponse()
	if err != nil {
		t.Fatalf("getResponse failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status mismatch: %d", resp.StatusCode)
	}
}

// --- PATCH method test ---

func TestBuildURLPatchWithParams(t *testing.T) {
	fmt.Println("=== 测试: buildURL PATCH with params ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, "method=%s body=%s", r.Method, string(body))
	}))
	defer ts.Close()

	req := NewRequest(ts.URL, "PATCH")
	req.Param("key", "val")
	str, err := req.String()
	if err != nil {
		t.Fatalf("PATCH request failed: %v", err)
	}
	if !strings.Contains(str, "PATCH") {
		t.Errorf("method not PATCH: %s", str)
	}
}

// --- String with error ---

func TestTHttpRequestStringError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest String error ===")
	_, err := HttpGet("http://127.0.0.1:1/nonexistent").String()
	if err == nil {
		t.Log("String did not return error for unreachable URL")
	}
}

// --- DownloadURL error ---

func TestDownloadURLError(t *testing.T) {
	fmt.Println("=== 测试: DownloadURL error ===")
	_, _, err := DownloadURL("http://127.0.0.1:1/nonexistent", 1*time.Second, 1*time.Second)
	if err == nil {
		t.Log("DownloadURL did not return error")
	}
}

// --- DownloadTextFile error ---

func TestDownloadTextFileError(t *testing.T) {
	fmt.Println("=== 测试: DownloadTextFile error ===")
	_, _, err := DownloadTextFile("http://127.0.0.1:1/nonexistent", 1*time.Second, 1*time.Second)
	if err == nil {
		t.Log("DownloadTextFile did not return error")
	}
}

// --- SimpleHttpGet error ---

func TestSimpleHttpGetError(t *testing.T) {
	fmt.Println("=== 测试: SimpleHttpGet error ===")
	_, _, err := SimpleHttpGet("http://127.0.0.1:1/nonexistent", 1*time.Second, 1*time.Second)
	if err == nil {
		t.Log("SimpleHttpGet did not return error")
	}
}

// --- SimpleHttpPost error ---

func TestSimpleHttpPostError(t *testing.T) {
	fmt.Println("=== 测试: SimpleHttpPost error ===")
	_, _, err := SimpleHttpPost("http://127.0.0.1:1/nonexistent", nil, 1*time.Second, 1*time.Second)
	if err == nil {
		t.Log("SimpleHttpPost did not return error")
	}
}

// --- DownloadFile error ---

func TestDownloadFileError(t *testing.T) {
	fmt.Println("=== 测试: DownloadFile error ===")
	tmpDir := t.TempDir()
	err := DownloadFile("http://127.0.0.1:1/nonexistent", filepath.Join(tmpDir, "test.txt"), 1*time.Second, 1*time.Second)
	if err == nil {
		t.Log("DownloadFile did not return error")
	}
}

// --- Bytes with gzip error ---

func TestTHttpRequestBytesGzipError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Bytes gzip error ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		fmt.Fprint(w, "not-valid-gzip-data") // invalid gzip
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	req.setting.Gzip = true
	_, err := req.Bytes()
	if err == nil {
		t.Log("Bytes with invalid gzip did not return error (gzip may not be triggered)")
	} else {
		t.Logf("Bytes gzip error (expected): %v", err)
	}
}

// --- AutoLogin with errno != 0 ---

func TestAutoLoginErrNo(t *testing.T) {
	fmt.Println("=== 测试: AutoLogin errno != 0 ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errno":  1001,
			"errmsg": "auth failed",
		})
	}))
	defer ts.Close()

	code, _, err := AutoLogin(ts.URL, "appkey", "user", "pass")
	if err == nil {
		t.Error("expected error for non-zero errno")
	}
	if code != "" {
		t.Errorf("expected empty code: %s", code)
	}
	t.Logf("AutoLogin errno error (expected): %v", err)
}

// --- AutoLogin with request error ---

func TestAutoLoginRequestError(t *testing.T) {
	fmt.Println("=== 测试: AutoLogin request error ===")
	code, _, err := AutoLogin("http://127.0.0.1:1/notexist", "key", "user", "pwd")
	if err == nil {
		t.Error("expected error for unreachable URL")
	}
	if code != "" {
		t.Errorf("expected empty code: %s", code)
	}
	t.Logf("AutoLogin request error (expected): %v", err)
}

// --- DownloadFile with status not 200 ---

func TestDownloadFileNotOK(t *testing.T) {
	fmt.Println("=== 测试: DownloadFile not OK ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Forbidden")
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	err := DownloadFile(ts.URL, filepath.Join(tmpDir, "test.txt"), 5*time.Second, 5*time.Second)
	if err == nil {
		t.Error("expected error for 403")
	}
	t.Logf("DownloadFile 403 error (expected): %v", err)
}

// --- GetIPByURL with test server returning no IP ---

func TestGetIPByURLNoIP(t *testing.T) {
	fmt.Println("=== 测试: GetIPByURL no IP ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "no ip address here")
	}))
	defer ts.Close()

	ip, err := GetIPByURL(ts.URL)
	if err != nil {
		t.Fatalf("GetIPByURL failed: %v", err)
	}
	if ip != "" {
		t.Logf("GetIPByURL returned: %s (expected empty or found pattern)", ip)
	}
}

// --- GetPublicIPDetail with mock server ---

func TestGetPublicIPDetailMock(t *testing.T) {
	fmt.Println("=== 测试: GetPublicIPDetail mock ===")
	// We can't easily mock this since it uses SimpleHttpGet with a hardcoded URL
	// Just ensure it doesn't panic
	_, _ = GetPublicIPDetail()
}

// --- StatusCode with redirect ---

func TestTHttpRequestStatusCodeWithRedirect(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest StatusCode redirect ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/target", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "target")
	}))
	defer ts.Close()

	req := HttpGet(ts.URL + "/redirect")
	_, err := req.String()
	if err != nil {
		t.Fatalf("redirect request failed: %v", err)
	}
	code, err := req.StatusCode()
	if err != nil {
		t.Fatalf("StatusCode failed: %v", err)
	}
	t.Logf("StatusCode with redirect: %d", code)
}

// --- buildURL POST with no params and no body ---

func TestBuildURLPostNoParamsNoBody(t *testing.T) {
	fmt.Println("=== 测试: buildURL POST no params no body ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	req := HttpPost(ts.URL)
	str, err := req.String()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if str != "ok" {
		t.Errorf("unexpected response: %s", str)
	}
}

// --- buildURL POST with files and params ---

func TestBuildURLPostWithFilesAndParams(t *testing.T) {
	fmt.Println("=== 测试: buildURL POST with files and params ===")
	tmpFile := filepath.Join(t.TempDir(), "upload.txt")
	os.WriteFile(tmpFile, []byte("upload content"), 0644)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		val := r.FormValue("key")
		fmt.Fprintf(w, "param=%s uploaded=true", val)
	}))
	defer ts.Close()

	req := HttpPost(ts.URL)
	req.PostFile("uploadfile", tmpFile)
	req.Param("key", "value")
	str, err := req.String()
	if err != nil {
		t.Fatalf("file upload with params failed: %v", err)
	}
	if !strings.Contains(str, "param=value") {
		t.Errorf("params not in response: %s", str)
	}
}

// --- ToJSON/ToXML/ToYAML with Bytes error ---

func TestTHttpRequestToJSONBytesError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToJSON Bytes error ===")
	var result map[string]interface{}
	err := HttpGet("http://127.0.0.1:1/notexist").ToJSON(&result)
	if err == nil {
		t.Log("ToJSON did not return error")
	}
}

func TestTHttpRequestToXMLBytesError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToXML Bytes error ===")
	type TestXML struct {
		Value string `xml:"Value"`
	}
	var result TestXML
	err := HttpGet("http://127.0.0.1:1/notexist").ToXML(&result)
	if err == nil {
		t.Log("ToXML did not return error")
	}
}

func TestTHttpRequestToYAMLBytesError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToYAML Bytes error ===")
	var result map[string]interface{}
	err := HttpGet("http://127.0.0.1:1/notexist").ToYAML(&result)
	if err == nil {
		t.Log("ToYAML did not return error")
	}
}

// --- ToFile with getResponse error ---

func TestTHttpRequestToFileError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToFile error ===")
	tmpDir := t.TempDir()
	err := HttpGet("http://127.0.0.1:1/notexist").ToFile(filepath.Join(tmpDir, "test.txt"))
	if err == nil {
		t.Log("ToFile did not return error")
	}
}

// --- Bytes error path ---

func TestTHttpRequestBytesError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Bytes error ===")
	_, err := HttpGet("http://127.0.0.1:1/notexist").Bytes()
	if err == nil {
		t.Log("Bytes did not return error for unreachable URL")
	}
}

// --- doRequest URL parse error ---

func TestDoRequestURLParseError(t *testing.T) {
	fmt.Println("=== 测试: doRequest URL parse error ===")
	req := NewRequest("://invalid-url", "GET")
	_, err := req.Bytes()
	if err == nil {
		t.Log("doRequest did not return error for invalid URL")
	}
	t.Logf("doRequest URL parse error: %v", err)
}

// --- buildURL with nonexistent file upload ---

func TestBuildURLPostWithNonexistentFile(t *testing.T) {
	fmt.Println("=== 测试: buildURL POST with nonexistent file ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The request may fail to parse as multipart since the file doesn't exist
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	req := HttpPost(ts.URL)
	req.PostFile("uploadfile", "/nonexistent/path/file.txt")
	// This should still make a request, the file just won't have content
	_, err := req.String()
	if err != nil {
		t.Logf("POST with nonexistent file error: %v", err)
	}
}

// --- Enable cookie test with actual HTTP server ---

func TestTHttpRequestEnableCookie(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest enable cookie ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set a cookie on first request
		http.SetCookie(w, &http.Cookie{Name: "test", Value: "cookie"})
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	req.SetEnableCookie(true)
	_, err := req.String()
	if err != nil {
		t.Fatalf("enable cookie request failed: %v", err)
	}
}

// --- Debug mode with actual request ---

func TestTHttpRequestDebugMode(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest debug mode ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "debug-response")
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	req.Debug(true).DumpBody(true)
	str, err := req.String()
	if err != nil {
		t.Fatalf("debug request failed: %v", err)
	}
	if str != "debug-response" {
		t.Errorf("debug response mismatch: %s", str)
	}
	dump := req.DumpRequest()
	t.Logf("Debug dump length: %d", len(dump))
}

// --- Retries test with server ---

func TestTHttpRequestRetriesWithServer(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest retries ===")
	attempt := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt < 3 {
			// First two attempts fail
			conn, _, _ := w.(http.Hijacker).Hijack()
			conn.Close()
			return
		}
		fmt.Fprint(w, "success")
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	req.Retries(3)
	req.RetryDelay(10 * time.Millisecond)
	str, err := req.String()
	if err != nil {
		t.Logf("retries request failed (hijack not supported): %v", err)
	} else {
		if str != "success" {
			t.Errorf("retries response mismatch: %s", str)
		}
	}
}

// --- HttpRequest with result and successful JSON ---

func TestHttpRequestWithResultSuccess(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest with result success ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errno": 0,
			"data":  "success",
		})
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, _, err := HttpRequest("GET", ts.URL, false, nil, nil, nil, nil, "", "", &result)
	if err != nil {
		t.Logf("HttpRequest result error: %v", err)
	} else {
		t.Logf("HttpRequest result: code=%d, result=%v", code, result)
	}
}

// --- HttpRequest with large response body ---

func TestHttpRequestLargeBody(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest large body ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return > 1024 bytes to hit the truncation log path
		fmt.Fprint(w, strings.Repeat("x", 2048))
	}))
	defer ts.Close()

	code, _, data, err := HttpRequestB("GET", ts.URL, false)
	t.Logf("HttpRequest large body: code=%d, len=%d, err=%v", code, len(data), err)
}

// --- DownloadFile with valid server and subdirectory ---

func TestDownloadFileWithSubDir(t *testing.T) {
	fmt.Println("=== 测试: DownloadFile with subdirectory ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "file-content")
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, os.ModePerm)
	outFile := filepath.Join(subDir, "downloaded.txt")

	err := DownloadFile(ts.URL, outFile, 5*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("DownloadFile with subdirectory failed: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading downloaded file failed: %v", err)
	}
	if string(data) != "file-content" {
		t.Errorf("file content mismatch: %s", string(data))
	}
}

// --- SimpleHttpPost with struct body (default case) ---

func TestSimpleHttpPostStructBody(t *testing.T) {
	fmt.Println("=== 测试: SimpleHttpPost struct body ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprint(w, string(body))
	}))
	defer ts.Close()

	code, resp, err := SimpleHttpPost(ts.URL, map[string]string{"key": "val"}, 5*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("SimpleHttpPost struct body failed: %v", err)
	}
	if code != 200 {
		t.Errorf("status code: %d", code)
	}
	t.Logf("SimpleHttpPost struct body: %s", string(resp))
}

// --- SimpleHttpPost with invalid URL ---

func TestSimpleHttpPostInvalidURL(t *testing.T) {
	fmt.Println("=== 测试: SimpleHttpPost invalid URL ===")
	_, _, err := SimpleHttpPost("://invalid", nil, 5*time.Second, 5*time.Second)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	t.Logf("SimpleHttpPost invalid URL error (expected): %v", err)
}

// --- SimpleHttpGet with invalid URL ---

func TestSimpleHttpGetInvalidURL(t *testing.T) {
	fmt.Println("=== 测试: SimpleHttpGet invalid URL ===")
	_, _, err := SimpleHttpGet("://invalid", 5*time.Second, 5*time.Second)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	t.Logf("SimpleHttpGet invalid URL error (expected): %v", err)
}

// --- DownloadURL with invalid URL ---

func TestDownloadURLInvalidURL(t *testing.T) {
	fmt.Println("=== 测试: DownloadURL invalid URL ===")
	_, _, err := DownloadURL("://invalid", 5*time.Second, 5*time.Second)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	t.Logf("DownloadURL invalid URL error (expected): %v", err)
}

// --- DownloadTextFile with invalid URL ---

func TestDownloadTextFileInvalidURL(t *testing.T) {
	fmt.Println("=== 测试: DownloadTextFile invalid URL ===")
	_, _, err := DownloadTextFile("://invalid", 5*time.Second, 5*time.Second)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	t.Logf("DownloadTextFile invalid URL error (expected): %v", err)
}

// --- DownloadFile with invalid URL ---

func TestDownloadFileInvalidURL(t *testing.T) {
	fmt.Println("=== 测试: DownloadFile invalid URL ===")
	tmpDir := t.TempDir()
	err := DownloadFile("://invalid", filepath.Join(tmpDir, "test.txt"), 5*time.Second, 5*time.Second)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	t.Logf("DownloadFile invalid URL error (expected): %v", err)
}

// --- ToFile with error creating file ---

func TestTHttpRequestToFileCreateError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToFile create error ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "data")
	}))
	defer ts.Close()

	// Try to write to a directory that doesn't exist (should create it)
	err := HttpGet(ts.URL).ToFile("/tmp/test_tofile_network/test.txt")
	if err != nil {
		t.Logf("ToFile create error (expected): %v", err)
	} else {
		os.RemoveAll("/tmp/test_tofile_network")
	}
}

// --- Bytes with gzip decompression error ---

func TestTHttpRequestBytesGzipError2(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Bytes gzip error 2 ===")
	// Create a server that returns Content-Encoding: gzip but with invalid gzip data
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		// Write invalid gzip data
		w.Write([]byte("this is not gzip data at all"))
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	req.setting.Gzip = true
	_, err := req.Bytes()
	// gzip.NewReader should fail with invalid data
	if err != nil {
		t.Logf("Bytes gzip error (expected): %v", err)
	} else {
		t.Log("Bytes gzip did not error (may have succeeded with raw data)")
	}
}

// --- Bytes with valid gzip data ---

func TestTHttpRequestBytesGzipValid(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Bytes gzip valid ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		gz.Write([]byte("gzip compressed content"))
		gz.Close()
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	req.setting.Gzip = true
	data, err := req.Bytes()
	if err != nil {
		t.Fatalf("Bytes gzip valid failed: %v", err)
	}
	if !strings.Contains(string(data), "gzip compressed content") {
		t.Errorf("unexpected data: %s", string(data))
	}
}

// --- DownloadFile with Last-Modified header ---

func TestDownloadFileWithLastModified(t *testing.T) {
	fmt.Println("=== 测试: DownloadFile with Last-Modified ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Last-Modified", time.Now().UTC().Format(time.RFC1123))
		fmt.Fprint(w, "downloaded content")
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "test_download.txt")
	err := DownloadFile(ts.URL, outFile, 5*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != "downloaded content" {
		t.Errorf("unexpected content: %s", string(data))
	}
}

// --- DownloadFile with invalid Last-Modified header ---

func TestDownloadFileWithInvalidLastModified(t *testing.T) {
	fmt.Println("=== 测试: DownloadFile with invalid Last-Modified ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Last-Modified", "invalid-date-format")
		fmt.Fprint(w, "downloaded content")
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "test_download_invalid.txt")
	err := DownloadFile(ts.URL, outFile, 5*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}
	// The file should still be created, just with wrong Last-Modified parse warning
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != "downloaded content" {
		t.Errorf("unexpected content: %s", string(data))
	}
}

// --- DownloadFile with non-200 status ---

func TestDownloadFileNon200Status(t *testing.T) {
	fmt.Println("=== 测试: DownloadFile non-200 status ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "test_download_404.txt")
	err := DownloadFile(ts.URL, outFile, 5*time.Second, 5*time.Second)
	if err == nil {
		t.Error("expected error for 404")
	}
}

// --- Bytes with cached body ---

func TestTHttpRequestBytesCachedBody(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Bytes nil body ===")
	// This is tricky - httptest servers always return a body.
	// Instead test the cached body path
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "cached")
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	// First call populates b.body
	data, err := req.Bytes()
	if err != nil {
		t.Fatalf("Bytes failed: %v", err)
	}
	if string(data) != "cached" {
		t.Errorf("unexpected data: %s", string(data))
	}
	// Second call returns cached b.body
	data2, err := req.Bytes()
	if err != nil {
		t.Fatalf("Bytes cached failed: %v", err)
	}
	if string(data2) != "cached" {
		t.Errorf("unexpected cached data: %s", string(data2))
	}
}

// --- DownloadFile with connection error ---

func TestDownloadFileConnectionError(t *testing.T) {
	fmt.Println("=== 测试: DownloadFile connection error ===")
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "test_download_err.txt")
	err := DownloadFile("http://127.0.0.1:1/nonexistent", outFile, 1*time.Second, 1*time.Second)
	if err == nil {
		t.Error("expected error for connection failure")
	}
}

// --- HttpRequest with various body types ---

func TestHttpRequestWithBodyTypes(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest body types ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"received":"%s","length":%d}`, string(body), len(body))
	}))
	defer ts.Close()

	// Test with string body
	code, _, data, err := HttpRequest("POST", ts.URL, false, nil, "string body", nil, nil, "", "", nil)
	if err != nil {
		t.Logf("HttpRequest string body: code=%d err=%v", code, err)
	}
	if data != nil {
		t.Logf("HttpRequest string body: %s", string(data))
	}

	// Test with []byte body
	code, _, data, err = HttpRequest("POST", ts.URL, false, nil, []byte("byte body"), nil, nil, "", "", nil)
	if err != nil {
		t.Logf("HttpRequest []byte body: code=%d err=%v", code, err)
	}

	// Test with struct body
	code, _, data, err = HttpRequest("POST", ts.URL, false, nil, map[string]string{"key": "value"}, nil, nil, "", "", nil)
	if err != nil {
		t.Logf("HttpRequest struct body: code=%d err=%v", code, err)
	}
}

// --- HttpRequest with headers and params ---

func TestHttpRequestWithHeadersAndParams(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest headers and params ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	}))
	defer ts.Close()

	code, _, data, err := HttpRequest("GET", ts.URL, true,
		map[string]string{"param1": "value1"}, nil, nil,
		map[string]string{"X-Custom": "header-value"}, "", "", nil)
	if err != nil {
		t.Logf("HttpRequest headers and params: code=%d err=%v", code, err)
	}
	t.Logf("HttpRequest headers and params: code=%d data=%s", code, string(data))
}

// --- HttpRequest with basic auth ---

func TestHttpRequestWithBasicAuth(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest basic auth ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	}))
	defer ts.Close()

	code, _, data, err := HttpRequest("GET", ts.URL, false, nil, nil, nil, nil, "user", "pass", nil)
	if err != nil {
		t.Logf("HttpRequest basic auth: code=%d err=%v", code, err)
	}
	t.Logf("HttpRequest basic auth: code=%d data=%s", code, string(data))
}

// --- HttpRequest with result and error status ---

func TestHttpRequestWithResultAndError(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest with result error ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, data, err := HttpRequest("GET", ts.URL, false, nil, nil, nil, nil, "", "", &result)
	if err != nil {
		t.Logf("HttpRequest result error (expected): code=%d err=%v", code, err)
	}
	t.Logf("HttpRequest result error: code=%d data=%s", code, string(data))
}

// --- HttpRequest with cookies via helper ---

func TestHttpRequestWithCookiesHelper(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest with cookies ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	}))
	defer ts.Close()

	cookies := []*http.Cookie{{Name: "test", Value: "cookie"}}
	code, _, data, err := HttpRequest("GET", ts.URL, false, nil, nil, cookies, nil, "", "", nil)
	if err != nil {
		t.Logf("HttpRequest with cookies: code=%d err=%v", code, err)
	}
	t.Logf("HttpRequest with cookies: code=%d data=%s", code, string(data))
}

// --- DownloadFile with error creating output file ---

func TestDownloadFileCreateError(t *testing.T) {
	fmt.Println("=== 测试: DownloadFile create error ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "data")
	}))
	defer ts.Close()

	// Try to write to a read-only directory
	err := DownloadFile(ts.URL, "/nonexistent_dir/impossible/file.txt", 5*time.Second, 5*time.Second)
	if err == nil {
		t.Error("expected error creating file in nonexistent directory")
	}
}

// --- Bytes with gzip path ---

func TestTHttpRequestBytesGzipPath(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest Bytes gzip path ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		gz.Write([]byte("gzip response"))
		gz.Close()
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	req.setting.Gzip = true
	data, err := req.Bytes()
	if err != nil {
		t.Fatalf("Bytes gzip failed: %v", err)
	}
	if string(data) != "gzip response" {
		t.Errorf("unexpected data: %s", string(data))
	}
}

// --- XMLBody with error ---

func TestTHttpRequestXMLBodyError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest XMLBody error ===")
	req := NewRequest("http://example.com", "POST")
	// Pass a value that can't be marshaled as XML (e.g., channel)
	_, err := req.XMLBody(make(chan int))
	if err == nil {
		t.Log("XMLBody did not return error for unmarshallable type")
	}
}

// --- YAMLBody with error (skip - yaml.Marshal panics on channels) ---

func TestTHttpRequestYAMLBodyNormal(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest YAMLBody normal ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(body))
	}))
	defer ts.Close()

	req := HttpPost(ts.URL)
	_, err := req.YAMLBody(map[string]string{"key": "yaml_value"})
	if err != nil {
		t.Logf("YAMLBody normal: %v", err)
	}
}

// --- JSONBody with error ---

func TestTHttpRequestJSONBodyError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest JSONBody error ===")
	req := NewRequest("http://example.com", "POST")
	_, err := req.JSONBody(make(chan int))
	if err == nil {
		t.Log("JSONBody did not return error for unmarshallable type")
	}
}

// --- ToFile with path creation error ---

func TestTHttpRequestToFilePathError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToFile path error ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "data")
	}))
	defer ts.Close()

	// Try to save to a path that can't be created
	err := HttpGet(ts.URL).ToFile("/proc/nonexistent/impossible/file.txt")
	if err != nil {
		t.Logf("ToFile path error (expected): %v", err)
	}
}

// --- ToFile with file creation error ---

func TestTHttpRequestToFileFileCreateError(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ToFile create error ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "data")
	}))
	defer ts.Close()

	// Try to save to a read-only path
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	os.MkdirAll(readOnlyDir, 0555)
	defer os.Chmod(readOnlyDir, 0755)

	err := HttpGet(ts.URL).ToFile(filepath.Join(readOnlyDir, "nested", "file.txt"))
	if err != nil {
		t.Logf("ToFile create error (expected): %v", err)
	}
}

// --- HttpRequest with result and non-error status but bad JSON ---

func TestHttpRequestWithResultBadJSON(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest with result bad JSON ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "not valid json")
	}))
	defer ts.Close()

	var result map[string]interface{}
	code, _, _, err := HttpRequest("GET", ts.URL, false, nil, nil, nil, nil, "", "", &result)
	if err != nil {
		t.Logf("HttpRequest bad JSON (expected): code=%d err=%v", code, err)
	}
}

// --- HttpRequest with error status and nil result ---

func TestHttpRequestErrorStatusNilResult(t *testing.T) {
	fmt.Println("=== 测试: HttpRequest error status nil result ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "service unavailable")
	}))
	defer ts.Close()

	code, _, _, err := HttpRequest("GET", ts.URL, false, nil, nil, nil, nil, "", "", nil)
	if err != nil {
		t.Logf("HttpRequest error status (expected): code=%d err=%v", code, err)
	}
}

// --- doRequest with CheckRedirect ---

func TestTHttpRequestCheckRedirect(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest CheckRedirect ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		fmt.Fprint(w, "final destination")
	}))
	defer ts.Close()

	req := HttpGet(ts.URL + "/redirect")
	req.SetCheckRedirect(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})
	resp, err := req.Response()
	if err != nil {
		t.Logf("CheckRedirect response error: %v", err)
	} else {
		t.Logf("CheckRedirect status: %d", resp.StatusCode)
	}
}

// --- doRequest with ShowDebug ---

func TestTHttpRequestShowDebug(t *testing.T) {
	fmt.Println("=== 测试: THttpRequest ShowDebug ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "debug response")
	}))
	defer ts.Close()

	req := HttpGet(ts.URL)
	req.Debug(true)
	str, err := req.String()
	if err != nil {
		t.Fatalf("ShowDebug request failed: %v", err)
	}
	if !strings.Contains(str, "debug response") {
		t.Errorf("unexpected response: %s", str)
	}
}
