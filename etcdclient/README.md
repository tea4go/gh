# etcdclient 包文档

## 包概述

`etcdclient` 包提供了对etcd键值存储的Go语言封装，支持单例模式、键值操作、前缀查询、范围查询等功能。

## 文件结构

- `etcdclient.go` - etcd客户端封装实现

## 常量定义

```go
const (
    ConnTimeout = time.Second * 3  // 连接超时时间
    OperTimeout = time.Second * 5  // 操作超时时间
)
```

## 主要结构体

### TEtcdClient
```go
type TEtcdClient struct {
    client *clientv3.Client  // etcd客户端实例
    head   string           // 键前缀
}
```
**功能**: etcd客户端封装类，提供完整的键值存储操作接口。

## 单例模式函数

**InitInstance()** `*TEtcdClient`
- **功能**: 初始化单例实例，连接到默认地址127.0.0.1:2379
- **返回**: etcd客户端实例
- **说明**: 一般用于程序启动时初始化

**GetInstance()** `*TEtcdClient`
- **功能**: 获取单例实例（非线程安全）
- **返回**: etcd客户端实例
- **说明**: 简单获取实例，不考虑线程安全

**GetSafeInstance()** `*TEtcdClient`
- **功能**: 获取线程安全的单例实例
- **返回**: etcd客户端实例
- **说明**: 使用sync.Once确保线程安全

## 连接函数

**Connect(etcdAddrs []string)** `(*TEtcdClient, error)`
- **功能**: 连接到指定的etcd集群
- **参数**: `etcdAddrs` - etcd集群地址列表
- **返回**: etcd客户端实例和错误信息
- **默认前缀**: "tea4go"

## 配置方法

**GetHead()** `string`
- **功能**: 获取当前键前缀

**SetHead(head string)**
- **功能**: 设置键前缀
- **参数**: `head` - 前缀字符串
- **说明**: 自动处理前缀格式，确保以"/"开头和结尾

## 键值操作方法

### 设置键值

**Set(key, value string, args ...int)** `error`
- **功能**: 设置键值对
- **参数**: 
  - `key` - 键名
  - `value` - 值
  - `args` - 可选的过期时间（秒）
- **返回**: 错误信息
- **说明**: 支持设置TTL过期时间

### 获取单个键

**Get(key string)** `(string, error)`
- **功能**: 获取单个键的值
- **参数**: `key` - 键名
- **返回**: 键值和错误信息

### 获取多个键

**GetAll(prefix string)** `(map[string]string, int64, error)`
- **功能**: 根据前缀获取所有匹配的键值对
- **参数**: `prefix` - 键前缀
- **返回**: 键值对map、总数量和错误信息

**GetLimit(prefix string, limit int)** `(map[string]string, int64, error)`
- **功能**: 根据前缀获取限制数量的键值对
- **参数**: 
  - `prefix` - 键前缀
  - `limit` - 返回数量限制
- **返回**: 键值对map、总数量和错误信息

**GetRange(startKey, endKey string)** `(map[string]string, int64, error)`
- **功能**: 获取指定范围内的键值对 [startKey, endKey)
- **参数**: 
  - `startKey` - 起始键
  - `endKey` - 结束键（不包含）
- **返回**: 键值对map、总数量和错误信息

**GetRangeLimit(startKey string, limit int)** `(map[string]string, int64, error)`
- **功能**: 从指定键开始获取限制数量的键值对
- **参数**: 
  - `startKey` - 起始键
  - `limit` - 返回数量限制
- **返回**: 键值对map、总数量和错误信息

### 查询功能

**GetMaxKey(prefix string)** `(string, error)`
- **功能**: 获取指定前缀下的最大键（按字典序）
- **参数**: `prefix` - 键前缀
- **返回**: 最大键名和错误信息
- **用途**: 常用于获取最大ID，如Key_001...Key_102中的Key_102

**Count(prefix string)** `(int64, error)`
- **功能**: 统计指定前缀下的键数量
- **参数**: `prefix` - 键前缀
- **返回**: 键数量和错误信息

### 删除操作

**Del(key string)** `(int64, error)`
- **功能**: 删除单个键
- **参数**: `key` - 键名
- **返回**: 删除的键数量和错误信息

**DelAll(prefix string)** `(int64, error)`
- **功能**: 删除指定前缀的所有键
- **参数**: `prefix` - 键前缀
- **返回**: 删除的键数量和错误信息

## 高级功能

**HGet(getResp *clientv3.GetResponse, err error)** `(map[string]string, int64, error)`
- **功能**: 内部方法，将etcd响应转换为map结构
- **参数**: 
  - `getResp` - etcd获取响应
  - `err` - 错误信息
- **返回**: 键值对map、总数量和错误信息
- **说明**: 自动去除键前缀，注意map不保证顺序

## 原生接口访问

**GetClient()** `*clientv3.Client`
- **功能**: 获取原生etcd客户端实例
- **返回**: etcd v3客户端
- **用途**: 需要使用原生API时

**GetKV()** `clientv3.KV`
- **功能**: 获取原生键值接口
- **返回**: etcd KV接口
- **用途**: 直接进行键值操作

## 使用示例

### 基本使用
```go
// 连接etcd
client, err := etcdclient.Connect([]string{"127.0.0.1:2379", "127.0.0.1:2380"})
if err != nil {
    log.Fatal(err)
}

// 设置前缀
client.SetHead("myapp")

// 设置键值
err = client.Set("user:1001", "john", 3600) // 1小时后过期
if err != nil {
    log.Printf("设置失败: %v", err)
}

// 获取单个键
value, err := client.Get("user:1001")
if err != nil {
    log.Printf("获取失败: %v", err)
}
fmt.Printf("用户信息: %s\n", value)
```

### 使用单例模式
```go
// 获取线程安全的单例
client := etcdclient.GetSafeInstance()

// 批量获取用户信息
users, count, err := client.GetAll("user:")
if err != nil {
    log.Printf("批量获取失败: %v", err)
}
fmt.Printf("找到 %d 个用户\n", count)
for k, v := range users {
    fmt.Printf("用户 %s: %s\n", k, v)
}
```

### 分页查询
```go
// 获取前10个用户
users, count, err := client.GetLimit("user:", 10)

// 从指定键开始获取5个
moreUsers, _, err := client.GetRangeLimit("user:1000", 5)
```

### 删除操作
```go
// 删除单个用户
deleted, err := client.Del("user:1001")
fmt.Printf("删除了 %d 个键\n", deleted)

// 删除所有用户
allDeleted, err := client.DelAll("user:")
fmt.Printf("批量删除了 %d 个键\n", allDeleted)
```

## 注意事项

1. 客户端支持单例模式，便于全局使用
2. 所有操作都设置了超时时间，避免长时间阻塞
3. 键前缀会自动格式化为标准格式（以/开头和结尾）
4. HGet方法返回的map不保证顺序，如需有序请使用原生接口
5. 支持TTL过期时间设置
6. 提供原生接口访问，满足高级需求