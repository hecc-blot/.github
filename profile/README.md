# Hecc-Blot

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/License-MIT-green)](https://github.com/hecc-blot/.github/blob/main/LICENSE)

**Hecc-Blot** 是一个基于 Go 的轻量级后端框架，采用**面向接口 + IOC 容器**组装。框架能力按模块拆分为独立仓库，通过 `go get` 按需引入。

## 框架特性

- **面向接口** — 所有组件通过接口契约交互，实现可替换
- **依赖注入** — 基于反射的 IOC 容器，`inject` tag 自动注入
- **统一响应** — 返回值自动包装为 `{code, message, data}`
- **参数校验** — 自动绑定与校验，支持自定义错误信息
- **多数据库** — MySQL / PostgreSQL / ClickHouse / Elasticsearch / MongoDB
- **双层缓存** — 本地内存 + Redis
- **链路追踪** — OpenTelemetry，OTLP 导出
- **可观测** — 统一日志、Prometheus 指标、SSE 实时推送
- **高可用** — 限流 / 熔断 / 幂等 / 分布式锁 / 定时调度

## 快速开始

每个模块是独立仓库，`go get` 引入后通过 IOC 容器组装：

```bash
go get github.com/hecc-blot/framework
```

各模块的接口定义、配置项、用法见对应仓库 README。

## 模块仓库

| 模块 | 职责 |
|------|------|
| [core](https://github.com/hecc-blot/core) | 最底层公共库：通用接口契约 + 默认实现（日志 zap） |
| [framework](https://github.com/hecc-blot/framework) | 框架内核：接口契约 + IOC 容器 + HTTP 内核 + 统一响应 |
| [db](https://github.com/hecc-blot/db) | 关系型数据库（GORM MySQL/PostgreSQL） |
| [db-clickhouse](https://github.com/hecc-blot/db-clickhouse) | ClickHouse 分析型数据库 |
| [db-es](https://github.com/hecc-blot/db-es) | Elasticsearch 搜索型数据库 |
| [db-mongo](https://github.com/hecc-blot/db-mongo) | MongoDB 文档型数据库 |
| [cache](https://github.com/hecc-blot/cache) | 双层缓存（本地 + Redis） |
| [trace](https://github.com/hecc-blot/trace) | 链路追踪（OpenTelemetry） |
| [log-sls](https://github.com/hecc-blot/log-sls) | 日志（阿里云 SLS） |
| [httpclient](https://github.com/hecc-blot/httpclient) | 统一 HTTP 客户端 |
| [mq](https://github.com/hecc-blot/mq) | 消息队列（Kafka/NSQ） |
| [scheduler](https://github.com/hecc-blot/scheduler) | 定时任务调度（cron） |
| [ratelimit](https://github.com/hecc-blot/ratelimit) | 限流（内存 + Redis，滑动窗口/令牌桶） |
| [sse](https://github.com/hecc-blot/sse) | SSE 实时推送 |
| [lock](https://github.com/hecc-blot/lock) | 分布式锁（SetNX + watchdog 续期） |
| [metrics](https://github.com/hecc-blot/metrics) | 监控指标（Prometheus 采集） |
| [idempotent](https://github.com/hecc-blot/idempotent) | 幂等（SetNX 占位 + Lua 原子完成） |
| [circuitbreaker](https://github.com/hecc-blot/circuitbreaker) | 熔断器（连续失败熔断 + 半开探测） |

## 设计原则

1. **依赖倒置** — 高层依赖抽象接口，而非具体实现
2. **接口隔离** — 每个接口只定义单一职责
3. **开闭原则** — 对扩展开放，对修改关闭

## 许可证

MIT License
