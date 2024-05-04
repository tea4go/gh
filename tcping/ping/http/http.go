package http

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	pkgurl "net/url"
	"strconv"
	"time"

	"github.com/tea4go/gh/tcping/ping"
)

var _ ping.IPing = (*TPing)(nil)

func New(method string, url string, op *ping.TOption, trace bool) (*TPing, error) {

	_, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("网址或方法无效, %w", err)
	}

	if method == "" {
		method = http.MethodGet
	}

	return &TPing{
		url:    url,
		method: method,
		trace:  trace,
		option: op,
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// disable redirect
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				Proxy: func(r *http.Request) (*pkgurl.URL, error) {
					if op.Proxy != nil {
						return op.Proxy, nil
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
	trace  bool

	option *ping.TOption
	method string

	url string
}

func (p *TPing) SetTarget(t *ping.TTarget) {
	t.Protocol = ping.TCP
	t.Proxy = p.option.Proxy.String()
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
	if p.trace {
		stats.Extra = &trace
	}
	start := time.Now()
	req, err := http.NewRequestWithContext(trace.WithTrace(ctx), p.method, p.url, nil)
	if err != nil {
		stats.Error = err
		return &stats
	}
	req.Header.Set("user-agent", p.option.UA)
	resp, err := p.client.Do(req)
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
	return &stats
}

type Int int

func (i Int) String() string {
	return strconv.Itoa(int(i))
}
