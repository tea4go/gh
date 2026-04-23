package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
	dingtalk "github.com/tea4go/gh/dingtalk"
	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/gh/network"
)

// 标准程序块
var appName string = "test_dingtalk_app"
var appVer string = "v0.0.4"
var IsBeta string
var BuildTime string

var app *dingtalk.TDingTalkApp
var clientID string
var clientSecret string
var corpId string
var agentId string

func initApp() {
	if clientID == "" || clientSecret == "" {
		fmt.Println("请设置环境变量：")
		fmt.Println("  DINGTALK_Client_ID    - 应用的 AppKey")
		fmt.Println("  DINGTALK_Client_Secret - 应用的 AppSecret")
		fmt.Println("  DINGTALK_Corp_ID      - 企业 CorpId")
		fmt.Println("  DINGTALK_Agent_ID     - 应用 AgentId")
		fmt.Println("或通过命令行参数指定：")
		fmt.Println("  --client-id / --client-secret / --corp-id / --agent-id")
		os.Exit(1)
	}

	// 标准程序块
	network.SetAppVersion(appName, appVer, IsBeta, BuildTime) //设置应用版本号，便于自动更新
	logs.StartLogger()
	network.StartSelfUpdate("http://wc192.yj2025.icu:8118", "http://nj.yj2025.icu:23432", "http://wc8.yj2025.icu:8118", "http://wc47.yj2025.icu:23431")
	// 标准程序块

	app = dingtalk.GetDingTalkApp(clientID, clientSecret, corpId, agentId)

	proxyURL, _ := url.Parse("http://192.168.190.163:32121")
	httpTransport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	app.SetTransport(httpTransport)
}

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("JSON序列化失败: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func printReportDept(dept *dingtalk.TDDV2ReportDept) {
	fmt.Printf("【%s】共 %d 人\n", dept.DeptName, len(dept.Users))
	for _, child := range dept.SubDepts {
		fmt.Printf("  %s (%d)\n", child.Name, child.DeptId)
	}

	for _, u := range dept.Users {
		fmt.Printf("  %s, %-4s (%s)\n", u.StaffCode, u.StaffName, u.UserId)
	}
}

func usage() {
	fmt.Println(`钉钉接口测试工具

用法: test_dingtalk_app <命令> [参数]

命令列表:
  access-token                                  获取 AccessToken
  jsapi-ticket                                  获取 JSAPI Ticket
  config <nonceStr> <timestamp> <url>           获取前端鉴权配置
  admins                                        获取管理员列表
  user <userid>                                 根据userid获取用户详情
  user-by-phone <phone>                         查手机号获取用户详情
  user-by-unionid <unionid>                     根据unionid获取用户详情
  users-by-name <name>                          根据姓名搜索用户列表
  org-name <userid>                             获取用户所属组织名称
  job-name <userid>                             获取用户岗位名称
  dept <deptid>                                 获取部门详情
  dept-fullname <deptid>                        获取部门完整路径名称
  dept-users <deptid>                           获取部门下所有用户
  sub-depts <deptid>                            获取子部门列表
  report-users <userid>                         获取用户所属部门的员工列表（按子部门分组）
  report-templates <userid>                     获取用户可用的日志模板列表
  report-template <userid> <template_name>      获取指定日志模板详情
  report-list <userid> <start> <end>            获取用户日志列表（详情）
  report-simple-list <userid> <start> <end>     获取用户日志列表（简要）
  create-report <userid> <template_id> <to_ids> <text_a> <text_b> 创建日志
  send-notify <userid> <msg_json>               发送工作通知`)
}

func main() {
	clientID = os.Getenv("DINGTALK_Client_ID")
	clientSecret = os.Getenv("DINGTALK_Client_Secret")
	corpId = os.Getenv("DINGTALK_Corp_ID")
	agentId = os.Getenv("DINGTALK_Agent_ID")

	flag.StringVar(&clientID, "client-id", clientID, "DingTalk Client ID (AppKey)")
	flag.StringVar(&clientSecret, "client-secret", clientSecret, "DingTalk Client Secret (AppSecret)")
	flag.StringVar(&corpId, "corp-id", corpId, "DingTalk Corp ID")
	flag.StringVar(&agentId, "agent-id", agentId, "DingTalk Agent ID")

	flag.Parse()
	positionalArgs := flag.Args()
	if len(positionalArgs) < 1 {
		usage()
		return
	}

	initApp()

	cmd := positionalArgs[0]
	args := positionalArgs[1:]

	switch cmd {
	case "access-token": // access-token 获取 AccessToken
		cmdAccessToken()
	case "jsapi-ticket": // jsapi-ticket 获取 JSAPI Ticket
		cmdJSAPITicket()
	case "config": // config <nonceStr> <timestamp> <url> 获取前端鉴权配置
		requireArgs(cmd, args, 3)
		cmdConfig(args[0], args[1], args[2])
	case "admins": // admins 获取管理员列表
		cmdAdmins() /*OK*/
	case "user": // user <userid> 根据userid获取用户详情
		requireArgs(cmd, args, 1)
		cmdUserInfo(args[0]) /*OK*/
	case "user-by-phone": // user-by-phone <phone> 根据手机号获取用户详情
		requireArgs(cmd, args, 1)
		cmdUserInfoByPhone(args[0]) /*OK*/
	case "user-by-unionid": // user-by-unionid <unionid> 根据unionid获取用户详情
		requireArgs(cmd, args, 1)
		cmdUserInfoByUnionId(args[0]) /*OK*/
	case "users-by-name": // users-by-name <name> 根据姓名搜索用户列表
		requireArgs(cmd, args, 1)
		cmdUsersByName(args[0]) /*OK*/
	case "org-name": // org-name <userid> 获取用户所属组织名称
		requireArgs(cmd, args, 1)
		cmdOrgName(args[0]) /*OK*/
	case "job-name": // job-name <userid> 获取用户岗位名称
		requireArgs(cmd, args, 1)
		cmdJobName(args[0]) /*OK*/
	case "dept": // dept <deptid> 获取部门详情
		requireArgs(cmd, args, 1)
		cmdDept(parseIntArg(args[0])) /*OK*/
	case "dept-fullname": // dept-fullname <deptid> 获取部门完整路径名称
		requireArgs(cmd, args, 1)
		cmdDeptFullName(parseIntArg(args[0])) /*OK*/
	case "dept-users": // dept-users <deptid> 获取部门用户(直接下级)
		requireArgs(cmd, args, 1)
		cmdDeptUsers(parseIntArg(args[0])) /*OK*/
	case "sub-depts": // sub-depts <deptid> 获取子部门列表
		requireArgs(cmd, args, 1)
		cmdSubDepts(parseIntArg(args[0])) /*OK*/
	case "report-users": // <userid> 获取用户所属部门的员工列表（按子部门分组）
		requireArgs(cmd, args, 1)
		cmdReportUsers(args[0])
	case "report-templates": // report-templates <userid> 获取用户可用的日志模板列表
		requireArgs(cmd, args, 1)
		if len(args) >= 2 {
			cmdReportTemplates(args[0], args[1])
		} else {
			cmdReportTemplates(args[0])
		}
	case "report-template": // report-template <userid> <template_name> 获取指定日志模板详情
		requireArgs(cmd, args, 2)
		cmdReportTemplate(args[0], args[1])
	case "report-list": // report-list <userid> <start> <end> 获取用户日志列表（详情）
		requireArgs(cmd, args, 3)
		cmdReportList(args[0], args[1], args[2])
	case "report-simple-list-by-template": // report-simple-list-by-template-id <template_name> <start> <end> 获取用户日志列表（模板）
		requireArgs(cmd, args, 2)
		cmdReportSimpleListByTemplate(args[0], args[1])
	case "report-simple-list": // report-simple-list <userid> <start> <end> 获取用户日志列表（简要）
		requireArgs(cmd, args, 3)
		cmdReportSimpleList(args[0], args[1], args[2])
	case "create-report": // create-report <userid> <template_id> <to_ids> <text_a> <text_b> 创建日志
		requireArgs(cmd, args, 5)
		cmdCreateReport(args[0], args[1], args[2], args[3], args[4])
	case "send-notify": // send-notify <userid> <msg_json> 发送工作通知
		requireArgs(cmd, args, 2)
		cmdSendNotify(args[0], args[1])
	default:
		fmt.Printf("未知命令: %s\n\n", cmd)
		usage()
	}
}

func requireArgs(cmd string, args []string, count int) {
	if len(args) < count {
		fmt.Printf("命令 %s 需要 %d 个参数，实际 %d 个\n", cmd, count, len(args))
		os.Exit(1)
	}
}

func parseIntArg(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("参数 %s 不是有效的整数\n", s)
		os.Exit(1)
	}
	return n
}

// ============ 命令实现 ============

func cmdAccessToken() {
	token, err := app.GetAccessToken()
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("AccessToken: %s\n", token)
}

func cmdJSAPITicket() {
	ticket, err := app.GetJSAPITicket()
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("JsapiTicket: %s\n", ticket)
}

func cmdConfig(nonceStr, timestamp, url string) {
	config, err := app.GetConfig(nonceStr, timestamp, url)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("Config: %s\n", config)
}

func cmdAdmins() {
	admins, err := app.GetAdmins()
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("管理员列表 (%d 人):\n", len(admins.Admins))
	for _, a := range admins.Admins {
		if a.SysLevel == 1 {
			fmt.Printf("  userid = %s (主管理员)\n", a.UserId)
		} else {
			fmt.Printf("  userid = %s\n", a.UserId)
		}
	}
}

func cmdUserInfo(userid string) {
	user, err := app.GetV2UserInfo(userid)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	printJSON(user)
}

func cmdUserInfoByPhone(phone string) {
	user, err := app.GetV2UserInfoByPhone(phone)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	printJSON(user)
}

func cmdUserInfoByUnionId(unionid string) {
	user, err := app.GetV2UserInfoByUnionId(unionid)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	printJSON(user)
}

func cmdUsersByName(name string) {
	users, err := app.GetV2UsersByName(name)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("搜索结果 (%d 人):\n", len(users))
	for _, u := range users {
		fmt.Printf("%7s, %s, %s (%s/%s)\n", u.UserId, u.StaffCode, u.StaffName, u.Attrs.Org, u.Attrs.Job)
	}
}

func cmdOrgName(userid string) {
	name, err := app.GetOrgName(userid)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("组织名称: %s\n", name)
}

func cmdJobName(userid string) {
	name, err := app.GetJobName(userid)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("岗位名称: %s\n", name)
}

func cmdDept(depid int) {
	dept, err := app.GetV2Department(depid)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	printJSON(dept)
}

func cmdDeptFullName(depid int) {
	name, err := app.GetFullDeptName(depid)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("部门全称: %s\n", name)
}

func cmdDeptUsers(depid int) {
	dept, err := app.GetV2Department(depid)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf(":: 部门 %s (本级=%d，上级=%d)\n", dept.Name, dept.Id, dept.PId)
	users, err := app.GetDeptUsers(depid)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("部门用户 (%d 人):\n", len(users))
	for _, u := range users {
		fmt.Printf("  %s, %s, %s\n", u.UserId, u.StaffCode, u.StaffName)
	}
}

func cmdSubDepts(depid int) {
	dept, err := app.GetV2Department(depid)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf(":: 部门 %s (本级=%d，上级=%d)\n", dept.Name, dept.Id, dept.PId)
	depts, err := app.GetSubDepts(depid)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf(":: 下级 (%d 个):\n", len(depts))
	for _, d := range depts {
		fmt.Printf("  %s (%d)\n", d.Name, d.DeptId)
	}
}

func cmdReportUsers(userid string) {
	report, err := app.GetV2ReportUsers(userid)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	totalDepts, totalUsers := 0, 0
	for _, dept := range report.Departments {
		totalDepts++
		totalUsers += len(dept.Users)
		printReportDept(dept)
	}
	fmt.Printf("\n共 %d 个子团队，共 %d 名员工(包含自己)\n", totalDepts, totalUsers)
}

func cmdReportTemplates(userid string, keyword ...string) {
	templates, err := app.GetV2ReportTemplateList(userid)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	filterKeyword := ""
	if len(keyword) > 0 {
		filterKeyword = strings.TrimSpace(keyword[0])
	}

	filteredTemplates := templates
	if filterKeyword != "" {
		filteredTemplates = make([]dingtalk.TDDReportTemplateItem, 0, len(templates))
		filterLower := strings.ToLower(filterKeyword)
		for _, t := range templates {
			if strings.Contains(strings.ToLower(t.Name), filterLower) {
				filteredTemplates = append(filteredTemplates, t)
			}
		}
	}

	fmt.Printf("日志模板列表 (%d 个):\n", len(filteredTemplates))
	for k, t := range filteredTemplates {
		fmt.Printf("%3d %s - %s\n", k, t.ReportCode, t.Name)
	}
}

func cmdReportTemplate(userid, templateName string) {
	template, err := app.GetV2ReportTemplate(userid, templateName)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	printJSON(template)
}

func cmdReportList(userid, start, end string) {
	list, err := app.GetV2ReportList(userid, start, end)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("日志列表 (%d 条):\n", len(list.DataList))
	for i, item := range list.DataList {
		fmt.Printf("  [%d] %s\n", i+1, item.TemplateName)
	}
}

func cmdReportSimpleListByTemplate(templateName, start string) {
	items, err := app.GetV2ReportSimpleListByTemplate(templateName, start)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("简要日志列表 (%d 条):\n", len(items))
	for i, item := range items {
		fmt.Printf("  [%3d] %s - %s - %s (%s)\n", i+1, item.ReportID, time.Unix(item.CreateTime/1000, 0).Format("2006-01-02 15:04"), item.CreatorName, item.DeptName)
	}
}

func cmdReportSimpleList(userid, start, end string) {
	items, err := app.GetV2ReportSimpleList(userid, start, end)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("简要日志列表 (%d 条):\n", len(items))
	for i, item := range items {
		fmt.Printf("  [%d] id=%s, creator=%s\n", i+1, item.ReportID, item.CreatorID)
	}
}

func cmdCreateReport(userid, templateId, toUserIds, textA, textB string) {
	id, err := app.CreateV2Report(userid, templateId, toUserIds, textA, textB)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("创建成功, report_id=%s\n", id)
}

func cmdSendNotify(userid, msgJson string) {
	taskId, err := app.SendWorkNotify(userid, msgJson)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("发送成功, task_id=%d\n", taskId)
}

func cmdLoginInfo(authcode string) {
	info, err := app.GetV2LoginInfo(authcode)
	if err != nil {
		if strings.Contains(err.Error(), "不存在的临时授权码") {
			fmt.Println("authcode 无效或已过期（预期错误）")
			return
		}
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("登录信息: %s\n", info)
}
