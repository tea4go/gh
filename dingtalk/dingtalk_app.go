package dingtalk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/gh/network"
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
	MUserIds    []string `json:"dept_manager_userid_list"` //部门的主管列表，取值为由主管的userid组成的字符串，不同的userid使用“\|”符号进行分割
	OwnerUserId string   `json:"org_dept_owner"`           //企业群群主
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

type TDDV2User struct {
	UserId    string `json:"userid"`
	StaffCode string `json:"job_number"`
	StaffName string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"mobile"`
	Remark    string `json:"remark"`
	Avatar    string `json:"avatar"`
	Attrs     TDDV2UserAttr
	AttrText  string `json:"extension,omitempty"`
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

type TDingTalkApp struct {
	ddurl             string
	appkey            string
	appsecret         string
	agent_id          string
	token             *TAccessToken
	timeout_connect   time.Duration
	timeout_readwrite time.Duration
}

// GetDingTalkApp 获取钉钉 App 实例
func GetDingTalkApp(appkey, appsecret string, agent_id string) *TDingTalkApp {
	return &TDingTalkApp{
		appkey:            appkey,
		appsecret:         appsecret,
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

// GetAdmins 获取管理员列表
// {"sys_level":2,"userid":"userid2"},
// https://oapi.dingtalk.com/user/get_admin?access_token=ACCESS_TOKEN
func (Self *TDingTalkApp) GetAdmins() (*TAdmins, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	//首先通过免authcode登授权码,获取当前登录userid
	req := network.HttpGet(Self.ddurl+"/user/get_admin").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
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
func (Self *TDingTalkApp) GetV2UserInfoByPhone(phone string) (string, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return "", err
	}

	ddurl := Self.ddurl + "/topapi/v2/user/getbymobile"
	req := network.HttpGet(ddurl).SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("mobile", phone)
	logs.Debug("访问接口：%s (获取用户标识)", ddurl)

	var dingResp TDingTalkResponse
	err = req.ToJSON(&dingResp)
	if err != nil {
		return "", err
	}

	switch dingResp.ErrCode {
	case 0:
		logs.Debug("返回数据：%s", string(dingResp.Result))
		var info TDDV2User
		err = json.Unmarshal(dingResp.Result, &info)
		if err != nil {
			return "", err
		}
		return info.UserId, nil
	case 503:
		Self.token = nil
		return Self.GetV2UserInfoByPhone(phone)
	default:
		return "", errors.New(dingResp.ErrMsg)
	}
}

// GetUserInfoByPhone 根据手机号获取用户信息
// https://oapi.dingtalk.com/user/get_by_mobile?access_token=ACCESS_TOKEN&mobile=1xxxxxxxxxx
func (Self *TDingTalkApp) GetUserInfoByPhone(phone string) (string, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return "", err
	}

	//首先通过免authcode登授权码,获取当前登录userid
	req := network.HttpGet(Self.ddurl+"/user/get_by_mobile").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("mobile", phone)

	var info TDDUser
	err = req.ToJSON(&info)
	if err != nil {
		return "", err
	}
	if info.ErrCode == 0 {
		return info.UserId, nil
	} else if info.ErrCode == 503 {
		Self.token = nil
		return Self.GetUserInfoByPhone(phone)
	} else {
		return "", errors.New(info.ErrMsg)
	}
}

// GetUserInfoByUnionId 根据 UnionId 获取用户信息
// https://oapi.dingtalk.com/user/get?access_token=ACCESS_TOKEN&userid=zhangsan
func (Self *TDingTalkApp) GetV2UserInfoByUnionId(unionid string) (*TDDUser, error) {
	logs.Debug("GetUserInfoByUnionId() : 获取钉钉用户信息")
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	req := network.HttpGet(Self.ddurl+"/user/getUseridByUnionid?").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("unionid", unionid)

	body_text, err := req.String()
	if err != nil {
		return nil, fmt.Errorf("发送请求失败，原因：%s", err.Error())
	} //{"errcode":0,"errmsg":"ok","contactType":0,"userid":"333"}

	info := struct {
		TResult
		ContactType int    `json:"contactType"`
		UserId      string `json:"userid"`
	}{}
	err = req.ToJSON(&info)
	if err != nil {
		return nil, fmt.Errorf("序列化接收消息失败，内容：%s，原因：%s", body_text, err.Error())
	}

	if info.ErrCode == 42001 {
		Self.token = nil
		return Self.GetUserInfoByUnionId(unionid)
	} else if info.ErrCode != 0 {
		return nil, errors.New(info.ErrMsg)
	}

	return Self.GetUserInfo(info.UserId)
}

// GetUserInfoByUnionId 根据 UnionId 获取用户信息
// https://oapi.dingtalk.com/user/get?access_token=ACCESS_TOKEN&userid=zhangsan
func (Self *TDingTalkApp) GetUserInfoByUnionId(unionid string) (*TDDUser, error) {
	logs.Debug("GetUserInfoByUnionId() : 获取钉钉用户信息")
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	req := network.HttpGet(Self.ddurl+"/user/getUseridByUnionid?").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("unionid", unionid)

	body_text, err := req.String()
	if err != nil {
		return nil, fmt.Errorf("发送请求失败，原因：%s", err.Error())
	} //{"errcode":0,"errmsg":"ok","contactType":0,"userid":"333"}

	info := struct {
		TResult
		ContactType int    `json:"contactType"`
		UserId      string `json:"userid"`
	}{}
	err = req.ToJSON(&info)
	if err != nil {
		return nil, fmt.Errorf("序列化接收消息失败，内容：%s，原因：%s", body_text, err.Error())
	}

	if info.ErrCode == 42001 {
		Self.token = nil
		return Self.GetUserInfoByUnionId(unionid)
	} else if info.ErrCode != 0 {
		return nil, errors.New(info.ErrMsg)
	}

	return Self.GetUserInfo(info.UserId)
}

// GetUserInfo 根据 UserID 获取用户信息
// https://oapi.dingtalk.com/user/get?access_token=ACCESS_TOKEN&userid=zhangsan
func (Self *TDingTalkApp) GetUserInfo(userid string) (*TDDUser, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	ddurl := Self.ddurl + "/user/get"
	req := network.HttpGet(ddurl).SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("userid", userid)
	logs.Debug("访问接口：%s (获取用户详情)", ddurl)

	var info TDDUser
	err = req.ToJSON(&info)
	if err != nil {
		return nil, errors.New("获取钉钉用户信息失败！")
	}

	if info.ErrCode == 0 {
		return &info, nil
	} else if info.ErrCode == 503 {
		Self.token = nil
		return Self.GetUserInfo(userid)
	} else {
		return nil, errors.New(info.ErrMsg)
	}
}

// GetUserInfo 根据 UserID 获取用户信息
// https://oapi.dingtalk.com/user/get?access_token=ACCESS_TOKEN&userid=zhangsan
func (Self *TDingTalkApp) GetV2UserInfo(userid string) (*TDDV2User, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	ddurl := Self.ddurl + "/topapi/v2/user/get"
	req := network.HttpGet(ddurl).SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("userid", userid)
	logs.Debug("访问接口：%s (获取用户详情)", ddurl)

	var dingResp TDingTalkResponse
	err = req.ToJSON(&dingResp)
	if err != nil {
		return nil, err
	}
	logs.Debug("返回数据：%s", string(dingResp.Result))

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
		return &info, nil
	case 503:
		Self.token = nil
		return Self.GetV2UserInfo(userid)
	default:
		return nil, errors.New(dingResp.ErrMsg)
	}
}

// GetDepartment 获取部门详情
// https://oapi.dingtalk.com/department/get?access_token=ACCESS_TOKEN&id=123
func (Self *TDingTalkApp) GetV2Department(depid int) (*TDeptInfo, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	ddurl := Self.ddurl + "/topapi/v2/department/get"
	req := network.HttpGet(ddurl).SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
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
		logs.Debug("返回数据：%s", string(dingResp.Result))
		var info TDeptInfo
		err = json.Unmarshal(dingResp.Result, &info)
		if err != nil {
			return nil, err
		}
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
	req := network.HttpGet(ddurl).SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("dept_id", strconv.Itoa(depid))
	logs.Debug("访问接口：%s (获取部门用户)", ddurl)

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
		for _, v := range info.UserIDList {
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
			return name + "/" + info.Name, nil
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
	req := network.HttpGet(ddurl).SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
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
	req := network.HttpPost(ddurl).SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
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

// GetAccessToken 获取 AccessToken
// https://oapi.dingtalk.com/gettoken?appkey=key&appsecret=secret
// {"errorCode":503,"success":false,"errorMsg":"不合法的access_token"}
func (Self *TDingTalkApp) GetAccessToken() (string, error) {
	if Self.token != nil && Self.token.IsValid() {
		return Self.token.AccessToken, nil
	}

	req := network.HttpGet(Self.ddurl+"/gettoken").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("appkey", Self.appkey)
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
		//logs.Debug("GetAccessToken() : 获取钉钉Token信息 ...... OK (%s)", info.AccessToken)
		return info.AccessToken, nil
	} else {
		logs.Debug("GetAccessToken() : 获取钉钉Token信息失败，%s(%d)", info.ErrMsg, info.ErrCode)
		return "", fmt.Errorf("%s(%d)", info.ErrMsg, info.ErrCode)
	}
}
