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

func TestGetConfig(t *testing.T) {
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
	for i := 0; i < 300; i++ {
		_, err := app.GetV2UserInfo("201")
		if err != nil {
			t.Fatalf("获取用户信息出错: %v", err)
		}
	}
	//	t.Logf("用户信息: %+v", user)
	//	ffmt.Puts(user)
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

	reportText := reportList.GetReportText("")
	if reportText == "" {
		t.Fatalf("获取报告内容出错")
	}
	t.Logf("报告内容: \n%s", reportText)
}

func TestGetV2ReportSimpleList(t *testing.T) {
	reportList, err := app.GetV2ReportSimpleList("1795", "2026-01-01", "2026-02-19")
	if err != nil {
		t.Fatalf("查询用户日志摘要出错: %v", err)
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

	for _, v := range reportList {
		t.Logf(v.ReportCode + " -> " + v.Name + "\n")
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
	a_text := `售前项目
1）商机支撑
- a、新电New Core：正在售前评估阶段，已对Redhat官方发起环境适配性验证申请，其目前正在进行审批流程，通过后开展验证；
- b、津巴布韦Ecoent AA项目：本周于2月2日完成割接上线，平台与GTSC保持高度响应持续保障；
- c、印尼XL AI项目：本周于2月2日完成客户演示，客户反馈很满意，5月份完成版本开发工作；
- d、荷兰KPN Spring升级：平台完成升级所需的安全治理工作，计划3月底之前完成ZCM/SGP发版。
2）售前应标
- a、Globe EAM RFP应标：平台参与应标启动会，初步评估约70%的定制开发量，投标截止为菲律宾时间2月24日。
3）硬件配置与部署
- a、硬件配置&amp;部署方
案：内部进一步梳理标准SOP以及响应案例，后续会进一步保持内外拉通同步。
4）方案培训
- a、法电CRM项目：支撑商发完成Scalability能力演示；
- b、格鲁Silknet AIOps：线上会议的形式给客户介绍Agent方案。
5）竞品分析
- a、启动AWS产品分析工作，正在分析AWS EKS用户指南与ZCM做产品特性之间的差异性对比，形成标准报告预计3月份完成。
产品质量：
- a、集成测试：推进产品自动化覆盖，容器云70%/监控75%/调用链70%/日志70%/zmq70%;其中AI自动化所有产品覆盖率16%
- b、高可用测试：围绕K8S、mysql、redis、业务网关等平台高可用架构展开测试，容器云进展10%，监控、调用链、日志均进展5%
- c、稳定性建设：本周实战故障演练100%触发告警，发现并修复监控产品influxdbutil获取bean顺序异常的稳定性隐患
- d、AI测试：顺利完成印尼XL AI场景客户演示，
受控到正式版本回归通过率10%;AI trace测试进展10%
售后项目：
- a、津巴布韦Econet项目2.1成功上线
- b、新电AKS 1.29应用无法启动完成根因分析并提供解决方案
- c、刚果布MTN项目2.8保障上线

`
	b_text := `- 1. Core94-Security 分支升级SpringBoot4.0
- 2. 可观测2.0 大盘中心和自运维功能设计
- 3. 调用链异常采样功能性能压测和内存调优
- 4. 规范规划及发布计划拉通，推动ZMQ开发规范的发布
- 5. 鉴于 Singtel 现场发现的问题，修订调用链、日志接入和运维规范，增加对 APM_INIT URL 的使用规范
- 6. 面向项目发布技术通知单，排查License到期风险
- 7. 各项目需求研发和交付
- 8. 其他在途专题推进
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
