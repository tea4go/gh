package weixin

import (
	"fmt"
	"time"

	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/gh/network"
	"github.com/tea4go/gh/utils"
)

type WorkWeixin struct {
	CorpID     string
	CorpSecret string
	AgentID    int
	Token      string
	BeginTime  time.Time
	EndTime    time.Time
}

type ResultWeixin struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"` //秒 默认返回是2小时
	InvalidUser string `json:"invaliduser"`
	ErrorCode   int    `json:"errcode"`
	ErrorMsg    string `json:"errmsg"`
}

// adb shell am start -n org.autojs.autojs/.external.open.RunIntentActivity -d /sdcard/AData/MyJS/auto_run.js -t application/x-javascript
func (w *WorkWeixin) Init(corpid string, corpsecret string, agentId int) {
	w.CorpID = corpid
	w.CorpSecret = corpsecret
	w.AgentID = agentId
	w.GetAccessToken()
}

func (w *WorkWeixin) SendMessage(text string) bool {
	return w.SendUserMessage("Tony", text)
}

func (w *WorkWeixin) SendUserMessage(touser, text string) bool {
	message_text := map[string]interface{}{
		"touser":  touser,
		"msgtype": "text",
		"agentid": w.AgentID,
		"text": map[string]interface{}{
			"content": text,
		},
	}

	//申请Token
	if w.Token == "" && w.GetAccessToken() == "" {
		return false
	}
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", w.Token)
	req := network.HttpPost(url).SetTimeout(15*time.Second, 10*time.Second)
	_, err := req.JSONBody(message_text)
	if err != nil {
		logs.Error("序列化发送消息失败，原因：%s", err.Error())
		return false
	}

	body_text, err := req.String()
	if err != nil {
		logs.Error("发送消息失败，原因：%s", err.Error())
		return false
	}

	var result ResultWeixin
	err = req.ToJSON(&result)
	if err != nil {
		logs.Error("序列化接收消息失败，内容：%s，原因：%s", body_text, err.Error())
		return false
	}

	if result.ErrorCode == 0 {
		return true
	} else if result.ErrorCode == 42001 {
		w.Token = ""
		return w.SendUserMessage(touser, text)
	} else {
		logs.Error("微信发送失败，%s", result.ErrorMsg)
		return false
	}

}

func (w *WorkWeixin) GetAccessToken() string {
	if w.Token != "" {
		return w.Token
	}
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", w.CorpID, w.CorpSecret)
	req := network.HttpGet(url).SetTimeout(15*time.Second, 10*time.Second)

	body_text, err := req.String()
	if err != nil {
		logs.Error("发送认证请求失败，原因：%s", err.Error())
		return ""
	}

	var token ResultWeixin
	err = req.ToJSON(&token)
	if err != nil {
		logs.Error("序列化接收消息失败，内容：%s，原因：%s", body_text, err.Error())
		return ""
	}

	if token.ErrorCode == 0 {
		w.Token = token.AccessToken
		w.BeginTime = utils.GetNow()
		h, _ := time.ParseDuration(fmt.Sprintf("+%ds", token.ExpiresIn))
		w.EndTime = w.BeginTime.Add(h)
		return w.Token
	} else {
		logs.Error("发送微信认证失败，%s", token.ErrorMsg)
		return ""
	}
}
