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

模块之间通过 IOC 容器组装，整体就三步：**先构造 → 再 `Set` 进容器 → 最后注册路由并 `Listen`**。

下面是一个最小可运行的 `main.go` 骨架。各模块 README 的「快速开始」都只是它的片段，公共部分以本页为准。

```go
package main

import (
	cacheContract "github.com/hecc-blot/cache/contract"
	cache "github.com/hecc-blot/cache/service"
	logContract "github.com/hecc-blot/core/contract/log"
	ioc "github.com/hecc-blot/core/service/ioc"
	log "github.com/hecc-blot/core/service/log"
	dbContract "github.com/hecc-blot/db/contract"
	db "github.com/hecc-blot/db/service"
	httpSvc "github.com/hecc-blot/framework/service/http"
	traceContract "github.com/hecc-blot/trace/contract"
	trace "github.com/hecc-blot/trace/service"
)

func main() {
	// 读配置（viper / yaml 自选），得到 config

	// ── 1. 日志：最先构造，后续所有组件的日志都依赖它
	logSvc, err := log.NewLogger(&config.Log.Local)
	if err != nil {
		panic(err)
	}

	// ── 2. 链路追踪：必须早于 db / cache / mq
	//    GORM 的 otel 插件与 mq 的生产者都在「创建时」捕获全局 TracerProvider，
	//    晚于本步构造就拿不到 tracer，span 会全部丢失。
	traceSvc, traceClearUp, err := trace.NewTraceSvc(&config.Trace)
	if err != nil {
		panic(err)
	}

	// ── 3. 数据与缓存
	dbFactory, dbClearUp, err := db.NewDbFactory(&config.Db, logSvc)
	if err != nil {
		panic(err)
	}
	cacheFactory := cache.NewCacheFactory(&config.Cache, traceSvc)

	// ── 4. 容器：先把依赖全部 Set 完，再注册路由
	//    注册路由时会对 API 结构体调用 Inject，依赖未注册会 panic。
	container := ioc.New()
	container.Set(new(logContract.ILog), logSvc)
	container.Set(new(traceContract.ITrace), traceSvc)
	container.Set(new(dbContract.IDbFactory), dbFactory)
	container.Set(new(cacheContract.ICacheFactory), cacheFactory)

	// ── 5. HTTP 内核与路由
	responseSvc := httpSvc.NewResponseSvc()
	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)
	// trace 中间件决定 traceId 能否进入日志与下游调用，不注册则日志里没有 traceId
	apiHandle.Middleware(trace.NewHttpMiddleware(traceSvc))
	// apiHandle.Post("account/add", &AddAccountApi{})

	// ── 6. 优雅停机：defer 是 LIFO，先关连接、再让 Listen 退出
	//    （Listen 收到 SIGINT/SIGTERM 后先跑 OnShutdown 注册的钩子再 Shutdown；
	//      长连接模块如 sse 已在构造时自注册，业务不必传参）
	defer func() {
		dbClearUp()
		traceClearUp()
	}()

	apiHandle.Listen() // 阻塞
}
```

### 组装顺序的三条硬约束

| 约束 | 原因 |
|---|---|
| **日志最先构造** | 其后的每个组件构造失败都要靠它报错 |
| **`trace` 早于 `db` / `cache` / `mq`** | 这些组件在**创建时**捕获全局 `TracerProvider`，晚建就拿不到 tracer |
| **`container.Set` 全部早于路由注册** | 注册路由会对 API 结构体 `Inject`，依赖未注册直接 panic |

## 如何阅读各模块 README

每份 README 结构一致，按需跳转：

| 章节 | 用途 |
|---|---|
| **快速开始** | 可复制的最小用法，先跑通再深入 |
| **接口定义** | 与 `contract/` 逐字对齐的契约 |
| **使用** | 按场景展开的用法与示例 |
| **配置** | yaml + 配置项表格（含默认值与 0 值语义） |
| **语义与约定** | 并发安全、panic 条件、幂等、错误语义、边界 |
| **相关模块** | 上下游依赖 |

## 模块仓库

| 模块 | 职责 |
|------|------|
| [core](https://github.com/hecc-blot/core) | 最底层公共库：零依赖原语 + 通用日志契约与实现 + IOC 容器 |
| [framework](https://github.com/hecc-blot/framework) | 框架内核：HTTP 内核（路由 / 中间件 / 参数校验 / 统一响应）+ 统一错误 + 分页 |
| [db](https://github.com/hecc-blot/db) | 关系型数据库（GORM MySQL/PostgreSQL） |
| [db-clickhouse](https://github.com/hecc-blot/db-clickhouse) | ClickHouse 分析型数据库 |
| [db-es](https://github.com/hecc-blot/db-es) | Elasticsearch 搜索型数据库 |
| [db-mongo](https://github.com/hecc-blot/db-mongo) | MongoDB 文档型数据库 |
| [cache](https://github.com/hecc-blot/cache) | 双层缓存（本地 + Redis）+ cache-aside 编排层 |
| [trace](https://github.com/hecc-blot/trace) | 链路追踪（OpenTelemetry） |
| [log-sls](https://github.com/hecc-blot/log-sls) | 日志（阿里云 SLS） |
| [httpclient](https://github.com/hecc-blot/httpclient) | 统一 HTTP 客户端（重试 / 熔断 / 降级） |
| [mq](https://github.com/hecc-blot/mq) | 消息队列（Kafka / NSQ / RocketMQ）+ 消费者注册器 |
| [scheduler](https://github.com/hecc-blot/scheduler) | 定时任务调度（cron） |
| [ratelimit](https://github.com/hecc-blot/ratelimit) | 限流（内存 + Redis） |
| [sse](https://github.com/hecc-blot/sse) | SSE 实时推送 |
| [lock](https://github.com/hecc-blot/lock) | 分布式锁（Redis / 内存） |
| [metrics](https://github.com/hecc-blot/metrics) | 监控指标（Prometheus 采集） |
| [idempotent](https://github.com/hecc-blot/idempotent) | 幂等（Redis / 内存） |
| [circuitbreaker](https://github.com/hecc-blot/circuitbreaker) | 熔断器（进程内状态） |

## 设计原则

1. **依赖倒置** — 高层依赖抽象接口，而非具体实现
2. **接口隔离** — 每个接口只定义单一职责
3. **开闭原则** — 对扩展开放，对修改关闭

## 许可证

MIT License
