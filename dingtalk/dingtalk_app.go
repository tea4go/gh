package dingtalk

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/gh/network"
	"github.com/tea4go/gh/utils"
)

//钉钉小程序 助手类

//curl 'https://oapi.dingtalk.com/robot/send?access_token=ddd' \
//   -H 'Content-Type: application/json' \
//   -d '
//  {"msgtype": "text",
//    "text": {
//        "content": "我就是我, 是不一样的烟火"
//     }
//  }'

type TAccessToken struct {
	TResult
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	CreateDate  time.Time
}

// String 转换为字符串
func (Self *TAccessToken) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

// IsValid 检查 Token 是否有效
func (Self *TAccessToken) IsValid() bool {
	return Self.AccessToken != "" && int(time.Now().Unix()-Self.CreateDate.Unix()) <= Self.ExpiresIn
}

type TJsapiTicket struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	RequestId   string `json:"requestid"`
	JsapiTicket string `json:"jsapiTicket"`
	ExpiresIn   int    `json:"expireIn"`
	CreateDate  time.Time
}

type TOpenConversationIDInfo struct {
	Code               string `json:"code,omitempty"`
	Message            string `json:"message,omitempty"`
	RequestID          string `json:"requestId,omitempty"`
	OpenConversationID string `json:"openConversationId,omitempty"`
}

// String 转换为字符串
func (Self *TJsapiTicket) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

// IsValid 检查 JsapiTicket 是否有效
func (Self *TJsapiTicket) IsValid() bool {
	return Self.JsapiTicket != "" && int(time.Now().Unix()-Self.CreateDate.Unix()) <= Self.ExpiresIn
}

type workNotify struct {
	TResult
	TaskId int `json:"task_id"`
}

// String 转换为字符串
func (Self *workNotify) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

type TDeptInfo struct {
	TResult
	Id          int      `json:"dept_id"`
	PId         int      `json:"parent_id"`
	Name        string   `json:"name"`
	MemberCount int      `json:"member_count"`
	MUserIds    []string `json:"dept_manager_userid_list"` //部门的主管列表
	OwnerUserId string   `json:"org_dept_owner"`           //企业群群主
}

type TDeptSubInfo struct {
	DeptId   int    `json:"dept_id"`
	Name     string `json:"name"`
	ParentId int    `json:"parent_id"`
}

// String 转换为字符串
func (Self *TDeptInfo) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

type TAdmin struct {
	UserId   string `json:"userid"`
	SysLevel int    `json:"sys_level"`
}

// String 转换为字符串
func (Self *TAdmin) String() string {
	if Self.SysLevel == 1 {
		return fmt.Sprintf("主 %s", Self.UserId)
	} else {
		return fmt.Sprintf("子 %s", Self.UserId)
	}
}

type TAdmins struct {
	TResult
	Admins []TAdmin `json:"adminList"`
}

// String 转换为字符串
func (Self *TAdmins) String() string {
	out := "管理员列表"
	for _, v := range Self.Admins {
		out = out + "\n" + v.String()
	}
	return out
}

type TDDUser struct {
	TResult
	UserId    string      `json:"userid"`
	StaffCode string      `json:"jobnumber"`
	StaffName string      `json:"name"`
	Email     string      `json:"email"`
	Phone     string      `json:"mobile"`
	Remark    string      `json:"remark"`
	Avatar    string      `json:"avatar"`
	Attrs     TDDUserAttr `json:"extattr"`
}

type TDDV2ReportDept struct {
	DeptId   int             `json:"dept_id"`
	DeptName string          `json:"dept_name"`
	Users    []*TDDV2User    `json:"users"`
	SubDepts []*TDeptSubInfo `json:"sub_depts"`
}

type TDDV2ReportUsers struct {
	Departments []*TDDV2ReportDept `json:"departments"`
}

type TDeptLeader struct {
	DeptId int  `json:"dept_id"`
	Leader bool `json:"leader"`
}

type TDDV2User struct {
	UserId       string        `json:"userid"`
	UnionId      string        `json:"unionid"`
	StaffCode    string        `json:"job_number"`
	StaffName    string        `json:"name"`
	Department   []int         `json:"dept_id_list"`
	LeaderInDept []TDeptLeader `json:"leader_in_dept"`
	Email        string        `json:"email"`
	Phone        string        `json:"mobile"`
	Remark       string        `json:"remark"`
	Avatar       string        `json:"avatar"`
	Attrs        TDDV2UserAttr
	AttrText     string `json:"extension,omitempty"`
}

type TDDV2UserAttr struct {
	JoinDate  string `json:"joinDate"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birthDate"`
	Age       string `json:"age"`
	Station   string `json:"station"`
	Belong    string `json:"belong"`
	Org       string `json:"org"`
	Dept      string `json:"dept"`
	Job       string `json:"job"`
	OrgId     string `json:"orgId"`
	DeptId    string `json:"dept_id"`
	JobId     string `json:"jobId"`
}

type TDDUserAttr struct {
	JoinDate  string `json:"joinDate"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birthDate"`
	Age       string `json:"age"`
	Station   string `json:"station"`
	Belong    string `json:"belong"`
	Org       string `json:"org"`
	Dept      string `json:"dept"`
	Job       string `json:"job"`
	OrgId     int    `json:"org_id"`
	DeptId    int    `json:"dept_id"`
	JobId     int    `json:"job_id"`
}

// String 转换为字符串
func (Self *TDDUser) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

// DisplayName 获取显示名称
func (Self *TDDUser) DisplayName() string {
	if Self.Attrs.Org == "" {
		return fmt.Sprintf("%s", Self.StaffName)
	} else {
		return fmt.Sprintf("%s(%s)", Self.StaffName, Self.Attrs.Org)
	}
}

// String 转换为字符串
func (Self *TDDV2User) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

// DisplayName 获取显示名称
func (Self *TDDV2User) DisplayName() string {
	if Self.Attrs.Org == "" {
		return fmt.Sprintf("%s", Self.StaffName)
	} else {
		return fmt.Sprintf("%s(%s)", Self.StaffName, Self.Attrs.Org)
	}
}

type TDDReportSimpleItem struct {
	CreateTime   int64  `json:"create_time"`
	CreatorID    string `json:"creator_id"`
	CreatorName  string `json:"creator_name"`
	DeptName     string `json:"dept_name"`
	Remark       string `json:"remark"`
	ReportID     string `json:"report_id"`
	TemplateName string `json:"template_name"`
}

func (Self *TDDReportSimpleItem) ToString() string {
	CreateTimeStr := time.Unix(Self.CreateTime/1000, 0).Format("2006-01-02 15:04")
	return fmt.Sprintf("%s的%s (%s)", Self.CreatorName, Self.TemplateName, CreateTimeStr)
}

type TDDReportSimpleList struct {
	DataList   []TDDReportSimpleItem `json:"data_list"`
	HasMore    bool                  `json:"has_more"`
	NextCursor int64                 `json:"next_cursor"`
	Size       int                   `json:"size"`
}

type TDDReportList struct {
	DataList   []TDDReportItem `json:"data_list"`
	HasMore    bool            `json:"has_more"`
	NextCursor int64           `json:"next_cursor"`
	Size       int             `json:"size"`
}

func (Self *TDDReportList) GetReportText(ReportID string) string {
	var out string
	for _, v := range Self.DataList {
		if ReportID == "" || v.ReportID == ReportID {
			out = out + v.ToString() + "\n"
			for _, c := range v.Contents {
				if c.Type == "1" {
					out = out + c.ToString() + "\n"
				}
			}
		}
		out = out + "\n"
	}
	return out
}

type TDDReportTemplateField struct {
	FieldName string `json:"field_name"`
	Type      int    `json:"type"`
	Sort      int    `json:"sort"`
}
type TDDReportTemplateDefaultReceiver struct {
	UserName string `json:"user_name"`
	UserID   string `json:"userid"`
}
type TDDReportTemplateDefaultReceivedConv struct {
	ConversationID string `json:"conversation_id"`
	Title          string `json:"title"`
}
type TDDReportTemplate struct {
	Fields               []TDDReportTemplateField               `json:"fields"`
	UserID               string                                 `json:"userid"`
	UserName             string                                 `json:"user_name"`
	TemplateID           string                                 `json:"id"`
	TemplateName         string                                 `json:"name"`
	DefaultReceivers     []TDDReportTemplateDefaultReceiver     `json:"default_receivers"`
	DefaultReceivedConvs []TDDReportTemplateDefaultReceivedConv `json:"default_received_convs"`
}

type TDDReportTemplateItem struct {
	Name       string `json:"name"`
	ReportCode string `json:"report_code"`
	URL        string `json:"url"`
}
type TDDReportTemplateList struct {
	TemplateList []TDDReportTemplateItem `json:"template_list"`
	NextCursor   int64                   `json:"next_cursor"`
}

const (
	DDReportStatisticsTypeRead    = 0
	DDReportStatisticsTypeComment = 1
	DDReportStatisticsTypeLike    = 2
)

type TDDReportStatisticsListByTypeRequest struct {
	ReportID string `json:"report_id"`
	Type     int    `json:"type"`
	Offset   int64  `json:"offset"`
	Size     int    `json:"size"`
}

type TDDReportStatisticsPage struct {
	HasMore    bool     `json:"has_more"`
	NextCursor int64    `json:"next_cursor"`
	UserIDList []string `json:"userid_list"`
}

type TDDReportItem struct {
	Contents     []TDDReportContent `json:"contents"`
	CreateTime   int64              `json:"create_time"`
	CreatorID    string             `json:"creator_id"`
	CreatorName  string             `json:"creator_name"`
	DeptName     string             `json:"dept_name"`
	Images       []string           `json:"images"`
	ModifiedTime int64              `json:"modified_time"`
	Remark       string             `json:"remark"`
	ReportID     string             `json:"report_id"`
	TemplateName string             `json:"template_name"`
}

func (Self *TDDReportItem) ToString() string {
	ModifiedTimeStr := time.Unix(Self.ModifiedTime/1000, 0).Format("2006-01-02 15:04")
	return fmt.Sprintf("# %s的%s (%s)", Self.CreatorName, Self.TemplateName, ModifiedTimeStr)
}

type TDDReportContent struct {
	Key   string `json:"key"`
	Sort  string `json:"sort"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (Self *TDDReportContent) ToString() string {
	return fmt.Sprintf("## %s\r\n%s", Self.Key, Self.Value)
}

// ReportContent 日志内容项
type TDDCreateReportContent struct {
	Key         string `json:"key"`          // 字段名称
	Type        int    `json:"type"`         // 类型
	Sort        int    `json:"sort"`         // 排序
	ContentType string `json:"content_type"` // 内容类型（如：markdown）
	Content     string `json:"content"`      // 内容
}

// CreateReportParam 创建日志参数
type TDDCreateReportParam struct {
	TemplateID string                   `json:"template_id"`       // 模板ID
	Contents   []TDDCreateReportContent `json:"contents"`          // 日志内容列表
	ToUserIDs  []string                 `json:"to_userids"`        // 接收人用户ID列表
	ToChat     bool                     `json:"to_chat"`           // 是否发送到群聊
	DDFrom     string                   `json:"dd_from,omitempty"` // 来源
	UserID     string                   `json:"userid"`            // 创建人用户ID
}

type TDDCreateReportParamV2 struct {
	TemplateID string                   `json:"template_id"`
	Contents   []TDDCreateReportContent `json:"contents"`
	ToUserIDs  []string                 `json:"to_userids"`
	ToCIDs     []string                 `json:"to_cids,omitempty"`
	ToChat     bool                     `json:"to_chat"`
	DDFrom     string                   `json:"dd_from,omitempty"`
	UserID     string                   `json:"userid"`
}

type TDingTalkApp struct {
	ddurl             string
	appkey            string
	appsecret         string
	corp_id           string
	agent_id          string
	token             *TAccessToken
	ticket            *TJsapiTicket
	transport         http.RoundTripper
	timeout_connect   time.Duration
	timeout_readwrite time.Duration
}

// String 转换为字符串
func (Self *TDingTalkApp) String() string {
	str_text := fmt.Sprintf(`CorpId:%s,AgentId:%s,AppKey:%s,AppSecret:%s`, utils.GetShowKey(Self.corp_id), Self.agent_id, utils.GetShowKey(Self.appkey), utils.GetShowKey(Self.appsecret))
	return str_text
}

// GetDingTalkApp 获取钉钉 App 实例
func GetDingTalkApp(appkey, appsecret, corp_id, agent_id string) *TDingTalkApp {
	return &TDingTalkApp{
		appkey:            appkey,
		appsecret:         appsecret,
		corp_id:           corp_id,
		agent_id:          agent_id,
		ddurl:             `https://oapi.dingtalk.com`,
		timeout_connect:   30 * time.Second,
		timeout_readwrite: 30 * time.Second,
	}
}

// SetAgentId 设置 Agent ID
func (Self *TDingTalkApp) SetAgentId(agent_id string) {
	Self.agent_id = agent_id
}

// SetTransport 为当前 DingTalk app 实例注入自定义 transport。
// 这样代理配置只作用于当前实例，不会污染 network 包的全局默认设置。
func (Self *TDingTalkApp) SetTransport(transport http.RoundTripper) {
	Self.transport = transport
}

// newGet/newPost 保留为向后兼容的轻量包装，真正的入口是 newRequest。
func (Self *TDingTalkApp) newGet(url string) *network.THttpRequest {
	var req *network.THttpRequest
	req = network.HttpGet(url)
	req.SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.SetTransport(Self.transport)
	return req
}

func (Self *TDingTalkApp) newPost(url string) *network.THttpRequest {
	var req *network.THttpRequest
	req = network.HttpPost(url)
	req.SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.SetTransport(Self.transport)
	return req
}

// GetConfig is to return config in json
func (c *TDingTalkApp) GetConfig(nonceStr string, timestamp string, url string) (string, error) {
	ticket, err := c.GetJSAPITicket()
	if err != nil {
		return "", err
	}

	s := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%s&url=%s", ticket, nonceStr, timestamp, url)
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)

	config := map[string]string{
		"url":       url,
		"nonceStr":  nonceStr,
		"agentId":   c.agent_id,
		"timeStamp": timestamp,
		"corpId":    c.corp_id,
		"ticket":    ticket,
		"signature": fmt.Sprintf("%x", bs),
	}
	// logs.Debug("TDingTalkApp::GetConfig() : CorpId = %s", utils.GetShowKey(config["corpId"]))
	// logs.Debug("TDingTalkApp::GetConfig() : AgentId = %s", config["agentId"])
	// logs.Debug("TDingTalkApp::GetConfig() : Ticket = %s", utils.GetShowKey(config["ticket"]))
	// logs.Debug("TDingTalkApp::GetConfig() : Signature = %s", utils.GetShowKey(config["signature"]))
	// logs.Debug("TDingTalkApp::GetConfig() : URL = %s", config["url"])
	return utils.GetJson(&config), nil
}

// GetAccessToken 获取 AccessToken
// https://oapi.dingtalk.com/gettoken?appkey=key&appsecret=secret
// {"errorCode":503,"success":false,"errorMsg":"不合法的access_token"}
func (Self *TDingTalkApp) GetAccessToken() (string, error) {
	if Self.token != nil && Self.token.IsValid() {
		return Self.token.AccessToken, nil
	}

	req := Self.newGet(Self.ddurl + "/gettoken")
	req.Param("appkey", Self.appkey)
	req.Param("appsecret", Self.appsecret)

	var info TAccessToken
	err := req.ToJSON(&info)
	if err != nil {
		return "", errors.New("获取钉钉临时令牌失败！")
	}
	if info.ErrCode == 0 {
		info.ExpiresIn = 7200
		info.CreateDate = time.Now()
		Self.token = &info
		//logs.Debug("GetAccessToken() : 获取钉钉Token信息 ...... OK (%s)", info.AccessToken)
		return info.AccessToken, nil
	} else {
		logs.Debug("GetAccessToken() : 获取钉钉Token信息失败，%s(%d)", info.ErrMsg, info.ErrCode)
		return "", fmt.Errorf("%s(%d)", info.ErrMsg, info.ErrCode)
	}
}

// 获取jsapiTicket (这里需要实现获取ticket的逻辑，可能是从缓存或钉钉API获取)
func (Self *TDingTalkApp) GetJSAPITicket() (string, error) {
	if Self.ticket != nil && Self.ticket.IsValid() {
		return Self.ticket.JsapiTicket, nil
	}

	_, err := Self.GetAccessToken()
	if err != nil {
		return "", err
	}

	ddurl := "https://api.dingtalk.com/v1.0/oauth2/jsapiTickets"

	req := Self.newPost(ddurl)
	req.Header("x-acs-dingtalk-access-token", Self.token.AccessToken)
	//req.Param("access_token", Self.token.AccessToken)
	logs.Debug("访问接口：%s (获取jsapiTicket)", ddurl)

	var reqData TJsapiTicket
	err = req.ToJSON(&reqData)
	if err != nil {
		return "", errors.New("获取钉钉Ticket失败！")
	}

	if reqData.Code != "" {
		return "", errors.New(reqData.Message)
	}

	reqData.CreateDate = time.Now()
	Self.ticket = &reqData

	logs.Debug("返回数据：%s 超时 %d (%s)", utils.GetShowKey(reqData.JsapiTicket), reqData.ExpiresIn, reqData.CreateDate.Add(time.Duration(reqData.ExpiresIn)*time.Second).Format(utils.DateTimeFormat))
	return reqData.JsapiTicket, nil
}

// GetAdmins 获取管理员列表
// {"sys_level":2,"userid":"userid2"},
// https://oapi.dingtalk.com/user/get_admin?access_token=ACCESS_TOKEN
// GetOpenConversationIDByChatID 鏍规嵁 chatId 鑾峰彇 openConversationId
func (Self *TDingTalkApp) GetOpenConversationIDByChatID(chatID string) (string, error) {
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return "", errors.New("chatID is empty")
	}

	_, err := Self.GetAccessToken()
	if err != nil {
		return "", err
	}

	escapedChatID := strings.ReplaceAll(chatID, "/", "%2F")
	ddurl := "https://api.dingtalk.com/v1.0/im/chat/" + escapedChatID + "/convertToOpenConversationId"
	req := Self.newPost(ddurl)
	req.Header("x-acs-dingtalk-access-token", Self.token.AccessToken)
	logs.Debug("璁块棶鎺ュ彛锛?s (鏍规嵁chatId鑾峰彇openConversationId)", ddurl)

	var info TOpenConversationIDInfo
	if err := req.ToJSON(&info); err != nil {
		return "", err
	}

	statusCode, err := req.StatusCode()
	if err != nil {
		return "", err
	}

	if statusCode == http.StatusUnauthorized {
		Self.token = nil
		return Self.GetOpenConversationIDByChatID(chatID)
	}
	if statusCode != http.StatusOK {
		if info.Message != "" {
			if info.RequestID != "" {
				return "", fmt.Errorf("%s,RequestID: %s", info.Message, info.RequestID)
			}
			return "", errors.New(info.Message)
		}
		return "", fmt.Errorf("http status %d", statusCode)
	}
	if info.Code != "" {
		if info.RequestID != "" {
			return "", fmt.Errorf("%s,RequestID: %s", info.Message, info.RequestID)
		}
		if info.Message != "" {
			return "", errors.New(info.Message)
		}
		return "", errors.New(info.Code)
	}

	info.OpenConversationID = strings.TrimSpace(info.OpenConversationID)
	if info.OpenConversationID == "" {
		return "", errors.New("openConversationId is empty")
	}

	logs.Debug("杩斿洖鏁版嵁锛?s", utils.GetShowKey(info.OpenConversationID))
	return info.OpenConversationID, nil
}

// GetAdmins 鑾峰彇绠＄悊鍛樺垪琛?
// {"sys_level":2,"userid":"userid2"},
// https://oapi.dingtalk.com/user/get_admin?access_token=ACCESS_TOKEN
func (Self *TDingTalkApp) GetAdmins() (*TAdmins, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	//首先通过免authcode登授权码,获取当前登录userid
	req := Self.newGet(Self.ddurl + "/user/get_admin")
	req.Param("access_token", Self.token.AccessToken)

	var info TAdmins
	err = req.ToJSON(&info)
	if err != nil {
		return nil, err
	}
	if info.ErrCode == 0 {
		return &info, nil
	} else if info.ErrCode == 503 {
		Self.token = nil
		return Self.GetAdmins()
	} else {
		return nil, errors.New(info.ErrMsg)
	}
}

// GetUserInfoByPhone 根据手机号获取用户信息
// https://oapi.dingtalk.com/user/get_by_mobile?access_token=ACCESS_TOKEN&mobile=1xxxxxxxxxx
func (Self *TDingTalkApp) GetV2UserInfoByPhone(phone string) (*TDDV2User, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}
	ddurl := Self.ddurl + "/topapi/v2/user/getbymobile"
	//req := network.HttpGet(ddurl).SetTimeout(Self.timeout_connect, Self.timeout_readwrite).SetTransport(Self.transport)
	req := Self.newGet(ddurl)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("mobile", phone)
	logs.Debug("访问接口：%s (获取用户标识)", ddurl)

	var dingResp TDingTalkResponse
	err = req.ToJSON(&dingResp)
	if err != nil {
		return nil, err
	}

	if dingResp.ErrCode == 42001 {
		Self.token = nil
		return Self.GetV2UserInfoByPhone(phone)
	} else if dingResp.ErrCode != 0 {
		return nil, errors.New(dingResp.ErrMsg)
	}

	logs.Debug("返回数据：%s", string(dingResp.Result))
	var info TDDV2User
	err = json.Unmarshal(dingResp.Result, &info)
	if err != nil {
		return nil, err
	}
	return Self.GetV2UserInfo(info.UserId)
}

// GetUserInfoByUnionId 根据 UnionId 获取用户信息
// https://oapi.dingtalk.com/topapi/user/getbyunionid?access_token=ACCESS_TOKEN&unionid=zhangsan
func (Self *TDingTalkApp) GetV2UserInfoByUnionId(unionid string) (*TDDV2User, error) {
	logs.Debug("GetUserInfoByUnionId() : 获取钉钉用户信息")
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}
	ddurl := Self.ddurl + "/topapi/user/getbyunionid"
	req := Self.newGet(ddurl)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("unionid", unionid)
	logs.Debug("访问接口：%s (获取用户标识)", ddurl)

	var dingResp TDingTalkResponse
	err = req.ToJSON(&dingResp)
	if err != nil {
		return nil, err
	}

	if dingResp.ErrCode == 42001 {
		Self.token = nil
		return Self.GetV2UserInfoByUnionId(unionid)
	} else if dingResp.ErrCode != 0 {
		return nil, errors.New(dingResp.ErrMsg)
	}

	logs.Debug("返回数据：%s", string(dingResp.Result))
	var info TDDV2User
	err = json.Unmarshal(dingResp.Result, &info)
	if err != nil {
		return nil, err
	}
	return Self.GetV2UserInfo(info.UserId)
}

// GetUsersByName 根据姓名获取用户列表
// https://oapi.dingtalk.com/user/get_by_name?access_token=ACCESS_TOKEN&name=张三
func (Self *TDingTalkApp) GetV2UsersByName(name string) ([]*TDDV2User, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	ddurl := "https://api.dingtalk.com/v1.0/contact/users/search"
	logs.Debug("访问接口：%s (根据姓名获取用户列表)", ddurl)

	header := make(map[string]string)
	header["x-acs-dingtalk-access-token"] = Self.token.AccessToken
	header["Content-Type"] = "application/json"
	var reqData struct {
		QueryWord      string `json:"queryWord"`      // 搜索关键字（驼峰命名）
		Offset         int    `json:"offset"`         // 分页偏移量，必需
		Size           int    `json:"size"`           // 分页大小，必需
		FullMatchField string `json:"fullMatchField"` // 完全匹配字段，可选
	}
	reqData.QueryWord = name
	reqData.Offset = 0
	reqData.Size = 200
	//reqData.FullMatchField = "1"
	var dingResp struct {
		ErrCode    string   `json:"code"`
		ErrMsg     string   `json:"message"`
		HasMore    bool     `json:"hasMore"`
		TotalCount int      `json:"totalCount"`
		List       []string `json:"list"`
	}
	var users []*TDDV2User
	var userids []string
	for {
		// 这里不能再走 HttpRequestBHB，否则会绕过实例级 transport，导致代理配置不生效。
		req := Self.newPost(ddurl)
		for key, value := range header {
			req.Header(key, value)
		}
		if _, err := req.JSONBody(reqData); err != nil {
			return nil, err
		}
		if err := req.ToJSON(&dingResp); err != nil {
			return nil, err
		}
		statusCode, err := req.StatusCode()
		if err != nil {
			return nil, err
		}
		if statusCode == 200 {
			logs.Debug("查询 %s 有 %d/%d 个用户", name, reqData.Offset, dingResp.TotalCount)
			for _, userid := range dingResp.List {
				userids = append(userids, userid)
			}
			if !dingResp.HasMore {
				break
			}
			reqData.Offset += reqData.Size
		} else {
			return nil, errors.New(dingResp.ErrMsg)
		}
	}

	for _, userid := range userids {

		info, err := Self.GetV2UserInfo(userid)
		if err != nil {
			return nil, err
		}
		users = append(users, info)
	}
	return users, nil
}

// GetSubDeptIds 获取部门的直属子部门
// https://oapi.dingtalk.com/topapi/v2/department/listsub
func (Self *TDingTalkApp) GetSubDeptIds(deptId int) ([]int, error) {
	depts, err := Self.GetSubDepts(deptId)
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(depts))
	for _, d := range depts {
		ids = append(ids, d.DeptId)
	}
	return ids, nil
}

func (Self *TDingTalkApp) GetSubDepts(deptId int) ([]*TDeptSubInfo, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	ddurl := Self.ddurl + "/topapi/v2/department/listsub"
	req := Self.newPost(ddurl)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("dept_id", strconv.Itoa(deptId))
	logs.Debug("访问接口：%s (获取子部门 - %d)", ddurl, deptId)

	var dingResp TDingTalkResponse
	err = req.ToJSON(&dingResp)
	if err != nil {
		return nil, err
	}

	switch dingResp.ErrCode {
	case 0:
		var deptList []*TDeptSubInfo
		err = json.Unmarshal(dingResp.Result, &deptList)
		if err != nil {
			return nil, err
		}
		logs.Debug("==> %d (共 %d 个子部门)", deptId, len(deptList))
		return deptList, nil
	case 503:
		Self.token = nil
		return Self.GetSubDepts(deptId)
	default:
		return nil, errors.New(dingResp.ErrMsg)
	}
}

// GetV2ReportUsers 获取用户所属部门的员工列表（按多级子部门分组）
// 如果用户是部门主管，返回该部门及所有层级子团队的全部员工（含主管）
// 如果用户不是部门主管，返回该部门及所有层级子团队除主管之外的员工
func (Self *TDingTalkApp) GetV2ReportUsers(userid string) (*TDDV2ReportUsers, error) {
	logs.Info("获取用户信息(userid=%s)", userid)
	userInfo, err := Self.GetV2UserInfo(userid)
	if err != nil {
		return nil, err
	}
	logs.Info("当前用户 %s 所属部门: %v", userid, userInfo.Department)

	report := &TDDV2ReportUsers{}

	for _, deptId := range userInfo.Department {
		if deptId == 1 {
			continue
		}
		logs.Info("获取部门 %d 信息", deptId)
		deptInfo, err := Self.GetV2Department(deptId)
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(deptInfo.Name, "HRBP_") || strings.HasSuffix(deptInfo.Name, "_HRBP") || strings.Contains(deptInfo.Name, "钉钉合作") ||
			strings.Contains(deptInfo.Name, "驻地") || strings.Contains(deptInfo.Name, "研发中心") || strings.Contains(deptInfo.Name, "互联网公司") {
			continue
		}

		logs.Info("获取部门 %d 信息 - %s, 主管=%v", deptId, deptInfo.Name, deptInfo.MUserIds)

		dept_users := &TDDV2ReportDept{}
		dept_users.DeptId = deptId
		dept_users.DeptName = deptInfo.Name

		//获取当前部门的子部门列表
		subDepts, err := Self.GetSubDepts(deptId)
		if err != nil {
			return nil, err
		}
		logs.Info("获取部门 %d 子部门 - %d 个", deptId, len(subDepts))

		sort.Slice(subDepts, func(i, j int) bool {
			di, dj := subDepts[i], subDepts[j]
			if di == nil {
				return false
			}
			if dj == nil {
				return true
			}
			ni, nj := di.Name, dj.Name
			if ni == "" && nj != "" {
				return false
			}
			if nj == "" && ni != "" {
				return true
			}
			return ni < nj
		})

		dept_users.SubDepts = subDepts
		//获取当前部门的子部门员工列表
		subDeptUsers, err := Self.GetDeptUsers(deptId)
		if err != nil {
			return nil, err
		}

		//如果用户不是主管，则过滤掉主管员工
		managerUserIDsText := "|" + strings.Join(deptInfo.MUserIds, "|") + "|"
		if managerUserIDsText != "||" && !strings.Contains(managerUserIDsText, "|"+userid+"|") {
			filtered := make([]*TDDV2User, 0, len(subDeptUsers))
			for _, u := range subDeptUsers {
				if u == nil || u.UserId == "" {
					continue
				}
				if strings.Contains(managerUserIDsText, "|"+u.UserId+"|") {
					continue
				}
				filtered = append(filtered, u)
			}
			subDeptUsers = filtered
		}

		//根据员工Code排序
		sort.Slice(subDeptUsers,
			func(i, j int) bool {
				ui, uj := subDeptUsers[i], subDeptUsers[j]
				if ui == nil {
					return false
				}
				if uj == nil {
					return true
				}
				si, sj := ui.StaffCode, uj.StaffCode
				if si == "" && sj != "" {
					return false
				}
				if sj == "" && si != "" {
					return true
				}
				return si < sj
			})

		logs.Info("获取部门 %d 子部门员工 - %d 人", deptId, len(subDeptUsers))
		dept_users.Users = subDeptUsers

		report.Departments = append(report.Departments, dept_users)
	}

	return report, nil
}

// GetV2UserInfo 根据 UserID 获取用户信息
// https://oapi.dingtalk.com/topapi/v2/user/get?access_token=ACCESS_TOKEN&userid=zhangsan
func (Self *TDingTalkApp) GetV2UserInfo(userid string) (*TDDV2User, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	ddurl := Self.ddurl + "/topapi/v2/user/get"
	req := Self.newGet(ddurl)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("userid", userid)
	//logs.Debug("访问接口：%s (获取用户详情)", ddurl)

	var dingResp TDingTalkResponse
	err = req.ToJSON(&dingResp)
	if err != nil {
		return nil, err
	}
	//logs.Debug("返回数据：%s", string(dingResp.Result))

	switch dingResp.ErrCode {
	case 0:
		var info TDDV2User
		err = json.Unmarshal(dingResp.Result, &info)
		if err != nil {
			return nil, err
		}
		if info.AttrText != "" {
			err = json.Unmarshal([]byte(info.AttrText), &info.Attrs)
			if err != nil {
				return nil, err
			}
		}
		info.AttrText = ""
		//logs.Debug("==> %s - %s", info.StaffCode, info.StaffName)
		return &info, nil
	case 503:
		Self.token = nil
		return Self.GetV2UserInfo(userid)
	default:
		return nil, errors.New(dingResp.ErrMsg)
	}
}

// GetDepartment 获取部门详情
// https://oapi.dingtalk.com/topapi/v2/department/get?access_token=ACCESS_TOKEN&dept_id=123
func (Self *TDingTalkApp) GetV2Department(depid int) (*TDeptInfo, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	ddurl := Self.ddurl + "/topapi/v2/department/get"
	req := Self.newGet(ddurl)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("dept_id", strconv.Itoa(depid))
	logs.Debug("访问接口：%s (获取部门详情)", ddurl)

	var dingResp TDingTalkResponse
	err = req.ToJSON(&dingResp)
	if err != nil {
		return nil, err
	}
	switch dingResp.ErrCode {
	case 0:
		//logs.Debug("返回数据：%s", string(dingResp.Result))
		var info TDeptInfo
		err = json.Unmarshal(dingResp.Result, &info)
		if err != nil {
			return nil, err
		}
		logs.Debug("==> %d - %s", info.Id, info.Name)
		return &info, nil
	case 503:
		Self.token = nil
		return Self.GetV2Department(depid)
	default:
		return nil, errors.New(dingResp.ErrMsg)
	}
}

// 获取部门详情
func (Self *TDingTalkApp) GetDeptUsers(depid int) ([]*TDDV2User, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}
	ddurl := Self.ddurl + "/topapi/user/listid"
	req := Self.newGet(ddurl)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("dept_id", strconv.Itoa(depid))
	logs.Debug("访问接口：%s (获取部门用户) - %d", ddurl, depid)

	var dingResp TDingTalkResponse
	err = req.ToJSON(&dingResp)
	if err != nil {
		return nil, err
	}

	switch dingResp.ErrCode {
	case 0:
		// 获取员工列表
		var info struct {
			UserIDList []string `json:"userid_list"`
		}
		err = json.Unmarshal(dingResp.Result, &info)
		if err != nil {
			return nil, err
		}

		var users []*TDDV2User
		for k, v := range info.UserIDList {
			if k >= 100 {
				logs.Warning("最多获取 100 人 (部门=%d)", depid)
				break
			}
			user, err := Self.GetV2UserInfo(v)
			if err != nil {
				return nil, err
			} else {
				users = append(users, user)
			}
		}
		return users, nil
	case 503:
		Self.token = nil
		return Self.GetDeptUsers(depid)
	default:
		return nil, errors.New(dingResp.ErrMsg)
	}
}

func (Self *TDingTalkApp) GetDeptUsersV2(deptId int) ([]*TDDV2User, error) {
	byUserID := map[string]*TDDV2User{}
	ordered := make([]*TDDV2User, 0, 128)
	visitedDepts := map[int]struct{}{}

	var walk func(id int) error
	walk = func(id int) error {
		if _, ok := visitedDepts[id]; ok {
			return nil
		}
		visitedDepts[id] = struct{}{}

		users, err := Self.GetDeptUsers(id)
		if err != nil {
			return err
		}
		for _, u := range users {
			if u == nil || u.UserId == "" {
				continue
			}
			if _, ok := byUserID[u.UserId]; ok {
				continue
			}
			byUserID[u.UserId] = u
			ordered = append(ordered, u)
		}

		subDeptIDs, err := Self.GetSubDeptIds(id)
		if err != nil {
			return err
		}
		for _, subID := range subDeptIDs {
			if err := walk(subID); err != nil {
				return err
			}
		}
		return nil
	}

	if err := walk(deptId); err != nil {
		return nil, err
	}
	return ordered, nil
}

// GetOrgName 获取组织名称
func (Self *TDingTalkApp) GetOrgName(userid string) (string, error) {
	logs.Debug("GetOrgName() : 获取钉钉部门信息")

	info, err := Self.GetV2UserInfo(userid)
	if err != nil {
		return "", err
	}

	return info.Attrs.Org, nil
}

// GetJobName 获取职位名称
func (Self *TDingTalkApp) GetJobName(userid string) (string, error) {
	logs.Debug("GetJobName() : 获取钉钉岗位信息")

	info, err := Self.GetV2UserInfo(userid)
	if err != nil {
		return "", err
	}
	return info.Attrs.Job, nil
}

// GetFullDepartmentName 获取完整部门名称
func (Self *TDingTalkApp) GetFullDeptName(depid int) (string, error) {
	info, err := Self.GetV2Department(depid)
	if err != nil {
		return "", err
	} else {
		if info.PId > 0 {
			name, err := Self.GetFullDeptName(info.PId)
			if err != nil {
				return "", err
			}
			if info.PId == 1 {
				return info.Name, nil
			} else {
				return name + "-" + info.Name, nil
			}
		} else {
			return "", nil
		}
	}
}

// GetLoginInfo 获取登录信息
// https://oapi.dingtalk.com/user/getuserinfo?access_token=access_token&code=code
func (Self *TDingTalkApp) GetV2LoginInfo(authcode string) (string, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return "", err
	}

	//首先通过authcode免登授权码,获取当前登录userid
	ddurl := Self.ddurl + "/topapi/v2/user/getuserinfo"
	req := Self.newGet(ddurl)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("code", authcode)
	logs.Debug("访问接口：%s (获取登录用户信息)", ddurl)

	var dingResp TDingTalkResponse
	err = req.ToJSON(&dingResp)
	if err != nil {
		return "", err
	}
	switch dingResp.ErrCode {
	case 0:
		logs.Debug("返回数据：%s", string(dingResp.Result))
		var info struct {
			UserId   string `json:"userid"`
			DeviceID string `json:"deviceid"`
			Name     string `json:"name"`
		}
		err = json.Unmarshal(dingResp.Result, &info)
		if err != nil {
			return "", err
		}
		logs.Debug("当前登录信息：%s - %s(%s)", info.DeviceID, info.Name, info.UserId)
		return info.UserId, nil
	case 503:
		Self.token = nil
		return Self.GetV2LoginInfo(authcode)
	default:
		return "", errors.New(dingResp.ErrMsg)
	}
}

// SendWorkNotify 发送工作通知
func (Self *TDingTalkApp) SendWorkNotify(user_id string, msg_text string) (int, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return -1, err
	}
	ddurl := Self.ddurl + "/topapi/message/corpconversation/asyncsend_v2"
	req := Self.newPost(ddurl)
	req.Param("access_token", Self.token.AccessToken)
	//req.Param("agent_id", Self.agent_id)
	req.Param("agent_id", Self.agent_id)
	req.Param("userid_list", user_id)
	req.Param("msg", msg_text)
	logs.Debug("访问接口：%s (发送工作通知)", ddurl)

	body_text, err := req.String()
	if err != nil {
		return -1, fmt.Errorf("发送钉钉工作通知消息失败，原因：%s", err.Error())
	}

	var info workNotify
	err = req.ToJSON(&info)
	if err != nil {
		return -1, fmt.Errorf("发送钉钉工作通知消息失败，内容：%s，原因：%s", body_text, err.Error())
	}
	switch info.ErrCode {
	case 0:
		logs.Debug("SendWorkNotify() : 发送钉钉工作通知消息成功(%d)", info.TaskId)
		return info.TaskId, nil
	case 503:
		Self.token = nil
		return Self.SendWorkNotify(user_id, msg_text)
	default:
		return -1, errors.New(info.ErrMsg)
	}
}

// GetV2ReportTemplateList 获取用户可见的日志模板列表（自动分页）
// https://oapi.dingtalk.com/topapi/report/template/listbyuserid?access_token=ACCESS_TOKEN&userid=userid&cursor=cursor&size=size
func (Self *TDingTalkApp) GetV2ReportTemplateList(userid string) ([]TDDReportTemplateItem, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	var allItems []TDDReportTemplateItem
	var cursor int64 = 0
	pageSize := 100

	for {
		ddurl := Self.ddurl + "/topapi/report/template/listbyuserid"
		req := Self.newGet(ddurl)
		req.Param("access_token", Self.token.AccessToken)
		req.Param("userid", userid)
		req.Param("offset", fmt.Sprintf("%d", cursor))
		req.Param("size", fmt.Sprintf("%d", pageSize))
		logs.Debug("访问接口：%s (获取用户可见的日志模板) cursor=%d", ddurl, cursor)

		var dingResp TDingTalkResponse
		err = req.ToJSON(&dingResp)
		if err != nil {
			return nil, err
		}
		switch dingResp.ErrCode {
		case 0:
			var page TDDReportTemplateList
			err = json.Unmarshal(dingResp.Result, &page)
			if err != nil {
				return nil, err
			}
			//fmt.Printf("page.TemplateList: %+v\n", page)
			allItems = append(allItems, page.TemplateList...)
			logs.Debug("返回数据：本页 %d 个日志模板，累计 %d 个", len(page.TemplateList), len(allItems))
			if page.NextCursor == 0 || len(page.TemplateList) == 0 {
				sort.SliceStable(allItems, func(i, j int) bool {
					return strings.Compare(allItems[i].Name, allItems[j].Name) < 0
				})
				return allItems, nil
			}
			cursor = page.NextCursor
		case 503:
			Self.token = nil
			_, err = Self.GetAccessToken()
			if err != nil {
				return nil, err
			}
			continue
		default:
			return nil, errors.New(dingResp.ErrMsg)
		}
	}
}

// GetV2ReportTemplate 获取用户日志模板
// https://oapi.dingtalk.com/topapi/report/template/getbyname?access_token=ACCESS_TOKEN&userid=userid&template_name=template_name
func (Self *TDingTalkApp) GetV2ReportTemplate(userid, template_name string) (*TDDReportTemplate, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}
	ddurl := Self.ddurl + "/topapi/report/template/getbyname"
	req := Self.newGet(ddurl)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("userid", userid)
	req.Param("template_name", template_name)
	logs.Debug("访问接口：%s (获取用户日志模板)", ddurl)

	var dingResp TDingTalkResponse
	err = req.ToJSON(&dingResp)
	if err != nil {
		return nil, err
	}
	switch dingResp.ErrCode {
	case 0:
		var info TDDReportTemplate
		err = json.Unmarshal(dingResp.Result, &info)
		if err != nil {
			return nil, err
		}
		logs.Debug("返回数据：%d 个字段", len(info.Fields))
		return &info, nil
	case 503:
		Self.token = nil
		return Self.GetV2ReportTemplate(userid, template_name)
	default:
		return nil, errors.New(dingResp.ErrMsg)
	}
}

// GetV2ReportList 获取用户在指定时间范围内的日志列表（自动分页）
// https://oapi.dingtalk.com/topapi/report/list?access_token=ACCESS_TOKEN&userid=userid&start_time=start_time&end_time=end_time
func (Self *TDingTalkApp) GetV2ReportList(userid, start_time, end_time string) (*TDDReportList, error) {
	// 处理时间格式：start_time 取当天 00:00:00，end_time 取当天 23:59:59
	layout := "2006-01-02"
	startTime, err := time.ParseInLocation(layout, start_time, time.Local)
	if err != nil {
		return nil, err
	}
	endTime, err := time.ParseInLocation(layout, end_time, time.Local)
	if err != nil {
		return nil, err
	}
	endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	if endTime.Sub(startTime).Hours() > 180*24 {
		return nil, fmt.Errorf("查询时间跨度不能超过180天")
	}

	_, err = Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	var allList TDDReportList
	var cursor int64 = 0
	pageSize := 20

	for {
		ddurl := Self.ddurl + "/topapi/report/list"
		req := Self.newGet(ddurl)
		req.Param("access_token", Self.token.AccessToken)
		req.Param("userid", userid)
		req.Param("start_time", fmt.Sprintf("%d", startTime.UnixMilli()))
		req.Param("end_time", fmt.Sprintf("%d", endTime.UnixMilli()))
		req.Param("cursor", fmt.Sprintf("%d", cursor))
		req.Param("size", fmt.Sprintf("%d", pageSize))
		logs.Debug("访问接口：%s (获取用户日志列表) cursor=%d, %s - %s", ddurl, cursor, start_time, end_time)

		var dingResp TDingTalkResponse
		err = req.ToJSON(&dingResp)
		if err != nil {
			return nil, err
		}
		switch dingResp.ErrCode {
		case 0:
			var page TDDReportList
			err = json.Unmarshal(dingResp.Result, &page)
			if err != nil {
				return nil, err
			}
			allList.DataList = append(allList.DataList, page.DataList...)
			logs.Debug("返回数据：本页 %d 个日志，累计 %d 个", len(page.DataList), len(allList.DataList))
			if !page.HasMore {
				allList.HasMore = false
				allList.NextCursor = 0
				allList.Size = len(allList.DataList)
				return &allList, nil
			}
			cursor = page.NextCursor
		case 503:
			Self.token = nil
			_, err = Self.GetAccessToken()
			if err != nil {
				return nil, err
			}
			continue
		default:
			return nil, errors.New(dingResp.ErrMsg)
		}
	}
}

// GetV2ReportSimpleListByTemplate 获取日志模板从指定日期到当天的日志（自动按180天分段查询）
func (Self *TDingTalkApp) GetV2ReportSimpleListByTemplate(template_name, start_time string) ([]TDDReportSimpleItem, error) {
	layout := "2006-01-02"
	startTime, err := time.ParseInLocation(layout, start_time, time.Local)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local)

	_, err = Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	var allItems []TDDReportSimpleItem
	segmentStart := startTime

	for {
		segmentEnd := segmentStart.AddDate(0, 0, 179)
		if segmentEnd.After(today) {
			segmentEnd = today
		}

		items, err := Self.getV2ReportSimpleListByTemplateSegment(template_name, segmentStart, segmentEnd)
		if err != nil {
			return nil, err
		}
		allItems = append(allItems, items...)
		if len(items) > 0 {
			logs.Info("分段查询完成：%s ~ %s，本段 %d 条，累计 %d 条", segmentStart.Format(layout), segmentEnd.Format(layout), len(items), len(allItems))
		}
		if segmentEnd.Equal(today) || segmentEnd.After(today) {
			break
		}
		segmentStart = segmentEnd.AddDate(0, 0, 1)
	}

	return allItems, nil
}

// getV2ReportSimpleListByTemplateSegment 查询单个时间段内的日志（不超过180天，自动分页）
func (Self *TDingTalkApp) getV2ReportSimpleListByTemplateSegment(template_name string, startTime, endTime time.Time) ([]TDDReportSimpleItem, error) {
	layout := "2006-01-02"
	var allItems []TDDReportSimpleItem
	var cursor int64 = 0
	pageSize := 20

	for {
		ddurl := Self.ddurl + "/topapi/report/simplelist"
		req := Self.newGet(ddurl)
		req.Param("access_token", Self.token.AccessToken)
		req.Param("template_name", template_name)
		req.Param("start_time", fmt.Sprintf("%d", startTime.UnixMilli()))
		req.Param("end_time", fmt.Sprintf("%d", endTime.UnixMilli()))
		req.Param("cursor", fmt.Sprintf("%d", cursor))
		req.Param("size", fmt.Sprintf("%d", pageSize))
		logs.Debug("访问接口：%s (获取用户日志概要列表) cursor=%d, %s - %s", ddurl, cursor, startTime.Format(layout), endTime.Format(layout))

		var dingResp TDingTalkResponse
		err := req.ToJSON(&dingResp)
		if err != nil {
			return nil, err
		}
		switch dingResp.ErrCode {
		case 0:
			var page TDDReportSimpleList
			err = json.Unmarshal(dingResp.Result, &page)
			if err != nil {
				return nil, err
			}
			allItems = append(allItems, page.DataList...)
			logs.Debug("返回数据：本页 %d 个日志概要，累计 %d 个", len(page.DataList), len(allItems))
			if !page.HasMore {
				return allItems, nil
			}
			cursor = page.NextCursor
		case 503:
			Self.token = nil
			_, err = Self.GetAccessToken()
			if err != nil {
				return nil, err
			}
			continue
		default:
			return nil, errors.New(dingResp.ErrMsg)
		}
	}
}

// GetV2ReportSimpleList 获取用户在指定时间范围内的日志概要列表（自动分页）
// https://oapi.dingtalk.com/topapi/report/simplelist?access_token=ACCESS_TOKEN&userid=userid&start_time=start_time&end_time=end_time
func (Self *TDingTalkApp) GetV2ReportSimpleList(userid, start_time, end_time string) ([]TDDReportSimpleItem, error) {
	// 处理时间格式：start_time 取当天 00:00:00，end_time 取当天 23:59:59
	layout := "2006-01-02"
	startTime, err := time.ParseInLocation(layout, start_time, time.Local)
	if err != nil {
		return nil, err
	}
	endTime, err := time.ParseInLocation(layout, end_time, time.Local)
	if err != nil {
		return nil, err
	}
	endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	if endTime.Sub(startTime).Hours() > 180*24 {
		return nil, fmt.Errorf("查询时间跨度不能超过180天")
	}

	_, err = Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	var allItems []TDDReportSimpleItem
	var cursor int64 = 0
	pageSize := 20

	for {
		ddurl := Self.ddurl + "/topapi/report/simplelist"
		req := Self.newGet(ddurl)
		req.Param("access_token", Self.token.AccessToken)
		req.Param("userid", userid)
		req.Param("start_time", fmt.Sprintf("%d", startTime.UnixMilli()))
		req.Param("end_time", fmt.Sprintf("%d", endTime.UnixMilli()))
		req.Param("cursor", fmt.Sprintf("%d", cursor))
		req.Param("size", fmt.Sprintf("%d", pageSize))
		logs.Debug("访问接口：%s (获取用户日志概要列表) cursor=%d, %s - %s", ddurl, cursor, start_time, end_time)

		var dingResp TDingTalkResponse
		err = req.ToJSON(&dingResp)
		if err != nil {
			return nil, err
		}
		switch dingResp.ErrCode {
		case 0:
			var page TDDReportSimpleList
			err = json.Unmarshal(dingResp.Result, &page)
			if err != nil {
				return nil, err
			}
			allItems = append(allItems, page.DataList...)
			logs.Debug("返回数据：本页 %d 个日志概要，累计 %d 个", len(page.DataList), len(allItems))
			if !page.HasMore {
				return allItems, nil
			}
			cursor = page.NextCursor
		case 503:
			Self.token = nil
			_, err = Self.GetAccessToken()
			if err != nil {
				return nil, err
			}
			continue
		default:
			return nil, errors.New(dingResp.ErrMsg)
		}
	}
}

// GetV2ReportStatisticsListByType 获取日志已读/评论/点赞人员列表（自动分页）
// https://oapi.dingtalk.com/topapi/report/statistics/listbytype?access_token=ACCESS_TOKEN
func (Self *TDingTalkApp) GetV2ReportStatisticsListByType(report_id string, listType int) ([]string, error) {
	if listType != DDReportStatisticsTypeRead && listType != DDReportStatisticsTypeComment && listType != DDReportStatisticsTypeLike {
		return nil, fmt.Errorf("无效的日志统计类型: %d", listType)
	}

	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	var allUserIDs []string
	var offset int64 = 0
	pageSize := 100

	for {
		ddurl := Self.ddurl + "/topapi/report/statistics/listbytype?access_token=" + Self.token.AccessToken
		logs.Debug("访问接口：%s (获取日志相关人员列表) report_id=%s, type=%d, offset=%d", ddurl, report_id, listType, offset)

		reqData := TDDReportStatisticsListByTypeRequest{
			ReportID: report_id,
			Type:     listType,
			Offset:   offset,
			Size:     pageSize,
		}

		var dingResp TDingTalkResponse
		req := Self.newPost(ddurl)
		if _, err := req.JSONBody(reqData); err != nil {
			return nil, err
		}
		if err := req.ToJSON(&dingResp); err != nil {
			return nil, err
		}
		statusCode, err := req.StatusCode()
		if err != nil {
			return nil, err
		}
		if statusCode != 200 {
			return nil, fmt.Errorf("http status %d", statusCode)
		}

		switch dingResp.ErrCode {
		case 0:
			var page TDDReportStatisticsPage
			err = json.Unmarshal(dingResp.Result, &page)
			if err != nil {
				return nil, err
			}
			allUserIDs = append(allUserIDs, page.UserIDList...)
			logs.Debug("返回数据：本页 %d 个用户，累计 %d 个", len(page.UserIDList), len(allUserIDs))
			if !page.HasMore || len(page.UserIDList) == 0 || page.NextCursor == offset {
				return allUserIDs, nil
			}
			offset = page.NextCursor
		case 503:
			Self.token = nil
			_, err = Self.GetAccessToken()
			if err != nil {
				return nil, err
			}
			continue
		default:
			return nil, errors.New(dingResp.ErrMsg)
		}
	}
}

// GetV2ReportReadUserList 获取日志已读人员列表（自动分页）
func (Self *TDingTalkApp) GetV2ReportReadUserList(report_id string) ([]string, error) {
	return Self.GetV2ReportStatisticsListByType(report_id, DDReportStatisticsTypeRead)
}

// GetV2ReportCommentUserList 获取日志评论人员列表（自动分页）
func (Self *TDingTalkApp) GetV2ReportCommentUserList(report_id string) ([]string, error) {
	return Self.GetV2ReportStatisticsListByType(report_id, DDReportStatisticsTypeComment)
}

// GetV2ReportLikeUserList 获取日志点赞人员列表（自动分页）
func (Self *TDingTalkApp) GetV2ReportLikeUserList(report_id string) ([]string, error) {
	return Self.GetV2ReportStatisticsListByType(report_id, DDReportStatisticsTypeLike)
}

// CreateV2Report 创建用户日志
// https://oapi.dingtalk.com/topapi/v2/report/create?access_token=ACCESS_TOKEN&userid=userid&start_time=start_time&end_time=end_time
func (Self *TDingTalkApp) CreateV2ReportWithContents(userid, templateID string, contents []TDDCreateReportContent, toUserIDs []string, toCIDs []string, toChat bool) (string, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return "", err
	}

	ddurl := Self.ddurl + "/topapi/report/create?access_token=" + Self.token.AccessToken
	logs.Debug("访问接口：%s (创建用户日志)", ddurl)

	reportParam := TDDCreateReportParamV2{
		UserID:     userid,
		TemplateID: templateID,
		Contents:   contents,
		ToUserIDs:  normalizeNonEmptyStrings(toUserIDs),
		ToCIDs:     normalizeNonEmptyStrings(toCIDs),
		DDFrom:     "ygx",
	}
	reportParam.ToChat = toChat

	var dingResp struct {
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
		Result    string `json:"result"`
		RequestID string `json:"request_id"`
	}
	reqData := map[string]any{
		"create_report_param": reportParam,
	}
	req := Self.newPost(ddurl)
	if _, err := req.JSONBody(reqData); err != nil {
		return "", err
	}
	if err := req.ToJSON(&dingResp); err != nil {
		return "", err
	}
	statusCode, err := req.StatusCode()
	if err != nil {
		return "", err
	}
	if statusCode != 200 {
		return "", fmt.Errorf("http status %d", statusCode)
	}

	switch dingResp.ErrCode {
	case 0:
		return dingResp.Result, nil
	case 503:
		Self.token = nil
		return Self.CreateV2ReportWithContents(userid, templateID, contents, toUserIDs, toCIDs, toChat)
	default:
		return "", errors.New(dingResp.ErrMsg + ",RequestID: " + dingResp.RequestID)
	}
}

func normalizeNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func (Self *TDingTalkApp) CreateV2Report(userid, template_id, to_userids, a_text, b_text string) (string, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return "", err
	}

	ddurl := Self.ddurl + "/topapi/report/create?access_token=" + Self.token.AccessToken
	logs.Debug("访问接口：%s (创建用户日志)", ddurl)
	var contents [2]TDDCreateReportContent
	contents[0].Key = "本周完成工作"
	contents[0].Type = 1
	contents[0].Sort = 2
	contents[0].Content = a_text
	contents[0].ContentType = "markdown"
	contents[1].Key = "下周工作计划"
	contents[1].Type = 1
	contents[1].Sort = 6
	contents[1].Content = b_text
	contents[1].ContentType = "markdown"
	var reportParam TDDCreateReportParam
	reportParam.UserID = userid
	reportParam.TemplateID = template_id
	reportParam.Contents = contents[:]
	reportParam.ToUserIDs = strings.Split(to_userids, ",")
	reportParam.ToChat = false
	reportParam.DDFrom = "report"
	var dingResp struct {
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
		Result    string `json:"result"`
		RequestID string `json:"request_id"`
	}
	reqData := make(map[string]any)
	reqData["create_report_param"] = reportParam
	// 创建日志同样必须走实例级请求封装，才能确保代理 transport 被实际使用。
	req := Self.newPost(ddurl)
	if _, err := req.JSONBody(reqData); err != nil {
		return "", err
	}
	if err := req.ToJSON(&dingResp); err != nil {
		return "", err
	}
	statusCode, err := req.StatusCode()
	if err != nil {
		return "", err
	}
	if statusCode != 200 {
		return "", fmt.Errorf("http status %d", statusCode)
	}

	switch dingResp.ErrCode {
	case 0:
		return string(dingResp.Result), nil
	case 503:
		Self.token = nil
		return Self.CreateV2Report(userid, template_id, to_userids, a_text, b_text)
	default:
		return "", errors.New(dingResp.ErrMsg + ",RequestID: " + dingResp.RequestID)
	}
}
