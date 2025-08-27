package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"time"

	"github.com/mdlayher/arp"
	"github.com/tea4go/gh/execplus"
)

type TIPData struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
	ISP     string `json:"isp"`
}

// 通过IP地址得到Mac地址，如果通过网关访问，得到的将是网关的mac。
func GetMacByIP_file(ip_text string) (string, error) {
	cmd := execplus.CommandString("arp |grep " + ip_text)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	cmd_out_text := out.String()
	if err != nil {
		return "", fmt.Errorf("没有找到IP对应的Mac地址。")
	}
	if err != nil {
		return "", fmt.Errorf("执行Arp命令出错，原因：%s", cmd_out_text)
	}

	reg_text := `([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}`
	reg, err := regexp.Compile(reg_text)
	if err != nil {
		return "", fmt.Errorf("正则表达式错误，表达式：%s，原因：%s", reg_text, err.Error())
	}
	mac := reg.FindString(cmd_out_text)
	if mac != "" {
		return mac, nil
	}

	reg_text = `([0-9A-Fa-f]{2}-){5}[0-9A-Fa-f]{2}`
	reg, err = regexp.Compile(reg_text)
	if err != nil {
		return "", fmt.Errorf("正则表达式错误，表达式：%s，原因：%s", reg_text, err.Error())
	}
	mac = reg.FindString(cmd_out_text)
	if mac != "" {
		return mac, nil
	}
	return "", fmt.Errorf("没有找到IP对应的Mac地址。")
}

func GetMacByIP_arp(ip_text string) (string, error) {
	var mac net.HardwareAddr

	ip, err := netip.ParseAddr(ip_text)
	if err != nil {
		return "", fmt.Errorf("传入的IP地址不正确。%s", err.Error())
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("得到网关列表出错，原因：%s", err.Error())
	}

	for _, inter := range interfaces {
		if inter.Flags&net.FlagUp <= 0 || inter.Flags&net.FlagBroadcast <= 0 {
			continue
		}
		//fmt.Println("====>", inter.Index, inter.Name, inter.HardwareAddr)

		// Set up ARP client with socket
		c, err := arp.Dial(&inter)
		if err != nil {
			return "", fmt.Errorf("创建Arp网络句柄出错，原因：%s", err.Error())
		}
		defer func() {
			// Clean up ARP client socket
			_ = c.Close()
		}()

		// Set request deadline from flag
		err = c.SetDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			return "", fmt.Errorf("设置网络超时出错，原因：%s", err.Error())
		}

		// Request hardware address for IP address
		//ip := net.ParseIP(*ipFlag).To4()
		mac, err = c.Resolve(ip)
		if err != nil {
			//return "", fmt.Errorf("获得Mac出错，原因：%s", err.Error())
			continue
		}

		if len(mac) > 0 {
			out_text := fmt.Sprintf("%s", mac)
			return out_text, nil
		}
	}
	return "", fmt.Errorf("没有找到IP对应的Mac地址。")
}

func GetMacByIP(ip_text string) (string, error) {
	mac, err := GetMacByIP_file(ip_text)
	if err != nil {
		mac, err = GetMacByIP_arp(ip_text)
		if err != nil {
			return "", err
		}
	}
	return mac, nil
}

// "http://myexternalip.com/raw"
func GetPublicIP() string {
	ok_chan := make(chan string)
	go func() {
		ip, err := GetIPByURL("http://myexternalip.com/raw")
		//fmt.Println("1 - ", ip)
		if err == nil {
			ok_chan <- ip
		}
	}()
	go func() {
		ip, err := GetIPByURL("http://myip.ipip.net")
		//fmt.Println("2 - ", ip)
		if err == nil {
			ok_chan <- ip
		}
	}()
	go func() {
		ip, err := GetIPByURL("http://httpbin.org/ip")
		//fmt.Println("3 - ", ip)
		if err == nil {
			ok_chan <- ip
		}
	}()
	go func() {
		ip, err := GetIPByURL("http://ifcfg.cn/echo")
		//fmt.Println("4 - ", ip)
		if err == nil {
			ok_chan <- ip
		}
	}()
	go func() {
		ip, err := GetIPByURL("https://ifconfig.me")
		//fmt.Println("5 - ", ip)
		if err == nil {
			ok_chan <- ip
		}
	}()
	t := time.NewTicker(5 * time.Second)
	select {
	case cc := <-ok_chan:
		return cc
	case <-t.C:
		return ""
	}
}

// "http://myexternalip.com/raw"
// 目前网址有问题，可以找其他网址
func GetPublicIPDetail() (*TIPData, error) {
	_, respBody, err := SimpleHttpGet("http://ip.taobao.com/service/getIpInfo2.php?ip=myip", 10*time.Second, 10*time.Second)
	if err != nil {
		return nil, err
	}

	type TResult struct {
		Code int     `json:"code"`
		Data TIPData `json:"data"`
	}

	var result TResult
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("返回数据异常！")
	} else {
		return &result.Data, nil
	}
}

// 取公网IP，通过解析返回的网页，取出IP地址
func GetIPByURL(url_str string) (string, error) {
	_, resp, err := SimpleHttpGet(url_str, 15*time.Second, 10*time.Second)
	if err != nil {
		return "", err
	}
	re, err := regexp.Compile(`\d{1,3}\.\d{1,3}\.\d{1,3}.\d{1,3}`)
	if err != nil {
		return "", err
	}
	return re.FindString(string(resp)), nil
}
