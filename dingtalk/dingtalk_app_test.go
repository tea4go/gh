package dingtalk

import (
	"os"
	"strings"
	"testing"

	logs "github.com/tea4go/gh/log4go"
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
	_, err := app.GetAccessToken()
	if err != nil {
		t.Fatalf("获取 AccessToken 出错: %v", err)
	}
}

func TestGetUserInfo(t *testing.T) {
	user, err := app.GetUserInfo("201")
	if err != nil {
		t.Fatalf("获取用户信息出错: %v", err)
	}
	t.Logf("用户信息: %+v", user)
}

func TestGetDepartment(t *testing.T) {
	dep, err := app.GetDepartment(919894208)
	if err != nil {
		t.Fatalf("获取部门信息出错: %v", err)
	}
	t.Logf("部门信息: %+v", dep.Name)
}

func TestGetLoginInfo(t *testing.T) {
	_, err := app.GetLoginInfo("authcode123")
	if err != nil && !strings.Contains(err.Error(), "不存在的临时授权码") {
		t.Fatalf("获取登录信息出错: %v", err)
	}
}

func TestSendWorkNotify(t *testing.T) {
	taskId, err := app.SendWorkNotify("201", `{"msgtype":"text","text":{"content":"hello"}}`)
	if err != nil {
		t.Fatalf("发送工作通知出错: %v", err)
	}
	if taskId != 42 {
		t.Fatalf("任务ID不符合预期: %d", taskId)
	}
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
	users, err := app.GetDeptUsers(920096052)
	if err != nil {
		t.Fatalf("获取部门用户出错: %v", err)
	}
	t.Logf("部门用户: %+v", users)

}
