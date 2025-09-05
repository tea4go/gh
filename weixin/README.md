# Weixin - 微信集成

## 概述

Weixin包提供了微信公众号、小程序、企业微信等微信生态系统的API集成功能。支持消息收发、用户管理、支付接口、素材管理等核心功能。

## 主要功能

### 1. 微信公众号
- 消息收发（文本、图片、语音、视频）
- 用户管理和标签
- 自定义菜单
- 素材管理
- 网页授权
- JS-SDK集成

### 2. 微信小程序
- 小程序登录
- 数据解密
- 支付接口
- 模板消息
- 客服消息

### 3. 企业微信
- 通讯录管理
- 消息推送
- 应用管理
- 企业支付

### 4. 微信支付
- 统一下单
- 支付结果通知
- 退款处理
- 对账单下载

## 使用示例

### 微信公众号基础配置
```go
package main

import (
    "fmt"
    "path/to/weixin"
)

func main() {
    // 创建公众号客户端
    client := weixin.NewOfficial(&weixin.Config{
        AppID:     "your_app_id",
        AppSecret: "your_app_secret",
        Token:     "your_token",
        EncodingAESKey: "your_aes_key",
    })
    
    // 获取Access Token
    token, err := client.GetAccessToken()
    if err != nil {
        fmt.Printf("获取Access Token失败: %v\n", err)
        return
    }
    
    fmt.Printf("Access Token: %s\n", token)
}
```

### 消息处理
```go
package main

import (
    "fmt"
    "net/http"
    "path/to/weixin"
)

func main() {
    client := weixin.NewOfficial(&weixin.Config{
        AppID:     "your_app_id",
        AppSecret: "your_app_secret",
        Token:     "your_token",
    })
    
    // 消息处理器
    handler := func(msg *weixin.Message) *weixin.Response {
        switch msg.MsgType {
        case "text":
            return &weixin.Response{
                ToUserName:   msg.FromUserName,
                FromUserName: msg.ToUserName,
                MsgType:      "text",
                Content:      fmt.Sprintf("您发送的消息是: %s", msg.Content),
            }
        case "image":
            return &weixin.Response{
                ToUserName:   msg.FromUserName,
                FromUserName: msg.ToUserName,
                MsgType:      "text",
                Content:      "收到您的图片消息",
            }
        default:
            return &weixin.Response{
                ToUserName:   msg.FromUserName,
                FromUserName: msg.ToUserName,
                MsgType:      "text",
                Content:      "暂不支持该类型消息",
            }
        }
    }
    
    // 设置消息处理器
    client.SetMessageHandler(handler)
    
    // 启动HTTP服务
    http.HandleFunc("/wechat", client.HandleMessage)
    fmt.Println("微信消息服务启动在 :8080/wechat")
    http.ListenAndServe(":8080", nil)
}
```

### 发送模板消息
```go
package main

import (
    "fmt"
    "path/to/weixin"
)

func main() {
    client := weixin.NewOfficial(&weixin.Config{
        AppID:     "your_app_id",
        AppSecret: "your_app_secret",
    })
    
    // 发送模板消息
    templateMsg := &weixin.TemplateMessage{
        ToUser:     "user_openid",
        TemplateID: "template_id",
        URL:        "https://example.com",
        Data: map[string]*weixin.TemplateData{
            "first": {
                Value: "尊敬的用户",
                Color: "#173177",
            },
            "keyword1": {
                Value: "订单编号12345",
                Color: "#173177",
            },
            "keyword2": {
                Value: "支付成功",
                Color: "#173177",
            },
            "remark": {
                Value: "感谢您的使用",
                Color: "#173177",
            },
        },
    }
    
    result, err := client.SendTemplateMessage(templateMsg)
    if err != nil {
        fmt.Printf("发送模板消息失败: %v\n", err)
        return
    }
    
    fmt.Printf("消息发送成功，消息ID: %d\n", result.MsgID)
}
```

### 用户管理
```go
package main

import (
    "fmt"
    "path/to/weixin"
)

func main() {
    client := weixin.NewOfficial(&weixin.Config{
        AppID:     "your_app_id",
        AppSecret: "your_app_secret",
    })
    
    // 获取用户列表
    userList, err := client.GetUserList("")
    if err != nil {
        fmt.Printf("获取用户列表失败: %v\n", err)
        return
    }
    
    fmt.Printf("总用户数: %d\n", userList.Total)
    fmt.Printf("本次返回用户数: %d\n", userList.Count)
    
    // 获取用户详细信息
    for _, openid := range userList.Data.OpenIDs {
        userInfo, err := client.GetUserInfo(openid)
        if err != nil {
            fmt.Printf("获取用户信息失败: %v\n", err)
            continue
        }
        
        fmt.Printf("用户: %s (OpenID: %s)\n", userInfo.Nickname, userInfo.OpenID)
    }
}
```

### 微信小程序集成
```go
package main

import (
    "fmt"
    "path/to/weixin"
)

func main() {
    // 创建小程序客户端
    miniapp := weixin.NewMiniProgram(&weixin.Config{
        AppID:     "miniapp_app_id",
        AppSecret: "miniapp_app_secret",
    })
    
    // 用户登录
    code := "js_code_from_frontend"
    session, err := miniapp.Login(code)
    if err != nil {
        fmt.Printf("小程序登录失败: %v\n", err)
        return
    }
    
    fmt.Printf("Session Key: %s\n", session.SessionKey)
    fmt.Printf("OpenID: %s\n", session.OpenID)
    
    // 解密用户数据
    encryptedData := "encrypted_user_data"
    iv := "iv_from_frontend"
    
    userInfo, err := miniapp.DecryptData(session.SessionKey, encryptedData, iv)
    if err != nil {
        fmt.Printf("解密用户数据失败: %v\n", err)
        return
    }
    
    fmt.Printf("用户信息: %+v\n", userInfo)
}
```

### 微信支付
```go
package main

import (
    "fmt"
    "path/to/weixin"
)

func main() {
    // 创建支付客户端
    pay := weixin.NewPay(&weixin.PayConfig{
        AppID:     "your_app_id",
        MchID:     "your_mch_id",
        APIKey:    "your_api_key",
        NotifyURL: "https://yoursite.com/notify",
    })
    
    // 统一下单
    order := &weixin.UnifiedOrder{
        Body:           "测试商品",
        OutTradeNo:     "order_20231225_001",
        TotalFee:       100, // 1元，单位为分
        SpbillCreateIP: "127.0.0.1",
        TradeType:      "JSAPI",
        OpenID:         "user_openid",
    }
    
    result, err := pay.UnifiedOrder(order)
    if err != nil {
        fmt.Printf("统一下单失败: %v\n", err)
        return
    }
    
    if result.ReturnCode == "SUCCESS" && result.ResultCode == "SUCCESS" {
        fmt.Printf("预支付订单创建成功: %s\n", result.PrepayID)
        
        // 生成前端支付参数
        payParams := pay.GetJSAPIPayParams(result.PrepayID)
        fmt.Printf("支付参数: %+v\n", payParams)
    } else {
        fmt.Printf("下单失败: %s\n", result.ReturnMsg)
    }
}
```

### 企业微信集成
```go
package main

import (
    "fmt"
    "path/to/weixin"
)

func main() {
    // 创建企业微信客户端
    work := weixin.NewWork(&weixin.WorkConfig{
        CorpID:     "corp_id",
        AgentID:    1000001,
        CorpSecret: "corp_secret",
    })
    
    // 发送消息给用户
    message := &weixin.WorkMessage{
        ToUser:  "user1|user2",
        MsgType: "text",
        AgentID: 1000001,
        Text: &weixin.WorkTextMessage{
            Content: "这是一条企业微信消息",
        },
    }
    
    result, err := work.SendMessage(message)
    if err != nil {
        fmt.Printf("发送消息失败: %v\n", err)
        return
    }
    
    fmt.Printf("消息发送结果: %+v\n", result)
    
    // 获取部门用户列表
    users, err := work.GetDepartmentUsers(1, true)
    if err != nil {
        fmt.Printf("获取部门用户失败: %v\n", err)
        return
    }
    
    for _, user := range users {
        fmt.Printf("用户: %s (%s)\n", user.Name, user.UserID)
    }
}
```

### 自定义菜单
```go
package main

import (
    "fmt"
    "path/to/weixin"
)

func main() {
    client := weixin.NewOfficial(&weixin.Config{
        AppID:     "your_app_id",
        AppSecret: "your_app_secret",
    })
    
    // 创建自定义菜单
    menu := &weixin.Menu{
        Button: []*weixin.Button{
            {
                Name: "功能菜单",
                SubButton: []*weixin.Button{
                    {
                        Type: "click",
                        Name: "点击事件",
                        Key:  "click_menu_1",
                    },
                    {
                        Type: "view",
                        Name: "跳转网页",
                        URL:  "https://example.com",
                    },
                },
            },
            {
                Type: "click",
                Name: "关于我们",
                Key:  "about_us",
            },
        },
    }
    
    err := client.CreateMenu(menu)
    if err != nil {
        fmt.Printf("创建菜单失败: %v\n", err)
        return
    }
    
    fmt.Println("菜单创建成功")
}
```

## 配置说明

### 公众号配置
```go
type Config struct {
    AppID          string // 应用ID
    AppSecret      string // 应用密钥
    Token          string // 令牌
    EncodingAESKey string // 消息加解密密钥
}
```

### 支付配置
```go
type PayConfig struct {
    AppID     string // 应用ID
    MchID     string // 商户号
    APIKey    string // API密钥
    NotifyURL string // 支付结果通知地址
    CertFile  string // 证书文件路径（退款时需要）
    KeyFile   string // 私钥文件路径（退款时需要）
}
```

## 适用场景

- **营销推广**: 模板消息、群发消息
- **客户服务**: 自动回复、人工客服
- **电商系统**: 订单通知、支付处理
- **企业应用**: 内部通知、审批流程
- **会员管理**: 用户积分、等级管理
- **数据分析**: 用户行为统计

## 注意事项

1. **接口限制**: 注意微信API的调用频率限制
2. **消息加密**: 生产环境建议启用消息加密
3. **token管理**: Access Token有效期2小时，需要定时刷新
4. **签名验证**: 确保正确验证消息签名
5. **错误处理**: 处理微信API返回的各种错误码
6. **安全考虑**: 保护好AppSecret等敏感信息

## 错误码处理

常见错误码及处理方案：
- **40001**: Access Token过期，重新获取
- **40003**: OpenID无效
- **45009**: 接口调用超过限制
- **48001**: API功能未授权

## 最佳实践

1. **token缓存**: 将Access Token缓存到Redis等存储中
2. **异步处理**: 消息处理使用异步队列
3. **日志记录**: 记录详细的API调用日志
4. **重试机制**: 实现API调用的重试机制
5. **监控告警**: 监控API调用成功率和响应时间