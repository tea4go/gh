# GH - Go语言企业级工具库

## 项目概述

GH是一个功能全面的Go语言企业级工具库集合，为现代应用程序开发提供了从基础工具到复杂系统集成的完整解决方案。该项目涵盖了数据存储、身份认证、消息通信、系统监控、网络工具等企业级应用开发的各个方面。

## 🎯 核心特性

- **全面性** - 17个功能模块，覆盖企业级应用开发的各个领域
- **企业级** - 支持高并发、分布式架构、安全认证、监控告警等企业需求
- **跨平台** - 支持Windows、Linux、macOS等主流操作系统
- **高性能** - 经过优化的算法和数据结构，满足高性能应用需求
- **易用性** - 详细的文档和使用示例，简化开发过程
- **可扩展** - 模块化设计，支持独立使用或组合使用

## 📦 模块架构

### 🔗 通信集成
| 模块 | 功能描述 | 适用场景 |
|------|----------|----------|
| [dingtalk](dingtalk/README.md) | 钉钉API集成包，支持小程序、群机器人、OAuth2授权 | 企业办公自动化、内部通知系统 |
| [weixin](weixin/README.md) | 微信生态API集成，支持公众号、小程序、企业微信 | 移动应用后端、营销推广、客户服务 |

### 🗄️ 数据存储
| 模块 | 功能描述 | 适用场景 |
|------|----------|----------|
| [etcdclient](etcdclient/README.md) | etcd键值存储客户端，支持分布式配置管理 | 微服务架构、配置中心、集群管理 |
| [redisclient](redisclient/README.md) | Redis客户端，支持连接池、集群、发布订阅 | 高性能缓存、会话管理、消息队列 |
| [nustdbclient](nustdbclient/README.md) | NutsDB嵌入式键值数据库客户端 | 嵌入式应用、配置存储、临时数据存储 |

### 🔐 身份认证
| 模块 | 功能描述 | 适用场景 |
|------|----------|----------|
| [ldapclient](ldapclient/README.md) | LDAP客户端，支持AD集成、用户管理 | 企业身份认证、统一用户管理、SSO集成 |
| [ldapserver](ldapserver/README.md) | LDAP服务器实现，支持TCP/TLS连接 | 自定义LDAP服务、目录服务代理 |
| [radius](radius/README.md) | RADIUS协议实现，支持网络设备AAA认证 | 网络设备管理、WiFi热点认证 |

### 📊 监控日志
| 模块 | 功能描述 | 适用场景 |
|------|----------|----------|
| [log4go](log4go/README.md) | 多级别、高性能日志系统 | 应用日志记录、系统监控、错误追踪 |
| [syslog](syslog/README.md) | 系统日志协议实现，支持远程日志传输 | 系统监控、合规性审计、日志集中化 |
| [fsnotify](fsnotify/README.md) | 跨平台文件系统事件监控 | 配置文件监控、热加载、自动同步 |

### 🌐 网络工具
| 模块 | 功能描述 | 适用场景 |
|------|----------|----------|
| [network](network/README.md) | 网络诊断工具集，包括Ping、DNS查询等 | 网络管理工具、监控系统、网络诊断 |
| [tcping](tcping/README.md) | TCP连接测试工具，支持延迟和可用性测试 | 网络故障诊断、服务监控 |

### 🛠️ 系统工具
| 模块 | 功能描述 | 适用场景 |
|------|----------|----------|
| [execplus](execplus/README.md) | 增强版进程执行器，支持跨平台命令执行 | 系统工具开发、DevOps自动化 |
| [timewheel](timewheel/README.md) | 高效时间轮算法实现，用于定时任务管理 | 超时管理、缓存过期、任务调度 |
| [image](image/README.md) | 图像处理工具，支持格式转换和EXIF提取 | Web应用图片处理、内容管理系统 |

### 🧰 基础工具
| 模块 | 功能描述 | 适用场景 |
|------|----------|----------|
| [utils](utils/README.md) | 综合工具函数库，包含加密、缓存、网络等功能 | 各种应用程序的基础工具支持 |

## 🚀 快速开始

### 安装

```bash
go get -u github.com/tea4go/gh
```

### 基本使用

```go
import (
    "github.com/tea4go/gh/log4go"
    "github.com/tea4go/gh/redisclient" 
    "github.com/tea4go/gh/utils"
)

func main() {
    // 日志记录
    logs.Info("应用启动...")
    
    // Redis操作
    client := redisclient.GetInstance()
    client.Set("key", "value", 3600)
    
    // 工具函数
    guid := utils.GenerateGUID()
    logs.Info("生成GUID: %s", guid)
}
```

## 🏗️ 应用场景

### 企业级Web应用
- 使用 `log4go` 进行日志记录
- 使用 `redisclient` 进行会话管理
- 使用 `dingtalk/weixin` 进行消息推送
- 使用 `utils` 提供基础工具支持

### 微服务架构
- 使用 `etcdclient` 进行服务发现和配置管理
- 使用 `log4go` 进行分布式日志记录
- 使用 `network` 进行健康检查
- 使用 `timewheel` 进行超时控制

### 系统监控平台
- 使用 `fsnotify` 监控配置文件变化
- 使用 `syslog` 收集系统日志
- 使用 `tcping` 进行连通性测试
- 使用 `network` 进行网络诊断

### 自动化工具
- 使用 `execplus` 执行系统命令
- 使用 `image` 进行图像批处理
- 使用 `utils` 提供各种工具函数

## 📋 模块分类

### 按功能分类

**通信集成** (2个模块)
- dingtalk - 钉钉集成
- weixin - 微信集成

**数据存储** (3个模块)
- etcdclient - 分布式键值存储
- redisclient - 内存数据库
- nustdbclient - 嵌入式数据库

**身份认证** (3个模块)
- ldapclient - LDAP客户端
- ldapserver - LDAP服务器
- radius - RADIUS认证

**监控日志** (3个模块)
- log4go - 应用日志
- syslog - 系统日志
- fsnotify - 文件监控

**网络工具** (2个模块)
- network - 网络诊断
- tcping - TCP测试

**系统工具** (3个模块)
- execplus - 进程执行
- timewheel - 定时任务
- image - 图像处理

**基础工具** (1个模块)
- utils - 工具函数集

### 按使用频率分类

**核心模块** - 大多数应用都会使用
- log4go, utils, redisclient

**企业模块** - 企业级应用常用
- dingtalk, weixin, ldapclient, etcdclient

**专用模块** - 特定场景使用
- radius, ldapserver, timewheel, image

**诊断模块** - 运维和诊断使用
- network, tcping, fsnotify, syslog

## 🔧 开发指南

### 代码规范
- 遵循Go语言官方编码规范
- 使用gofmt格式化代码
- 添加必要的注释和文档
- 编写单元测试

### 错误处理
- 统一的错误处理机制
- 详细的错误信息
- 支持错误链传递

### 性能优化
- 连接池管理
- 缓存机制
- 并发控制
- 内存优化

### 安全考虑
- 加密传输支持
- 输入参数验证
- SQL注入防护
- 权限控制

## 📈 版本历史

- **v1.0.0** - 初始版本，包含基础功能模块
- **v1.1.0** - 新增微信和钉钉集成模块
- **v1.2.0** - 完善LDAP和RADIUS认证功能
- **v1.3.0** - 优化性能，增加更多工具函数

## 🤝 贡献指南

我们欢迎社区贡献！请遵循以下流程：

1. Fork项目仓库
2. 创建功能分支
3. 编写代码和测试
4. 提交Pull Request
5. 等待代码审核

## 📄 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](LICENSE) 文件。

## 📞 支持与反馈

- 🐛 **Bug报告**: 请在GitHub Issues中提交
- 💡 **功能建议**: 欢迎提交Enhancement请求
- 📚 **文档完善**: 帮助我们改进文档
- 💬 **技术讨论**: 欢迎在Discussions中讨论

## 🌟 致谢

感谢所有贡献者的努力和社区的支持！

---

> **注意**: 本工具库持续更新中，建议关注项目动态获取最新功能和安全更新。

**技术栈标签**: `Go` `企业级` `工具库` `微服务` `分布式` `高性能` `跨平台`