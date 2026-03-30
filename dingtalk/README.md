# dingtalk 包文档

## 包概述

`dingtalk` 包提供了钉钉API的Go语言封装，支持钉钉小程序、群机器人、OAuth2授权、工作通知、日志管理和JSAPI集成等功能。

## 文件结构

- `common.go` - 通用数据结构和消息类型定义
- `dingtalk_app.go` - 钉钉小程序应用功能（用户管理、部门管理、消息发送、日志管理等）
- `dingtalk_robot.go` - 钉钉群机器人和SNS/OAuth2授权功能
- `dingtalk_app_test.go` - 单元测试

## 主要结构体

### TResult
```go
type TResult struct {
    ErrCode int    `json:"errcode,omitempty"`
    ErrMsg  string `json:"errmsg,omitempty"`
}
```
**功能**: 钉钉API统一错误返回结构体

### TDingTalkResponse
```go
type TDingTalkResponse struct {
    ErrCode   int             `json:"errcode"`
    ErrMsg    string          `json:"errmsg"`
    Result    json.RawMessage `json:"result"`
    RequestID string          `json:"request_id"`
}
```
**功能**: 钉钉V2 API统一响应结构体

### TAccessToken
```go
type TAccessToken struct {
    TResult
    AccessToken string    `json:"access_token"`
    ExpiresIn   int       `json:"expires_in"`
    CreateDate  time.Time
}
```
**功能**: 钉钉访问令牌信息
- **IsValid()** `bool` - 检查令牌是否有效

### TJsapiTicket
```go
type TJsapiTicket struct {
    Code        string `json:"code"`
    Message     string `json:"message"`
    RequestId   string `json:"requestid"`
    JsapiTicket string `json:"jsapiTicket"`
    ExpiresIn   int    `json:"expireIn"`
    CreateDate  time.Time
}
```
**功能**: JSAPI Ticket信息，用于钉钉H5微应用JSAPI鉴权
- **IsValid()** `bool` - 检查Ticket是否有效

### TDingTalkApp
钉钉小程序应用操作类，提供用户管理、部门管理、消息发送、日志管理等功能。

#### 构造方法

**GetDingTalkApp(appkey, appsecret, corp_id, agent_id string)** `*TDingTalkApp`
- **功能**: 创建钉钉小程序应用实例
- **参数**:
  - `appkey` - 应用key
  - `appsecret` - 应用密钥
  - `corp_id` - 企业ID
  - `agent_id` - 应用代理ID

#### 基础方法

**SetAgentId(agent_id string)**
- **功能**: 设置应用代理ID

**GetAccessToken()** `(string, error)`
- **功能**: 获取钉钉访问令牌（自动缓存和刷新）
- **返回**: 令牌字符串和错误信息

**GetJSAPITicket()** `(string, error)`
- **功能**: 获取JSAPI Ticket（自动缓存和刷新）
- **返回**: Ticket字符串和错误信息

**GetConfig(nonceStr, timestamp, url string)** `(string, error)`
- **功能**: 获取钉钉H5微应用JSAPI鉴权配置
- **参数**:
  - `nonceStr` - 随机字符串
  - `timestamp` - 时间戳
  - `url` - 当前页面URL
- **返回**: JSON格式的鉴权配置字符串

#### 用户管理

**GetV2UserInfo(userid string)** `(*TDDV2User, error)`
- **功能**: 根据用户ID获取用户详细信息（V2接口）
- **参数**: `userid` - 用户ID

**GetV2UserInfoByPhone(phone string)** `(*TDDV2User, error)`
- **功能**: 根据手机号获取用户信息（V2接口）
- **参数**: `phone` - 手机号

**GetV2UserInfoByUnionId(unionid string)** `(*TDDV2User, error)`
- **功能**: 根据UnionID获取用户信息（V2接口）
- **参数**: `unionid` - 用户统一ID

**GetV2UsersByName(name string)** `([]*TDDV2User, error)`
- **功能**: 根据姓名搜索用户列表
- **参数**: `name` - 用户姓名

**GetV2LoginInfo(authcode string)** `(string, error)`
- **功能**: 通过免登授权码获取登录用户ID（V2接口）
- **参数**: `authcode` - 免登授权码

#### 部门管理

**GetV2Department(depid int)** `(*TDeptInfo, error)`
- **功能**: 获取部门详细信息（V2接口）
- **参数**: `depid` - 部门ID

**GetDeptUsers(depid int)** `([]*TDDV2User, error)`
- **功能**: 获取部门下的所有用户列表
- **参数**: `depid` - 部门ID

**GetFullDeptName(depid int)** `(string, error)`
- **功能**: 获取完整的部门层级名称（如：总公司/技术部/开发组）
- **参数**: `depid` - 部门ID

#### 组织信息

**GetOrgName(userid string)** `(string, error)`
- **功能**: 获取用户所属组织名称
- **参数**: `userid` - 用户ID

**GetJobName(userid string)** `(string, error)`
- **功能**: 获取用户职位名称
- **参数**: `userid` - 用户ID

**GetAdmins()** `(*TAdmins, error)`
- **功能**: 获取企业管理员列表

#### 工作通知

**SendWorkNotify(user_id, msg_text string)** `(int, error)`
- **功能**: 发送工作通知消息
- **参数**:
  - `user_id` - 接收用户ID
  - `msg_text` - 消息内容
- **返回**: 任务ID和错误信息

#### 日志管理

**GetV2ReportTemplateList(userid string)** `([]TDDReportTemplateItem, error)`
- **功能**: 获取用户可见的日志模板列表
- **参数**: `userid` - 用户ID

**GetV2ReportTemplate(userid, template_name string)** `(*TDDReportTemplate, error)`
- **功能**: 根据模板名称获取日志模板详情
- **参数**:
  - `userid` - 用户ID
  - `template_name` - 模板名称

**GetV2ReportList(userid, start_time, end_time string)** `(*TDDReportList, error)`
- **功能**: 获取用户在指定时间范围内的日志列表
- **参数**:
  - `userid` - 用户ID
  - `start_time` - 开始时间（格式：2006-01-02）
  - `end_time` - 结束时间（格式：2006-01-02）
- **注意**: 查询时间跨度不能超过180天

**GetV2ReportSimpleList(userid, start_time, end_time string)** `([]TDDReportSimpleItem, error)`
- **功能**: 获取用户在指定时间范围内的日志概要列表
- **参数**: 同 `GetV2ReportList`

**CreateV2Report(userid, template_id, to_userids, a_text, b_text string)** `(string, error)`
- **功能**: 创建用户日志
- **参数**:
  - `userid` - 创建人用户ID
  - `template_id` - 日志模板ID
  - `to_userids` - 接收人用户ID（多个用逗号分隔）
  - `a_text` - 本周完成工作内容
  - `b_text` - 下周工作计划内容
- **返回**: 日志ID和错误信息

### TDingTalkRobot
钉钉群机器人操作类，用于向群组发送消息。

#### 主要方法

**Init(access_token, secret string)**
- **功能**: 初始化单个机器人配置
- **参数**:
  - `access_token` - 机器人访问令牌
  - `secret` - 机器人加签密钥

**Inits()**
- **功能**: 初始化多个机器人配置（批量发送场景）

**SendTextMessage(text string, all bool)** `bool`
- **功能**: 发送文本消息
- **参数**:
  - `text` - 消息内容
  - `all` - 是否@所有人
- **返回**: 发送是否成功

**SendTextMessages(text string, all bool)** `int`
- **功能**: 向多个机器人配置发送文本消息
- **参数**: 同 `SendTextMessage`
- **返回**: 成功发送的数量

**SendMDMessage(title, text string)** `bool`
- **功能**: 发送Markdown格式消息
- **参数**:
  - `title` - 消息标题
  - `text` - Markdown内容

**SendMessage(msg interface{})** `bool`
- **功能**: 发送自定义格式消息
- **参数**: `msg` - 消息对象（支持text/markdown类型）

### TDingTalkSns
钉钉SNS扫码登录授权类（旧版API）。

#### 主要方法

**GetDingTalkSns(appkey, appsecret string)** `*TDingTalkSns`
- **功能**: 创建SNS授权实例

**Init(appKey, appSecret string)**
- **功能**: 初始化SNS配置

**GetAppKey()** `string`
- **功能**: 获取当前AppKey

**GetAccessToken()** `(string, error)`
- **功能**: 获取SNS访问令牌

**GetUserByUnionId(code string)** `(bool, string, error)`
- **功能**: 通过临时授权码获取用户UnionID
- **参数**: `code` - 临时授权码
- **返回**: 成功标识、UnionID和错误信息

### TDingTalkOAuth2
钉钉OAuth2.0授权类（新版API）。

#### 主要方法

**GetDingTalkOAuth2(appkey, appsecret string)** `*TDingTalkOAuth2`
- **功能**: 创建OAuth2授权实例

**GetAppKey()** `string`
- **功能**: 获取当前AppKey

**GetAccessToken()** `(string, error)`
- **功能**: 获取OAuth2访问令牌

**GetUserByUnionId(code string)** `(bool, string, error)`
- **功能**: 通过授权码获取用户UnionID
- **参数**: `code` - 授权码
- **返回**: 成功标识、UnionID和错误信息

## 用户数据结构

### TDDV2User
```go
type TDDV2User struct {
    UserId    string       `json:"userid"`
    UnionId   string       `json:"unionid"`
    StaffCode string       `json:"job_number"`
    StaffName string       `json:"name"`
    Email     string       `json:"email"`
    Phone     string       `json:"mobile"`
    Remark    string       `json:"remark"`
    Avatar    string       `json:"avatar"`
    Attrs     TDDV2UserAttr
    AttrText  string       `json:"extension,omitempty"`
}
```
**功能**: 钉钉V2用户信息
- **DisplayName()** `string` - 获取用户显示名称（含组织名称）

### TDDV2UserAttr
```go
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
```
**功能**: 钉钉V2用户扩展属性（组织、部门、职位等）

### TDDUser
```go
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
```
**功能**: 钉钉V1用户信息（兼容旧接口）
- **DisplayName()** `string` - 获取用户显示名称

### TDeptInfo
```go
type TDeptInfo struct {
    TResult
    Id          int      `json:"dept_id"`
    PId         int      `json:"parent_id"`
    Name        string   `json:"name"`
    MemberCount int      `json:"member_count"`
    MUserIds    []string `json:"dept_manager_userid_list"`
    OwnerUserId string   `json:"org_dept_owner"`
}
```
**功能**: 部门信息

### TAdmin / TAdmins
```go
type TAdmin struct {
    UserId   string `json:"userid"`
    SysLevel int    `json:"sys_level"`
}
type TAdmins struct {
    TResult
    Admins []TAdmin `json:"adminList"`
}
```
**功能**: 企业管理员信息（SysLevel=1为主管理员，其他为子管理员）

## 日志数据结构

### TDDReportTemplateItem
```go
type TDDReportTemplateItem struct {
    Name       string `json:"name"`
    ReportCode string `json:"report_code"`
}
```
**功能**: 日志模板概要信息

### TDDReportTemplate
```go
type TDDReportTemplate struct {
    Fields       []TDDReportTemplateField `json:"fields"`
    UserID       string                   `json:"userid"`
    UserName     string                   `json:"user_name"`
    TemplateID   string                   `json:"id"`
    TemplateName string                   `json:"name"`
}
```
**功能**: 日志模板详情

### TDDReportItem / TDDReportContent
```go
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
type TDDReportContent struct {
    Key   string `json:"key"`
    Sort  string `json:"sort"`
    Type  string `json:"type"`
    Value string `json:"value"`
}
```
**功能**: 日志详情及内容项

### TDDReportSimpleItem
```go
type TDDReportSimpleItem struct {
    CreateTime   int64  `json:"create_time"`
    CreatorID    string `json:"creator_id"`
    CreatorName  string `json:"creator_name"`
    DeptName     string `json:"dept_name"`
    Remark       string `json:"remark"`
    ReportID     string `json:"report_id"`
    TemplateName string `json:"template_name"`
}
```
**功能**: 日志概要信息
- **ToString()** `string` - 格式化为可读字符串

## 消息类型

### MessageText
```go
type MessageText struct {
    At      MessageTextAt  `json:"at"`
    Type    string         `json:"msgtype"`
    Message MessageTextSub `json:"text"`
}
```
**功能**: 文本消息结构

### MessageMarkdown
```go
type MessageMarkdown struct {
    Type    string             `json:"msgtype"`
    Message MessageMarkdownSub `json:"markdown"`
}
```
**功能**: Markdown消息结构

## 使用示例

### 创建小程序应用并获取用户信息
```go
app := dingtalk.GetDingTalkApp("appkey", "appsecret", "corp_id", "agent_id")
user, err := app.GetV2UserInfo("userid123")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("用户名: %s, 邮箱: %s, 组织: %s\n", user.StaffName, user.Email, user.Attrs.Org)
```

### 获取JSAPI鉴权配置
```go
app := dingtalk.GetDingTalkApp("appkey", "appsecret", "corp_id", "agent_id")
config, err := app.GetConfig("nonceStr", "1234567890", "https://example.com")
if err != nil {
    log.Fatal(err)
}
fmt.Println(config)
```

### 使用机器人发送消息
```go
robot := &dingtalk.TDingTalkRobot{}
robot.Init("access_token", "secret")
success := robot.SendTextMessage("Hello, DingTalk!", false)
```

### OAuth2用户授权
```go
oauth2 := dingtalk.GetDingTalkOAuth2("appkey", "appsecret")
ok, unionId, err := oauth2.GetUserByUnionId("auth_code")
if ok {
    fmt.Printf("用户UnionID: %s\n", unionId)
}
```

### 日志管理
```go
app := dingtalk.GetDingTalkApp("appkey", "appsecret", "corp_id", "agent_id")
// 获取日志概要列表
reports, err := app.GetV2ReportSimpleList("userid", "2024-01-01", "2024-01-31")
if err != nil {
    log.Fatal(err)
}
for _, r := range reports {
    fmt.Println(r.ToString())
}
// 创建日志
reportId, err := app.CreateV2Report("userid", "template_id", "receiver1,receiver2", "本周完成内容", "下周计划")
```

## 注意事项

1. 所有API调用都需要有效的访问令牌，系统会自动缓存和刷新
2. 令牌过期（错误码503或42001）时会自动重试获取新令牌
3. 错误处理采用统一的TResult结构
4. 支持多种消息格式（文本、Markdown等）
5. 提供新旧两套授权API（SNS和OAuth2），建议使用OAuth2版本
6. 日志查询时间跨度不能超过180天
7. 用户相关接口已升级至V2版本，返回TDDV2User结构体
