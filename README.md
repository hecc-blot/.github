# Hecc-Blot

[![Go Version](https://img.shields.io/badge/Go-1.26.1-00ADD8?logo=go&logoColor=white)](https://github.com/hecc-blot/guide)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![GitHub](https://img.shields.io/badge/GitHub-Hecc--Blot-181717?logo=github&logoColor=white)](https://github.com/hecc-blot/guide)
[![Gitee](https://img.shields.io/badge/Gitee-Hecc--Blot-C71D23?logo=gitee&logoColor=white)](https://gitee.com/hecc-blot/hecc-blot-guide)
[![English](https://img.shields.io/badge/English-README_EN-blue)](README_EN.md)

Hecc-Blot 是一个基于 Go 语言的轻量级后端框架，采用面向接口的设计理念，提供依赖注入、路由注册、参数校验、统一响应等核心功能。本仓库是框架的**组装入口**，各功能模块已拆分为独立仓库，通过 `go get` 引入。

## 框架特性

- **面向接口**: 所有组件通过接口契约交互，易于替换和扩展
- **依赖注入**: 基于反射实现的 IOC 容器，通过 `inject` tag 自动注入
- **路由管理**: 基于 Gin 框架，支持 GET/POST 路由注册和中间件链
- **参数校验**: 自动参数绑定和校验，支持自定义校验错误信息
- **统一响应**: 自动包装返回值为 `{code, message, data}` 统一格式
- **多数据库**: 同时支持 MySQL 和 PostgreSQL，运行时可动态切换
- **事务支持**: 链式调用风格的数据库事务 API
- **双层缓存**: 本地内存缓存 + Redis 缓存，支持过期清理
- **链路追踪**: 基于 OpenTelemetry，支持 OTLP 导出到 Jaeger
- **SSE 推送**: 支持 Server-Sent Events，与 API 共享端口，适用于实时数据推送
- **可替换**: 所有组件可独立替换，只需实现对应接口并注册到 IOC

## 快速开始

完整可运行示例见 [`example/`](example/)，组装入口在 `main.go`，各组件示例在 `demo/` 子包。

```bash
cd example
go run .
```

## 目录结构

```
├── example/                # 完整使用示例（go run ./example）
│   ├── main.go             #   组装入口（package main）
│   ├── config.go           #   配置结构
│   └── demo/               #   各组件示例（package demo）
├── example.http            # 全部路由端点的请求文件
└── README.md
```

> 各功能模块已拆分为独立仓库，通过 `go get` 引入（见下方「模块仓库」），不再包含在本仓库中。各模块的接口定义、用法、配置说明均在其仓库 README 中。

## 模块仓库

| 模块 | 职责 | 仓库 |
|------|------|------|
| framework | 框架内核（接口契约 + IOC 容器 + HTTP 内核 + 统一错误 + 本地日志 + 分页） | [hecc-blot-framework](https://github.com/hecc-blot/framework) |
| ratelimit | 限流（内存 + Redis，滑动窗口/令牌桶） | [hecc-blot-ratelimit](https://github.com/hecc-blot/ratelimit) |
| log-sls | 日志（阿里云 SLS） | [hecc-blot-log-sls](https://github.com/hecc-blot/log-sls) |
| sse | SSE 推送 | [hecc-blot-sse](https://github.com/hecc-blot/sse) |
| db | 数据库（GORM MySQL/PostgreSQL） | [hecc-blot-db](https://github.com/hecc-blot/db) |
| cache | 缓存（本地 + Redis） | [hecc-blot-cache](https://github.com/hecc-blot/cache) |
| trace | 链路追踪（OpenTelemetry） | [hecc-blot-trace](https://github.com/hecc-blot/trace) |
| httpclient | 统一 HTTP 客户端 | [hecc-blot-httpclient](https://github.com/hecc-blot/httpclient) |
| mq | 消息队列（Kafka/NSQ） | [hecc-blot-mq](https://github.com/hecc-blot/mq) |

## 组装骨架

模块间通过接口契约 + IOC 容器组装，一个最小可运行入口如下：

```go
config := initConf("config.yaml")

// 日志：本地（core）与 SLS（log-sls）二选一，业务方显式指定
logSvc := must(log.NewLogger(&config.Log.Local))

traceSvc, traceClearUp := must2(trace.NewTraceSvc(&config.Trace))
dbFactory, dbClearUp := must2(db.NewDbFactory(&config.Db, logSvc))
cacheFactory := cache.NewCacheFactory(&config.Cache, traceSvc)
responseSvc := httpSvc.NewResponseSvc()
defer func() { dbClearUp(); traceClearUp() }()

// 注册到 IOC 容器
container := ioc.New()
container.Set(new(dbContract.IDbFactory), dbFactory)
container.Set(new(logContract.ILog), logSvc)
container.Set(new(cacheContract.ICacheFactory), cacheFactory)
container.Set(new(iCoreApi.IResponse), responseSvc)
container.Set(new(traceContract.ITrace), traceSvc)

// 创建 API 处理器并注册中间件 / 路由
apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)
apiHandle.Middleware(trace.NewHttpMiddleware(traceSvc))
apiHandle.Middleware(&RateLimitMiddleware{}) // 业务方引入 ratelimit 后自行实现
registerRoutes(apiHandle)

// SSE 共享 API 的 Engine
sseHandle := sse.NewSseSvc(apiHandle.Engine(), container)
sseHandle.Middleware(trace.NewSseMiddleware(traceSvc))
registerSseRoutes(sseHandle)

apiHandle.Listen(sseHandle.Shutdown)
```

各模块的详细接口定义、配置项、用法示例见对应模块仓库 README。

## 示例代码导航

`example/` 的组装入口（`main.go`）+ `demo/` 子包（各组件示例），可作为框架功能的活文档使用：

| 文件 | 演示内容 | 详文 |
|------|----------|------|
| main.go | 组装入口：初始化 → IOC → 路由注册 → 启动 | [framework](https://github.com/hecc-blot/framework) |
| config.go | 配置结构定义（Log/Db/Cache/Server/Trace/RateLimit） | [framework](https://github.com/hecc-blot/framework) |
| demo/model.go | Model 定义：IDbModel 接口、TableName、多 Model | [db](https://github.com/hecc-blot/db) |
| demo/validation.go | 请求参数与校验：binding tag、GetMessages() | [framework](https://github.com/hecc-blot/framework) |
| demo/middleware.go | 中间件：Token 校验、限流、SSE Accept/CORS | [framework](https://github.com/hecc-blot/framework) · [ratelimit](https://github.com/hecc-blot/ratelimit) |
| demo/db.go | 数据库 CRUD + 多数据库切换 | [db](https://github.com/hecc-blot/db) |
| demo/cache.go | 缓存操作：Local/Redis 读写删、Hash、读穿透 | [cache](https://github.com/hecc-blot/cache) |
| demo/trace.go | 链路追踪：Span/SetAttribute/RecordError/子 Span | [trace](https://github.com/hecc-blot/trace) |
| demo/paginator.go | 分页：Offset 分页 + 游标分页 | [framework](https://github.com/hecc-blot/framework) |
| demo/sse.go | SSE 推送：ISse 接口、心跳、Writer 写入 | [sse](https://github.com/hecc-blot/sse) |
| demo/httpclient.go | HTTP 客户端：NewHttpClient、Get/Post、重试与结构化日志 | [httpclient](https://github.com/hecc-blot/httpclient) |
| demo/mq.go | 消息队列：Producer 发送、Consumer 订阅、延迟/顺序能力断言 | [mq](https://github.com/hecc-blot/mq) |

## 设计原则

1. **依赖倒置**: 高层模块依赖抽象接口，而非具体实现
2. **接口隔离**: 每个接口只定义单一职责
3. **开闭原则**: 对扩展开放，对修改关闭

## 感谢

如果你觉得 Hecc-Blot 对你有帮助，欢迎给我们一个 ⭐️

### 反馈与贡献

- **Bug 反馈和功能建议**: 欢迎提交 [Issue](https://github.com/hecc-blot/guide/issues)
- **代码贡献**: 欢迎提交 Pull Request

### 致谢

- [Gin](https://github.com/gin-gonic/gin) — 高性能 Go Web 框架
- [GORM](https://github.com/go-gorm/gorm) — Go ORM 库
- [Zap](https://github.com/uber-go/zap) — 高性能日志库
- [OpenTelemetry](https://opentelemetry.io/) — 分布式追踪标准

## 许可证

MIT License
