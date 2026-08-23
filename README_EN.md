# Hecc-Blot

[![Go Version](https://img.shields.io/badge/Go-1.26.1-00ADD8?logo=go&logoColor=white)](https://github.com/hecc-blot/guide)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![GitHub](https://img.shields.io/badge/GitHub-Hecc--Blot-181717?logo=github&logoColor=white)](https://github.com/hecc-blot/guide)
[![Gitee](https://img.shields.io/badge/Gitee-Hecc--Blot-C71D23?logo=gitee&logoColor=white)](https://gitee.com/hecc-blot/hecc-blot-guide)
[![简体中文](https://img.shields.io/badge/简体中文-README-blue)](README.md)

Hecc-Blot is a lightweight Go backend framework built around interface-oriented design, providing dependency injection, route registration, parameter validation, and unified responses.

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

See [`example/example.go`](example/example.go) for a complete runnable example covering all features, organized by module.

```bash
cd example
go run .
```

## Project Layout

```
├── example/                # full usage example (go run ./example)
├── docs/                   # per-module documentation (currently in Chinese)
├── feature.md              # roadmap and optimization plan
└── README.md
```

> Each functional module lives in its own repository (see "Module Repositories" below) and is pulled in via `go get` — they are no longer part of this repo.

## Module Repositories

| Module | Responsibility | Repository |
|--------|----------------|------------|
| core | contract SDK (contract/entity/enum/util) | [hecc-blot-core](https://github.com/hecc-blot/core) |
| ioc | dependency injection container (zero deps) | [hecc-blot-ioc](https://github.com/hecc-blot/ioc) |
| api | HTTP core (routes + response + trace middleware) | [hecc-blot-api](https://github.com/hecc-blot/api) |
| error | unified errors | [hecc-blot-error](https://github.com/hecc-blot/error) |
| sse | SSE push | [hecc-blot-sse](https://github.com/hecc-blot/sse) |
| db | database (GORM MySQL/PostgreSQL) | [hecc-blot-db](https://github.com/hecc-blot/db) |
| cache | cache (local + Redis) | [hecc-blot-cache](https://github.com/hecc-blot/cache) |
| log | logging (Zap + SLS) | [hecc-blot-log](https://github.com/hecc-blot/log) |
| trace | tracing (OpenTelemetry) | [hecc-blot-trace](https://github.com/hecc-blot/trace) |

## Documentation

> **Note**: the docs under `docs/` are currently written in Chinese.

### Example Walkthrough

`example/example.go` is divided into 11 sections and serves as living documentation:

| # | Section | Demonstrates | Details |
|---|---------|--------------|---------|
| 1 | Entry point | main() skeleton: init → IOC → routes → start | [Quick Start](docs/quick_start.md) |
| 2 | Config loading | viper reads config.yaml | [Config](docs/config.md) |
| 3 | Model definition | IDbModel interface, TableName, multiple models | [Database](docs/database.md) |
| 4 | Request & validation | binding tags, GetMessages() | [Routes & Middleware](docs/routes_middleware.md) |
| 5 | Middleware | Authorization check, inject injection | [Routes & Middleware](docs/routes_middleware.md) |
| 6 | Database CRUD | Add/Take/Find/Save/Remove/Count/transactions | [Database](docs/database.md) |
| 7 | Multi-database | MySQL ↔ PostgreSQL switching | [Database](docs/database.md) |
| 8 | Cache operations | Local/Redis read-write-delete, Hash, read-through | [Cache](docs/cache.md) |
| 9 | Tracing | Span/SetAttribute/RecordError/sub-span | [Tracing](docs/trace.md) |
| 10 | Pagination | offset + cursor pagination | [Pagination](docs/paginator.md) |
| 11 | SSE | ISse interface, heartbeat, Flusher assertion | [SSE](docs/sse.md) |

### Getting Started

| Doc | Description |
|-----|-------------|
| [Quick Start Guide](docs/quick_start.md) | full tutorial for building a project from scratch |
| [Config Reference](docs/config.md) | all config.yaml options |

### Core Mechanisms

| Doc | Description |
|-----|-------------|
| [Routes & Middleware](docs/routes_middleware.md) | route registration, middleware, auto-validation, response wrapping |
| [IOC Injection](docs/ioc_injection.md) | injection principles, rules, named injection |
| [Component Replacement](docs/component_replacement.md) | full examples of swapping log/db/cache components |

### Component Usage

| Doc | Description |
|-----|-------------|
| [Logging](docs/logging.md) | local file logging, Alibaba Cloud SLS |
| [Database](docs/database.md) | CRUD, transactions, multi-database, model definition |
| [Cache](docs/cache.md) | local cache, Redis, expiry cleanup, tracing integration |
| [Tracing](docs/trace.md) | OpenTelemetry integration, span operations, cross-service propagation |
| [SSE](docs/sse.md) | SSE usage, route registration, middleware reuse, error handling |
| [Pagination](docs/paginator.md) | offset/limit and cursor pagination |

## Component Overview

### IOC Container

Auto-inject dependencies via the `inject:""` tag — no manual wiring. → [IOC Injection](docs/ioc_injection.md)

### API Service

Routes automatically perform binding, validation and response wrapping. → [Routes & Middleware](docs/routes_middleware.md)

### Database Service

MySQL and PostgreSQL support, chainable queries, transactions. → [Database](docs/database.md)

### Cache Service

In-memory + Redis two-tier cache with Hash operations and read-through. → [Cache](docs/cache.md)

### Logging Service

Local file logging (Zap + lumberjack rotation) and Alibaba Cloud SLS. → [Logging](docs/logging.md)

### Tracing

OpenTelemetry-based, auto-traces HTTP requests and correlates logs. → [Tracing](docs/trace.md)

### SSE Real-time Push

Shares the API port and pushes from the server via the `ISse` interface. → [SSE](docs/sse.md)

### Pagination

Offset/limit and cursor pagination. → [Pagination](docs/paginator.md)

## Design Principles

1. **Dependency inversion**: high-level modules depend on abstractions, not concrete implementations
2. **Interface segregation**: each interface defines a single responsibility
3. **Open/closed principle**: open for extension, closed for modification

## Roadmap

See [feature.md](feature.md) for the framework's optimization plan.

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
