package dingtalk

import (
	"os"
	"strings"
	"testing"

	logs "github.com/tea4go/gh/log4go"
	"gopkg.in/ffmt.v1"
)

var app *TDingTalkApp

func TestMain(m *testing.M) {
	logs.SetLogger("console", `{"color":true,"level":7}`)
	clientID := os.Getenv("DINGTALK_Client_ID")
	clientSecret := os.Getenv("DINGTALK_Client_Secret")
	app = GetDingTalkApp(clientID, clientSecret, "3867271950")
	m.Run()
}

func TestGetAccessToken(t *testing.T) {
	pkey, err := app.GetAccessToken()
	if err != nil {
		t.Fatalf("获取 AccessToken 出错: %v", err)
	}

	t.Logf("AccessToken: %s", pkey)
}

func TestGetV2UserInfo(t *testing.T) {
	user, err := app.GetV2UserInfo("201")
	if err != nil {
		t.Fatalf("获取用户信息出错: %v", err)
	}
	t.Logf("用户信息: %+v", user)
	ffmt.Puts(user)
}

func TestGetV2Department(t *testing.T) {
	dep, err := app.GetV2Department(919894208)
	if err != nil {
		t.Fatalf("获取部门信息出错: %v", err)
	}
	t.Logf("部门信息: %+v", dep)
}

func TestGetV2LoginInfo(t *testing.T) {
	_, err := app.GetV2LoginInfo("authcode123")

	if err != nil && !strings.Contains(err.Error(), "不存在的临时授权码") {
		t.Fatalf("获取登录信息出错: %v", err)
	} else {
		t.Logf("本测试函数通过")
	}
}

func TestSendWorkNotify(t *testing.T) {
	taskId, err := app.SendWorkNotify("201", `{"msgtype":"text","text":{"content":"hello 1234"}}`)
	if err != nil {
		t.Fatalf("发送工作通知出错: %v", err)
	}
	t.Logf("任务ID: %d", taskId)
}

func TestGetOrgName(t *testing.T) {
	name, err := app.GetOrgName("201")
	if err != nil {
		t.Fatalf("获取组织名称出错: %v", err)
	}
	t.Logf("组织名称: %s", name)
}

func TestGetJobName(t *testing.T) {
	name, err := app.GetJobName("201")
	if err != nil {
		t.Fatalf("获取职位名称出错: %v", err)
	}
	t.Logf("职位名称: %s", name)
}

func TestGetFullDeptName(t *testing.T) {
	full, err := app.GetFullDeptName(919894208)
	if err != nil {
		t.Fatalf("获取完整部门名称出错: %v", err)
	}
	t.Logf("完整部门名称: %s", full)
}

func TestGetDeptUsers(t *testing.T) {
	users, err := app.GetDeptUsers(919894208)
	if err != nil {
		t.Fatalf("获取部门用户出错: %v", err)
	}
	//t.Logf("部门用户: %+v", users)
	for _, v := range users {
		t.Logf("%s - %s", v.StaffCode, v.StaffName)
	}
}

func TestGetV2UserInfoByPhone(t *testing.T) {
	userinfo, err := app.GetV2UserInfoByPhone("13016985150")
	if err != nil {
		t.Fatalf("获取用户标识出错: %v", err)
	}
	t.Logf("用户信息: %+v", userinfo)
}

func TestGetV2UserInfoByUnionId(t *testing.T) {
	userinfo, err := app.GetV2UserInfoByUnionId("dxUDiP03drHsiE")
	if err != nil {
		t.Fatalf("获取用户标识出错: %v", err)
	}
	t.Logf("用户信息: %+v", userinfo)
}
