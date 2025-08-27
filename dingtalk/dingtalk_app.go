package dingtalk

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

func (Self *TAccessToken) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

func (Self *TAccessToken) IsValid() bool {
	return Self.AccessToken != "" && int(time.Now().Unix()-Self.CreateDate.Unix()) <= Self.ExpiresIn
}

type workNotify struct {
	TResult
	TaskId int `json:"task_id"`
}

func (Self *workNotify) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

type TDeptInfo struct {
	TResult
	Id      int    `json:"id"`
	PId     int    `json:"parentid"`
	Name    string `json:"name"`
	MUserId string `json:"deptManagerUseridList"` //部门的主管列表，取值为由主管的userid组成的字符串，不同的userid使用“\|”符号进行分割
	IsSub   bool   `json:"groupContainSubDept"`   //部门群是否包含子部门
}

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

func (Self *TDDUser) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

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
func (Self *TDingTalkApp) SetAgentId(agent_id string) {
	Self.agent_id = agent_id
}

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

// https://oapi.dingtalk.com/user/get?access_token=ACCESS_TOKEN&userid=zhangsan
func (Self *TDingTalkApp) GetUserInfo(userid string) (*TDDUser, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	//首先通过免authcode登授权码,获取当前登录userid
	req := network.HttpGet(Self.ddurl+"/user/get").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("userid", userid)

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

// https://oapi.dingtalk.com/department/get?access_token=ACCESS_TOKEN&id=123
// 获取部门详情
func (Self *TDingTalkApp) GetDepartment(depid int) (*TDeptInfo, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return nil, err
	}

	//首先通过免authcode登授权码,获取当前登录userid
	req := network.HttpGet(Self.ddurl+"/department/get").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("id", strconv.Itoa(depid))

	var info TDeptInfo
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

func (Self *TDingTalkApp) GetOrgName(depids []int) (string, error) {
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

func (Self *TDingTalkApp) GetJobName(depids []int) (string, error) {
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

func (Self *TDingTalkApp) GetFullDepartmentName(depid int) (string, error) {
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
func (Self *TDingTalkApp) GetLoginInfo(authcode string) (string, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return "", err
	}

	//首先通过authcode免登授权码,获取当前登录userid
	req := network.HttpGet(Self.ddurl+"/user/getuserinfo").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("code", authcode)

	var info struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		UserId  string `json:"userid"`
	}
	err = req.ToJSON(&info)
	if err != nil {
		return "", err
	}
	if info.ErrCode == 0 {
		return info.UserId, nil
	} else if info.ErrCode == 503 {
		Self.token = nil
		return Self.GetLoginInfo(authcode)
	} else {
		return "", errors.New(info.ErrMsg)
	}
}

func (Self *TDingTalkApp) SendWorkNotify(user_id string, msg_text string) (int, error) {
	_, err := Self.GetAccessToken()
	if err != nil {
		return -1, err
	}

	req := network.HttpPost(Self.ddurl+"/topapi/message/corpconversation/asyncsend_v2").SetTimeout(Self.timeout_connect, Self.timeout_readwrite)
	req.Param("access_token", Self.token.AccessToken)
	req.Param("agent_id", Self.agent_id)
	req.Param("userid_list", user_id)
	req.Param("msg", msg_text)

	body_text, err := req.String()
	if err != nil {
		return -1, fmt.Errorf("发送钉钉工作通知消息失败，原因：%s", err.Error())
	}

	var info workNotify
	err = req.ToJSON(&info)
	if err != nil {
		return -1, fmt.Errorf("发送钉钉工作通知消息失败，内容：%s，原因：%s", body_text, err.Error())
	}
	if info.ErrCode == 0 {
		logs.Debug("SendWorkNotify() : 发送钉钉工作通知消息成功(%d)", info.TaskId)
		return info.TaskId, nil
	} else if info.ErrCode == 503 {
		Self.token = nil
		return Self.SendWorkNotify(user_id, msg_text)
	} else {
		return -1, errors.New(info.ErrMsg)
	}
}

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
		logs.Debug("GetAccessToken() : 获取钉钉Token信息 ...... Not OK %s(%d)", info.ErrMsg, info.ErrCode)
		return "", fmt.Errorf("%s(%d)", info.ErrMsg, info.ErrCode)
	}
}
