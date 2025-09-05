# Log4go - 日志系统

## 概述

Log4go是一个功能强大、高性能的Go语言日志系统，基于Beego日志框架开发。支持多种输出适配器、异步日志、多级别日志记录等特性，适用于各种规模的应用程序。

## 主要功能

### 1. 多级别日志
- **LevelEmergency**: 紧急事故
- **LevelAlert**: 警报
- **LevelCritical**: 危险
- **LevelError**: 错误
- **LevelWarning**: 警告
- **LevelNotice**: 通知
- **LevelInfo**: 信息
- **LevelDebug**: 调试
- **LevelPrint**: 直接打印（无前缀）

### 2. 多种输出适配器
- **Console**: 控制台输出
- **File**: 单文件输出
- **MultiFile**: 多文件分级输出
- **SMTP**: 邮件输出
- **Conn**: 网络连接输出

### 3. 异步日志支持
- 缓冲通道机制
- 可配置队列长度
- 优雅关闭和刷新

## 核心结构

### TLogger - 主日志器
```go
type TLogger struct {
    lock          sync.Mutex      // 互斥锁
    init_flag     bool           // 初始化标志
    funcCallDepth int            // 函数调用深度
    Async_flag    bool           // 异步标志
    msgChanLen    int64          // 消息通道长度
    msgChan       chan *tLogMsg  // 消息通道
    signalChan    chan string    // 信号通道
    lastTime      time.Time      // 最后日志时间
    wg            sync.WaitGroup // 等待组
    outputs       []*nameLogger  // 输出器列表
}
```

### ILogger - 日志接口
```go
type ILogger interface {
    Init(config string) error
    SetLevel(l int)
    GetLevel() int
    WriteMsg(fileName string, fileLine int, callLevel int, callFunc string, 
             logLevel int, when time.Time, msg string) error
    Destroy()
    Flush()
}
```

## 使用示例

### 基本使用
```go
package main

import (
    "path/to/log4go"
)

func main() {
    // 使用全局日志器
    logs.SetLogger("console")
    
    // 记录不同级别的日志
    logs.Emergency("系统紧急事故")
    logs.Alert("系统警报")
    logs.Critical("危险情况")
    logs.Error("错误信息")
    logs.Warning("警告信息")
    logs.Notice("通知信息")
    logs.Info("普通信息")
    logs.Debug("调试信息")
    
    // 直接打印（无格式）
    logs.Print("直接打印内容")
    
    // 刷新日志缓冲
    logs.Flush()
}
```

### 创建自定义日志器
```go
package main

import (
    "path/to/log4go"
)

func main() {
    // 创建新的日志器
    logger := logs.NewLogger(1000) // 通道大小为1000
    
    // 设置日志级别
    logger.SetLevel(logs.LevelInfo)
    
    // 添加控制台输出
    logger.SetLogger("console", `{"level":4}`)
    
    // 添加文件输出
    logger.SetLogger("file", `{"filename":"app.log","level":6}`)
    
    // 使用日志器
    logger.Info("应用程序启动")
    logger.Error("发生错误: %s", "示例错误")
    
    // 关闭日志器
    logger.Close()
}
```

### 文件日志配置
```go
package main

import (
    "path/to/log4go"
)

func main() {
    logger := logs.NewLogger()
    
    // 单文件日志
    fileConfig := `{
        "filename": "app.log",
        "level": 6,
        "maxsize": 1048576,
        "maxdays": 7,
        "rotate": true,
        "perm": "0600"
    }`
    logger.SetLogger("file", fileConfig)
    
    logger.Info("文件日志测试")
    logger.Flush()
}
```

### 多文件分级日志
```go
package main

import (
    "path/to/log4go"
)

func main() {
    logger := logs.NewLogger()
    
    // 多文件日志配置
    multiFileConfig := `{
        "filename": "logs/app.log",
        "separate": [
            "emergency", "alert", "critical", "error", 
            "warning", "notice", "info", "debug"
        ],
        "level": 7,
        "maxsize": 1048576,
        "maxdays": 7,
        "rotate": true
    }`
    logger.SetLogger("multifile", multiFileConfig)
    
    // 不同级别会写入不同文件
    logger.Error("错误信息") // 写入 logs/app.error.log
    logger.Info("信息")     // 写入 logs/app.info.log
    logger.Debug("调试")    // 写入 logs/app.debug.log
    
    logger.Flush()
}
```

### SMTP邮件日志
```go
package main

import (
    "path/to/log4go"
)

func main() {
    logger := logs.NewLogger()
    
    // SMTP日志配置
    smtpConfig := `{
        "username": "sender@example.com",
        "password": "password",
        "host": "smtp.example.com:587",
        "sendTos": ["admin@example.com"],
        "subject": "应用程序告警",
        "level": 3
    }`
    logger.SetLogger("smtp", smtpConfig)
    
    // 只有错误级别及以上才发邮件
    logger.Error("严重错误，需要立即处理")
    logger.Critical("系统危险状态")
    
    logger.Flush()
}
```

### 网络连接日志
```go
package main

import (
    "path/to/log4go"
)

func main() {
    logger := logs.NewLogger()
    
    // 网络连接日志配置
    connConfig := `{
        "net": "tcp",
        "addr": "192.168.1.100:8080",
        "level": 6,
        "reconnectOnMsg": false
    }`
    logger.SetLogger("conn", connConfig)
    
    logger.Info("通过网络发送日志")
    logger.Error("网络日志错误测试")
    
    logger.Flush()
}
```

### 异步日志配置
```go
package main

import (
    "path/to/log4go"
    "time"
)

func main() {
    logger := logs.NewLogger()
    
    // 设置异步模式，通道大小为10000
    logger.SetAsync(10000)
    
    // 添加输出器
    logger.SetLogger("console")
    logger.SetLogger("file", `{"filename":"async.log"}`)
    
    // 高频日志写入
    for i := 0; i < 1000; i++ {
        logger.Info("异步日志测试 %d", i)
    }
    
    // 等待一段时间让异步日志处理完成
    time.Sleep(time.Second)
    logger.Close()
}
```

### 自定义日志格式
```go
package main

import (
    "path/to/log4go"
)

func main() {
    logger := logs.NewLogger()
    
    // 控制台彩色输出
    consoleConfig := `{
        "level": 7,
        "color": true
    }`
    logger.SetLogger("console", consoleConfig)
    
    // 文件详细格式
    fileConfig := `{
        "filename": "detailed.log",
        "level": 7,
        "maxsize": 1048576,
        "rotate": true
    }`
    logger.SetLogger("file", fileConfig)
    
    // 设置调用深度（显示准确的调用位置）
    logger.SetFuncCallDepth(3)
    
    logger.Debug("调试信息，包含调用位置")
    logger.Info("信息日志")
    logger.Warning("警告日志")
    
    logger.Close()
}
```

### 组合使用示例
```go
package main

import (
    "os"
    "path/to/log4go"
)

func main() {
    // 初始化全局日志器
    logs.SetLevel(logs.LevelInfo)
    
    // 添加控制台输出（开发环境）
    if os.Getenv("ENV") == "development" {
        logs.SetLogger("console", `{"level":7,"color":true}`)
    }
    
    // 添加文件输出（生产环境）
    logs.SetLogger("file", `{
        "filename": "app.log",
        "level": 6,
        "maxsize": 10485760,
        "maxdays": 30,
        "rotate": true
    }`)
    
    // 添加错误邮件通知
    logs.SetLogger("smtp", `{
        "username": "app@company.com",
        "password": "password",
        "host": "smtp.company.com:587",
        "sendTos": ["admin@company.com"],
        "subject": "[应用告警] 系统错误",
        "level": 3
    }`)
    
    // 应用程序逻辑
    logs.Info("应用程序启动成功")
    
    // 模拟业务逻辑
    if err := businessLogic(); err != nil {
        logs.Error("业务逻辑执行失败: %v", err)
    }
    
    logs.Info("应用程序结束")
    logs.Flush()
}

func businessLogic() error {
    logs.Debug("开始执行业务逻辑")
    
    // 模拟一些处理过程
    logs.Info("处理步骤 1 完成")
    logs.Info("处理步骤 2 完成")
    
    // 模拟警告情况
    logs.Warning("检测到潜在问题，继续执行")
    
    logs.Debug("业务逻辑执行完成")
    return nil
}
```

## 全局API函数

### 基础配置
```go
func Reset()                                    // 重置日志器
func SetLevel(l int, adapters ...string)       // 设置日志级别
func GetLevel(adapter ...string) int           // 获取日志级别
func SetLogFuncCallDepth(d int)                // 设置函数调用深度
func SetConsole2Stderr(f bool)                 // 设置控制台输出到stderr
func SetLogger(adapter string, config ...string) error  // 添加日志输出器
func DelLogger(adapter string) error           // 删除日志输出器
```

### 日志记录
```go
func Emergency(format string, v ...interface{})
func Alert(format string, v ...interface{})
func Critical(format string, v ...interface{})
func Error(format string, v ...interface{})
func Warning(format string, v ...interface{})
func Notice(format string, v ...interface{})
func Info(format string, v ...interface{})
func Debug(format string, v ...interface{})
func Print(v ...interface{})
```

### 控制操作
```go
func Flush()                    // 刷新日志缓冲
func GetLastLogTime() time.Time // 获取最后日志时间
```

## 配置详解

### 文件日志配置
```json
{
  "filename": "app.log",      // 日志文件名
  "level": 6,                 // 日志级别
  "maxsize": 1048576,         // 最大文件大小(字节)
  "maxdays": 7,               // 保留天数
  "rotate": true,             // 是否轮转
  "perm": "0600"             // 文件权限
}
```

### 控制台日志配置
```json
{
  "level": 7,                 // 日志级别
  "color": true               // 是否彩色输出
}
```

### 邮件日志配置
```json
{
  "username": "sender@example.com",
  "password": "password",
  "host": "smtp.example.com:587",
  "sendTos": ["admin@example.com"],
  "subject": "应用程序告警",
  "level": 3
}
```

## 性能优化

1. **异步模式**: 使用`SetAsync()`启用异步日志提高性能
2. **合理级别**: 生产环境避免使用Debug级别
3. **文件轮转**: 配置合适的文件大小和保留天数
4. **缓冲区大小**: 根据日志量调整通道大小
5. **及时刷新**: 在关键点调用`Flush()`确保日志写入

## 注意事项

1. **内存管理**: 异步模式需要合理设置通道大小避免内存溢出
2. **文件权限**: 确保日志文件目录有写权限
3. **线程安全**: 所有操作都是线程安全的
4. **优雅关闭**: 程序退出前调用`Flush()`和`Close()`
5. **错误处理**: 配置错误可能导致日志丢失，需要检查返回值

## 最佳实践

- 开发环境使用console输出便于调试
- 生产环境使用file输出并配置轮转
- 重要错误同时发送邮件通知
- 使用结构化日志格式便于日志分析
- 定期清理过期日志文件释放磁盘空间