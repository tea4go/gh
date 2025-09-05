# execplus 包文档

## 包概述

`execplus` 包是对Go标准库 `os/exec` 的增强版本，提供了更强大的进程执行功能。主要特性包括：
1. 支持多次调用Wait()而不会出错
2. 提供Terminate()函数强制终止进程
3. 跨平台支持（Windows/POSIX）
4. 字符编码转换支持
5. 环境变量设置和用户切换功能

## 文件结构

- `execplus.go` - 核心功能实现
- `execplus_windows.go` - Windows平台特定功能
- `execplus_posix.go` - POSIX系统（Linux/macOS等）特定功能
- `execplus_test.go` - 测试文件

## 主要结构体

### CmdPlus
```go
type CmdPlus struct {
    *exec.Cmd                    // 内嵌标准Cmd
    cancelFunc context.CancelFunc // 取消函数
    errChs     []chan error       // 错误通道列表
    err        error              // 错误信息
    finished   bool               // 完成标识
    once       sync.Once          // 确保只执行一次
    mu         sync.Mutex         // 互斥锁
}
```
**功能**: 增强版命令执行器，支持更可靠的进程管理。

## 全局配置

**SetShellName(name string)**
- **功能**: 设置默认shell名称
- **参数**: `name` - shell程序名称
- **默认值**: Windows: "cmd", POSIX: "/bin/bash"

## 命令创建函数

### 执行程序命令

**Command(name string, arg ...string)** `*CmdPlus`
- **功能**: 创建命令执行实例
- **参数**: 
  - `name` - 程序名称
  - `arg` - 程序参数
- **返回**: CmdPlus实例
- **特性**: 支持取消操作

**CommandContext(ctx context.Context, name, arg ...string)** `*CmdPlus`
- **功能**: 创建带上下文的命令执行实例
- **参数**: 
  - `ctx` - 上下文
  - `name` - 程序名称
  - `arg` - 程序参数
- **返回**: CmdPlus实例

### 执行Shell命令

**CommandString(command string)** `*CmdPlus`
- **功能**: 创建Shell命令执行实例
- **参数**: `command` - Shell命令字符串
- **返回**: CmdPlus实例
- **平台差异**: 
  - Windows: 使用 `cmd /c`
  - POSIX: 使用 `bash -c`

**CommandStringContext(ctx context.Context, command string)** `*CmdPlus`
- **功能**: 创建带上下文的Shell命令执行实例
- **参数**: 
  - `ctx` - 上下文
  - `command` - Shell命令字符串
- **返回**: CmdPlus实例

## CmdPlus 方法

### 进程控制

**Terminate()**
- **功能**: 强制终止进程
- **说明**: 通过取消上下文来终止进程执行

**Wait()** `error`
- **功能**: 等待进程结束
- **返回**: 错误信息
- **特性**: 支持多次调用而不会出错（与标准库不同）

### 显示控制

**ShowConsole(flag bool)**
- **功能**: 控制是否显示控制台输出
- **参数**: `flag` - true显示输出，false隐藏输出

**HideWindow()**
- **功能**: 隐藏程序窗口
- **平台差异**: 
  - Windows: 设置HideWindow属性
  - POSIX: 空实现

### 环境变量操作

**SetEnv(key, value string)** `bool`
- **功能**: 设置环境变量
- **参数**: 
  - `key` - 环境变量名
  - `value` - 环境变量值
- **返回**: 是否为新增变量（true=新增，false=修改）
- **特性**: 支持跨平台大小写处理

### 用户权限

**SetUser(name string)** `error`
- **功能**: 设置进程运行用户
- **参数**: `name` - 用户名
- **返回**: 错误信息
- **平台差异**: 
  - Windows: 不支持，返回nil
  - POSIX: 设置用户和组ID

## 工具函数

**ConvertByte2String(byte []byte, charset string)** `string`
- **功能**: 字节转字符串编码转换
- **参数**: 
  - `byte` - 字节数据
  - `charset` - 字符编码（"GB18030", "UTF-8"）
- **返回**: 转换后的字符串
- **用途**: 处理中文编码问题

## 平台特性

### Windows 特性
- 默认Shell: `cmd`
- 支持隐藏CMD窗口
- 不支持用户切换
- 环境变量名不区分大小写

### POSIX 特性  
- 默认Shell: `/bin/bash`
- 支持进程组管理（Setsid）
- 支持用户和组切换
- 环境变量名区分大小写

## 使用示例

### 基本命令执行
```go
// 执行程序
cmd := execplus.Command("ls", "-la")
cmd.ShowConsole(true)
err := cmd.Start()
if err != nil {
    log.Fatal(err)
}
err = cmd.Wait()
if err != nil {
    log.Printf("命令执行失败: %v", err)
}
```

### Shell命令执行
```go
// 执行Shell命令
cmd := execplus.CommandString("echo 'Hello World'")
output, err := cmd.Output()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("输出: %s\n", output)
```

### 带超时的命令执行
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

cmd := execplus.CommandContext(ctx, "ping", "google.com")
err := cmd.Start()
if err != nil {
    log.Fatal(err)
}

// 可以强制终止
go func() {
    time.Sleep(10 * time.Second)
    cmd.Terminate()
}()

err = cmd.Wait()
if err != nil {
    log.Printf("命令被终止或失败: %v", err)
}
```

### 设置环境变量和用户
```go
cmd := execplus.CommandString("env | grep MYVAR")

// 设置环境变量
isNew := cmd.SetEnv("MYVAR", "myvalue")
fmt.Printf("是否新增变量: %v\n", isNew)

// POSIX系统下设置运行用户
if runtime.GOOS != "windows" {
    err := cmd.SetUser("nobody")
    if err != nil {
        log.Printf("设置用户失败: %v", err)
    }
}

// 隐藏窗口（Windows）
cmd.HideWindow()

err := cmd.Start()
// ... 处理执行结果
```

### 多次等待支持
```go
cmd := execplus.CommandString("sleep 5")
err := cmd.Start()
if err != nil {
    log.Fatal(err)
}

// 第一次等待
go func() {
    err1 := cmd.Wait()
    log.Printf("第一次等待结果: %v", err1)
}()

// 第二次等待（不会出错）
go func() {
    err2 := cmd.Wait()
    log.Printf("第二次等待结果: %v", err2)
}()

time.Sleep(6 * time.Second)
```

### 字符编码处理
```go
cmd := execplus.CommandString("dir") // Windows命令
cmd.HideWindow()
output, err := cmd.Output()
if err != nil {
    log.Fatal(err)
}

// 转换编码（处理中文）
result := execplus.ConvertByte2String(output, "GB18030")
fmt.Printf("目录列表:\n%s", result)
```

## 注意事项

1. **线程安全**: Wait()方法支持多协程并发调用
2. **平台差异**: 注意Windows和POSIX系统的功能差异
3. **资源管理**: 使用完毕后建议调用Terminate()确保资源清理
4. **编码问题**: Windows下可能需要使用ConvertByte2String处理中文编码
5. **权限要求**: POSIX系统下切换用户需要相应权限
6. **上下文管理**: 合理使用Context进行超时和取消控制