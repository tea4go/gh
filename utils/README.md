# Utils - 工具函数库

## 概述

Utils是一个综合性的工具函数库，包含加密、缓存、网络、数据库、同步、文件操作等各类常用功能。为应用程序开发提供基础的工具支持。

## 主要功能

### 1. 加密工具 (aes.go)
- AES加密/解密
- Base64编码
- 密钥管理

### 2. 缓存系统 (lrucache.go)
- LRU缓存实现
- 过期时间管理
- 线程安全操作

### 3. GUID/UUID (guid.go, uuid.go)
- GUID生成和解析
- UUID生成
- Windows GUID支持

### 4. 网络工具 (net_utils.go)
- IP地址处理
- 网络连接检测
- URL处理工具

### 5. 数据库工具 (db_utils.go)
- 数据库连接管理
- SQL工具函数
- 事务处理辅助

### 6. 同步工具 (sync_helper.go)
- 并发控制工具
- 同步原语封装
- 协程管理

### 7. 单例模式 (singleton.go)
- 单例模式实现
- 线程安全单例
- 延迟初始化

### 8. 堆栈工具 (stack.go)
- 调用堆栈获取
- 错误追踪
- 调试信息

### 9. 通用辅助 (helper.go)
- 文件操作
- 字符串处理
- 类型转换
- 时间处理

## 使用示例

### AES加密
```go
package main

import (
    "fmt"
    "path/to/utils"
)

func main() {
    key := "your-secret-key"
    plaintext := "hello world"
    
    // 加密
    encrypted, err := utils.AESEncrypt(plaintext, key)
    if err != nil {
        fmt.Printf("加密失败: %v\n", err)
        return
    }
    
    // 解密
    decrypted, err := utils.AESDecrypt(encrypted, key)
    if err != nil {
        fmt.Printf("解密失败: %v\n", err)
        return
    }
    
    fmt.Printf("原文: %s\n", plaintext)
    fmt.Printf("密文: %s\n", encrypted)
    fmt.Printf("解密: %s\n", decrypted)
}
```

### LRU缓存
```go
package main

import (
    "fmt"
    "time"
    "path/to/utils"
)

func main() {
    // 创建容量为100的LRU缓存
    cache := utils.NewLRUCache(100)
    
    // 设置缓存项
    cache.Set("key1", "value1", time.Minute*5)
    cache.Set("key2", "value2", time.Hour)
    
    // 获取缓存项
    if value, ok := cache.Get("key1"); ok {
        fmt.Printf("缓存值: %s\n", value)
    }
    
    // 检查缓存
    if cache.Contains("key2") {
        fmt.Println("key2 存在于缓存中")
    }
    
    // 删除缓存项
    cache.Delete("key1")
    
    // 清空缓存
    cache.Clear()
}
```

### GUID操作
```go
package main

import (
    "fmt"
    "path/to/utils"
)

func main() {
    // 生成新GUID
    guid := utils.NewGUID()
    fmt.Printf("新GUID: %s\n", guid.String())
    
    // 解析GUID字符串
    guidStr := "{12345678-1234-5678-9ABC-DEF012345678}"
    parsedGuid, err := utils.ParseGUID(guidStr)
    if err != nil {
        fmt.Printf("解析失败: %v\n", err)
        return
    }
    
    fmt.Printf("解析的GUID: %s\n", parsedGuid.String())
}
```

### 网络工具
```go
package main

import (
    "fmt"
    "path/to/utils"
)

func main() {
    // 检查网络连接
    if utils.IsNetworkAvailable() {
        fmt.Println("网络连接可用")
    }
    
    // 获取本机IP
    ip, err := utils.GetLocalIP()
    if err != nil {
        fmt.Printf("获取IP失败: %v\n", err)
        return
    }
    fmt.Printf("本机IP: %s\n", ip)
    
    // 验证IP地址
    if utils.IsValidIP("192.168.1.1") {
        fmt.Println("IP地址有效")
    }
}
```

### 文件操作
```go
package main

import (
    "fmt"
    "path/to/utils"
)

func main() {
    // 检查文件存在
    if utils.FileExists("test.txt") {
        fmt.Println("文件存在")
    }
    
    // 创建目录
    err := utils.CreateDir("./data")
    if err != nil {
        fmt.Printf("创建目录失败: %v\n", err)
    }
    
    // 获取文件大小
    size, err := utils.GetFileSize("test.txt")
    if err != nil {
        fmt.Printf("获取文件大小失败: %v\n", err)
    } else {
        fmt.Printf("文件大小: %d bytes\n", size)
    }
    
    // 复制文件
    err = utils.CopyFile("source.txt", "dest.txt")
    if err != nil {
        fmt.Printf("复制文件失败: %v\n", err)
    }
}
```

### 字符串处理
```go
package main

import (
    "fmt"
    "path/to/utils"
)

func main() {
    // 生成随机字符串
    randomStr := utils.RandomString(10)
    fmt.Printf("随机字符串: %s\n", randomStr)
    
    // 字符串转换
    result := utils.ToCamelCase("hello_world")
    fmt.Printf("驼峰命名: %s\n", result)
    
    // 字符串截取
    truncated := utils.TruncateString("这是一个很长的字符串", 10)
    fmt.Printf("截取字符串: %s\n", truncated)
    
    // MD5哈希
    hash := utils.MD5Hash("hello world")
    fmt.Printf("MD5哈希: %s\n", hash)
}
```

### 时间处理
```go
package main

import (
    "fmt"
    "time"
    "path/to/utils"
)

func main() {
    // 格式化时间
    now := time.Now()
    formatted := utils.FormatTime(now, "2006-01-02 15:04:05")
    fmt.Printf("格式化时间: %s\n", formatted)
    
    // 解析时间字符串
    timeStr := "2023-12-25 10:30:00"
    parsedTime, err := utils.ParseTime(timeStr, "2006-01-02 15:04:05")
    if err != nil {
        fmt.Printf("时间解析失败: %v\n", err)
    } else {
        fmt.Printf("解析时间: %s\n", parsedTime.String())
    }
    
    // 时间差计算
    duration := utils.TimeSince(parsedTime)
    fmt.Printf("时间差: %s\n", duration.String())
}
```

### 单例模式
```go
package main

import (
    "fmt"
    "path/to/utils"
)

type Config struct {
    AppName string
    Version string
}

var configInstance *Config

func GetConfig() *Config {
    return utils.GetSingleton(&configInstance, func() *Config {
        return &Config{
            AppName: "MyApp",
            Version: "1.0.0",
        }
    }).(*Config)
}

func main() {
    config1 := GetConfig()
    config2 := GetConfig()
    
    fmt.Printf("配置1: %+v\n", config1)
    fmt.Printf("配置2: %+v\n", config2)
    fmt.Printf("是同一个实例: %t\n", config1 == config2)
}
```

## 常用常量

```go
// 时间格式
const (
    DateTimeFormat = "2006-01-02 15:04:05"
    DateFormat     = "2006-01-02"
    TimeFormat     = "15:04:05"
)

// 文件大小单位
const (
    KB = 1024
    MB = KB * 1024
    GB = MB * 1024
    TB = GB * 1024
)
```

## 性能优化

1. **缓存使用**: 合理使用LRU缓存减少重复计算
2. **字符串处理**: 使用strings.Builder进行字符串拼接
3. **并发控制**: 使用sync_helper中的工具进行并发控制
4. **内存管理**: 及时释放不再使用的资源

## 注意事项

1. **线程安全**: 标记了线程安全的函数可以并发使用
2. **错误处理**: 所有可能失败的操作都返回error
3. **资源清理**: 使用完毕后及时清理资源
4. **配置验证**: 传入参数需要进行有效性验证

## 适用场景

- Web应用开发
- API服务开发  
- 系统工具开发
- 数据处理程序
- 网络服务程序

Utils包提供了开发中常用的基础工具，可以显著提高开发效率和代码质量。