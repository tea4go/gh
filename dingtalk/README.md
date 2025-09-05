# dingtalk 包文档

## 包概述

`dingtalk` 包提供了钉钉API的Go语言封装，支持钉钉小程序、群机器人、OAuth2授权和工作通知等功能。

## 文件结构

- `dingtalk_robot.go` - 钉钉机器人和SNS/OAuth2授权功能
- `dingtalk_app.go` - 钉钉小程序应用功能
- `common.go` - 通用数据结构和消息类型定义

## 主要结构体

### TResult
```go
type TResult struct {
    ErrCode int    `json:"errcode"`
    ErrMsg  string `json:"errmsg"`
}
```
**功能**: 钉钉API统一错误返回结构体

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

### TDingTalkApp
钉钉小程序应用操作类，提供用户管理、部门管理、消息发送等功能。

#### 主要方法

**GetDingTalkApp(appkey, appsecret, agent_id string)** `*TDingTalkApp`
- **功能**: 创建钉钉小程序应用实例
- **参数**: 
  - `appkey` - 应用key
  - `appsecret` - 应用密钥
  - `agent_id` - 应用代理ID

**GetAccessToken()** `(string, error)`
- **功能**: 获取钉钉访问令牌
- **返回**: 令牌字符串和错误信息

**GetUserInfo(userid string)** `(*TDDUser, error)`
- **功能**: 根据用户ID获取用户详细信息
- **参数**: `userid` - 用户ID

**GetUserInfoByPhone(phone string)** `(string, error)`
- **功能**: 根据手机号获取用户ID
- **参数**: `phone` - 手机号

**GetUserInfoByUnionId(unionid string)** `(*TDDUser, error)`
- **功能**: 根据UnionID获取用户信息
- **参数**: `unionid` - 用户统一ID

**GetDepartment(depid int)** `(*TDeptInfo, error)`
- **功能**: 获取部门详细信息
- **参数**: `depid` - 部门ID

**GetAdmins()** `(*TAdmins, error)`
- **功能**: 获取企业管理员列表

**SendWorkNotify(user_id, msg_text string)** `(int, error)`
- **功能**: 发送工作通知消息
- **参数**: 
  - `user_id` - 接收用户ID
  - `msg_text` - 消息内容
- **返回**: 任务ID和错误信息

**GetLoginInfo(authcode string)** `(string, error)`
- **功能**: 通过授权码获取登录用户ID
- **参数**: `authcode` - 授权码

### TDingTalkRobot
钉钉群机器人操作类，用于向群组发送消息。

#### 主要方法

**Init(access_token, secret string)**
- **功能**: 初始化机器人配置
- **参数**: 
  - `access_token` - 机器人访问令牌
  - `secret` - 机器人加签密钥

**SendTextMessage(text string, all bool)** `bool`
- **功能**: 发送文本消息
- **参数**: 
  - `text` - 消息内容
  - `all` - 是否@所有人
- **返回**: 发送是否成功

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

**GetAccessToken()** `(string, error)`
- **功能**: 获取OAuth2访问令牌

**GetUserByUnionId(code string)** `(bool, string, error)`
- **功能**: 通过授权码获取用户UnionID
- **参数**: `code` - 授权码
- **返回**: 成功标识、UnionID和错误信息

## 用户数据结构

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
**功能**: 钉钉用户信息
- **DisplayName()** `string` - 获取用户显示名称

### TDeptInfo
```go
type TDeptInfo struct {
    TResult
    Id      int    `json:"id"`
    PId     int    `json:"parentid"`
    Name    string `json:"name"`
    MUserId string `json:"deptManagerUseridList"`
    IsSub   bool   `json:"groupContainSubDept"`
}
```
**功能**: 部门信息

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
app := dingtalk.GetDingTalkApp("appkey", "appsecret", "agentid")
user, err := app.GetUserInfo("userid123")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("用户名: %s, 邮箱: %s\n", user.StaffName, user.Email)
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

## 注意事项

1. 所有API调用都需要有效的访问令牌
2. 令牌有过期时间限制，系统会自动刷新
3. 错误处理采用统一的TResult结构
4. 支持多种消息格式（文本、Markdown等）
5. 提供新旧两套授权API，建议使用OAuth2版本