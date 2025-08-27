package weixin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/gh/network"
)

//curl 'https://oapi.dingtalk.com/robot/send?access_token=xxx' \
//   -H 'Content-Type: application/json' \
//   -d '
//  {"msgtype": "text",
//    "text": {
//        "content": "我就是我, 是不一样的烟火"
//     }
//  }'

type TResult struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type TAccessToken struct {
	TResult
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	CreateDate  time.Time
}

func (Self *TAccessToken) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

func (Self *TAccessToken) IsValid() bool {
	return Self.AccessToken != "" && int(time.Now().Unix()-Self.CreateDate.Unix()) <= Self.ExpiresIn
}

type TLoginInfo struct {
	TResult
	UserId string `json:"userid"`
}

func (Self *TLoginInfo) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

type TWorkNotify struct {
	TResult
	TaskId int `json:"task_id"`
}

func (Self *TWorkNotify) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

type TUserInfo struct {
	TResult
	UUID        string `json:"unionid"`
	UserId      string `json:"userid"`
	StaffCode   string `json:"jobnumber"`
	StaffName   string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"mobile"`
	Job         string `json:"job"`
	Org         string `json:"org"`
	Departments []int  `json:"department"` //成员所属部门id列表
	Position    string `json:"position"`   //职位信息
	Avatar      string `json:"avatar"`     //头像url
}

// func (Self *TUserInfo) String() string {
// 	return fmt.Sprintf("%s %s (%s)", Self.StaffCode, Self.StaffName, Self.Job)
// }

type TDepartmentInfo struct {
	TResult
	Id      int    `json:"id"`
	PId     int    `json:"parentid"`
	Name    string `json:"name"`
	MUserId string `json:"deptManagerUseridList"` //部门的主管列表，取值为由主管的userid组成的字符串，不同的userid使用“\|”符号进行分割
	IsSub   bool   `json:"groupContainSubDept"`   //部门群是否包含子部门
}

func (Self *TDepartmentInfo) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

type TAdmin struct {
	UserId   string `json:"userid"`
	SysLevel int    `json:"sys_level"`
}

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

func (Self *TAdmins) String() string {
	out := "管理员列表"
	for _, v := range Self.Admins {
		out = out + "\n" + v.String()
	}
	return out
}

type TDingTalkClient struct {
	ddurl             string
	appkey            string
	appsecret         string
	token             *TAccessToken
	timeout_connect   time.Duration
	timeout_readwrite time.Duration
}

func GetDingTalkClient(appkey, appsecret string) *TDingTalkClient {
	return &TDingTalkClient{
		appkey:            appkey,
		appsecret:         appsecret,
		ddurl:             `https://oapi.dingtalk.com`,
		timeout_connect:   30 * time.Second,
		timeout_readwrite: 30 * time.Second,
	}
}

// {"sys_level":2,"userid":"userid2"},
// https://oapi.dingtalk.com/user/get_admin?access_token=ACCESS_TOKEN
func (Self *TDingTalkClient) GetAdmins() (*TAdmins, error) {
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

// https://oapi.dingtalk.com/user/get_by_mobile?access_token=ACCESS_TOKEN&mobile=1xxxxxxxxxx
func (Self *TDingTalkClient) GetUserInfoByPhone(phone string) (*TUserInfo, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	//首先通过免authcode登授权码,获取当前登录userid
	req := network.HttpGet(Self.ddurl+"/user/get_by_mobile").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("mobile", phone)

	var info TUserInfo
	err = req.ToJSON(&info)
	if err != nil {
		return nil, err
	}
	if info.ErrCode == 0 {
		info.Job, _ = Self.GetJobName(info.Departments)
		info.Org, _ = Self.GetOrgName(info.Departments)
		return &info, nil
	} else if info.ErrCode == 503 {
		Self.token = nil
		return Self.GetUserInfoByPhone(phone)
	} else {
		return nil, errors.New(info.ErrMsg)
	}
}

// https://oapi.dingtalk.com/user/get?access_token=ACCESS_TOKEN&userid=zhangsan
func (Self *TDingTalkClient) GetUserInfo(userid string) (*TUserInfo, error) {
	logs.Debug("GetUserInfo() : 获取钉钉用户信息")
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	//首先通过免authcode登授权码,获取当前登录userid
	req := network.HttpGet(Self.ddurl+"/user/get").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("userid", userid)

	var info TUserInfo
	err = req.ToJSON(&info)
	if err != nil {
		return nil, errors.New("获取钉钉用户信息失败！")
	}
	if info.ErrCode == 0 {
		logs.Debug("GetLoginInfo() : 获取钉钉用户信息 ...... %s-%s", info.StaffCode, info.StaffName)
		//info.Job, _ = Self.GetJobName(info.Departments)
		//info.Org, _ = Self.GetOrgName(info.Departments)
		//logs.Debug("GetLoginInfo() : 获取钉钉用户信息 ...... %s-%s", info.Job, info.Org)
		return &info, nil
	} else if info.ErrCode == 503 {
		Self.token = nil
		return Self.GetUserInfo(userid)
	} else {
		return nil, errors.New(info.ErrMsg)
	}
}

// https://oapi.dingtalk.com/department/get?access_token=ACCESS_TOKEN&id=123
// 获取部门详情
func (Self *TDingTalkClient) GetDepartment(depid int) (*TDepartmentInfo, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	//首先通过免authcode登授权码,获取当前登录userid
	req := network.HttpGet(Self.ddurl+"/department/get").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("id", strconv.Itoa(depid))

	var info TDepartmentInfo
	err = req.ToJSON(&info)
	if err != nil {
		return nil, err
	}
	if info.ErrCode == 0 {
		return &info, nil
	} else if info.ErrCode == 503 {
		Self.token = nil
		return Self.GetDepartment(depid)
	} else {
		return nil, errors.New(info.ErrMsg)
	}
}

func (Self *TDingTalkClient) GetOrgName(depids []int) (string, error) {
	logs.Debug("GetJobName() : 获取钉钉部门信息")
	var name string
	for _, v := range depids {
		dep, err := Self.GetDepartment(v)
		if err != nil {
			return "", err
		} else {
			if strings.Contains(dep.Name, "HRBP") {
				if name == "" {
					name = dep.Name
				} else {
					name = name + "|" + dep.Name
				}
			}
		}
	}
	return name, nil
}

func (Self *TDingTalkClient) GetJobName(depids []int) (string, error) {
	logs.Debug("GetJobName() : 获取钉钉岗位信息")
	var name string
	for _, v := range depids {
		dep, err := Self.GetDepartment(v)
		if err != nil {
			return "", err
		} else {
			if !strings.Contains(dep.Name, "HRBP") && !strings.Contains(dep.Name, "考勤") {
				if name == "" {
					name = dep.Name
				} else {
					name = name + "|" + dep.Name
				}
			}
		}
	}
	return name, nil
}

func (Self *TDingTalkClient) GetFullDepartmentName(depid int) (string, error) {
	info, err := Self.GetDepartment(depid)
	if err != nil {
		return "", err
	} else {
		if info.PId > 0 {
			name, err := Self.GetFullDepartmentName(info.PId)
			if err != nil {
				return "", err
			}
			return name + "/" + info.Name, nil
		} else {
			return "/", nil
		}
	}
}

// https://oapi.dingtalk.com/user/getuserinfo?access_token=access_token&code=code
func (Self *TDingTalkClient) GetLoginInfo(authcode string) (*TLoginInfo, error) {
	logs.Debug("GetLoginInfo() : 获取钉钉登录信息")
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	//首先通过authcode免登授权码,获取当前登录userid
	req := network.HttpGet(Self.ddurl+"/user/getuserinfo").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("code", authcode)

	var info TLoginInfo
	err = req.ToJSON(&info)
	if err != nil {
		return nil, err
	}
	if info.ErrCode == 0 {
		logs.Debug("GetLoginInfo() : 获取钉钉登录信息 ...... %s", info.UserId)
		return &info, nil
	} else if info.ErrCode == 503 {
		Self.token = nil
		return Self.GetLoginInfo(authcode)
	} else {
		return nil, errors.New(info.ErrMsg)
	}
}

func (Self *TDingTalkClient) SendWorkNotify(msg string) (int, error) {
	logs.Debug("GetAccessToken() : 发送钉钉工作通知消息")
	_, err := Self.GetAccessToken()
	if err != nil {
		return -1, err
	}

	req := network.HttpPost(Self.ddurl+"/topapi/message/corpconversation/asyncsend_v2").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("agent_id", "615063230")
	req.Param("userid_list", "201")
	req.Param("msg", `{"msgtype":"text","text":{"content":"消息内容"}}`)

	var info TWorkNotify
	err = req.ToJSON(&info)
	if err != nil {
		return -1, errors.New("发送钉钉工作通知消息失败！")
	}
	if info.ErrCode == 0 {
		logs.Debug("GetLoginInfo() : 发送钉钉工作通知消息成功 ...... %s", info.TaskId)
		return info.TaskId, nil
	} else if info.ErrCode == 503 {
		Self.token = nil
		return Self.SendWorkNotify(msg)
	} else {
		return -1, errors.New(info.ErrMsg)
	}
}

// https://oapi.dingtalk.com/gettoken?appkey=key&appsecret=secret
// {"errorCode":503,"success":false,"errorMsg":"不合法的access_token"}
func (Self *TDingTalkClient) GetAccessToken() (string, error) {
	if Self.token != nil && Self.token.IsValid() {
		return Self.token.AccessToken, nil
	}
	logs.Debug("GetAccessToken() : 获取钉钉Token信息")

	req := network.HttpGet(Self.ddurl+"/gettoken").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
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
		logs.Debug("GetAccessToken() : 获取钉钉Token信息 ...... OK (%s)", info.AccessToken)
		return info.AccessToken, nil
	} else {
		logs.Debug("GetAccessToken() : 获取钉钉Token信息 ...... Not OK %s(%d)", info.ErrMsg, info.ErrCode)
		return "", fmt.Errorf("%s(%d)", info.ErrMsg, info.ErrCode))
	}
}
