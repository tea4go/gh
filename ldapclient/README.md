# LDAPClient - LDAP客户端

## 概述

LDAPClient是一个功能全面的LDAP（轻量级目录访问协议）客户端包，专门用于与Active Directory等LDAP服务器进行交互。提供用户认证、搜索、创建、修改、删除等完整的目录服务操作功能。

## 主要功能

### 1. 连接和认证
- LDAP/LDAPS连接支持
- StartTLS加密连接
- 用户绑定和认证
- 连接状态管理

### 2. 用户管理
- 用户创建、删除、搜索
- 用户账户启用/禁用
- 密码管理和修改
- 用户属性更新

### 3. 组织架构管理
- 组织单元(OU)创建和删除
- 用户组创建和删除
- 组成员管理

## 核心结构

### TLdapClient - LDAP客户端结构
```go
type TLdapClient struct {
    Addr       string   `json:"addr"`       // LDAP服务器地址
    BaseDn     string   `json:"baseDn"`     // 基础DN
    BindDn     string   `json:"bindDn"`     // 绑定DN
    BindPass   string   `json:"bindPass"`   // 绑定密码
    AuthFilter string   `json:"authFilter"` // 认证过滤器
    Attributes []string `json:"attributes"` // 查询属性
    MailDomain string   `json:"mailDomain"` // 邮箱域名
    TLS        bool     `json:"tls"`        // TLS连接
    StartTLS   bool     `json:"startTLS"`   // StartTLS
    Conn       *ldap.Conn                   // LDAP连接
}
```

### TLdapUser - LDAP用户结构
```go
type TLdapUser struct {
    StaffCode string `json:"code"`    // 员工编码
    StaffName string `json:"name"`    // 员工姓名
    Email     string `json:"mail"`    // 邮箱地址
    Org       string `json:"org"`     // 组织
    Dept      string `json:"dept"`    // 部门
    Phone     string `json:"phone"`   // 手机号
    Company   string `json:"belong"`  // 公司
    Station   string `json:"station"` // 属地
}
```

### TResultLdap - LDAP查询结果
```go
type TResultLdap struct {
    DN    string              `json:"dn"`         // 区别名
    Attrs map[string][]string `json:"attributes"` // 属性映射
}
```

## 主要功能方法

### 连接管理
```go
func (lc *TLdapClient) Connect() error                        // 连接LDAP服务器
func (lc *TLdapClient) Close()                               // 关闭连接
func (lc *TLdapClient) IsClosing() bool                      // 检查连接状态
func (lc *TLdapClient) Bind(username, password string) (bool, error) // 用户绑定
```

### 认证和搜索
```go
func (lc *TLdapClient) Auth(username, password string) (bool, error)  // 用户认证
func (lc *TLdapClient) SearchUser(username string) (*TResultLdap, error) // 搜索用户
func (lc *TLdapClient) Search(SearchFilter string, scope int) ([]*TResultLdap, error) // 通用搜索
```

### 用户管理
```go
func (lc *TLdapClient) CreateUser(path string, ldapUser *TLdapUser) (string, string, error) // 创建用户
func (lc *TLdapClient) DeleteUser(path string, staff_code string) error                    // 删除用户
func (lc *TLdapClient) EnableAccount(staff_code string) error                              // 启用账户
func (lc *TLdapClient) DisableAccount(staff_code string) error                             // 禁用账户
func (lc *TLdapClient) ChangePassword(path string, staff_code, new_password string) error // 修改密码
```

### 组织架构管理
```go
func (lc *TLdapClient) CreatePath(path string, pathName string) error        // 创建OU
func (lc *TLdapClient) DeletePath(path string, path_name string) error       // 删除OU
func (lc *TLdapClient) CreateGroup(path string, groupName string) error      // 创建用户组
func (lc *TLdapClient) DeleteGroup(path string, group_name string) error     // 删除用户组
func (lc *TLdapClient) AddGroupUser(path string, groupName, userDN string) error // 添加组成员
func (lc *TLdapClient) DelGroupUser(path string, groupName, userDN string) error // 删除组成员
```

## 使用示例

### 基本连接和认证
```go
package main

import (
    "fmt"
    "log"
    "path/to/ldapclient"
)

func main() {
    // 创建LDAP客户端
    client := &ldapclient.TLdapClient{
        Addr:       "ldap.example.com:389",
        BaseDn:     "dc=example,dc=com",
        BindDn:     "cn=admin,dc=example,dc=com",
        BindPass:   "adminpassword",
        AuthFilter: "(sAMAccountName=%s)",
        Attributes: []string{"cn", "mail", "sAMAccountName", "displayName"},
        MailDomain: "@example.com",
        TLS:        false,
        StartTLS:   true,
    }

    // 连接到LDAP服务器
    err := client.Connect()
    if err != nil {
        log.Fatal("连接失败:", err)
    }
    defer client.Close()

    // 用户认证
    success, err := client.Auth("testuser", "password123")
    if err != nil {
        log.Fatal("认证失败:", err)
    }
    if success {
        fmt.Println("用户认证成功")
    }
}
```

### 用户搜索
```go
package main

import (
    "fmt"
    "log"
    "path/to/ldapclient"
)

func main() {
    client := &ldapclient.TLdapClient{
        Addr:       "ldap.example.com:389",
        BaseDn:     "dc=example,dc=com",
        BindDn:     "cn=admin,dc=example,dc=com",
        BindPass:   "adminpassword",
        AuthFilter: "(sAMAccountName=%s)",
        Attributes: []string{"*"},
    }

    // 使用便捷函数搜索用户
    user, err := ldapclient.SearchUser(client, "testuser")
    if err != nil {
        log.Fatal("搜索失败:", err)
    }

    fmt.Printf("用户DN: %s\n", user.DN)
    fmt.Printf("用户邮箱: %s\n", user.GetAttr("mail"))
    fmt.Printf("显示名称: %s\n", user.GetAttr("displayName"))
    fmt.Printf("所有属性: %s\n", user.String())
}
```

### 创建用户
```go
package main

import (
    "fmt"
    "log"
    "path/to/ldapclient"
)

func main() {
    client := &ldapclient.TLdapClient{
        Addr:       "ldap.example.com:389",
        BaseDn:     "dc=example,dc=com",
        BindDn:     "cn=admin,dc=example,dc=com",
        BindPass:   "adminpassword",
        MailDomain: "@example.com",
    }

    err := client.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 创建新用户
    newUser := &ldapclient.TLdapUser{
        StaffCode: "user001",
        StaffName: "张三",
        Email:     "zhangsan@example.com",
        Org:       "技术部",
        Dept:      "开发部",
        Phone:     "13800138000",
        Company:   "示例公司",
        Station:   "北京",
    }

    dn, nickname, err := client.CreateUser("ou=users", newUser)
    if err != nil {
        log.Fatal("创建用户失败:", err)
    }

    fmt.Printf("用户创建成功:\n")
    fmt.Printf("DN: %s\n", dn)
    fmt.Printf("昵称: %s\n", nickname)
}
```

### 用户密码管理
```go
package main

import (
    "fmt"
    "log"
    "path/to/ldapclient"
)

func main() {
    client := &ldapclient.TLdapClient{
        Addr:     "ldap.example.com:389",
        BaseDn:   "dc=example,dc=com",
        BindDn:   "cn=admin,dc=example,dc=com",
        BindPass: "adminpassword",
    }

    err := client.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 修改用户密码
    err = client.ChangePassword("ou=users", "user001", "NewPassword123!")
    if err != nil {
        log.Fatal("密码修改失败:", err)
    }

    fmt.Println("密码修改成功")

    // 启用用户账户
    err = client.EnableAccount("user001")
    if err != nil {
        log.Fatal("账户启用失败:", err)
    }

    fmt.Println("账户启用成功")
}
```

### 组织架构管理
```go
package main

import (
    "fmt"
    "log"
    "path/to/ldapclient"
)

func main() {
    client := &ldapclient.TLdapClient{
        Addr:     "ldap.example.com:389",
        BaseDn:   "dc=example,dc=com",
        BindDn:   "cn=admin,dc=example,dc=com",
        BindPass: "adminpassword",
    }

    err := client.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 创建组织单元
    err = client.CreatePath("", "TechDepartment")
    if err != nil {
        log.Fatal("创建OU失败:", err)
    }
    fmt.Println("组织单元创建成功")

    // 创建用户组
    err = client.CreateGroup("ou=TechDepartment", "Developers")
    if err != nil {
        log.Fatal("创建组失败:", err)
    }
    fmt.Println("用户组创建成功")

    // 将用户添加到组
    userDN := "cn=user001,ou=users"
    err = client.AddGroupUser("ou=TechDepartment", "Developers", userDN)
    if err != nil {
        log.Fatal("添加组成员失败:", err)
    }
    fmt.Println("用户添加到组成功")
}
```

### 批量用户操作
```go
package main

import (
    "fmt"
    "log"
    "path/to/ldapclient"
)

func main() {
    client := &ldapclient.TLdapClient{
        Addr:       "ldap.example.com:389",
        BaseDn:     "dc=example,dc=com",
        BindDn:     "cn=admin,dc=example,dc=com",
        BindPass:   "adminpassword",
        AuthFilter: "(sAMAccountName=%s)",
        Attributes: []string{"*"},
        MailDomain: "@example.com",
    }

    err := client.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 批量创建用户
    users := []*ldapclient.TLdapUser{
        {"user001", "张三", "zhangsan", "技术部", "开发部", "13800138001", "示例公司", "北京"},
        {"user002", "李四", "lisi", "技术部", "测试部", "13800138002", "示例公司", "上海"},
        {"user003", "王五", "wangwu", "市场部", "销售部", "13800138003", "示例公司", "深圳"},
    }

    for _, user := range users {
        dn, nickname, err := client.CreateUser("ou=users", user)
        if err != nil {
            log.Printf("创建用户 %s 失败: %v", user.StaffCode, err)
            continue
        }
        
        fmt.Printf("用户 %s 创建成功: %s (%s)\n", user.StaffCode, dn, nickname)
        
        // 启用账户
        err = client.EnableAccount(user.StaffCode)
        if err != nil {
            log.Printf("启用用户 %s 失败: %v", user.StaffCode, err)
        } else {
            fmt.Printf("用户 %s 账户已启用\n", user.StaffCode)
        }
    }
}
```

## 高级搜索功能

### 不同范围的搜索
```go
package main

import (
    "fmt"
    "log"
    "path/to/ldapclient"
)

func main() {
    client := &ldapclient.TLdapClient{
        Addr:       "ldap.example.com:389",
        BaseDn:     "dc=example,dc=com",
        BindDn:     "cn=admin,dc=example,dc=com",
        BindPass:   "adminpassword",
        Attributes: []string{"cn", "mail", "sAMAccountName"},
    }

    err := client.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 搜索所有用户
    filter := "(objectClass=user)"
    
    // 基础对象搜索
    results, err := client.SearchBase(filter)
    if err != nil {
        log.Printf("基础搜索失败: %v", err)
    } else {
        fmt.Printf("基础搜索找到 %d 个结果\n", len(results))
    }

    // 单层级搜索
    results, err = client.SearchSubOne(filter)
    if err != nil {
        log.Printf("单层搜索失败: %v", err)
    } else {
        fmt.Printf("单层搜索找到 %d 个结果\n", len(results))
    }

    // 全子树搜索
    results, err = client.SearchSubAll(filter)
    if err != nil {
        log.Printf("全子树搜索失败: %v", err)
    } else {
        fmt.Printf("全子树搜索找到 %d 个结果\n", len(results))
        
        // 显示前5个结果
        for i, result := range results {
            if i >= 5 {
                break
            }
            fmt.Printf("  %d. %s (%s)\n", i+1, 
                result.GetAttr("cn"), 
                result.GetAttr("mail"))
        }
    }
}
```

## 配置和最佳实践

### 1. 连接配置
- 根据网络环境选择合适的TLS配置
- 使用专用服务账户进行绑定
- 设置合适的连接超时时间

### 2. 搜索优化
- 使用精确的过滤器减少网络传输
- 限制返回的属性列表
- 根据需要选择合适的搜索范围

### 3. 安全考虑
- 使用加密连接(LDAPS或StartTLS)
- 妥善管理绑定账户凭据
- 验证用户输入防止LDAP注入

### 4. 错误处理
- 实现连接重试机制
- 处理网络中断和超时
- 记录详细的错误日志

## 依赖库

- `github.com/go-ldap/ldap/v3` - LDAP协议支持
- `github.com/mozillazg/go-pinyin` - 中文拼音转换
- `golang.org/x/text/encoding/unicode` - Unicode编码支持

## 注意事项

1. **连接管理**: 及时关闭连接避免资源泄漏
2. **权限控制**: 确保绑定账户具有必要的操作权限
3. **密码策略**: 遵循组织的密码复杂性要求
4. **时间转换**: Active Directory时间格式需要特殊处理
5. **字符编码**: 密码需要UTF-16编码后传输
6. **并发安全**: 连接对象不是线程安全的，需要同步访问

## 故障排除

常见问题及解决方案：

- **连接失败**: 检查网络连通性和端口开放情况
- **认证失败**: 验证绑定DN和密码的正确性
- **搜索无结果**: 检查BaseDN和过滤器的正确性
- **权限不足**: 确保绑定账户有足够的权限
- **编码问题**: 确保字符串使用正确的编码格式