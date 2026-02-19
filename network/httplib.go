// 参考：http://beego.me/docs/module/httplib.md
package network

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/gh/utils"
	"gopkg.in/yaml.v2"
)

var defaultSetting = THttpSettings{
	UserAgent:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Safari/537.36 Edg/91.0.864.48",
	ConnectTimeout:   5 * time.Second,
	ReadWriteTimeout: 3 * time.Second,
	EnableCookie:     false,
	Gzip:             true,
	DumpBody:         true,
}

var defaultCookieJar http.CookieJar
var settingMutex sync.Mutex

var doRequestFilter = func(ctx context.Context, req *THttpRequest) (*http.Response, error) {
	return req.doRequest(ctx)
}

// createDefaultCookie creates a global cookiejar to store cookies.
func createDefaultCookie() {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	defaultCookieJar, _ = cookiejar.New(nil)
}

func SetDefaultSettingByTimeout(connectTimeout, readWriteTimeout time.Duration) {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	defaultSetting.ConnectTimeout = connectTimeout
	defaultSetting.ReadWriteTimeout = readWriteTimeout
}

// SetDefaultSetting 覆盖默认设置
func SetDefaultSetting(setting THttpSettings) {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	defaultSetting = setting
}

// 获取HTTP请求的客户端真实IP地址
func GetHttpRemoteAddr(req *http.Request) string {
	addr := ""
	if req != nil {
		addr = req.Header.Get("X-Forwarded-For")
		if addr == "" {
			addr = req.RemoteAddr
			addr = strings.ReplaceAll(addr, "[::1]", "127.0.0.1")
			if strings.Contains(addr, ":") {
				return fmt.Sprintf("%s", strings.Split(addr, ":")[0])
			} else {
				return addr
			}
		} else {
			return addr
		}
	} else {
		return addr
	}
}

// GetHttpRemoteAddrPort 获取HTTP请求的客户端真实IP地址和端口
func GetHttpRemoteAddrPort(req *http.Request) string {
	addr := ""
	if req != nil {
		addr = req.Header.Get("X-Forwarded-For")
		if addr == "" {
			addr = req.RemoteAddr
			addr = strings.ReplaceAll(addr, "[::1]", "127.0.0.1")
			return addr
		} else {
			return addr
		}
	} else {
		return addr
	}
}

// NewRequest 返回指定方法的 *THttpRequest
func NewRequest(rawurl, method string) *THttpRequest {
	var resp http.Response
	u, err := url.Parse(rawurl)
	if err != nil {
		logs.Warning("解析URL报错，%s", err.Error())
	}
	req := http.Request{
		URL:        u,
		Method:     method,
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	return &THttpRequest{
		url:     rawurl,
		req:     &req,
		params:  map[string][]string{},
		files:   map[string]string{},
		setting: defaultSetting,
		resp:    &resp,
	}
}

// HttpGet 返回 GET 方法的 *THttpRequest
func HttpGet(url string) *THttpRequest {
	return NewRequest(url, "GET")
}

// HttpPost 返回 POST 方法的 *THttpRequest
func HttpPost(url string) *THttpRequest {
	return NewRequest(url, "POST")
}

// HttpPut 返回 PUT 方法的 *THttpRequest
func HttpPut(url string) *THttpRequest {
	return NewRequest(url, "PUT")
}

// Delete returns *THttpRequest DELETE method.
func HttpDelete(url string) *THttpRequest {
	return NewRequest(url, "DELETE")
}

// HttpHead 返回 HEAD 方法的 *THttpRequest
func HttpHead(url string) *THttpRequest {
	return NewRequest(url, "HEAD")
}

type FilterChain func(next Filter) Filter

type Filter func(ctx context.Context, req *THttpRequest) (*http.Response, error)

// THTTPSettings is the http.Client setting
type THttpSettings struct {
	ShowDebug        bool
	UserAgent        string
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
	TLSClientConfig  *tls.Config
	Proxy            func(*http.Request) (*url.URL, error)
	Transport        http.RoundTripper
	CheckRedirect    func(req *http.Request, via []*http.Request) error
	EnableCookie     bool
	Gzip             bool
	DumpBody         bool
	Retries          int // if set to -1 means will retry forever
	RetryDelay       time.Duration
	FilterChains     []FilterChain
}

// THTTPRequest provides more useful methods for requesting one url than http.Request.
type THttpRequest struct {
	url     string
	req     *http.Request
	params  map[string][]string
	files   map[string]string
	setting THttpSettings
	resp    *http.Response
	body    []byte
	dump    []byte // debug模式会复制body
}

// GetRequest return the request object
func (b *THttpRequest) GetRequest() *http.Request {
	return b.req
}

// Response executes request client gets response mannually.
func (b *THttpRequest) GetResponse() *http.Response {
	return b.resp
}

// Setting Change request settings
func (b *THttpRequest) Setting(setting THttpSettings) *THttpRequest {
	b.setting = setting
	return b
}

// SetBasicAuth sets the request's Authorization header to use HTTP Basic Authentication with the provided username and password.
func (b *THttpRequest) SetBasicAuth(username, password string) *THttpRequest {
	b.req.SetBasicAuth(username, password)
	return b
}

// 是否会话保持，保存Cookied
func (b *THttpRequest) SetEnableCookie(enable bool) *THttpRequest {
	b.setting.EnableCookie = enable
	return b
}

// SetUserAgent sets User-Agent header field
func (b *THttpRequest) SetUserAgent(useragent string) *THttpRequest {
	b.setting.UserAgent = useragent
	return b
}

// Debug sets show debug or not when executing request.
func (b *THttpRequest) Debug(isdebug bool) *THttpRequest {
	b.setting.ShowDebug = isdebug
	return b
}

// Retries sets Retries times.
// default is 0 means no retried.
// -1 means retried forever.
// others means retried times.
func (b *THttpRequest) Retries(times int) *THttpRequest {
	b.setting.Retries = times
	return b
}

// RetryDelay sets the time to sleep between reconnection attempts
func (b *THttpRequest) RetryDelay(delay time.Duration) *THttpRequest {
	b.setting.RetryDelay = delay
	return b
}

// DumpBody setting whether need to Dump the Body.
func (b *THttpRequest) DumpBody(isdump bool) *THttpRequest {
	b.setting.DumpBody = isdump
	return b
}

// DumpRequest return the DumpRequest
func (b *THttpRequest) DumpRequest() []byte {
	return b.dump
}

// SetTimeout sets connect time out and read-write time out for TRequest.
func (b *THttpRequest) SetTimeout(connectTimeout, readWriteTimeout time.Duration) *THttpRequest {
	b.setting.ConnectTimeout = connectTimeout
	b.setting.ReadWriteTimeout = readWriteTimeout
	return b
}

// SetTLSClientConfig sets tls connection configurations if visiting https url.
func (b *THttpRequest) SetTLSClientConfig(config *tls.Config) *THttpRequest {
	b.setting.TLSClientConfig = config
	return b
}

// Header add header item string in request.
func (b *THttpRequest) Header(key, value string) *THttpRequest {
	b.req.Header.Set(key, value)
	return b
}

// SetHost set the request host
func (b *THttpRequest) SetHost(host string) *THttpRequest {
	b.req.Host = host
	return b
}

// SetProtocolVersion Set the protocol version for incoming requests.
// Client requests always use HTTP/1.1.
func (b *THttpRequest) SetProtocolVersion(vers string) *THttpRequest {
	if len(vers) == 0 {
		vers = "HTTP/1.1"
	}

	major, minor, ok := http.ParseHTTPVersion(vers)
	if ok {
		b.req.Proto = vers
		b.req.ProtoMajor = major
		b.req.ProtoMinor = minor
	}

	return b
}

// SetCookie add cookie into request.
func (b *THttpRequest) SetCookie(cookie *http.Cookie) *THttpRequest {
	b.req.Header.Add("Cookie", cookie.String())
	return b
}

// SetTransport set the setting transport
func (b *THttpRequest) SetTransport(transport http.RoundTripper) *THttpRequest {
	b.setting.Transport = transport
	return b
}

// SetProxy set the http proxy
// example:
//
//	func(req *http.Request) (*url.URL, error) {
//		u, _ := url.ParseRequestURI("http://127.0.0.1:8118")
//		return u, nil
//	}
func (b *THttpRequest) SetProxy(proxy func(*http.Request) (*url.URL, error)) *THttpRequest {
	b.setting.Proxy = proxy
	return b
}

// SetCheckRedirect specifies the policy for handling redirects.
//
// If CheckRedirect is nil, the Client uses its default policy,
// which is to stop after 10 consecutive requests.
func (b *THttpRequest) SetCheckRedirect(redirect func(req *http.Request, via []*http.Request) error) *THttpRequest {
	b.setting.CheckRedirect = redirect
	return b
}

// SetFilters will use the filter as the invocation filters
func (b *THttpRequest) SetFilters(fcs ...FilterChain) *THttpRequest {
	b.setting.FilterChains = fcs
	return b
}

// AddFilters adds filter
func (b *THttpRequest) AddFilters(fcs ...FilterChain) *THttpRequest {
	b.setting.FilterChains = append(b.setting.FilterChains, fcs...)
	return b
}

// Param adds query param in to request.
// params build query string as ?key1=value1&key2=value2...
func (b *THttpRequest) Param(key, value string) *THttpRequest {
	if param, ok := b.params[key]; ok {
		b.params[key] = append(param, value)
	} else {
		b.params[key] = []string{value}
	}
	return b
}

// PostFile add a post file to the request
func (b *THttpRequest) PostFile(formname, filename string) *THttpRequest {
	b.files[formname] = filename
	return b
}

// Body adds request raw body.
// it supports string and []byte.
func (b *THttpRequest) Body(data interface{}) *THttpRequest {
	switch t := data.(type) {
	case string:
		bf := bytes.NewBufferString(t)
		b.req.Body = ioutil.NopCloser(bf)
		b.req.ContentLength = int64(len(t))
	case []byte:
		bf := bytes.NewBuffer(t)
		b.req.Body = ioutil.NopCloser(bf)
		b.req.ContentLength = int64(len(t))
	}
	return b
}

// XMLBody adds request raw body encoding by XML.
func (b *THttpRequest) XMLBody(obj interface{}) (*THttpRequest, error) {
	if b.req.Body == nil && obj != nil {
		byts, err := xml.Marshal(obj)
		if err != nil {
			return b, err
		}
		b.req.Body = ioutil.NopCloser(bytes.NewReader(byts))
		b.req.ContentLength = int64(len(byts))
		b.req.Header.Set("Content-Type", "application/xml")
	}
	return b, nil
}

// YAMLBody adds request raw body encoding by YAML.
func (b *THttpRequest) YAMLBody(obj interface{}) (*THttpRequest, error) {
	if b.req.Body == nil && obj != nil {
		byts, err := yaml.Marshal(obj)
		if err != nil {
			return b, err
		}
		b.req.Body = ioutil.NopCloser(bytes.NewReader(byts))
		b.req.ContentLength = int64(len(byts))
		b.req.Header.Set("Content-Type", "application/x+yaml")
	}
	return b, nil
}

// JSONBody adds request raw body encoding by JSON.
func (b *THttpRequest) JSONBody(obj interface{}) (*THttpRequest, error) {
	if b.req.Body == nil && obj != nil {
		byts, err := json.Marshal(obj)
		if err != nil {
			return b, err
		}
		b.req.Body = ioutil.NopCloser(bytes.NewReader(byts))
		b.req.ContentLength = int64(len(byts))
		b.req.Header.Set("Content-Type", "application/json")
	}
	return b, nil
}

func (b *THttpRequest) buildURL(paramBody string) {
	// build GET url with query string
	if b.req.Method == "GET" && len(paramBody) > 0 {
		if strings.Contains(b.url, "?") {
			b.url += "&" + paramBody
		} else {
			b.url = b.url + "?" + paramBody
		}
		return
	}

	// build POST/PUT/PATCH url and body
	if (b.req.Method == "POST" || b.req.Method == "PUT" || b.req.Method == "PATCH" || b.req.Method == "DELETE") && b.req.Body == nil {
		// with files
		if len(b.files) > 0 {
			pr, pw := io.Pipe()
			bodyWriter := multipart.NewWriter(pw)
			go func() {
				for formname, filename := range b.files {
					//CreateFormFile 是一个围绕 CreatePart 的便捷包装器。
					//它使用提供的字段名和文件名创建一个新的表单数据头。
					fileWriter, err := bodyWriter.CreateFormFile(formname, filename)
					if err != nil {
						logs.Warning("创建新的表单数据头报错，%s", err.Error())
					}
					fh, err := os.Open(filename)
					if err != nil {
						logs.Warning("打开文件(%s)报错，%s", filename, err.Error())
					}
					//iocopy
					_, err = io.Copy(fileWriter, fh)
					fh.Close()
					if err != nil {
						logs.Warning("上传文件(%s)报错，%s", filename, err.Error())
					}
				}
				for k, v := range b.params {
					for _, vv := range v {
						bodyWriter.WriteField(k, vv)
					}
				}
				bodyWriter.Close()
				pw.Close()
			}()
			b.Header("Content-Type", bodyWriter.FormDataContentType())
			b.req.Body = ioutil.NopCloser(pr)
			return
		}

		// with params
		if len(paramBody) > 0 {
			b.Header("Content-Type", "application/x-www-form-urlencoded")
			b.Body(paramBody)
		}
	}
}

func (b *THttpRequest) getResponse() (*http.Response, error) {
	if b.resp.StatusCode != 0 {
		return b.resp, nil
	}
	resp, err := b.DoRequest()
	if err != nil {
		return nil, err
	}
	b.resp = resp
	return resp, nil
}

// DoRequest executes client.Do
func (b *THttpRequest) DoRequest() (resp *http.Response, err error) {
	return b.DoRequestWithCtx(context.Background())
}

func (b *THttpRequest) DoRequestWithCtx(ctx context.Context) (resp *http.Response, err error) {

	root := doRequestFilter
	if len(b.setting.FilterChains) > 0 {
		for i := len(b.setting.FilterChains) - 1; i >= 0; i-- {
			root = b.setting.FilterChains[i](root)
		}
	}
	return root(ctx, b)
}

func (b *THttpRequest) doRequest(ctx context.Context) (resp *http.Response, err error) {
	var paramBody string
	if len(b.params) > 0 {
		var buf bytes.Buffer
		for k, v := range b.params {
			for _, vv := range v {
				buf.WriteString(url.QueryEscape(k))
				buf.WriteByte('=')
				buf.WriteString(url.QueryEscape(vv))
				buf.WriteByte('&')
			}
		}
		paramBody = buf.String()
		paramBody = paramBody[0 : len(paramBody)-1]
	}

	b.buildURL(paramBody)
	urlParsed, err := url.Parse(b.url)
	if err != nil {
		return nil, err
	}

	b.req.URL = urlParsed

	trans := b.setting.Transport

	if trans == nil {
		// create default transport
		trans = &http.Transport{
			TLSClientConfig:     b.setting.TLSClientConfig,
			Proxy:               b.setting.Proxy,
			Dial:                TimeoutDialer(b.setting.ConnectTimeout, b.setting.ReadWriteTimeout),
			MaxIdleConnsPerHost: 100,
		}
	} else {
		// if b.transport is *http.Transport then set the settings.
		if t, ok := trans.(*http.Transport); ok {
			if t.TLSClientConfig == nil {
				t.TLSClientConfig = b.setting.TLSClientConfig
			}
			if t.Proxy == nil {
				t.Proxy = b.setting.Proxy
			}
			if t.Dial == nil {
				t.Dial = TimeoutDialer(b.setting.ConnectTimeout, b.setting.ReadWriteTimeout)
			}
		}
	}

	var jar http.CookieJar
	if b.setting.EnableCookie {
		if defaultCookieJar == nil {
			createDefaultCookie()
		}
		jar = defaultCookieJar
	}

	client := &http.Client{
		Transport: trans,
		Jar:       jar,
	}

	if b.setting.UserAgent != "" && b.req.Header.Get("User-Agent") == "" {
		b.req.Header.Set("User-Agent", b.setting.UserAgent)
	}

	if b.setting.CheckRedirect != nil {
		client.CheckRedirect = b.setting.CheckRedirect
	}

	if b.setting.ShowDebug {
		dump, err := httputil.DumpRequest(b.req, b.setting.DumpBody)
		if err != nil {
			logs.Warning("同步数据报错，%s", err.Error())
		}
		b.dump = dump
	}
	// retries default value is 0, it will run once.
	// retries equal to -1, it will run forever until success
	// retries is setted, it will retries fixed times.
	for i := 0; b.setting.Retries == -1 || i <= b.setting.Retries; i++ {
		resp, err = client.Do(b.req)
		if err == nil {
			break
		}
	}
	return resp, err
}

// String returns the body string in response.
// it calls Response inner.
func (b *THttpRequest) String() (string, error) {
	data, err := b.Bytes()
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Bytes returns the body []byte in response.
// it calls Response inner.
func (b *THttpRequest) Bytes() ([]byte, error) {
	if b.body != nil {
		return b.body, nil
	}
	resp, err := b.getResponse()
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, nil
	}
	defer resp.Body.Close()

	if b.setting.Gzip && resp.Header.Get("Content-Encoding") == "gzip" {
		logs.FDebug("<=== 读取GZIP数据")
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		b.body, err = ioutil.ReadAll(reader)
		return b.body, err
	}

	b.body, err = ioutil.ReadAll(resp.Body)
	logs.FDebug("<=== 读取数据长度为 %d 字节", len(b.body))
	return b.body, err
}

// ToFile saves the body data in response to one file.
// it calls Response inner.
func (b *THttpRequest) ToFile(filename string) error {
	resp, err := b.getResponse()
	if err != nil {
		return err
	}
	if resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()
	err = pathExistAndMkdir(filename)
	if err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// ToJSON returns the map that marshals from the body bytes as json in response .
// it calls Response inner.
func (b *THttpRequest) ToJSON(v interface{}) error {
	data, err := b.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToXML returns the map that marshals from the body bytes as xml in response .
// it calls Response inner.
func (b *THttpRequest) ToXML(v interface{}) error {
	data, err := b.Bytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// ToYAML returns the map that marshals from the body bytes as yaml in response .
// it calls Response inner.
func (b *THttpRequest) ToYAML(v interface{}) error {
	data, err := b.Bytes()
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

// 返回状态码（可以真实的返回Http状态）
func (b *THttpRequest) StatusCode() (int, error) {
	if b.resp.Request != nil && b.resp.Request.Response != nil {
		return b.resp.Request.Response.StatusCode, nil
	} else {
		return b.resp.StatusCode, nil
	}
}

// Response executes request client gets response mannually.
func (b *THttpRequest) Response() (*http.Response, error) {
	return b.getResponse()
}

// TimeoutDialer returns functions of connection dialer with timeout settings for http.Transport Dial field.
func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		err = conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, err
	}
}

// Check that the file directory exists, there is no automatically created
func pathExistAndMkdir(filename string) (err error) {
	filename = path.Dir(filename)
	_, err = os.Stat(filename)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(filename, os.ModePerm)
		if err == nil {
			return nil
		}
	}
	return err
}

// 单句调用Http请求(有返回值)
func HttpRequestA(method, url string, is_cookie bool, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, nil, nil, nil, nil, "", "", result)
}

// 单句调用Http请求(无返回值)
func HttpRequestB(method, url string, is_cookie bool) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, nil, nil, nil, nil, "", "", nil)
}

// 单句调用Http请求(带Header，有返回值)
func HttpRequestHA(method, url string, is_cookie bool, header map[string]string, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, nil, nil, nil, header, "", "", result)
}

// 单句调用Http请求(带Header，无返回值)
func HttpRequestHB(method, url string, is_cookie bool, header map[string]string) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, nil, nil, nil, header, "", "", nil)
}

// 单句调用Http请求(带Params，有返回值)
func HttpRequestPB(method, url string, is_cookie bool, params map[string]string, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, params, nil, nil, nil, "", "", result)
}

// 单句调用Http请求(带Params，Cookies，有返回值)
func HttpRequestPC(method, url string, is_cookie bool, params map[string]string, cookies []*http.Cookie, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, params, nil, cookies, nil, "", "", result)
}

// 单句调用Http请求(带Params，用户密码，有返回值)
func HttpRequestPD(method, url string, is_cookie bool, params map[string]string, username, password string, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, params, nil, nil, nil, username, password, result)
}

// 单句调用Http请求(带Params，Header，有返回值)
func HttpRequestPHB(method, url string, is_cookie bool, params map[string]string, header map[string]string, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, params, nil, nil, header, "", "", result)
}

// 单句调用Http请求(带Params，Cookies，Header，有返回值)
func HttpRequestPHC(method, url string, is_cookie bool, params map[string]string, cookies []*http.Cookie, header map[string]string, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, params, nil, cookies, header, "", "", result)
}

// 单句调用Http请求(带Params，Header，用户密码，有返回值)
func HttpRequestPHD(method, url string, is_cookie bool, params map[string]string, header map[string]string, username, password string, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, params, nil, nil, header, username, password, result)
}

// 单句调用Http请求(带Body，Header，有返回值)
func HttpRequestBHB(method, url string, is_cookie bool, body interface{}, header map[string]string, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, nil, body, nil, header, "", "", result)
}

// 单句调用Http请求(带Body，Cookies，Header，有返回值)
func HttpRequestBHC(method, url string, is_cookie bool, body interface{}, cookies []*http.Cookie, header map[string]string, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, nil, body, cookies, header, "", "", result)
}

// 单句调用Http请求(带Body，Header，用户密码，有返回值)
func HttpRequestBHD(method, url string, is_cookie bool, body interface{}, header map[string]string, username, password string, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, nil, body, nil, header, username, password, result)
}

// 单句调用Http请求(带Body，有返回值)
func HttpRequestBB(method, url string, is_cookie bool, body interface{}, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, nil, body, nil, nil, "", "", result)
}

// 单句调用Http请求(带Body，Cookies，有返回值)
func HttpRequestBC(method, url string, is_cookie bool, body interface{}, cookies []*http.Cookie, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, nil, body, cookies, nil, "", "", result)
}

// 单句调用Http请求(带Body，用户密码，有返回值)
func HttpRequestBD(method, url string, is_cookie bool, body interface{}, username, password string, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, nil, body, nil, nil, username, password, result)
}

// 单句调用Http请求(带Body，Header，有返回值)
func HttpRequestPBHB(method, url string, is_cookie bool, params map[string]string, body interface{}, header map[string]string, result interface{}) (int, *http.Response, []byte, error) {
	return HttpRequest(method, url, is_cookie, params, body, nil, header, "", "", result)
}

func HttpRequest(method, url string, is_cookie bool,
	params map[string]string, body interface{},
	cookies []*http.Cookie, header map[string]string, username, password string,
	result interface{}) (int, *http.Response, []byte, error) {

	logs.FDebug("= %s - %s", method, url)
	//新建Http请求
	req := NewRequest(url, method)

	if len(header) > 0 {
		for k, v := range header {
			logs.FDebug("===> SetHeader(%s) : %v", k, v)
			req.Header(k, v)
		}
	}

	if body != nil {
		req.Header("Content-Type", "application/json")
		req.Header("accept", "*/*")
		logs.FDebug("===> SetBody")
		var body_str string
		switch body.(type) {
		case string:
			body_str = body.(string)
		case []byte:
			body_str = string(body.([]byte))
		default:
			body_str = utils.GetJson(body)
		}
		req.Body(body_str)
	}
	if username+password != "" {
		logs.FDebug("===> SetBasicAuth(%s/%s)", username, password)
		req.SetBasicAuth(username, password)
	}

	logs.FDebug("===> SetEnableCookie(%v)", is_cookie)
	req.SetEnableCookie(is_cookie)

	if len(cookies) > 0 {
		logs.FDebug("===> SetCookie")
		for _, cookie := range cookies {
			req.SetCookie(cookie)
		}
	}
	if len(params) > 0 {
		for k, v := range params {
			logs.FDebug("===> SetParams(%s) : %v", k, v)
			req.Param(k, v)
		}
	}

	data, err := req.Bytes()
	if err != nil {
		return 0, req.GetResponse(), nil, fmt.Errorf("获取返回数据错误，%s", utils.GetNetError(err))
	}
	if len(data) > 1024 {
		logs.FDebug("<=== 返回部分数据：[%s]", string(data[:1024]))
	} else {
		logs.FDebug("<=== 返回全部数据：[%s]", string(data))
	}

	state_code, err := req.StatusCode()
	if err != nil {
		return 0, req.GetResponse(), data, fmt.Errorf("获取返回码错误，%s", err.Error())
	}
	logs.FDebug("<=== 返回状态码：%d", state_code)
	if result == nil {
		return state_code, req.GetResponse(), data, nil
	}

	//如果没有返回数据，刚直接返回
	if len(data) == 0 {
		return state_code, req.GetResponse(), data, fmt.Errorf("服务请求错误，没有返回数据 (%d)", state_code)
	}

	err_json := json.Unmarshal(data, result)
	//fmt.Println(err_json, utils.GetJson(result))
	if state_code >= 300 {
		//logs.Debug("服务请求错误，返回码：%d\n%s", state_code, string(data))
		return state_code, req.GetResponse(), data, fmt.Errorf("服务请求错误，返回码：%d", state_code)
	} else {
		if err_json != nil {
			return state_code, req.GetResponse(), data, fmt.Errorf("返回数据解析错误(%d)，%s", state_code, utils.GetNetError(err_json))
		} else {
			return state_code, req.GetResponse(), data, nil
		}
	}
}

func SimpleHttpPost(url_str string, body interface{}, conn_timeout, rw_timeout time.Duration) (int, []byte, error) {
	//fmt.Printf("SimpleHttpPost(%s)", url_str)
	//fmt.Println(body)
	c := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, conn_timeout) //设置建立连接超时
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(rw_timeout)) //设置发送接收数据超时
				return conn, nil
			},
		},
	}

	req, err := http.NewRequest(http.MethodPost, url_str, nil)
	if err != nil {
		return 598, nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	if body != nil {
		var body_str string
		switch body.(type) {
		case string:
			body_str = body.(string)
		case []byte:
			body_str = string(body.([]byte))
		default:
			body_str = utils.GetJson(body)
		}
		bf := bytes.NewBufferString(body_str)
		req.Body = ioutil.NopCloser(bf)
		req.ContentLength = int64(len(body_str))
	}
	resp, err := c.Do(req)
	if err != nil {
		return 599, nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, respBody, nil
}

// DownloadURL 下载URL内容并按行分割
func DownloadURL(url_str string, conn_timeout, rw_timeout time.Duration) (int, []string, error) {
	c := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, conn_timeout) //设置建立连接超时
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(rw_timeout)) //设置发送接收数据超时
				return conn, nil
			},
		},
	}

	req, err := http.NewRequest(http.MethodGet, url_str, nil)
	if err != nil {
		return 598, nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return 599, nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 597, nil, err
	}

	lines := strings.Split(string(respBody), "\n")
	return resp.StatusCode, lines, nil
}

func DownloadFile(url_str string, out_file string, conn_timeout, rw_timeout time.Duration) error {
	c := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, conn_timeout) //设置建立连接超时
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(rw_timeout)) //设置发送接收数据超时
				return conn, nil
			},
		},
	}
	req, err := http.NewRequest(http.MethodGet, url_str, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}

	// 创建一个新的文件
	out, err := os.Create(out_file)
	if err != nil {
		return err
	}
	defer func() {
		out.Close()

		mod_time := resp.Header.Get("Last-Modified")
		if mod_time != "" {
			// 使用time包的Parse函数解析日期字符串
			t, err := time.Parse(time.RFC1123, mod_time)
			if err != nil {
				logs.Warning("解析Last-Modified报错(%s)，%s", mod_time, err.Error())
				return
			}
			// 修改文件的访问时间和修改时间
			err = os.Chtimes(out_file, time.Now(), t.Local())
			if err != nil {
				logs.Warning("修改下载文件时间失败，%s", err.Error())
				return
			}
		}
	}()

	// 将响应的内容写入文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// DownloadTextFile 下载文本文件并按行分割
func DownloadTextFile(url_str string, conn_timeout, rw_timeout time.Duration) (int, []string, error) {
	c := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, conn_timeout) //设置建立连接超时
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(rw_timeout)) //设置发送接收数据超时
				return conn, nil
			},
		},
	}

	req, err := http.NewRequest(http.MethodGet, url_str, nil)
	if err != nil {
		return 598, nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return 599, nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 597, nil, err
	}

	lines := strings.Split(string(respBody), "\n")
	return resp.StatusCode, lines, nil
}

// SimpleHttpGet 简单的HTTP GET请求
func SimpleHttpGet(url_str string, conn_timeout, rw_timeout time.Duration) (int, []byte, error) {
	c := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, conn_timeout) //设置建立连接超时
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(rw_timeout)) //设置发送接收数据超时
				return conn, nil
			},
		},
	}

	req, err := http.NewRequest(http.MethodGet, url_str, nil)
	if err != nil {
		return 598, nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return 599, nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, respBody, nil
}

func AutoLogin(url, app_key, user_name, pass_word string) (string, *http.Response, error) {
	params := make(map[string]string)
	params["app_key"] = app_key
	params["username"] = user_name
	params["password"] = pass_word

	result := struct {
		ErrNo  int    `json:"errno"`
		Code   string `json:"code,omitempty"`
		Reason string `json:"errmsg,omitempty"`
	}{}

	state_code, resq, _, err := HttpRequestPB("GET", url, true, params, &result) //登录
	if err != nil {
		return "", resq, err
	}

	if result.ErrNo != 0 {
		return "", resq, fmt.Errorf("%s (%d)", result.Reason, state_code)
	}

	return result.Code, resq, nil
}

func webManagerAPI(w http.ResponseWriter, r *http.Request) {
	msg := ""
	wmsg := "{errno:0}"
	query := r.URL.Query()
	level_str := query.Get("log_level")
	if level_str != "" {
		var log_name string
		var log_level int
		var err error
		pps := strings.Split(level_str, "=")
		if len(pps) == 2 {
			log_name = pps[0]
			log_level, err = strconv.Atoi(pps[1])
		} else {
			log_level, err = strconv.Atoi(level_str)
		}
		if err == nil && log_level <= logs.LevelDebug && log_level >= logs.LevelEmergency {
			msg = fmt.Sprintf("[%s]设置日志级别 - %s%s", GetHttpRemoteAddr(r), log_name, logs.GetLevelName(log_level))
			logs.SetLevel(log_level, log_name)
			wmsg = fmt.Sprintf(`{errno:0,errmsg:"%s"}`, msg)
		} else {
			if err != nil {
				msg = fmt.Sprintf("设置日志级别失败，%s", err.Error())
			} else {
				msg = fmt.Sprintf("设置日志级别失败，输入错误级别(%d)", log_level)
			}
			wmsg = fmt.Sprintf(`{errno:53010,errmsg:"%s"}`, msg)
		}
		fmt.Println(msg)
		fmt.Fprintln(w, wmsg)
	} else {
		if WebManagerAPI != nil {
			WebManagerAPI(w, r)
		} else {
			fmt.Fprintln(w, wmsg)
		}
	}
}
func webAutoTest(w http.ResponseWriter, r *http.Request) {
	result := make(map[string]interface{})
	md, _ := utils.GetFileModTime(os.Args[0])
	result["app_name"] = "统一接口中心(WebManager)"
	result["local_ip"] = utils.GetIPAdress()
	result["build_time"] = md.Format(utils.DateTimeFormat)
	result["create_time"] = time.Now().Format(utils.DateTimeFormat)
	result["log_level"] = logs.GetLevel()
	result["log_level_name"] = logs.GetLevelName(logs.GetLevel())
	result["log_time"] = logs.GetLastLogTime().Format(utils.DateTimeFormat)
	if WebAutoTestAPI != nil {
		result = WebAutoTestAPI(w, r, result)
	}
	fmt.Fprintln(w, utils.GetJson(result))
}

func webManagerHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to ManagerHome!")
}

var WebManagerAPI func(w http.ResponseWriter, r *http.Request)
var WebAutoTestAPI func(w http.ResponseWriter, r *http.Request, re map[string]interface{}) map[string]interface{}

// OpenWebManager 开启Web管理服务
func OpenWebManager(ports ...int) *http.ServeMux {
	ports = append(ports, 54321)
	port := ports[0]

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(webManagerHome))
	mux.Handle("/manager", http.HandlerFunc(webManagerAPI))
	mux.Handle("/autotest", http.HandlerFunc(webAutoTest))
	logs.Info("进程管理服务 ...... 0.0.0.0:%d", port)
	go func(port int, mux *http.ServeMux) {
		logs.Warning(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), mux))
	}(port, mux)
	return mux
}

// 用户凭据
var validUsers = map[string]string{}

func SetUserAndPwd(AUser, APass string) {
	AUser = strings.ToLower(AUser)
	APass = strings.ToLower(APass)
	logs.Debug("增加用户 (%s:%s)", AUser, utils.GetShowPassword(APass))
	for k, _ := range validUsers {
		if k == AUser {
			validUsers[k] = APass
			return
		}
	}
	validUsers[AUser] = APass
}

// SetBasicAuth 设置Basic认证信息
func SetBasicAuth(Auth string) {
	Auth = strings.ToLower(Auth)
	auths := strings.Split(Auth, ":")
	if len(auths) == 2 {
		auths[0] = strings.TrimSpace(auths[0])
		auths[1] = strings.TrimSpace(auths[1])
		logs.Debug("增加用户 (%s:%s)", auths[0], utils.GetShowPassword(auths[1]))
		validUsers[auths[0]] = auths[1]
	}
}

// 中间件函数，用于验证Basic Auth
func BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(validUsers) == 0 {
			next(w, r)
			return
		}

		var msg string
		// 从请求头获取Basic Auth凭据
		username, password, ok := r.BasicAuth()
		//fmt.Println("username, password, ok := r.BasicAuth()", username, password, ok)
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			msg = "获取鉴权信息失败"
			//logs.Error(msg)
			http.Error(w, `{"errno":401,"errmsg":"`+msg+`"}`, http.StatusUnauthorized)
			return
		}

		logs.Debug("Http.BasicAuth() : 共有 %d 个可鉴权用户，传入参数：%s/%s", len(validUsers), username, password)

		// 验证用户名和密码
		username = strings.ToLower(username)
		password = strings.ToLower(password)
		validPassword, userExists := validUsers[username]
		if !userExists || password != validPassword {
			msg = fmt.Sprintf("用户或密码错误(%s:%s)", username, password)
			logs.Error(msg)
			http.Error(w, `{"errno":402,"errmsg":"`+msg+`"}`, http.StatusUnauthorized)
			return
		}

		msg = "验证通过"
		logs.Debug("Http.BasicAuth() : %s (%s/%s)", msg, username, password)
		// 验证通过，调用下一个处理器
		next(w, r)
	}
}

// BasicAuth2 Basic认证中间件（Handler版本）
func BasicAuth2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(validUsers) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// 从请求头获取Basic Auth凭据
		stemp := w.Header().Get("Authorization")
		username, password, ok := r.BasicAuth()

		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			msg := fmt.Sprintf("鉴权失败(%s)", stemp)
			http.Error(w, `{"errno":401,"errmsg":"`+msg+`"}`, http.StatusUnauthorized)
			return
		}

		// 验证用户名和密码
		validPassword, userExists := validUsers[username]
		if !userExists || password != validPassword {
			msg := fmt.Sprintf("用户或密码错误(%s:%s)", username, password)
			http.Error(w, `{"errno":402,"errmsg":"`+msg+`"}`, http.StatusUnauthorized)
			return
		}

		// 验证通过，调用下一个处理器
		next.ServeHTTP(w, r)
	})
}
