# Hecc-Blot

[![Go Version](https://img.shields.io/badge/Go-1.26.1-00ADD8?logo=go&logoColor=white)](https://github.com/hecc-blot/guide)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![GitHub](https://img.shields.io/badge/GitHub-Hecc--Blot-181717?logo=github&logoColor=white)](https://github.com/hecc-blot/guide)
[![Gitee](https://img.shields.io/badge/Gitee-Hecc--Blot-C71D23?logo=gitee&logoColor=white)](https://gitee.com/hecc-blot/hecc-blot-guide)
[![简体中文](https://img.shields.io/badge/简体中文-README-blue)](README.md)

Hecc-Blot is a lightweight Go backend framework built around interface-oriented design, providing dependency injection, route registration, parameter validation, and unified responses. This repository is the **assembly entry point**; each functional module lives in its own repository and is pulled in via `go get`.

## Features

- **Interface-oriented**: all components interact through interface contracts, easy to replace and extend
- **Dependency injection**: reflection-based IOC container, auto-inject via the `inject` tag
- **Routing**: built on Gin, supports GET/POST route registration and middleware chains
- **Validation**: automatic binding and validation with customizable error messages
- **Unified response**: wraps return values into a `{code, message, data}` format
- **Multi-database**: supports MySQL and PostgreSQL with runtime switching
- **Transactions**: chainable transaction API
- **Two-tier cache**: in-memory cache + Redis, with expiry cleanup
- **Tracing**: OpenTelemetry-based, OTLP export to Jaeger
- **SSE**: Server-Sent Events sharing the API port, for real-time push
- **Replaceable**: every component can be swapped by implementing its interface and registering it with the IOC

## Quick Start

See [`example/`](example/) for a complete runnable example — entry point in `main.go`, component examples in the `demo/` subpackage.

```bash
cd example
go run .
```

## Project Layout

```
├── example/                # full usage example (go run ./example)
│   ├── main.go             #   entry point (package main)
│   ├── config.go           #   config structs
│   └── demo/               #   component examples (package demo)
├── example.http            # request file for all route endpoints
└── README.md
```

> Each functional module lives in its own repository (see "Module Repositories" below) and is pulled in via `go get` — they are no longer part of this repo. Each module's interface definitions, usage, and configuration are documented in its own repository README.

## Module Repositories

| Module | Responsibility | Repository |
|--------|----------------|------------|
| framework | framework core (contracts + IOC container + HTTP kernel + unified errors + local logging + pagination) | [hecc-blot-framework](https://github.com/hecc-blot/framework) |
| ratelimit | rate limiting (in-memory + Redis, sliding window / token bucket) | [hecc-blot-ratelimit](https://github.com/hecc-blot/ratelimit) |
| log-sls | logging (Alibaba Cloud SLS) | [hecc-blot-log-sls](https://github.com/hecc-blot/log-sls) |
| sse | SSE push | [hecc-blot-sse](https://github.com/hecc-blot/sse) |
| db | database (GORM MySQL/PostgreSQL) | [hecc-blot-db](https://github.com/hecc-blot/db) |
| cache | cache (local + Redis) | [hecc-blot-cache](https://github.com/hecc-blot/cache) |
| trace | tracing (OpenTelemetry) | [hecc-blot-trace](https://github.com/hecc-blot/trace) |
| httpclient | unified HTTP client | [hecc-blot-httpclient](https://github.com/hecc-blot/httpclient) |
| mq | message queue (Kafka/NSQ) | [hecc-blot-mq](https://github.com/hecc-blot/mq) |

## Assembly Skeleton

Modules are wired together through interface contracts and the IOC container. A minimal runnable entry point looks like:

```go
config := initConf("config.yaml")

// Logging: local (core) or SLS (log-sls), chosen explicitly by the business
logSvc := must(log.NewLogger(&config.Log.Local))

traceSvc, traceClearUp := must2(trace.NewTraceSvc(&config.Trace))
dbFactory, dbClearUp := must2(db.NewDbFactory(&config.Db, logSvc))
cacheFactory := cache.NewCacheFactory(&config.Cache, traceSvc)
responseSvc := httpSvc.NewResponseSvc()
defer func() { dbClearUp(); traceClearUp() }()

// Register with the IOC container
container := ioc.New()
container.Set(new(dbContract.IDbFactory), dbFactory)
container.Set(new(logContract.ILog), logSvc)
container.Set(new(cacheContract.ICacheFactory), cacheFactory)
container.Set(new(iCoreApi.IResponse), responseSvc)
container.Set(new(traceContract.ITrace), traceSvc)

// Create the API handler and register middleware / routes
apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)
apiHandle.Middleware(trace.NewHttpMiddleware(traceSvc))
apiHandle.Middleware(&RateLimitMiddleware{}) // implemented by you via ratelimit
registerRoutes(apiHandle)

// SSE reuses the API route registration
sseHandle := sse.NewSseSvc(apiHandle, container)
sseHandle.Middleware(trace.NewSseMiddleware(traceSvc))
registerSseRoutes(sseHandle)

apiHandle.Listen(sseHandle.Shutdown)
```

See each module's repository README for full interface definitions, configuration options, and usage examples.

## Example Walkthrough

The `example/` entry point (`main.go`) plus the `demo/` subpackage (component examples) serve as living documentation:

| File | Demonstrates | Details |
|------|--------------|---------|
| main.go | entry point: init → IOC → route registration → start | [framework](https://github.com/hecc-blot/framework) |
| config.go | config structs (Log/Db/Cache/Server/Trace/RateLimit) | [framework](https://github.com/hecc-blot/framework) |
| demo/model.go | model definition: IDbModel, TableName, multiple models | [db](https://github.com/hecc-blot/db) |
| demo/validation.go | request & validation: binding tags, GetMessages() | [framework](https://github.com/hecc-blot/framework) |
| demo/middleware.go | middleware: token check, rate limiting, SSE Accept/CORS | [framework](https://github.com/hecc-blot/framework) · [ratelimit](https://github.com/hecc-blot/ratelimit) |
| demo/db.go | database CRUD + multi-database switching | [db](https://github.com/hecc-blot/db) |
| demo/cache.go | cache: Local/Redis read-write-delete, Hash, read-through | [cache](https://github.com/hecc-blot/cache) |
| demo/trace.go | tracing: Span/SetAttribute/RecordError/sub-span | [trace](https://github.com/hecc-blot/trace) |
| demo/paginator.go | pagination: offset + cursor | [framework](https://github.com/hecc-blot/framework) |
| demo/sse.go | SSE: ISse interface, heartbeat, Writer write | [sse](https://github.com/hecc-blot/sse) |
| demo/httpclient.go | HTTP client: NewHttpClient, Get/Post, retry & structured logging | [httpclient](https://github.com/hecc-blot/httpclient) |
| demo/mq.go | message queue: Producer send, Consumer subscribe, delayed/ordered assertion | [mq](https://github.com/hecc-blot/mq) |

## Design Principles

1. **Dependency inversion**: high-level modules depend on abstractions, not concrete implementations
2. **Interface segregation**: each interface defines a single responsibility
3. **Open/closed principle**: open for extension, closed for modification

## Thanks

If Hecc-Blot helps you, a ⭐️ is appreciated.

### Feedback & Contributing

- **Bug reports and feature requests**: open an [Issue](https://github.com/hecc-blot/guide/issues)
- **Code contributions**: pull requests are welcome

### Credits

- [Gin](https://github.com/gin-gonic/gin) — high-performance Go web framework
- [GORM](https://github.com/go-gorm/gorm) — Go ORM library
- [Zap](https://github.com/uber-go/zap) — high-performance logging library
- [OpenTelemetry](https://opentelemetry.io/) — distributed tracing standard

## License

MIT License
