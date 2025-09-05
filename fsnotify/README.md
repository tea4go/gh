# FSNotify - 文件系统监控

## 概述

FSNotify是一个跨平台的文件系统事件监控包，提供对文件和目录变化的实时监控功能。支持Linux、Windows、BSD等多个操作系统，使用各系统的原生API（如Linux的inotify、Windows的ReadDirectoryChanges）实现高效的文件系统监控。

## 主要功能

### 1. 事件类型
- **FSN_CREATE**: 文件/目录创建
- **FSN_MODIFY**: 文件/目录修改
- **FSN_DELETE**: 文件/目录删除
- **FSN_RENAME**: 文件/目录重命名
- **FSN_ALL**: 所有事件类型

### 2. 核心结构

#### Watcher 结构
```go
type Watcher struct {
    mu            sync.Mutex        // 映射访问锁
    watches       map[string]*watch // 监控映射
    fsnFlags      map[string]uint32 // 文件标志映射
    fsnmut        sync.Mutex        // 标志保护锁
    internalEvent chan *FileEvent   // 内部事件通道
    Event         chan *FileEvent   // 外部事件通道
    Error         chan error        // 错误通道
    done          chan bool         // 完成信号通道
    isClosed      bool              // 关闭状态
}
```

#### FileEvent 结构
```go
type FileEvent struct {
    mask   uint32 // 事件掩码
    cookie uint32 // 唯一标识符（用于重命名关联）
    Name   string // 文件名
}
```

### 3. 主要方法

#### 创建监控器
```go
func NewWatcher() (*Watcher, error)
```

#### 添加监控
```go
func (w *Watcher) Watch(path string) error                    // 监控所有事件
func (w *Watcher) WatchFlags(path string, flags uint32) error // 监控指定事件
```

#### 移除监控
```go
func (w *Watcher) RemoveWatch(path string) error
```

#### 关闭监控器
```go
func (w *Watcher) Close() error
```

#### 事件判断方法
```go
func (e *FileEvent) IsCreate() bool  // 是否为创建事件
func (e *FileEvent) IsDelete() bool  // 是否为删除事件
func (e *FileEvent) IsModify() bool  // 是否为修改事件
func (e *FileEvent) IsRename() bool  // 是否为重命名事件
func (e *FileEvent) IsAttrib() bool  // 是否为属性变化事件
func (e *FileEvent) GetMask() string // 获取事件掩码描述
```

## 使用示例

### 基本监控示例
```go
package main

import (
    "fmt"
    "log"
    "path/to/fsnotify"
)

func main() {
    // 创建监控器
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    // 添加监控路径
    err = watcher.Watch("/path/to/directory")
    if err != nil {
        log.Fatal(err)
    }

    // 处理事件
    for {
        select {
        case event := <-watcher.Event:
            fmt.Printf("事件: %s\n", event)
            if event.IsCreate() {
                fmt.Printf("创建文件: %s\n", event.Name)
            }
            if event.IsModify() {
                fmt.Printf("修改文件: %s\n", event.Name)
            }
            if event.IsDelete() {
                fmt.Printf("删除文件: %s\n", event.Name)
            }
        case err := <-watcher.Error:
            log.Printf("错误: %v", err)
        }
    }
}
```

### 监控特定事件类型
```go
package main

import (
    "fmt"
    "log"
    "path/to/fsnotify"
)

func main() {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    // 只监控创建和删除事件
    flags := fsnotify.FSN_CREATE | fsnotify.FSN_DELETE
    err = watcher.WatchFlags("/path/to/directory", flags)
    if err != nil {
        log.Fatal(err)
    }

    for {
        select {
        case event := <-watcher.Event:
            fmt.Printf("事件详情: %s\n", event.GetMask())
            fmt.Printf("文件: %s\n", event.Name)
        case err := <-watcher.Error:
            log.Printf("错误: %v", err)
        }
    }
}
```

### 多路径监控
```go
package main

import (
    "fmt"
    "log"
    "path/to/fsnotify"
)

func main() {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    // 添加多个监控路径
    paths := []string{
        "/path/to/dir1",
        "/path/to/dir2",
        "/path/to/file.txt",
    }

    for _, path := range paths {
        err = watcher.Watch(path)
        if err != nil {
            log.Printf("添加监控失败 %s: %v", path, err)
            continue
        }
        fmt.Printf("开始监控: %s\n", path)
    }

    // 事件处理循环
    for {
        select {
        case event := <-watcher.Event:
            handleEvent(event)
        case err := <-watcher.Error:
            log.Printf("监控错误: %v", err)
        }
    }
}

func handleEvent(event *fsnotify.FileEvent) {
    fmt.Printf("事件: %s, 文件: %s\n", event.GetMask(), event.Name)
    
    switch {
    case event.IsCreate():
        fmt.Println("  -> 文件被创建")
    case event.IsModify():
        fmt.Println("  -> 文件被修改")
    case event.IsDelete():
        fmt.Println("  -> 文件被删除")
    case event.IsRename():
        fmt.Println("  -> 文件被重命名")
    case event.IsAttrib():
        fmt.Println("  -> 文件属性被修改")
    }
}
```

## 平台特性

### Linux 平台
- 使用 inotify 系统调用
- 支持递归目录监控
- 支持符号链接监控
- 事件掩码详细丰富

### Windows 平台
- 使用 ReadDirectoryChanges API
- 支持完整的目录变化监控
- 使用完成端口机制提高性能
- 自动处理长文件名

### BSD 平台
- 使用 kqueue 机制
- 高效的事件通知
- 支持文件描述符监控

## 注意事项

1. **资源管理**: 使用完毕后必须调用 `Close()` 方法释放系统资源
2. **错误处理**: 始终监听 `Error` 通道以处理可能的错误
3. **事件过滤**: 合理使用 `WatchFlags` 避免不必要的事件通知
4. **路径处理**: 路径会被自动清理和标准化
5. **并发安全**: 所有公开方法都是并发安全的
6. **系统限制**: 各平台对监控数量可能有限制（如Linux的inotify限制）

## 最佳实践

1. **及时处理事件**: 避免事件通道阻塞导致事件丢失
2. **合理设置缓冲**: 根据需要调整事件通道缓冲区大小
3. **错误恢复**: 实现错误恢复机制，在监控失败时重新建立监控
4. **路径验证**: 添加监控前验证路径的有效性
5. **资源清理**: 在程序退出时确保正确清理监控资源

## 错误处理

常见错误类型：
- 路径不存在或无权限访问
- 系统资源不足
- 监控器已关闭
- 平台特定的系统调用错误

建议的错误处理策略：
```go
select {
case event := <-watcher.Event:
    // 处理事件
case err := <-watcher.Error:
    log.Printf("监控错误: %v", err)
    // 根据错误类型决定是否重试或退出
}
```

## 版本兼容性

本实现基于标准Go库进行开发，兼容Go 1.0+版本，支持主流操作系统包括Linux、Windows、macOS、FreeBSD等。