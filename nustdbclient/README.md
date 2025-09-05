# NutsDBClient - NutsDB客户端

## 概述

NutsDBClient是NutsDB数据库的客户端包装库，提供了简化的接口来操作NutsDB嵌入式键值数据库。支持多种数据结构操作，包括字符串、列表、集合等。

## 主要功能

### 1. 基础操作
- 数据库连接管理
- 键值对操作
- 事务支持
- 批量操作

### 2. 数据结构支持
- String: 字符串操作
- List: 列表操作  
- Set: 集合操作
- Hash: 哈希表操作

### 3. 高级功能
- TTL过期设置
- 范围查询
- 索引操作
- 备份恢复

## 使用示例

### 基本操作
```go
package main

import (
    "fmt"
    "path/to/nustdbclient"
)

func main() {
    // 连接数据库
    client, err := nustdbclient.Open("/path/to/db")
    if err != nil {
        fmt.Printf("连接失败: %v\n", err)
        return
    }
    defer client.Close()
    
    // 设置键值
    err = client.Set("key1", "value1")
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

## 适用场景

- 嵌入式应用
- 本地缓存
- 配置存储
- 临时数据存储

## 注意事项

1. **文件锁定**: 同一时间只能有一个进程访问数据库
2. **数据持久化**: 数据存储在本地文件系统
3. **内存管理**: 合理控制缓存大小
4. **备份策略**: 定期备份数据库文件