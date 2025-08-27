package network

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"

	"github.com/miekg/dns"
	"github.com/tea4go/gh/utils"
)

// 通过域名得到IP地址，返回：真实的域名，IP列表，域名服务器，错误
func GetIPByDomain(query_name string, args ...interface{}) (string, []string, string, error) {
	ip_port := ""
	if len(args) >= 1 {
		ip_port = args[0].(string)
	}
	if ip_port != "" {
		//通过参数传入DNS域名服务器与端口
		if !strings.Contains(ip_port, ":") {
			ip_port = fmt.Sprintf("%s:53", ip_port)
		}
	} else {
		//使用系统自带的域名配置文件获得域名服务器与端口
		resolv_name := "/etc/resolv.conf"
		if utils.FileIsExist(resolv_name) {
			config, err := dns.ClientConfigFromFile(resolv_name)
			if err != nil {
				return query_name, nil, ip_port, fmt.Errorf("解析系统DNS配置失败，原因：%s\n", err.Error())
			}
			if len(config.Servers) == 0 {
				return query_name, nil, ip_port, fmt.Errorf("解析系统DNS配置失败，文件%s没有配置DNS服务器，请手工指定。\n", resolv_name)
			}
			ip_port = net.JoinHostPort(config.Servers[0], config.Port)
		} else {
			ip_port = "114.114.114.114:53"
		}
	}

	c := dns.Client{}

	//新建解析域名的消息体
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(query_name), dns.TypeCNAME)
	m.SetQuestion(dns.Fqdn(query_name), dns.TypeA)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, ip_port)
	if r == nil {
		return query_name, nil, ip_port, fmt.Errorf("解析域名(%s)失败。 \n原因：%s\n", query_name, err.Error())
	}

	if r.Rcode != dns.RcodeSuccess {
		return query_name, nil, ip_port, fmt.Errorf("解析域名(%s)失败，错误码：%d\n", query_name, r.Rcode)
	}

	addres := make([]string, 0)
	err = nil
	for _, answer := range r.Answer {
		rv := reflect.ValueOf(answer)
		cname, ok := rv.Interface().(*dns.CNAME)
		if ok {
			query_name = cname.Target
			continue
		}

		a, ok := rv.Interface().(*dns.A)
		if ok {
			addres = append(addres, a.A.String())
			continue
		} else {
			err = errors.New(answer.String())
		}
	}
	return query_name, addres, ip_port, err
}
