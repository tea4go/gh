package dingtalk

import (
	"os"
	"strings"
	"testing"
	"time"

	logs "github.com/tea4go/gh/log4go"
)

var app *TDingTalkApp

func TestMain(m *testing.M) {
	logs.SetLogger("console", `{"color":true,"level":7}`)
	clientID := os.Getenv("DINGTALK_Client_ID")
	clientSecret := os.Getenv("DINGTALK_Client_Secret")
	app = GetDingTalkApp(clientID, clientSecret, "6150")
	m.Run()
}

func TestAccessToken_IsValid(t *testing.T) {
	tk := &TAccessToken{AccessToken: "abc", ExpiresIn: 10, CreateDate: time.Now().Add(-5 * time.Second)}
	if !tk.IsValid() {
		t.Fatalf("expected token to be valid")
	}
}

func TestGetAccessToken(t *testing.T) {
	token, err := app.GetAccessToken()
	if err != nil {
		t.Fatalf("GetAccessToken error: %v", err)
	}
	if token != "testtoken" {
		t.Fatalf("unexpected token: %s", token)
	}
	if app.token == nil || app.token.AccessToken != "testtoken" {
		t.Fatalf("app.token not set correctly")
	}
}

func TestGetUserInfo(t *testing.T) {
	user, err := app.GetUserInfo("u1")
	if err != nil {
		t.Fatalf("GetUserInfo error: %v", err)
	}
	if user.UserId != "u1" || user.StaffName != "Alice" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestGetDepartment(t *testing.T) {
	dep, err := app.GetDepartment(2)
	if err != nil {
		t.Fatalf("GetDepartment error: %v", err)
	}
	if dep.Id != 2 || dep.PId != 1 || dep.Name != "Team" {
		t.Fatalf("unexpected department: %+v", dep)
	}
}

func TestGetDeptUsers(t *testing.T) {
	users, err := app.GetDeptUsers(1)
	if err != nil {
		t.Fatalf("GetDeptUsers error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestGetLoginInfo(t *testing.T) {
	userid, err := app.GetLoginInfo("authcode123")
	if err != nil {
		t.Fatalf("GetLoginInfo error: %v", err)
	}
	if userid != "u1" {
		t.Fatalf("unexpected userid: %s", userid)
	}
}

func TestSendWorkNotify(t *testing.T) {
	taskId, err := app.SendWorkNotify("u1", `{"msgtype":"text","text":{"content":"hello"}}`)
	if err != nil {
		t.Fatalf("SendWorkNotify error: %v", err)
	}
	if taskId != 42 {
		t.Fatalf("unexpected task id: %d", taskId)
	}
}

func TestGetOrgName(t *testing.T) {
	name, err := app.GetOrgName([]int{3, 4, 6})
	if err != nil {
		t.Fatalf("GetOrgName error: %v", err)
	}
	parts := strings.Split(name, "|")
	if len(parts) != 2 || !strings.Contains(name, "HRBP East") || !strings.Contains(name, "HRBP West") {
		t.Fatalf("unexpected org name: %s", name)
	}
}

func TestGetJobName(t *testing.T) {
	name, err := app.GetJobName([]int{3, 5, 6})
	if err != nil {
		t.Fatalf("GetJobName error: %v", err)
	}
	if name != "Engineering" {
		t.Fatalf("unexpected job name: %s", name)
	}
}

func TestGetFullDepartmentName(t *testing.T) {
	full, err := app.GetFullDepartmentName(2)
	if err != nil {
		t.Fatalf("GetFullDepartmentName error: %v", err)
	}
	if full != "//Team" {
		t.Fatalf("unexpected full name: %s", full)
	}
}
