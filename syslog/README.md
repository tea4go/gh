# Syslog - 系统日志

## 概述

Syslog包实现了系统日志协议（RFC3164/RFC5424），提供了向本地和远程syslog服务器发送日志消息的功能。支持不同的日志级别、设施和传输协议。

## 主要功能

### 1. 日志级别
- Emergency: 紧急事件
- Alert: 需要立即处理
- Critical: 危险状况
- Error: 错误信息
- Warning: 警告信息
- Notice: 正常但重要的事件
- Info: 信息消息
- Debug: 调试信息

### 2. 传输协议
- UDP传输
- TCP传输
- Unix Domain Socket
- 本地syslog

### 3. 日志设施
- Kernel: 内核消息
- Mail: 邮件系统
- Daemon: 系统守护进程
- Auth: 安全/权限消息
- User: 用户级消息
- Local0-7: 本地使用

## 使用示例

### 基本日志记录
```go
package main

import (
    "fmt"
    "path/to/syslog"
)

func main() {
    // 连接到本地syslog
    logger, err := syslog.New(syslog.LOG_INFO|syslog.LOG_USER, "myapp")
    if err != nil {
        fmt.Printf("连接syslog失败: %v\n", err)
        return
    }
    defer logger.Close()
    
    // 记录不同级别的日志
    logger.Info("应用程序启动")
    logger.Warning("这是一个警告消息")
    logger.Err("这是一个错误消息")
}
```

### 远程syslog
```go
package main

import (
    "fmt"
    "path/to/syslog"
)

func main() {
    // 连接到远程syslog服务器
    logger, err := syslog.Dial("udp", "192.168.1.100:514", 
                              syslog.LOG_INFO|syslog.LOG_USER, "myapp")
    if err != nil {
        fmt.Printf("连接远程syslog失败: %v\n", err)
        return
    }
    defer logger.Close()
    
    // 发送日志到远程服务器
    logger.Info("远程日志消息")
    logger.Alert("需要立即关注的事件")
}
```

### 自定义syslog客户端
```go
package main

import (
    "fmt"
    "time"
    "path/to/syslog"
)

type CustomLogger struct {
    writer *syslog.Writer
}

func NewCustomLogger(network, addr string) (*CustomLogger, error) {
    writer, err := syslog.Dial(network, addr, 
                              syslog.LOG_INFO|syslog.LOG_LOCAL0, "custom")
    if err != nil {
        return nil, err
    }
    
    return &CustomLogger{writer: writer}, nil
}

func (c *CustomLogger) LogEvent(level syslog.Priority, event string) error {
    timestamp := time.Now().Format("2006-01-02 15:04:05")
    message := fmt.Sprintf("[%s] %s", timestamp, event)
    
    switch level {
    case syslog.LOG_EMERG:
        return c.writer.Emerg(message)
    case syslog.LOG_ALERT:
        return c.writer.Alert(message)
    case syslog.LOG_CRIT:
        return c.writer.Crit(message)
    case syslog.LOG_ERR:
        return c.writer.Err(message)
    case syslog.LOG_WARNING:
        return c.writer.Warning(message)
    case syslog.LOG_NOTICE:
        return c.writer.Notice(message)
    case syslog.LOG_INFO:
        return c.writer.Info(message)
    case syslog.LOG_DEBUG:
        return c.writer.Debug(message)
    default:
        return c.writer.Info(message)
    }
}

func (c *CustomLogger) Close() error {
    return c.writer.Close()
}

func main() {
    logger, err := NewCustomLogger("tcp", "localhost:514")
    if err != nil {
        fmt.Printf("创建日志器失败: %v\n", err)
        return
    }
    defer logger.Close()
    
    // 记录各种级别的事件
    logger.LogEvent(syslog.LOG_INFO, "系统初始化完成")
    logger.LogEvent(syslog.LOG_WARNING, "内存使用率较高")
    logger.LogEvent(syslog.LOG_ERR, "数据库连接失败")
}
```

## 适用场景

- 系统监控和告警
- 安全事件记录
- 应用程序日志集中化
- 合规性审计日志
- 网络设备日志收集
- 容器化应用日志

## 注意事项

1. **网络连接**: 远程syslog需要确保网络连接稳定
2. **消息格式**: 遵循RFC3164或RFC5424格式规范
3. **缓冲处理**: 考虑网络中断时的消息缓冲
4. **安全性**: 传输敏感日志时考虑加密
5. **性能影响**: 同步日志可能影响应用性能

## 配置示例

### rsyslog配置
```
# 接收UDP日志
$ModLoad imudp
$UDPServerRun 514
$UDPServerAddress 0.0.0.0

# 接收TCP日志  
$ModLoad imtcp
$InputTCPServerRun 514

# 日志文件配置
local0.*    /var/log/myapp.log
```

### syslog-ng配置
```
source s_network {
    network(ip(0.0.0.0) port(514) transport(udp));
    network(ip(0.0.0.0) port(514) transport(tcp));
};

destination d_myapp {
    file("/var/log/myapp.log");
};

log {
    source(s_network);
    filter(f_local0);
    destination(d_myapp);
};
```

## 最佳实践

1. **结构化日志**: 使用结构化格式便于解析和分析
2. **日志轮转**: 配置日志轮转避免磁盘空间不足
3. **监控告警**: 设置重要事件的实时告警
4. **日志聚合**: 使用ELK等工具进行日志聚合分析
5. **性能优化**: 异步发送日志减少对业务的影响