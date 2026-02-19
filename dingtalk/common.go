package dingtalk

import (
	"bytes"
	"encoding/json"
)

//钉钉发消息（群）

type TResult struct {
	ErrCode int    `json:"errcode,omitempty"`
	ErrMsg  string `json:"errmsg,omitempty"`
}

type TDingTalkResponse struct {
	ErrCode   int             `json:"errcode"`
	ErrMsg    string          `json:"errmsg"`
	Result    json.RawMessage `json:"result"`
	RequestID string          `json:"request_id"`
}

type MessageTextSub struct {
	Text string `json:"content"`
}

type MessageTextAt struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
}

type MessageText struct {
	At      MessageTextAt  `json:"at"`
	Type    string         `json:"msgtype"`
	Message MessageTextSub `json:"text"`
}

// String 转换为字符串
func (Self *MessageText) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

type MessageMarkdownSub struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type MessageMarkdown struct {
	Type    string             `json:"msgtype"`
	Message MessageMarkdownSub `json:"markdown"`
}

// String 转换为字符串
func (Self *MessageMarkdown) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

// 用户信息
type TReSnsUser struct {
	TResult
	UserInfo struct {
		NickName string `json:"nick"`
		UnionId  string `json:"unionid"`
		DingId   string `json:"dingId"`
		OpenId   string `json:"openid"`
	} `json:"user_info"`
}
