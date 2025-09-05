# LDAPServer - LDAP服务器

## 概述

LDAPServer是一个完整的LDAP（轻量级目录访问协议）服务器实现，支持TCP和TLS连接，提供路由管理、客户端访问控制、并发处理等企业级功能。基于Go语言开发，支持自定义处理器来实现各种LDAP操作。

## 主要功能

### 1. 服务器管理
- TCP/TLS双重连接支持
- 优雅启动和关闭
- 客户端连接管理
- 代理协议支持
- 白名单访问控制

### 2. 路由系统
- 基于操作类型的路由匹配
- 支持复杂的过滤条件
- 自定义处理函数
- 未匹配请求处理

### 3. 并发处理
- 每客户端独立协程
- 消息队列管理
- 请求生命周期管理
- 优雅的连接关闭

## 核心结构

### Server - LDAP服务器
```go
type Server struct {
    Listener      net.Listener
    proxyListener *proxyproto.Listener
    ReadTimeout   time.Duration
    WriteTimeout  time.Duration
    wg            sync.WaitGroup
    chDone        chan bool
    client_nets   []net.IPNet
    client_ips    sync.Map
    onNewConnection func(c net.Conn) error
    Handler       Handler
}
```

### RouteMux - 路由多路复用器
```go
type RouteMux struct {
    routes        []*route
    notFoundRoute *route
}
```

### route - 路由规则
```go
type route struct {
    label       string      // 路由标签
    operation   string      // 操作类型
    handler     HandlerFunc // 处理函数
    exoName     string      // 扩展请求名称
    sBasedn     string      // 基础DN
    sFilter     string      // 搜索过滤器
    sScope      int         // 搜索范围
    sAuthChoice string      // 认证选择
}
```

## 主要接口和方法

### 服务器管理
```go
func NewServer() *Server                                                          // 创建新服务器
func (s *Server) Handle(h Handler)                                               // 注册处理器
func (s *Server) ListenAndServe(addr string, ch chan error, options ...func(*Server)) // HTTP服务
func (s *Server) ListenAndServeTLS(addr, certFile, keyFile string, ch chan error, options ...func(*Server)) // HTTPS服务
func (s *Server) Stop()                                                          // 停止服务器
```

### 客户端访问控制
```go
func (s *Server) SetClients(ip_text string) error     // 设置允许的客户端IP
func (s *Server) CheckClient(ip_text string) bool     // 检查客户端是否被允许
func (s *Server) GetClients() string                  // 获取客户端列表
func (s *Server) GetClientIPs() *sync.Map             // 获取客户端IP计数
```

### 路由管理
```go
func NewRouteMux() *RouteMux                          // 创建路由器
func (h *RouteMux) Bind(handler HandlerFunc) *route   // 绑定请求处理
func (h *RouteMux) Search(handler HandlerFunc) *route // 搜索请求处理
func (h *RouteMux) Add(handler HandlerFunc) *route    // 添加请求处理
func (h *RouteMux) Modify(handler HandlerFunc) *route // 修改请求处理
func (h *RouteMux) Delete(handler HandlerFunc) *route // 删除请求处理
func (h *RouteMux) Compare(handler HandlerFunc) *route // 比较请求处理
func (h *RouteMux) Extended(handler HandlerFunc) *route // 扩展请求处理
func (h *RouteMux) NotFound(handler HandlerFunc) *route // 未找到处理
```

### 路由条件设置
```go
func (r *route) Label(label string) *route                    // 设置标签
func (r *route) BaseDn(dn string) *route                     // 设置基础DN
func (r *route) Filter(pattern string) *route                // 设置过滤器
func (r *route) Scope(scope int) *route                      // 设置搜索范围
func (r *route) AuthenticationChoice(choice string) *route   // 设置认证类型
func (r *route) RequestName(name ldap.LDAPOID) *route        // 设置请求名称
```

## 使用示例

### 基本LDAP服务器
```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "path/to/ldapserver"
)

func main() {
    // 创建服务器
    server := ldapserver.NewServer()
    
    // 创建路由器
    routes := ldapserver.NewRouteMux()
    
    // 注册绑定处理器
    routes.Bind(handleBind).Label("Bind Handler")
    
    // 注册搜索处理器  
    routes.Search(handleSearch).Label("Search Handler")
    
    // 注册处理器到服务器
    server.Handle(routes)
    
    // 设置允许的客户端IP
    err := server.SetClients("127.0.0.1;192.168.1.0/24")
    if err != nil {
        log.Fatal("设置客户端白名单失败:", err)
    }
    
    // 启动服务器
    ch := make(chan error)
    go server.ListenAndServe(":389", ch)
    
    if err := <-ch; err != nil {
        log.Fatal("服务器启动失败:", err)
    }
    
    fmt.Println("LDAP服务器启动成功，监听端口 :389")
    
    // 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    fmt.Println("正在关闭服务器...")
    server.Stop()
    fmt.Println("服务器已关闭")
}

// 绑定请求处理器
func handleBind(w ldapserver.ResponseWriter, r *ldapserver.Message) {
    res := ldapserver.NewBindResponse(ldapserver.LDAPResultSuccess)
    w.Write(res)
}

// 搜索请求处理器
func handleSearch(w ldapserver.ResponseWriter, r *ldapserver.Message) {
    res := ldapserver.NewSearchResultDoneResponse(ldapserver.LDAPResultSuccess)
    w.Write(res)
}
```

### 带TLS支持的LDAP服务器
```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "path/to/ldapserver"
)

func main() {
    server := ldapserver.NewServer()
    routes := ldapserver.NewRouteMux()
    
    // 设置处理器
    routes.Bind(handleBind)
    routes.Search(handleSearch)
    routes.NotFound(handleNotFound)
    
    server.Handle(routes)
    
    // 设置客户端白名单
    server.SetClients("0.0.0.0/0") // 允许所有IP
    
    // 启动TLS服务器
    ch := make(chan error)
    go server.ListenAndServeTLS(":636", "cert.pem", "key.pem", ch)
    
    if err := <-ch; err != nil {
        log.Fatal("TLS服务器启动失败:", err)
    }
    
    fmt.Println("LDAP TLS服务器启动成功，监听端口 :636")
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    server.Stop()
}

func handleBind(w ldapserver.ResponseWriter, r *ldapserver.Message) {
    bindRequest := r.ProtocolOp().(ldap.BindRequest)
    
    // 简单的认证逻辑
    if string(bindRequest.Name()) == "admin" {
        res := ldapserver.NewBindResponse(ldapserver.LDAPResultSuccess)
        w.Write(res)
    } else {
        res := ldapserver.NewBindResponse(ldapserver.LDAPResultInvalidCredentials)
        w.Write(res)
    }
}

func handleSearch(w ldapserver.ResponseWriter, r *ldapserver.Message) {
    // 返回搜索完成响应
    res := ldapserver.NewSearchResultDoneResponse(ldapserver.LDAPResultSuccess)
    w.Write(res)
}

func handleNotFound(w ldapserver.ResponseWriter, r *ldapserver.Message) {
    res := ldapserver.NewResponse(ldapserver.LDAPResultOperationsError)
    res.SetDiagnosticMessage("不支持的操作")
    w.Write(res)
}
```

### 复杂路由配置
```go
package main

import (
    "fmt"
    "path/to/ldapserver"
)

func main() {
    server := ldapserver.NewServer()
    routes := ldapserver.NewRouteMux()
    
    // 用户认证路由
    routes.Bind(handleUserAuth).
        Label("User Authentication").
        AuthenticationChoice("simple")
    
    // 根据BaseDN的搜索路由
    routes.Search(handleUserSearch).
        Label("User Search").
        BaseDn("ou=users,dc=example,dc=com").
        Filter("(objectclass=person)")
    
    routes.Search(handleGroupSearch).
        Label("Group Search").
        BaseDn("ou=groups,dc=example,dc=com").
        Scope(1) // 单层搜索
    
    // 添加用户操作
    routes.Add(handleAddUser).
        Label("Add User")
    
    // 修改用户操作
    routes.Modify(handleModifyUser).
        Label("Modify User")
    
    // 删除用户操作
    routes.Delete(handleDeleteUser).
        Label("Delete User")
    
    server.Handle(routes)
    
    ch := make(chan error)
    go server.ListenAndServe(":389", ch)
    
    if err := <-ch; err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("复杂路由LDAP服务器启动成功")
    select {} // 保持运行
}

func handleUserAuth(w ldapserver.ResponseWriter, r *ldapserver.Message) {
    bindReq := r.ProtocolOp().(ldap.BindRequest)
    
    username := string(bindReq.Name())
    password := string(bindReq.AuthenticationSimple())
    
    // 实际认证逻辑
    if authenticateUser(username, password) {
        res := ldapserver.NewBindResponse(ldapserver.LDAPResultSuccess)
        w.Write(res)
    } else {
        res := ldapserver.NewBindResponse(ldapserver.LDAPResultInvalidCredentials)
        res.SetDiagnosticMessage("用户名或密码错误")
        w.Write(res)
    }
}

func handleUserSearch(w ldapserver.ResponseWriter, r *ldapserver.Message) {
    searchReq := r.ProtocolOp().(ldap.SearchRequest)
    
    // 获取搜索参数
    basedn := string(searchReq.BaseObject())
    filter := searchReq.FilterString()
    scope := int(searchReq.Scope())
    
    fmt.Printf("搜索请求 - BaseDN: %s, Filter: %s, Scope: %d\n", 
               basedn, filter, scope)
    
    // 执行搜索并返回结果
    // ... 搜索逻辑 ...
    
    res := ldapserver.NewSearchResultDoneResponse(ldapserver.LDAPResultSuccess)
    w.Write(res)
}

func handleGroupSearch(w ldapserver.ResponseWriter, r *ldapserver.Message) {
    // 处理组搜索
    res := ldapserver.NewSearchResultDoneResponse(ldapserver.LDAPResultSuccess)
    w.Write(res)
}

func handleAddUser(w ldapserver.ResponseWriter, r *ldapserver.Message) {
    res := ldapserver.NewAddResponse(ldapserver.LDAPResultSuccess)
    w.Write(res)
}

func handleModifyUser(w ldapserver.ResponseWriter, r *ldapserver.Message) {
    res := ldapserver.NewModifyResponse(ldapserver.LDAPResultSuccess)
    w.Write(res)
}

func handleDeleteUser(w ldapserver.ResponseWriter, r *ldapserver.Message) {
    res := ldapserver.NewDelResponse(ldapserver.LDAPResultSuccess)
    w.Write(res)
}

func authenticateUser(username, password string) bool {
    // 实际的用户认证逻辑
    return username == "admin" && password == "secret"
}
```

### 带中间件的服务器
```go
package main

import (
    "fmt"
    "log"
    "time"
    "path/to/ldapserver"
)

func main() {
    server := ldapserver.NewServer()
    
    // 设置读写超时
    server.ReadTimeout = time.Second * 30
    server.WriteTimeout = time.Second * 30
    
    routes := ldapserver.NewRouteMux()
    
    // 使用中间件包装处理器
    routes.Bind(withLogging(withAuth(handleBind)))
    routes.Search(withLogging(handleSearch))
    
    server.Handle(routes)
    server.SetClients("0.0.0.0/0")
    
    ch := make(chan error)
    go server.ListenAndServe(":389", ch)
    
    if err := <-ch; err != nil {
        log.Fatal(err)
    }
    
    select {} // 保持运行
}

// 日志中间件
func withLogging(next ldapserver.HandlerFunc) ldapserver.HandlerFunc {
    return func(w ldapserver.ResponseWriter, r *ldapserver.Message) {
        start := time.Now()
        
        fmt.Printf("[%s] 开始处理 %s 请求，来自 %s\n", 
                   time.Now().Format("15:04:05"), 
                   r.ProtocolOpName(), 
                   r.Client.Addr())
        
        next(w, r)
        
        fmt.Printf("[%s] 完成处理 %s 请求，耗时 %v\n", 
                   time.Now().Format("15:04:05"), 
                   r.ProtocolOpName(), 
                   time.Since(start))
    }
}

// 认证中间件
func withAuth(next ldapserver.HandlerFunc) ldapserver.HandlerFunc {
    return func(w ldapserver.ResponseWriter, r *ldapserver.Message) {
        // 简单的IP验证
        clientIP := r.Client.Addr().String()
        if !isAllowedIP(clientIP) {
            res := ldapserver.NewBindResponse(ldapserver.LDAPResultInsufficientAccessRights)
            res.SetDiagnosticMessage("客户端IP未授权")
            w.Write(res)
            return
        }
        
        next(w, r)
    }
}

func handleBind(w ldapserver.ResponseWriter, r *ldapserver.Message) {
    res := ldapserver.NewBindResponse(ldapserver.LDAPResultSuccess)
    w.Write(res)
}

func handleSearch(w ldapserver.ResponseWriter, r *ldapserver.Message) {
    res := ldapserver.NewSearchResultDoneResponse(ldapserver.LDAPResultSuccess)
    w.Write(res)
}

func isAllowedIP(ip string) bool {
    // 实际的IP验证逻辑
    return true
}
```

## 支持的LDAP操作

### 1. 绑定操作 (Bind)
- 简单认证
- SASL认证
- 匿名绑定

### 2. 搜索操作 (Search)
- 基础对象搜索
- 单层级搜索  
- 全子树搜索
- 复杂过滤器支持

### 3. 修改操作
- 添加条目 (Add)
- 修改条目 (Modify)
- 删除条目 (Delete)
- 修改DN (ModifyDN)

### 4. 比较操作 (Compare)
- 属性值比较

### 5. 扩展操作 (Extended)
- StartTLS支持
- 自定义扩展操作

### 6. 放弃操作 (Abandon)
- 取消正在进行的操作

## 配置和优化

### 1. 性能配置
```go
server.ReadTimeout = time.Second * 30
server.WriteTimeout = time.Second * 30
```

### 2. 安全配置
```go
// 设置客户端白名单
server.SetClients("192.168.1.0/24;10.0.0.0/8")

// TLS配置
server.ListenAndServeTLS(":636", "cert.pem", "key.pem", ch)
```

### 3. 监控配置
```go
// 获取连接统计
clientIPs := server.GetClientIPs()
clientIPs.Range(func(key, value interface{}) bool {
    fmt.Printf("客户端 %s: %d 次连接\n", key, value)
    return true
})
```

## 依赖库

- `github.com/openstandia/goldap/message` - LDAP协议消息处理
- `github.com/pires/go-proxyproto` - 代理协议支持

## 最佳实践

1. **错误处理**: 为所有操作提供详细的错误响应
2. **资源管理**: 及时关闭连接和释放资源
3. **安全考虑**: 使用TLS加密和客户端验证
4. **性能优化**: 合理设置超时时间和并发限制
5. **日志记录**: 记录详细的操作日志便于调试

## 注意事项

1. **并发安全**: 所有公开方法都是并发安全的
2. **协议兼容**: 完全兼容LDAP v3协议
3. **内存管理**: 大量客户端连接时注意内存使用
4. **网络处理**: 正确处理网络异常和超时
5. **优雅关闭**: 使用Stop()方法优雅关闭服务器