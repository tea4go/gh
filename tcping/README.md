# TCPing - TCP连接测试

## 概述

TCPing是一个TCP连接测试工具，类似于ping命令但测试的是TCP连接而不是ICMP。可以测试特定端口的连通性、延迟和可用性，适用于网络诊断和服务监控。

## 主要功能

### 1. TCP连接测试
- 指定主机和端口的连接测试
- 连接时间测量
- 连接成功率统计
- 超时设置

### 2. 批量测试
- 多个主机/端口组合测试
- 并发连接测试
- 结果汇总分析

### 3. 监控功能
- 持续监控模式
- 连接状态变化检测
- 告警阈值设置

## 使用示例

### 基本TCP连接测试
```go
package main

import (
    "fmt"
    "time"
    "path/to/tcping"
)

func main() {
    // 测试TCP连接
    result, err := tcping.Ping("google.com", 80, time.Second*5)
    if err != nil {
        fmt.Printf("连接测试失败: %v\n", err)
        return
    }
    
    if result.Success {
        fmt.Printf("连接成功 - 延迟: %v\n", result.Duration)
    } else {
        fmt.Printf("连接失败: %s\n", result.Error)
    }
}
```

### 批量端口测试
```go
package main

import (
    "fmt"
    "path/to/tcping"
)

func main() {
    // 测试多个端口
    host := "example.com"
    ports := []int{80, 443, 22, 25, 110, 995}
    
    for _, port := range ports {
        result, err := tcping.Ping(host, port, time.Second*3)
        if err != nil {
            fmt.Printf("端口 %d: 测试失败 - %v\n", port, err)
            continue
        }
        
        if result.Success {
            fmt.Printf("端口 %d: 开放 - 延迟 %v\n", port, result.Duration)
        } else {
            fmt.Printf("端口 %d: 关闭或超时\n", port)
        }
    }
}
```

### 连续监控
```go
package main

import (
    "fmt"
    "time"
    "path/to/tcping"
)

func main() {
    host := "www.google.com"
    port := 443
    
    fmt.Printf("开始监控 %s:%d\n", host, port)
    
    ticker := time.NewTicker(time.Second * 10)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            result, err := tcping.Ping(host, port, time.Second*5)
            if err != nil {
                fmt.Printf("[%s] 测试失败: %v\n", 
                          time.Now().Format("15:04:05"), err)
                continue
            }
            
            if result.Success {
                fmt.Printf("[%s] 连接正常 - 延迟: %v\n", 
                          time.Now().Format("15:04:05"), result.Duration)
            } else {
                fmt.Printf("[%s] 连接失败\n", 
                          time.Now().Format("15:04:05"))
            }
        }
    }
}
```

### 统计分析
```go
package main

import (
    "fmt"
    "time"
    "path/to/tcping"
)

func main() {
    host := "github.com"
    port := 443
    testCount := 10
    
    var successCount int
    var totalDelay time.Duration
    var minDelay, maxDelay time.Duration = time.Hour, 0
    
    fmt.Printf("正在测试 %s:%d (%d次)...\n", host, port, testCount)
    
    for i := 0; i < testCount; i++ {
        result, err := tcping.Ping(host, port, time.Second*5)
        if err != nil {
            fmt.Printf("测试 %d: 失败 - %v\n", i+1, err)
            continue
        }
        
        if result.Success {
            successCount++
            totalDelay += result.Duration
            
            if result.Duration < minDelay {
                minDelay = result.Duration
            }
            if result.Duration > maxDelay {
                maxDelay = result.Duration
            }
            
            fmt.Printf("测试 %d: 成功 - %v\n", i+1, result.Duration)
        } else {
            fmt.Printf("测试 %d: 失败\n", i+1)
        }
        
        time.Sleep(time.Second)
    }
    
    // 输出统计结果
    fmt.Println("\n=== 统计结果 ===")
    fmt.Printf("总计测试: %d 次\n", testCount)
    fmt.Printf("成功次数: %d 次\n", successCount)
    fmt.Printf("成功率: %.1f%%\n", float64(successCount)/float64(testCount)*100)
    
    if successCount > 0 {
        avgDelay := totalDelay / time.Duration(successCount)
        fmt.Printf("最小延迟: %v\n", minDelay)
        fmt.Printf("最大延迟: %v\n", maxDelay)
        fmt.Printf("平均延迟: %v\n", avgDelay)
    }
}
```

### 并发测试多个服务
```go
package main

import (
    "fmt"
    "sync"
    "time"
    "path/to/tcping"
)

type ServiceResult struct {
    Service string
    Host    string
    Port    int
    Success bool
    Delay   time.Duration
    Error   error
}

func testService(service, host string, port int, wg *sync.WaitGroup, results chan<- ServiceResult) {
    defer wg.Done()
    
    result, err := tcping.Ping(host, port, time.Second*5)
    
    results <- ServiceResult{
        Service: service,
        Host:    host,
        Port:    port,
        Success: result != nil && result.Success,
        Delay:   func() time.Duration {
            if result != nil {
                return result.Duration
            }
            return 0
        }(),
        Error: err,
    }
}

func main() {
    services := []struct {
        name string
        host string
        port int
    }{
        {"Google HTTPS", "www.google.com", 443},
        {"GitHub HTTPS", "github.com", 443},
        {"Baidu HTTP", "www.baidu.com", 80},
        {"Local SSH", "localhost", 22},
        {"DNS", "8.8.8.8", 53},
    }
    
    var wg sync.WaitGroup
    results := make(chan ServiceResult, len(services))
    
    fmt.Println("并发测试多个服务...")
    
    // 启动并发测试
    for _, service := range services {
        wg.Add(1)
        go testService(service.name, service.host, service.port, &wg, results)
    }
    
    // 等待所有测试完成
    wg.Wait()
    close(results)
    
    // 输出结果
    fmt.Println("\n=== 测试结果 ===")
    for result := range results {
        status := "失败"
        delayInfo := ""
        
        if result.Success {
            status = "成功"
            delayInfo = fmt.Sprintf(" - %v", result.Delay)
        }
        
        fmt.Printf("%-15s %-20s:%d - %s%s\n", 
                   result.Service, result.Host, result.Port, status, delayInfo)
        
        if result.Error != nil {
            fmt.Printf("                错误: %v\n", result.Error)
        }
    }
}
```

## 适用场景

- 网络连通性测试
- 服务端口监控
- 防火墙规则验证  
- 负载均衡器检查
- 服务可用性监控
- 网络故障诊断

## 注意事项

1. **超时设置**: 合理设置连接超时时间
2. **频率控制**: 避免过于频繁的测试影响目标服务
3. **错误处理**: 区分网络错误和服务错误
4. **资源管理**: 及时关闭TCP连接避免资源泄漏
5. **防火墙**: 某些防火墙可能阻止连接测试

## 命令行工具

如果需要命令行版本，可以这样实现：

```bash
# 基本用法
tcping google.com 80

# 指定超时时间
tcping -t 5s github.com 443

# 持续监控
tcping -c 10 -i 2s localhost 3306

# 输出详细信息
tcping -v -c 5 example.com 80
```

## 性能考虑

1. **并发限制**: 控制并发连接数避免资源耗尽
2. **连接复用**: 对于频繁测试考虑连接池
3. **缓存DNS**: 缓存域名解析结果提高性能
4. **异步操作**: 大量测试时使用异步模式