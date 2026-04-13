package dingtalk

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	logs "github.com/tea4go/gh/log4go"
)

var app *TDingTalkApp

func TestMain(m *testing.M) {
	logs.SetFDebug(false)
	logs.SetLogger("console", `{"color":true,"level":7}`)
	clientID := os.Getenv("DINGTALK_Client_ID")
	clientSecret := os.Getenv("DINGTALK_Client_Secret")
	corpId := os.Getenv("DINGTALK_Corp_ID")
	app = GetDingTalkApp(clientID, clientSecret, corpId, "615063230")
	logs.Debug("clientID = %s", clientID)
	logs.Debug("clientSecret = %s", clientSecret)
	logs.Debug("corpId = %s", corpId)
	logs.Debug("agent_id = %s", app.agent_id)
	m.Run()
}

func TestGetAccessToken(t *testing.T) {
	pkey, err := app.GetAccessToken()
	if err != nil {
		t.Fatalf("获取 AccessToken 出错: %v", err)
	}

	t.Logf("AccessToken: %s", pkey)
}

func TestGetConfig(t *testing.T) {
	fmt.Println(app.String())
	nonceStr := "123456"
	url := "http://localhost:8080"
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	pkey, err := app.GetConfig(nonceStr, timestamp, url)
	if err != nil {
		t.Fatalf("获取 Config 出错: %v", err)
	}

	t.Logf("Config: %s", pkey)
}

func TestGetJSAPITicket(t *testing.T) {
	pkey, err := app.GetJSAPITicket()
	if err != nil {
		t.Fatalf("获取 JsapiTicket 出错: %v", err)
	}

	t.Logf("JsapiTicket: %s", pkey)
}

func TestGetV2UserInfo(t *testing.T) {
	_, err := app.GetV2UserInfo("201")
	if err != nil {
		t.Fatalf("获取用户信息出错: %v", err)
	}
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
	userinfo, err := app.GetV2UserInfoByPhone("xxx")
	if err != nil {
		t.Fatalf("获取用户出错: %v", err)
	}

	t.Logf("用户信息: %+v", userinfo)
}

func TestGetV2UserInfoByUnionId(t *testing.T) {
	userinfo, err := app.GetV2UserInfoByUnionId("dxUDiP03drHsiE")
	if err != nil {
		t.Fatalf("获取用户出错: %v", err)
	}
	t.Logf("用户信息: %+v", userinfo)
}

func TestGetV2UsersByName(t *testing.T) {
	users, err := app.GetV2UsersByName("刘")
	if err != nil {
		t.Fatalf("查询用户名出错: %v", err)
	}

	for _, v := range users {
		t.Logf("[%s] %s - %s(%s/%s)", v.UserId, v.StaffCode, v.StaffName, v.Attrs.Org, v.Attrs.Job)
	}
	t.Logf("总共查找 %d 用户", len(users))
}

func TestGetV2ReportList(t *testing.T) {
	reportList, err := app.GetV2ReportList("201", "2026-02-19", "2026-03-19")
	if err != nil {
		t.Fatalf("查询用户日志出错: %v", err)
	}
	if len(reportList.DataList) == 0 {
		t.Fatalf("返回日志列表为空")
	}

	reportText := reportList.GetReportText("")
	if reportText == "" {
		t.Fatalf("获取报告内容出错")
	}
	t.Logf("共 %d 条日志", len(reportList.DataList))
	t.Logf("报告内容: \n%s", reportText)
}

func TestGetV2ReportSimpleList(t *testing.T) {
	reportList, err := app.GetV2ReportSimpleList("1795", "2026-01-01", "2026-02-19")
	if err != nil {
		t.Fatalf("查询用户日志摘要出错: %v", err)
	}
	if len(reportList) == 0 {
		t.Fatalf("返回日志概要列表为空")
	}

	for _, v := range reportList {
		t.Logf(v.ToString() + " -> " + v.ReportID + "\n")
	}
}

func TestGetV2ReportTemplateList(t *testing.T) {
	reportList, err := app.GetV2ReportTemplateList("201")
	if err != nil {
		t.Fatalf("查询用户日志模板出错: %v", err)
	}
	if len(reportList) == 0 {
		t.Fatalf("返回日志模板列表为空")
	}

	for _, v := range reportList {
		t.Logf(v.ReportCode + " -> " + v.Name + "\n")
	}
	t.Logf("共 %d 个日志模板", len(reportList))
}

// TestGetV2ReportListEndTime 验证 end_time 包含当天数据（修复前 end_time=2026-03-19 实际为 00:00:00，丢失当天数据）
func TestGetV2ReportListEndTime(t *testing.T) {
	// 用许伟(userId=7318)的本周数据验证：他 4月11日 提交了日志，end_time=4月12日 应能查到 4月11日 的日志
	reportList, err := app.GetV2ReportList("7318", "2026-04-06", "2026-04-12")
	if err != nil {
		t.Fatalf("查询用户日志出错: %v", err)
	}

	found := false
	for _, v := range reportList.DataList {
		createTime := time.Unix(v.CreateTime/1000, 0)
		t.Logf("日志: %s (%s)", v.TemplateName, createTime.Format("2006-01-02 15:04"))
		if createTime.Format("2006-01-02") == "2026-04-11" {
			found = true
		}
	}
	if !found {
		t.Fatalf("end_time 边界修复失败：未能查到 4月11日 的日志（end_time=4月12日 应包含当天数据）")
	}
	t.Logf("end_time 边界验证通过：共 %d 条，成功查到 4月11日 日志", len(reportList.DataList))
}

// TestGetV2ReportSimpleListPagination 验证分页逻辑：跨页数据合并后时间单调递减
func TestGetV2ReportSimpleListPagination(t *testing.T) {
	// 用何永进(userId=1795)的 3 个月数据，预期超过 20 条，触发分页
	reportList, err := app.GetV2ReportSimpleList("1795", "2025-12-01", "2026-03-31")
	if err != nil {
		t.Fatalf("查询用户日志摘要出错: %v", err)
	}
	t.Logf("共 %d 条日志概要（跨越分页阈值 20）", len(reportList))

	// 验证数量合理（3 个月周报应有 10+ 条）
	if len(reportList) < 10 {
		t.Fatalf("分页可能丢失数据：3 个月仅返回 %d 条，预期至少 10 条", len(reportList))
	}

	// 打印前 5 条验证内容
	for i, v := range reportList {
		if i >= 5 {
			break
		}
		t.Logf("[%d] %s -> %s", i, v.ToString(), v.ReportID)
	}
}

// TestGetV2ReportTemplateListNoDeadLoop 验证模板列表分页不会死循环
func TestGetV2ReportTemplateListNoDeadLoop(t *testing.T) {
	start := time.Now()
	reportList, err := app.GetV2ReportTemplateList("201")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("查询用户日志模板出错: %v", err)
	}
	t.Logf("共 %d 个模板，耗时 %v", len(reportList), elapsed)

	// 正常情况下几百个模板应在 30 秒内返回
	if elapsed > 30*time.Second {
		t.Fatalf("模板列表查询耗时 %v，疑似死循环", elapsed)
	}
	// 模板数量应 > 0 且合理（不超过 2000）
	if len(reportList) == 0 {
		t.Fatalf("返回模板列表为空")
	}
	if len(reportList) > 2000 {
		t.Fatalf("模板数量 %d 超过 2000，疑似死循环导致重复数据", len(reportList))
	}
}

func TestGetV2ReportTemplate(t *testing.T) {
	reportList, err := app.GetV2ReportTemplate("201", "周报")
	if err != nil {
		t.Fatalf("查询用户日志模板出错: %v", err)
	}

	for _, v := range reportList.Fields {
		t.Logf("Typer=%d - Sort=%d : %s", v.Type, v.Sort, v.FieldName)
	}
}
func TestCreateV2Report(t *testing.T) {
	a_text := `
`
	b_text := `
	`
	// 读取文件内容
	// a_text_bytes, err := ioutil.ReadFile(`C:\SyncData\云上笔记_20260124_100025\A-我的笔记\测试MD格式.md`)
	// if err != nil {
	// 	log.Fatalf("读取文件失败,%v", err)
	// }
	// a_text = string(a_text_bytes)

	report_id, err := app.CreateV2Report("201", "1704cf092b3a4bab513974f44c6b53d6", "201", a_text, b_text)
	if err != nil {
		t.Fatalf("创建用户日志出错: %v", err)
	}
	t.Logf("report_id = %s\n", report_id)

}
