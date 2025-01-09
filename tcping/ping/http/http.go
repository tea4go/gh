package http

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	pkgurl "net/url"
	"strconv"
	"time"

	"github.com/tea4go/gh/tcping/ping"
)

var _ ping.IPing = (*TPing)(nil)

func NewHttp(url string, op *ping.TOption) (*TPing, error) {
	method := op.HttpMethod
	if method == "" {
		method = http.MethodGet
	}
	_, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("网址或方法无效, %w", err)
	}
	return &TPing{
		url:    url,
		method: method,
		option: op,
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// disable redirect
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				Proxy: func(r *http.Request) (*pkgurl.URL, error) {
					if op.HttpProxy != nil {
						return op.HttpProxy, nil
					}
					return http.ProxyFromEnvironment(r)
				},
				DialContext: (&net.Dialer{
					Resolver: op.Resolver,
				}).DialContext,
				DisableKeepAlives: true,
				ForceAttemptHTTP2: false,
			},
		},
	}, nil
}

type TPing struct {
	client *http.Client
	url    string
	method string
	option *ping.TOption
}

func (p *TPing) SetTarget(t *ping.TTarget) {
	if p.option.HttpProxy != nil {
		t.Proxy = p.option.HttpProxy.String()
	}
	t.Timeout = p.option.Timeout
}

func (p *TPing) Ping(ctx context.Context) *ping.TStats {
	timeout := ping.DefaultTimeout
	if p.option.Timeout > 0 {
		timeout = p.option.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	stats := ping.TStats{
		Meta: map[string]fmt.Stringer{},
	}
	trace := Trace{}
	if p.option.IsMeta {
		stats.Extra = &trace
	}

	//#region 开始 Ping 操作
	start := time.Now()
	req, err := http.NewRequestWithContext(trace.WithTrace(ctx), p.method, p.url, nil)
	if err != nil {
		stats.Error = err
		return &stats
	}
	req.Header.Set("user-agent", p.option.UserAgent)
	resp, err := p.client.Do(req)
	//#endregion

	//#region 开始 Ping 统计
	stats.DNSDuration = trace.DNSDuration
	stats.Address = trace.address
	if err != nil {
		stats.Error = err
		stats.Duration = time.Since(start)
	} else {
		stats.Meta["status"] = Int(resp.StatusCode)
		stats.Connected = true
		bodyStart := time.Now()
		defer resp.Body.Close()
		n, err := io.Copy(io.Discard, resp.Body)
		trace.BodyDuration = time.Since(bodyStart)
		if n > 0 {
			stats.Meta["bytes"] = Int(n)
		}
		stats.Duration = time.Since(start)
		if err != nil {
			stats.Connected = false
			stats.Error = fmt.Errorf("读取Http返回包失败， %w", err)
		}
	}
	//#endregion
	return &stats
}

type Int int

func (i Int) String() string {
	return strconv.Itoa(int(i))
}

func init() {
	ping.Register(ping.HTTP,
		func(url *url.URL, op *ping.TOption) (ping.IPing, error) {
			return NewHttp(url.String(), op)
		})
	ping.Register(ping.HTTPS,
		func(url *url.URL, op *ping.TOption) (ping.IPing, error) {
			return NewHttp(url.String(), op)
		})
}
