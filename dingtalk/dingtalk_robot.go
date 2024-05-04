package dingtalk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/gh/network"
	"github.com/tea4go/gh/utils"
)

/*
钉钉几个关键值：
1、钉钉小程序有三个字段，例如：安全令牌
   AgentId：615063230
   AppKey：dingwfevtfsjwgda14sf
   AppSecret：C93qxESMeqK1XgHc8uBpXWoxy53Ix6EBDZUIWrZ1lOco685Ln-ra4NETc8oGbE_F
2、钉钉群自定义机器人
   Webhook：https://oapi.dingtalk.com/robot/send?access_token=a32607882f95e2f1acc71564711332f4b843a0acac865e324d6914ba75548984
   加签：SEC9ca0fd5714d2b58d6d8474251b06cb20710f90e64ff8d735b13a23419ec3c858
   例如：钉钉发消息（群）
	 curl 'https://oapi.dingtalk.com/robot/send?access_token=c12a607195f4b06add917091bd4b7fe4eb71227f5cf975f517a952f7e092796b&timestamp=1593791380034&sign=C2w5aElJlpcXN9sjE1TDnDKnfjufANybcNo5WrSoTO0%3D' \
	    -H 'Content-Type: application/json' \
	    -d '
	   {"msgtype": "text",
	     "text": {
	         "content": "我就是我, 是不一样的烟火"
	      }
	   }'
	官网地址：https://developers.dingtalk.com/document/app/custom-robot-access?spm=ding_open_doc.document.0.0.6d9d28e15V2W03#topic-2026027
3、三方系统扫码登录
   appId:dingoa0ewoncpltwsurfqv
   appSecret:dzhAcgJTx7gvQxDQhlGqScZa6GrwpTHYiHLVJxma4pSPKeqnxFGyB9vDwm825LaQ

   获取access_token(appId,appSecret)
   https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s
   服务端通过临时授权码获取授权用户的个人信息
   https://oapi.dingtalk.com/sns/getuserinfo_bycode?accessKey=%s&timestamp=%s&signature=%s
   根据unionid获取userid
   https://oapi.dingtalk.com/user/getUseridByUnionid?access_token=%s&unionid=%s
*/

// 已经不在使用
type TDingTalkSDK struct {
	ddurl             string
	timeout_connect   time.Duration
	timeout_readwrite time.Duration
}

func (w *TDingTalkSDK) genSignedURL(format, secret string) string {
	timeStr := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	sign := fmt.Sprintf("%s\n%s", timeStr, secret)
	signData := w.computeHmacSha256(sign, secret)
	encodeURL := url.QueryEscape(signData)
	return fmt.Sprintf(format, timeStr, encodeURL)
}

func (w *TDingTalkSDK) computeHmacSha256(sign string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(sign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (w *TDingTalkSns) Init(appId string, appSecret string) {
	w.appkey = appId
	w.appsecret = appSecret
}

type TDingTalkSns struct {
	TDingTalkSDK
	appkey    string
	appsecret string
	token     *TAccessToken
}

func GetDingTalkSns(appkey, appsecret string) *TDingTalkSns {
	sns := &TDingTalkSns{
		appkey:    appkey,
		appsecret: appsecret,
	}
	sns.ddurl = `https://oapi.dingtalk.com`
	sns.timeout_connect = 30 * time.Second
	sns.timeout_readwrite = 30 * time.Second
	return sns
}

func (Self *TDingTalkSns) GetAppKey() string {
	return Self.appkey
}

// https://oapi.dingtalk.com/gettoken?appkey=key&appsecret=secret
// {"errorCode":503,"success":false,"errorMsg":"不合法的access_token"}
func (Self *TDingTalkSns) GetAccessToken() (string, error) {
	if Self.token != nil && Self.token.IsValid() {
		return Self.token.AccessToken, nil
	}
	//logs.Debug("GetAccessTokenBySns() : 获取钉钉Token信息")

	req := network.HttpGet(Self.ddurl+"/sns/gettoken").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("appid", Self.appkey)
	req.Param("appsecret", Self.appsecret)

	var info TAccessToken
	err := req.ToJSON(&info)
	if err != nil {
		fmt.Println(err.Error())
		return "", errors.New("获取钉钉临时令牌失败！")
	}
	if info.ErrCode == 0 {
		info.ExpiresIn = 7200
		info.CreateDate = time.Now()
		Self.token = &info
		//logs.Debug("GetAccessTokenBySns() : 获取钉钉Token信息 ...... OK (%s)", info.AccessToken)
		return info.AccessToken, nil
	} else {
		logs.Debug("GetAccessTokenBySns() : 获取钉钉Token信息 ...... Not OK %s(%d)", info.ErrMsg, info.ErrCode)
		return "", errors.New(fmt.Sprintf("%s(%d)", info.ErrMsg, info.ErrCode))
	}
}

// 根据sns临时授权码获取用户信息
// https://developers.dingtalk.com/document/app/obtain-the-user-information-based-on-the-sns-temporary-authorization
func (Self *TDingTalkSns) GetUserByUnionId(code string) (bool, string, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return false, "", err
	}
	//Self.token.AccessToken = "7f5cad8ab10a3df682b7182d51a5c034"
	//获取永久授权码
	dd_url := fmt.Sprintf(Self.ddurl+"/sns/get_persistent_code?access_token=%s", Self.token.AccessToken)
	req := network.HttpPost(dd_url).SetTimeout(Self.timeout_connect, Self.timeout_readwrite)

	msg := struct {
		Tmp_auth_code string `json:"tmp_auth_code"`
	}{code}
	_, err = req.JSONBody(msg)
	if err != nil {
		return false, "", errors.New(fmt.Sprintf("序列化包体失败，原因：%s", err.Error()))
	}

	body_text, err := req.String()
	if err != nil {
		return false, "", errors.New(fmt.Sprintf("发送请求失败，原因：%s", err.Error()))
	} //{"errcode":0,"errmsg":"ok","unionid":"dxUDiP03drHsiE","openid":"dxNuFxiS4hmkiE","persistent_code":"InVVQOiRWZxwX0Q_2JPIVkWUS7VtCkS7C3jkE-fvEKQ8V9979-utknjrplZWIClU"}

	info := struct {
		TResult
		UnionId        string `json:"unionid"`
		OpenId         string `json:"openid"`
		PersistentCode string `json:"persistent_code"`
	}{}
	err = req.ToJSON(&info)
	if err != nil {
		return false, "", errors.New(fmt.Sprintf("序列化接收消息失败，内容：%s，原因：%s", body_text, err.Error()))
	}

	if info.ErrCode == 42001 {
		Self.token = nil
		return Self.GetUserByUnionId(code)
	} else if info.ErrCode != 0 {
		return false, "", errors.New(info.ErrMsg)
	}

	//获取用户授权的SNS_TOKEN
	dd_url = fmt.Sprintf(Self.ddurl+"/sns/get_sns_token?access_token=%s", Self.token.AccessToken)
	req = network.HttpPost(dd_url).SetTimeout(Self.timeout_connect, Self.timeout_readwrite)

	msg1 := struct {
		Openid          string `json:"openid"`
		Persistent_code string `json:"persistent_code"`
	}{info.OpenId, info.PersistentCode}
	_, err = req.JSONBody(msg1)
	if err != nil {
		return false, "", errors.New(fmt.Sprintf("序列化包体失败，原因：%s", err.Error()))
	}

	body_text, err = req.String()
	if err != nil {
		return false, "", errors.New(fmt.Sprintf("发送请求失败，原因：%s", err.Error()))
	} //{"errcode":0,"errmsg":"ok","sns_token":"d89bb16dffe431778d806b1c4c30e3b9","expires_in":7200}

	info1 := struct {
		TResult
		SnsToken  string `json:"sns_token"`
		ExpiresIn int    `json:"expires_in"`
	}{}
	err = req.ToJSON(&info1)
	if err != nil {
		return false, "", errors.New(fmt.Sprintf("序列化接收消息失败，内容：%s，原因：%s", body_text, err.Error()))
	}

	if info1.ErrCode != 0 {
		return false, "", errors.New(info1.ErrMsg)
	}

	//获取用户unionid
	dd_url = fmt.Sprintf(Self.ddurl+"/sns/getuserinfo?sns_token=%s", info1.SnsToken)
	req = network.HttpGet(dd_url).SetTimeout(Self.timeout_connect, Self.timeout_readwrite)

	body_text, err = req.String()
	if err != nil {
		return false, "", errors.New(fmt.Sprintf("发送请求失败，原因：%s", err.Error()))
	} //{"errcode":0,"errmsg":"ok","user_info":{"nick":"刘启","unionid":"dxUDiP03drHsiE","dingId":"$:LWCP_v1:$WpQSgBHxN/6rm4v40aWRAA==","openid":"dxNuFxiS4hmkiE"}}

	info2 := TReSnsUser{}
	err = req.ToJSON(&info2)
	if err != nil {
		return false, "", errors.New(fmt.Sprintf("序列化接收消息失败，内容：%s，原因：%s", body_text, err.Error()))
	}

	if info2.ErrCode != 0 {
		return false, "", errors.New(info2.ErrMsg)
	}

	return true, info2.UserInfo.UnionId, nil
}

type TDingTalkOAuth2 struct {
	TDingTalkSDK
	appkey    string
	appsecret string
}

func GetDingTalkOAuth2(appkey, appsecret string) *TDingTalkOAuth2 {
	sns := &TDingTalkOAuth2{
		appkey:    appkey,
		appsecret: appsecret,
	}
	sns.ddurl = `https://api.dingtalk.com`
	sns.timeout_connect = 30 * time.Second
	sns.timeout_readwrite = 30 * time.Second
	return sns
}

func (Self *TDingTalkOAuth2) GetAppKey() string {
	return Self.appkey
}

// https://oapi.dingtalk.com/gettoken?appkey=key&appsecret=secret
// {"errorCode":503,"success":false,"errorMsg":"不合法的access_token"}
func (Self *TDingTalkOAuth2) GetAccessToken() (string, error) {
	//logs.Debug("GetAccessTokenBySns() : 获取钉钉Token信息")
	BodyParam := struct {
		AppKey    string `json:"appKey"`
		AppSecret string `json:"appSecret,omitempty"`
	}{}
	BodyParam.AppKey = Self.appkey
	BodyParam.AppSecret = Self.appsecret

	reat := struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int    `json:"expireIn"`
	}{}

	state_code, _, data, err := network.HttpRequestBB("POST", Self.ddurl+"/v1.0/oauth2/accessToken", true, &BodyParam, &reat)
	if err != nil {
		ReError := struct {
			Code    string `json:"code"`
			Message string `json:"message,omitempty"`
		}{}
		err_json := json.Unmarshal(data, &ReError)
		if err_json != nil {
			logs.Debug("返回数据解析错误，返回码：%d\n%s", state_code, string(data))
			return "", fmt.Errorf("获取钉钉临时令牌失败，%s (%d)", utils.GetNetError(err_json), state_code)
		} else {
			return "", fmt.Errorf("获取钉钉临时令牌失败，%s (%d)", ReError.Message, state_code)
		}
	}

	return reat.AccessToken, nil
}

// 根据临时授权码获取用户信息
// https://open.dingtalk.com/document/orgapp/obtain-user-token
func (Self *TDingTalkOAuth2) GetUserByUnionId(code string) (bool, string, error) {
	BodyParam := struct {
		ClientId     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
		Code         string `json:"code"`
		RefreshToken string `json:"refreshToken"`
		GrantType    string `json:"grantType"`
	}{}
	BodyParam.ClientId = Self.appkey
	BodyParam.ClientSecret = Self.appsecret
	BodyParam.Code = code
	BodyParam.GrantType = "authorization_code"

	//{"expireIn":7200,"accessToken":"dcf09b6491c633659d5e55bc69d70227","refreshToken":"fc7c89c8a12a34289fd0eb6dbf12e09a"}
	//{"expireIn":7200,"accessToken":"118f4b38962a305ca6d03fa72f67e55b","refreshToken":"2a4c0c87c8ba35469835772a40b510fb"}
	reat := struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int    `json:"expireIn"`
	}{}

	state_code, _, data, err := network.HttpRequestBB("POST", Self.ddurl+"/v1.0/oauth2/userAccessToken", true, &BodyParam, &reat)
	if err != nil {
		ReError := struct {
			Code    string `json:"code"`
			Message string `json:"message,omitempty"`
		}{}
		err_json := json.Unmarshal(data, &ReError)
		if err_json != nil {
			logs.Debug("返回数据解析错误，返回码：%d\n%s", state_code, string(data))
			return false, "", fmt.Errorf("获取钉钉用户令牌失败，%s (%d)", utils.GetNetError(err_json), state_code)
		} else {
			return false, "", fmt.Errorf("获取钉钉用户令牌失败，%s (%d)", ReError.Message, state_code)
		}
	}

	// {
	// 	"nick":"刘启",
	// 	"unionId":"dxUDiP03drHsiE",
	// 	"avatarUrl":"https://static-legacy.dingtalk.com/media/lADPBbCc1VsAkMLNAoDNAoA_640_640.jpg",
	// 	"openId":"yxdBZAHRsaQiE"
	// }
	reuserid := struct {
		UnionId   string `json:"unionId"`
		AvatarUrl string `json:"avatarUrl"`
		Nick      string `json:"nick"`
		OpenId    string `json:"openId"`
	}{}
	header := make(map[string]string)
	header["x-acs-dingtalk-access-token"] = reat.AccessToken
	state_code, _, data, err = network.HttpRequestHA("GET", Self.ddurl+"/v1.0/contact/users/me", true, header, &reuserid)

	if err != nil {
		ReError := struct {
			Code    string `json:"code"`
			Message string `json:"message,omitempty"`
		}{}
		err_json := json.Unmarshal(data, &ReError)
		if err_json != nil {
			logs.Debug("返回数据解析错误，返回码：%d\n%s", state_code, string(data))
			return false, "", fmt.Errorf("获取钉钉用户标识失败，%s (%d)", utils.GetNetError(err_json), state_code)
		} else {
			return false, "", fmt.Errorf("获取钉钉用户标识失败，%s (%d)", ReError.Message, state_code)
		}
	}

	return true, reuserid.UnionId, nil
}

// 钉钉群的配置
type robotTokenConfig struct {
	AccessToken string
	Secret      string
}

type TDingTalkRobot struct {
	TDingTalkSDK
	accessToken string
	secret      string
	config      []robotTokenConfig
}

func (w *TDingTalkRobot) Init(access_token string, secret string) {
	w.accessToken = access_token
	w.secret = secret
}

func (w *TDingTalkRobot) Inits() {
	w.config = make([]robotTokenConfig, 0)
	c1 := robotTokenConfig{} //工具01
	c1.AccessToken = "bdaf896947630f45256977260a4c75c8b9946bb53118f30f851e0de5d34e7a51"
	c1.Secret = "SEC709222f0aac2a839ad6fa34b9e978f6423565c71c335b1dc097f726ca3c2b4d4"
	c2 := robotTokenConfig{} // 工具02
	c2.AccessToken = "c12a607195f4b06add917091bd4b7fe4eb71227f5cf975f517a952f7e092796b"
	c2.Secret = "SEC12e9f06572b85c007f35b528ca1f09c4c3f66e0f6b3fcfc14894a4c3b16ea951"

	c3 := robotTokenConfig{} // 高等
	c3.AccessToken = "cf507d2334cd6e40751a19636eadc7f4328b82481dc6b3af40203b0e2f0c2894"
	c3.Secret = "SEC13a69ec3c473306963f70ae9878d09471249b405cc2d9acd53c3d948378048c6"

	c4 := robotTokenConfig{} // 机房
	c4.AccessToken = "f43efe4afb909b4525a99b5f348efdfcdcafc3d9f79484f01a9c7898f2656c25"
	c4.Secret = "SEC8cb3795ba61170f7dd367f5e79cafed342f50a8e8149c1dcedfa86564a33a8e2"

	c5 := robotTokenConfig{} //
	c5.AccessToken = "a32607882f95e2f1acc71564711332f4b843a0acac865e324d6914ba75548984"
	c5.Secret = "SEC9ca0fd5714d2b58d6d8474251b06cb20710f90e64ff8d735b13a23419ec3c858"

	w.config = append(w.config, c1)
	w.config = append(w.config, c2)
	w.config = append(w.config, c3)
	w.config = append(w.config, c4)
	w.config = append(w.config, c5)
}

func (w *TDingTalkRobot) SendMDMessage(title, text string) bool {
	var message_text MessageMarkdown
	message_text.Type = "markdown"
	message_text.Message.Title = title
	message_text.Message.Text = text
	return w.SendMessage(message_text)
}

func (w *TDingTalkRobot) SendTextMessage(text string, all bool) bool {
	var message_text MessageText
	message_text.Type = "text"
	mb := make([]string, 0)
	//mb = append(mb, "13951774197")
	//mb = append(mb, "13016985150")
	message_text.At = MessageTextAt{IsAtAll: all, AtMobiles: mb}
	message_text.Message.Text = text
	return w.SendMessage(message_text)
}

func (w *TDingTalkRobot) SendTextMessages(text string, all bool) int {
	re := 0
	for _, r := range w.config {
		w.accessToken = r.AccessToken
		w.secret = r.Secret
		if w.SendTextMessage(text, all) {
			re = re + 1
		}
	}
	return re
}

func (w *TDingTalkRobot) SendMessage(msg interface{}) bool {
	dd_url := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s%s", w.accessToken, w.genSignedURL("&timestamp=%s&sign=%s", w.secret))
	req := network.HttpPost(dd_url).SetTimeout(10*time.Second, 5*time.Second)

	_, err := req.JSONBody(msg)
	if err != nil {
		logs.Error("序列化发送消息失败，原因：%s", err.Error())
		return false
	}

	body_text, err := req.String()
	if err != nil {
		logs.Error("发送消息失败，原因：%s", err.Error())
		return false
	}

	var result TResult
	err = req.ToJSON(&result)
	if err != nil {
		logs.Error("序列化接收消息失败，内容：%s，原因：%s", body_text, err.Error())
		return false
	}

	if result.ErrCode != 0 {
		logs.Error("发送消息失败，" + result.ErrMsg)
	}

	return result.ErrCode == 0
}
