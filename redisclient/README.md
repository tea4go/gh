# RedisClient - Redis客户端

## 概述

RedisClient是Redis数据库的客户端包装库，提供了简化的接口来操作Redis服务器。支持连接池管理、集群操作、发布订阅等高级功能。

## 主要功能

### 1. 连接管理
- 单实例连接
- 连接池管理
- 集群连接支持
- 哨兵模式支持

### 2. 数据操作
- String操作
- Hash操作
- List操作  
- Set操作
- Sorted Set操作
- 过期时间管理

### 3. 高级功能
- 事务支持
- 管道操作
- 发布/订阅
- Lua脚本执行
- 分布式锁

## 使用示例

### 基本操作
```go
package main

import (
    "fmt"
    "path/to/redisclient"
)

func main() {
    // 连接Redis
    client := redisclient.NewClient(&redisclient.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })
    defer client.Close()
    
    // 设置键值
    err := client.Set("key1", "value1", 0)
    if err != nil {
        fmt.Printf("设置失败: %v\n", err)
        return
    }
    
    // 获取值
    value, err := client.Get("key1")
    if err != nil {
        fmt.Printf("获取失败: %v\n", err)
        return
    }
    
    fmt.Printf("值: %s\n", value)
}
```

### 连接池使用
```go
package main

import (
    "fmt"
    "time"
    "path/to/redisclient"
)

func main() {
    // 创建连接池
    pool := redisclient.NewPool(&redisclient.PoolOptions{
        MaxIdle:     10,
        MaxActive:   100,
        IdleTimeout: time.Minute * 5,
        Addr:        "localhost:6379",
    })
    defer pool.Close()
    
    // 从池中获取连接
    conn := pool.Get()
    defer conn.Close()
    
    // 使用连接
    reply, err := conn.Do("SET", "key1", "value1")
    if err != nil {
        fmt.Printf("操作失败: %v\n", err)
        return
    }
    
    fmt.Printf("结果: %v\n", reply)
}
```

### 发布订阅
```go
package main

import (
    "fmt"
    "path/to/redisclient"
)

func main() {
    client := redisclient.NewClient(&redisclient.Options{
        Addr: "localhost:6379",
    })
    defer client.Close()
    
    // 订阅频道
    pubsub := client.Subscribe("channel1")
    defer pubsub.Close()
    
    // 监听消息
    ch := pubsub.Channel()
    for msg := range ch {
        fmt.Printf("收到消息: %s\n", msg.Payload)
    }
}
```

### 分布式锁
```go
package main

import (
    "fmt"
    "time"
    "path/to/redisclient"
)

func main() {
    client := redisclient.NewClient(&redisclient.Options{
        Addr: "localhost:6379",
    })
    defer client.Close()
    
    // 获取分布式锁
    lock := client.NewLock("lock_key", time.Second*10)
    
    // 尝试加锁
    if lock.Lock() {
        fmt.Println("获取锁成功")
        
        // 执行业务逻辑
        time.Sleep(time.Second * 2)
        
        // 释放锁
        lock.Unlock()
        fmt.Println("释放锁成功")
    } else {
        fmt.Println("获取锁失败")
    }
}
```

## 适用场景

- 缓存存储
- 会话管理
- 分布式锁
- 消息队列
- 计数器
- 排行榜
- 实时数据分析

## 注意事项

1. **连接管理**: 合理配置连接池大小
2. **内存使用**: 监控Redis内存使用情况
3. **数据持久化**: 配置适当的持久化策略
4. **网络延迟**: 考虑网络延迟对性能的影响
5. **错误处理**: 处理网络中断等异常情况