# TimeWheel - 时间轮

## 概述

TimeWheel实现了高效的时间轮算法，用于管理大量的定时任务。时间轮是一种高效的定时器数据结构，特别适合处理大量超时任务的场景，如网络连接超时、缓存过期等。

## 主要功能

### 1. 时间轮算法
- 多层时间轮支持
- O(1)时间复杂度的任务插入
- 高精度时间管理
- 内存占用优化

### 2. 定时任务管理
- 任务调度
- 延时执行
- 周期性任务
- 任务取消

### 3. 高性能特性
- 批量任务处理
- 无锁设计选项
- 内存池管理
- 时间精度控制

## 核心概念

### 时间轮结构
```
Level 0: 精确到毫秒 (1000个槽位)
Level 1: 精确到秒   (60个槽位)  
Level 2: 精确到分钟 (60个槽位)
Level 3: 精确到小时 (24个槽位)
```

## 使用示例

### 基本定时器
```go
package main

import (
    "fmt"
    "time"
    "path/to/timewheel"
)

func main() {
    // 创建时间轮，精度为100毫秒
    tw := timewheel.New(time.Millisecond*100, 1000)
    tw.Start()
    defer tw.Stop()
    
    // 添加定时任务
    timer := tw.AfterFunc(time.Second*3, func() {
        fmt.Println("定时任务执行了")
    })
    
    fmt.Println("定时任务已添加，3秒后执行")
    
    // 等待任务执行
    time.Sleep(time.Second * 5)
    
    // 可以取消还未执行的任务
    timer.Cancel()
}
```

### 周期性任务
```go
package main

import (
    "fmt"
    "time"
    "path/to/timewheel"
)

func main() {
    tw := timewheel.New(time.Millisecond*10, 6000)
    tw.Start()
    defer tw.Stop()
    
    // 添加周期性任务
    ticker := tw.NewTicker(time.Second, func() {
        fmt.Printf("[%s] 周期性任务执行\n", 
                  time.Now().Format("15:04:05"))
    })
    
    // 运行10秒后停止
    time.Sleep(time.Second * 10)
    ticker.Stop()
    
    fmt.Println("周期性任务已停止")
}
```

### 超时管理
```go
package main

import (
    "fmt"
    "sync"
    "time"
    "path/to/timewheel"
)

type Connection struct {
    ID      string
    timer   timewheel.Timer
    mu      sync.Mutex
    closed  bool
}

func (c *Connection) ResetTimeout(tw *timewheel.TimeWheel, timeout time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.closed {
        return
    }
    
    // 取消旧的超时器
    if c.timer != nil {
        c.timer.Cancel()
    }
    
    // 设置新的超时器
    c.timer = tw.AfterFunc(timeout, func() {
        fmt.Printf("连接 %s 超时，自动关闭\n", c.ID)
        c.Close()
    })
}

func (c *Connection) Close() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.closed {
        return
    }
    
    c.closed = true
    if c.timer != nil {
        c.timer.Cancel()
    }
    
    fmt.Printf("连接 %s 已关闭\n", c.ID)
}

func main() {
    tw := timewheel.New(time.Second, 3600)
    tw.Start()
    defer tw.Stop()
    
    // 模拟连接管理
    conn1 := &Connection{ID: "conn-001"}
    conn2 := &Connection{ID: "conn-002"}
    
    // 设置30秒超时
    conn1.ResetTimeout(tw, time.Second*30)
    conn2.ResetTimeout(tw, time.Second*30)
    
    // 模拟conn1有活动，重置超时
    time.Sleep(time.Second * 20)
    conn1.ResetTimeout(tw, time.Second*30)
    fmt.Println("conn-001 有活动，超时时间重置")
    
    // 等待查看超时效果
    time.Sleep(time.Second * 40)
}
```

### 缓存过期管理
```go
package main

import (
    "fmt"
    "sync"
    "time"
    "path/to/timewheel"
)

type CacheItem struct {
    Key    string
    Value  interface{}
    Timer  timewheel.Timer
}

type Cache struct {
    items map[string]*CacheItem
    mu    sync.RWMutex
    tw    *timewheel.TimeWheel
}

func NewCache() *Cache {
    tw := timewheel.New(time.Second, 3600)
    tw.Start()
    
    return &Cache{
        items: make(map[string]*CacheItem),
        tw:    tw,
    }
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // 如果已存在，先取消旧的过期定时器
    if item, exists := c.items[key]; exists && item.Timer != nil {
        item.Timer.Cancel()
    }
    
    // 创建新的缓存项
    item := &CacheItem{
        Key:   key,
        Value: value,
    }
    
    // 设置过期定时器
    if ttl > 0 {
        item.Timer = c.tw.AfterFunc(ttl, func() {
            c.delete(key)
        })
    }
    
    c.items[key] = item
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if item, exists := c.items[key]; exists {
        return item.Value, true
    }
    return nil, false
}

func (c *Cache) delete(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if item, exists := c.items[key]; exists {
        if item.Timer != nil {
            item.Timer.Cancel()
        }
        delete(c.items, key)
        fmt.Printf("缓存项 %s 已过期删除\n", key)
    }
}

func (c *Cache) Delete(key string) {
    c.delete(key)
}

func (c *Cache) Size() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return len(c.items)
}

func (c *Cache) Close() {
    c.tw.Stop()
}

func main() {
    cache := NewCache()
    defer cache.Close()
    
    // 设置一些缓存项
    cache.Set("key1", "value1", time.Second*5)
    cache.Set("key2", "value2", time.Second*10)
    cache.Set("key3", "value3", time.Second*15)
    
    fmt.Printf("缓存大小: %d\n", cache.Size())
    
    // 读取缓存
    if value, exists := cache.Get("key1"); exists {
        fmt.Printf("key1 = %s\n", value)
    }
    
    // 等待过期
    time.Sleep(time.Second * 6)
    fmt.Printf("6秒后缓存大小: %d\n", cache.Size())
    
    time.Sleep(time.Second * 5)
    fmt.Printf("11秒后缓存大小: %d\n", cache.Size())
    
    time.Sleep(time.Second * 5)
    fmt.Printf("16秒后缓存大小: %d\n", cache.Size())
}
```

### 批量任务管理
```go
package main

import (
    "fmt"
    "time"
    "path/to/timewheel"
)

type TaskManager struct {
    tw     *timewheel.TimeWheel
    tasks  map[string]timewheel.Timer
}

func NewTaskManager() *TaskManager {
    tw := timewheel.New(time.Millisecond*100, 36000)
    tw.Start()
    
    return &TaskManager{
        tw:    tw,
        tasks: make(map[string]timewheel.Timer),
    }
}

func (tm *TaskManager) AddTask(id string, delay time.Duration, task func()) {
    // 如果任务已存在，先取消
    if timer, exists := tm.tasks[id]; exists {
        timer.Cancel()
    }
    
    // 添加新任务
    timer := tm.tw.AfterFunc(delay, func() {
        fmt.Printf("执行任务: %s\n", id)
        task()
        // 任务执行后从管理器中移除
        delete(tm.tasks, id)
    })
    
    tm.tasks[id] = timer
}

func (tm *TaskManager) CancelTask(id string) bool {
    if timer, exists := tm.tasks[id]; exists {
        timer.Cancel()
        delete(tm.tasks, id)
        return true
    }
    return false
}

func (tm *TaskManager) PendingCount() int {
    return len(tm.tasks)
}

func (tm *TaskManager) Close() {
    // 取消所有待执行的任务
    for id, timer := range tm.tasks {
        timer.Cancel()
        delete(tm.tasks, id)
    }
    tm.tw.Stop()
}

func main() {
    tm := NewTaskManager()
    defer tm.Close()
    
    // 添加多个定时任务
    for i := 1; i <= 5; i++ {
        taskID := fmt.Sprintf("task-%d", i)
        delay := time.Second * time.Duration(i*2)
        
        tm.AddTask(taskID, delay, func() {
            fmt.Printf("任务完成时间: %s\n", 
                      time.Now().Format("15:04:05"))
        })
    }
    
    fmt.Printf("已添加 %d 个任务\n", tm.PendingCount())
    
    // 3秒后取消task-3
    time.Sleep(time.Second * 3)
    if tm.CancelTask("task-3") {
        fmt.Println("task-3 已取消")
    }
    
    // 等待所有任务完成
    time.Sleep(time.Second * 12)
    fmt.Printf("剩余任务数: %d\n", tm.PendingCount())
}
```

## 适用场景

- **网络服务**: 连接超时管理、会话过期
- **缓存系统**: 数据过期清理、TTL管理  
- **消息队列**: 消息延迟投递、重试机制
- **游戏服务器**: 技能冷却、buff超时
- **监控系统**: 健康检查、告警延迟
- **分布式系统**: 租约管理、心跳检测

## 性能优势

1. **时间复杂度**: 插入和删除都是O(1)操作
2. **内存效率**: 相比红黑树等结构占用内存更少
3. **批量处理**: 同一时间槽的任务可以批量执行
4. **扩展性**: 支持多层时间轮处理长时间延迟

## 注意事项

1. **精度权衡**: 时间精度和内存占用的权衡
2. **任务执行**: 避免在定时器回调中执行耗时操作
3. **内存管理**: 及时取消不需要的定时器
4. **线程安全**: 多线程环境下的并发访问控制
5. **停止处理**: 程序退出时正确停止时间轮

## 配置参数

- **tick**: 时间轮的最小时间单位
- **wheelSize**: 每层时间轮的槽位数量
- **levels**: 时间轮的层数（自动计算）

合理配置这些参数可以在精度和性能之间找到最佳平衡点。