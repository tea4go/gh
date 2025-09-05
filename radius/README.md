# Radius - RADIUS认证

## 概述

Radius包实现了RADIUS（Remote Authentication Dial In User Service）协议的客户端和服务器功能，用于网络设备的身份认证、授权和计费（AAA）。

## 主要功能

### 1. RADIUS客户端
- 认证请求发送
- 授权请求处理
- 计费记录上报
- 重试和超时机制

### 2. RADIUS服务器
- 用户认证处理
- 授权策略管理
- 计费记录收集
- 多客户端支持

### 3. 协议支持
- PAP认证
- CHAP认证
- EAP认证
- 标准属性支持

## 使用示例

### RADIUS客户端
```go
package main

import (
    "fmt"
    "path/to/radius"
)

func main() {
    // 创建RADIUS客户端
    client := radius.NewClient("192.168.1.1:1812", "secret")
    
    // 用户认证
    result, err := client.Authenticate("username", "password")
    if err != nil {
        fmt.Printf("认证失败: %v\n", err)
        return
    }
    
    if result.Success {
        fmt.Println("认证成功")
    } else {
        fmt.Println("认证失败")
    }
}
```

### RADIUS服务器
```go
package main

import (
    "fmt"
    "path/to/radius"
)

func main() {
    // 创建RADIUS服务器
    server := radius.NewServer(":1812")
    
    // 设置认证处理器
    server.SetAuthHandler(func(username, password string) bool {
        // 实现用户认证逻辑
        return username == "admin" && password == "secret"
    })
    
    // 启动服务器
    fmt.Println("RADIUS服务器启动...")
    err := server.ListenAndServe()
    if err != nil {
        fmt.Printf("服务器启动失败: %v\n", err)
    }
}
```

## 适用场景

- 网络设备认证
- VPN接入控制
- WiFi热点认证
- 企业网络管理
- ISP用户管理

## 注意事项

1. **共享密钥**: 确保客户端和服务器使用相同的共享密钥
2. **网络安全**: RADIUS报文需要在安全网络环境中传输
3. **超时设置**: 合理设置请求超时和重试次数
4. **日志记录**: 记录认证和计费信息便于审计