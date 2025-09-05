# Network - 网络工具

## 概述

Network包提供了一系列网络相关的工具函数，包括Ping测试、DNS查询、HTTP请求、ARP查询、自更新等功能。适用于网络诊断、监控和管理任务。

## 主要功能

### 1. Ping工具
- IPv4/IPv6 Ping支持
- 延迟统计
- 丢包率计算
- 批量Ping测试

### 2. DNS工具
- DNS记录查询
- 域名解析
- 反向DNS查询

### 3. HTTP客户端
- HTTP请求封装
- 自定义头部和参数
- 响应处理

### 4. ARP查询
- 根据IP获取MAC地址
- 网络设备发现

### 5. 网络信息
- IP地理位置查询
- 网络接口信息

## 核心结构

### TPingResult - Ping结果
```go
type TPingResult struct {
    Domain       string  // 域名
    IPAddr       string  // IP地址
    DelayShort   float64 // 最短延迟
    DelayLong    float64 // 最长延迟
    DelayAverage float64 // 平均延迟
    Lost         int     // 丢包数
}
```

### TIPData - IP信息
```go
type TIPData struct {
    IP      string `json:"ip"`      // IP地址
    Country string `json:"country"` // 国家
    Region  string `json:"region"`  // 地区
    City    string `json:"city"`    // 城市
    ISP     string `json:"isp"`     // ISP供应商
}
```

## 使用示例

### Ping测试
```go
package main

import (
    "fmt"
    "path/to/network"
)

func main() {
    // Ping测试
    result, err := network.Ping("google.com", 4)
    if err != nil {
        fmt.Printf("Ping失败: %v\n", err)
        return
    }
    
    fmt.Printf("Ping结果: %s\n", result.String())
    fmt.Printf("平均延迟: %.2f ms\n", result.DelayAverage)
    fmt.Printf("丢包率: %d%%\n", result.Lost)
}
```

### MAC地址查询
```go
package main

import (
    "fmt"
    "path/to/network"
)

func main() {
    // 根据IP获取MAC地址
    mac, err := network.GetMacByIP_file("192.168.1.1")
    if err != nil {
        fmt.Printf("获取MAC失败: %v\n", err)
        return
    }
    
    fmt.Printf("IP: 192.168.1.1, MAC: %s\n", mac)
}
```

## 主要功能

- **网络诊断**: Ping、Traceroute等网络连通性测试
- **设备发现**: ARP扫描、设备信息获取
- **HTTP工具**: 简化的HTTP客户端
- **DNS工具**: 域名解析和查询
- **网络监控**: 连接状态检测
- **自动更新**: 程序自更新机制

## 注意事项

1. **权限要求**: 某些操作需要管理员权限
2. **跨平台兼容**: 考虑不同操作系统的差异
3. **网络超时**: 合理设置网络操作超时时间
4. **错误处理**: 网络操作具有不确定性，需要充分的错误处理

## 依赖库

- `github.com/mdlayher/arp` - ARP协议支持
- 标准库net、net/http等

## 适用场景

- 网络监控系统
- 网络诊断工具
- 自动化部署脚本
- 设备管理系统