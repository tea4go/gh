package ping

import (
	"bytes"
	"container/list"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math"
	"net"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	logs "github.com/tea4go/gh/log4go"

	"gopkg.in/ffmt.v1"
	//ffmt "gopkg.in/ffmt.v1"
)

var pinger = map[SProtocol]Factory{}

type Factory func(url *url.URL, op *TOption) (IPing, error)

func Register(p SProtocol, factory Factory) {
	logs.Debug("= 注册组件 ...... %s", p)
	pinger[p] = factory
}

func Load(p SProtocol) Factory {
	return pinger[p]
}

// Protocol ...
type SProtocol int

func (p SProtocol) String() string {
	switch p {
	case TCP:
		return "tcp"
	case HTTP:
		return "http"
	case HTTPS:
		return "https"
	}
	return "unknown"
}

const (
	// TCP is tcp protocol
	TCP SProtocol = iota
	// HTTP is http protocol
	HTTP
	// HTTPS is https protocol
	HTTPS
)

// NewProtocol convert protocol string to Protocol
func NewProtocol(protocol string) (SProtocol, error) {
	switch strings.ToLower(protocol) {
	case TCP.String():
		return TCP, nil
	case HTTP.String():
		return HTTP, nil
	case HTTPS.String():
		return HTTPS, nil
	}
	return 0, fmt.Errorf("无效协议(%s)", protocol)
}

type TOption struct {
	Timeout    time.Duration // 连接超时
	Resolver   *net.Resolver // 自定义DNS域名解析
	HttpProxy  *url.URL      // Http代理(格式：http://192.168.100.1:32126）(Http/Https)
	HttpMethod string        // 请求方法 (Http/Https)
	IsMeta     bool          // 是否启用Meta信息采集 (Http/Https)
	UserAgent  string        // 浏览器UA标识 (Http/Https)
	IsTls      bool          // 是否启用TLS (TCP)
}

type TStats struct {
	Connected   bool                    `json:"connected"`
	Duration    time.Duration           `json:"ping_duration"`
	DNSDuration time.Duration           `json:"dns_Duration"`
	Address     string                  `json:"address"`
	Meta        map[string]fmt.Stringer `json:"meta"`
	Extra       fmt.Stringer            `json:"extra"`
	Error       error                   `json:"error"`
}

func (s *TStats) FormatMeta() string {
	keys := make([]string, 0, len(s.Meta))
	for key := range s.Meta {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for i, key := range keys {
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(s.Meta[key].String())
		if i < len(keys)-1 {
			builder.WriteString(" ")
		}
	}
	return builder.String()
}

type IPing interface {
	Ping(ctx context.Context) *TStats
	SetTarget(t *TTarget)
}

func NewPinger(out io.Writer, url *url.URL, ping IPing, interval time.Duration, counter int) *TPinger {
	pger := &TPinger{
		stopC:    make(chan struct{}),
		counter:  counter,
		interval: interval,
		out:      out,
		url:      url,
		ping:     ping,
	}
	pger.result.Target.Protocol = url.Scheme
	pger.result.Target.URL = url.String()
	pger.result.Target.Interval = interval
	pger.result.Target.Counter = counter
	pger.result.Items = make([]*TData, 0)
	ping.SetTarget(&pger.result.Target)
	return pger
}

type TPinger struct {
	ping          IPing         //Ping的实际动作
	stopOnce      sync.Once     //退出
	stopC         chan struct{} //退出
	out           io.Writer     //输出信息
	url           *url.URL      //Ping的目标地址
	interval      time.Duration //间隔
	counter       int           //Ping的次数
	result        TResult       //Ping的返回信息
	minDuration   time.Duration
	maxDuration   time.Duration
	totalDuration time.Duration
	total         int
	failedTotal   int
	items         list.List
}

func (p *TPinger) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopC)
	})
}

func (p *TPinger) Done() <-chan struct{} {
	return p.stopC
}

func (p *TPinger) Ping() {
	defer p.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-p.Done()
		cancel()
	}()

	interval := DefaultInterval
	if p.interval > 0 {
		interval = p.interval
	}
	timer := time.NewTimer(1)
	defer timer.Stop()

	stop := false
	p.minDuration = time.Duration(math.MaxInt64)
	for !stop {
		select {
		case <-timer.C:
			stats := p.ping.Ping(ctx)
			p.SetResult(stats)
			if p.total++; p.counter > 0 && p.total > p.counter-1 {
				stop = true
			}
			timer.Reset(interval)
		case <-p.Done():
			stop = true
		}
	}
}
func (p *TPinger) GetStats() ([]TStats, error) {
	return nil, nil
}
func (p *TPinger) PingServer() {
	defer p.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-p.Done()
		cancel()
	}()

	interval := DefaultInterval
	if p.interval > 0 {
		interval = p.interval
	}
	timer := time.NewTimer(1)
	defer timer.Stop()

	stop := false
	p.minDuration = time.Duration(math.MaxInt64)
	for !stop {
		select {
		case <-timer.C:
			stats := p.ping.Ping(ctx)
			p.items.PushFront(stats)
			ffmt.Puts(stats)
			p.total++
			if p.total > 5 {
				p.items.Remove(p.items.Back())
			}
			fmt.Println(p.items.Len())
			timer.Reset(interval)
		case <-p.Done():
			stop = true
		}
	}
}

func (p *TPinger) Summarize() {
	const tpl = `
Ping statistics for %s
	%d probes sent.
	%d successful, %d failed.
Approximate trip times:
	Minimum = %s, Maximum = %s, Average = %s
`
	if p.total == 0 {
		return
	}
	_, _ = fmt.Fprintf(p.out, tpl, p.url.String(), p.total, p.total-p.failedTotal, p.failedTotal, p.minDuration, p.maxDuration, p.totalDuration/time.Duration(p.total))
}

func (p *TPinger) SetResult(st *TStats) {
	if st.Duration < p.minDuration {
		p.minDuration = st.Duration
	}
	if st.Duration > p.maxDuration {
		p.maxDuration = st.Duration
	}
	p.totalDuration += st.Duration
	if st.Error != nil {
		p.failedTotal++
		if errors.Is(st.Error, context.Canceled) {
			// ignore cancel
			return
		}
	}
	status := "Failed"
	if st.Connected {
		status = "Connected"
	}

	if st.Error != nil {
		_, _ = fmt.Fprintf(p.out, "Ping %s (%s) %s(%s) - time=%-10s dns=%-9s",
			p.url.String(), st.Address, status, FormatError(st.Error), st.Duration.String(), st.DNSDuration)
	} else {
		_, _ = fmt.Fprintf(p.out, "Ping %s (%s) %s - time=%-10s dns=%-9s",
			p.url.String(), st.Address, status, st.Duration.String(), st.DNSDuration)
	}
	if len(st.Meta) > 0 {
		_, _ = fmt.Fprintf(p.out, " %s", st.FormatMeta())
	}
	_, _ = fmt.Fprint(p.out, "\n")
	if st.Extra != nil {
		_, _ = fmt.Fprintf(p.out, "%s\n", strings.TrimSpace(st.Extra.String()))
	}
}

type TData struct {
}
type TTarget struct {
	Protocol string
	URL      string
	IP       string
	Port     int
	Proxy    string
	Counter  int
	Interval time.Duration
	Timeout  time.Duration
}

func (p TTarget) String() string {
	return fmt.Sprintf("%s://%s:%d", p.Protocol, p.IP, p.Port)
}

type TResult struct {
	Counter       int           //总数
	CounterOK     int           //成功数
	Target        TTarget       //Ping的目标信息
	Items         []*TData      //Ping的历史记录
	MinDuration   time.Duration //Ping最短时长
	MaxDuration   time.Duration //Ping最长时长
	TotalDuration time.Duration //Ping总时长
}

// Avg return the average time of ping
func (p TResult) Avg() time.Duration {
	if p.CounterOK == 0 {
		return 0
	}
	return p.TotalDuration / time.Duration(p.CounterOK)
}

// Failed return failed counter
func (p TResult) Failed() int {
	return p.Counter - p.CounterOK
}

func (p TResult) String() string {
	const resultTpl = `
Ping statistics {{.Target}}
	{{.Counter}} probes sent.
	{{.SuccessCounter}} successful, {{.Failed}} failed.
Approximate trip times:
	Minimum = {{.MinDuration}}, Maximum = {{.MaxDuration}}, Average = {{.Avg}}`
	t := template.Must(template.New("p").Parse(resultTpl))
	res := bytes.NewBufferString("")
	_ = t.Execute(res, p)
	return res.String()
}
