package main

import (
	cacheContract "github.com/hecc-blot/cache/contract"
	cache "github.com/hecc-blot/cache/service"
	logContract "github.com/hecc-blot/core/contract/log"
	log "github.com/hecc-blot/core/service/log"
	dbClickhouseContract "github.com/hecc-blot/db-clickhouse/contract"
	dbClickhouse "github.com/hecc-blot/db-clickhouse/service"
	dbEsContract "github.com/hecc-blot/db-es/contract"
	dbEs "github.com/hecc-blot/db-es/service"
	dbMongoContract "github.com/hecc-blot/db-mongo/contract"
	dbMongo "github.com/hecc-blot/db-mongo/service"
	dbContract "github.com/hecc-blot/db/contract"
	db "github.com/hecc-blot/db/service"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	httpSvc "github.com/hecc-blot/framework/service/http"
	ioc "github.com/hecc-blot/framework/service/ioc"
	"github.com/hecc-blot/guide/example/demo"
	"github.com/hecc-blot/guide/example/internal/app"
	httpClientContract "github.com/hecc-blot/httpclient/contract"
	httpClientService "github.com/hecc-blot/httpclient/service"
	idempotentContract "github.com/hecc-blot/idempotent/contract"
	idempotentService "github.com/hecc-blot/idempotent/service"
	lockContract "github.com/hecc-blot/lock/contract"
	lockService "github.com/hecc-blot/lock/service"
	metricsConfig "github.com/hecc-blot/metrics/config"
	metricsContract "github.com/hecc-blot/metrics/contract"
	metrics "github.com/hecc-blot/metrics/service"
	mqContract "github.com/hecc-blot/mq/contract"
	mqService "github.com/hecc-blot/mq/service"
	ratelimitConfig "github.com/hecc-blot/ratelimit/config"
	ratelimitContract "github.com/hecc-blot/ratelimit/contract"
	"github.com/hecc-blot/ratelimit/enum/algorithm"
	ratelimitSvc "github.com/hecc-blot/ratelimit/service"
	scheduler "github.com/hecc-blot/scheduler/service"
	sseContract "github.com/hecc-blot/sse/contract"
	sse "github.com/hecc-blot/sse/service"
	traceContract "github.com/hecc-blot/trace/contract"
	trace "github.com/hecc-blot/trace/service"

	"context"
	"net/http"

	"github.com/redis/go-redis/v9"
)

// ===== 启动入口 =====
// 演示：框架初始化全流程 — 配置 → 日志 → 追踪 → 数据库 → 缓存 → IOC → 路由 → 启动
// 各组件示例见 demo/ 子包，本文件只负责组装。

func main() {
	config := app.InitConf("config.yaml")

	// 日志：本地日志（core 默认）与 SLS（log-sls）二选一，业务方按需显式指定，不做 enable 自动切换。
	// 此处用本地日志；改用 SLS 时换成 logsls.NewLogger(&config.Log.Sls)（见 log-sls 模块）。
	logSvc := app.Must(log.NewLogger(&config.Log.Local))
	traceSvc, traceClearUp := app.Must2(trace.NewTraceSvc(&config.Trace))
	dbFactory, dbClearUp := app.Must2(db.NewDbFactory(&config.Db, logSvc))

	// 单库直连：仅使用 MySQL 的业务无需工厂，直接构造并注入 IDb（IDbFactory 仅多库切换场景使用）
	mysqlDb, mysqlClearUp := app.Must2(db.NewMysql(config.Db.Mysql, logSvc))

	// 可选后端数据库：分析型 / 文档型 / 搜索型，按配置二选一装配，未配置时跳过（对应路由返回友好提示）
	var (
		clickhouseDb    dbClickhouseContract.IDbClickhouse
		clickhouseClear func()
		mongoDb         dbMongoContract.IDbDocument
		mongoClear      func()
		esDb            dbEsContract.IDbSearch
		esClear         func()
	)
	if config.Clickhouse.Ip != "" {
		clickhouseDb, clickhouseClear = app.Must2(dbClickhouse.NewClickhouse(&config.Clickhouse, logSvc))
	}
	if config.Mongo.Uri != "" || config.Mongo.Ip != "" {
		mongoDb, mongoClear = app.Must2(dbMongo.NewMongo(&config.Mongo, logSvc))
	}
	if len(config.Es.Addresses) > 0 {
		esDb, esClear = app.Must2(dbEs.NewEs(&config.Es, logSvc))
	}

	cacheFactory := cache.NewCacheFactory(&config.Cache, traceSvc)
	responseSvc := httpSvc.NewResponseSvc()

	// HTTP 客户端：统一出站请求（内置重试 + 结构化日志 + 敏感头脱敏），无外部依赖
	httpClient := httpClientService.NewHttpClient(config.HttpClient, logSvc)

	// 监控指标：Prometheus 采集端点 + 请求 QPS/延迟/错误率中间件（无外部依赖）
	metricsCfg := metricsConfig.Normalize(config.Metrics)
	metricsSvc := metrics.NewMetrics(&metricsCfg)

	// 定时任务：cron 调度器，覆盖超时处理/周期对账/批量清理（无外部依赖，配置无效时跳过）
	schedulerSvc, err := scheduler.NewScheduler(&config.Scheduler, logSvc, traceSvc, nil)
	if err != nil {
		logSvc.Warn(context.Background(), "调度器未启用：配置无效，跳过", "err", err)
	} else {
		_ = schedulerSvc.Add("cleanup", "0 */5 * * * *", demo.NewCleanupJob(logSvc))
		schedulerSvc.Start()
		defer schedulerSvc.Stop()
	}

	// defer 注册退出清理（LIFO 顺序执行）
	defer func() {
		if esClear != nil {
			esClear()
		}
		if mongoClear != nil {
			mongoClear()
		}
		if clickhouseClear != nil {
			clickhouseClear()
		}
		mysqlClearUp()
		dbClearUp()
		traceClearUp()
		if cacheFactory.Redis() != nil {
			_ = cacheFactory.Redis().Close()
		}
	}()

	// 注册到 IOC 容器（顺序无关，但必须在路由注册之前）
	container := ioc.New()

	container.Set(new(dbContract.IDbFactory), dbFactory)
	container.Set(new(dbContract.IDb), mysqlDb)
	if clickhouseDb != nil {
		container.Set(new(dbClickhouseContract.IDbClickhouse), clickhouseDb)
	}
	if mongoDb != nil {
		container.Set(new(dbMongoContract.IDbDocument), mongoDb)
	}
	if esDb != nil {
		container.Set(new(dbEsContract.IDbSearch), esDb)
	}
	container.Set(new(logContract.ILog), logSvc)
	container.Set(new(cacheContract.ICacheFactory), cacheFactory)
	container.Set(new(iCoreApi.IResponse), responseSvc)
	container.Set(new(traceContract.ITrace), traceSvc)
	container.Set(new(httpClientContract.IHttpClient), httpClient)
	container.Set(new(metricsContract.IMetrics), metricsSvc)

	// 分布式锁：按需加载，复用 cache 的 redis 连接（IRedisCache 已实现 SetNX/Eval 原子原语）
	container.Set(new(lockContract.ILocker), lockService.NewRedisLocker(cacheFactory.Redis()))

	// 幂等：按需加载，复用 cache 的 redis 连接（IRedisCache 已实现 SetNX/Get/Eval 原子原语）
	container.Set(new(idempotentContract.IIdempotent), idempotentService.NewRedisIdempotent(cacheFactory.Redis()))

	// 消息队列：依赖 Kafka/NSQ broker，未配置时跳过（不影响其余示例启动）
	if config.Mq.Driver != "" {
		mqFactory, mqCleanup, err := mqService.NewMqFactory(&config.Mq, logSvc, traceSvc)
		if err != nil {
			logSvc.Warn(context.Background(), "MQ 未启用：broker 配置无效，跳过注册", "err", err)
		} else {
			container.Set(new(mqContract.IMqFactory), mqFactory)
			defer mqCleanup()
		}
	}

	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)
	// 链路追踪：由组装层显式注册中间件（api 不感知 trace）
	apiHandle.Middleware(trace.NewHttpMiddleware(traceSvc))
	// 监控指标：自动采集 QPS/延迟/错误率（path 用路由模板避免高基数）
	apiHandle.Middleware(metrics.NewHttpMiddleware(metricsSvc))
	// Prometheus 采集端点
	apiHandle.Handle(http.MethodGet, metricsCfg.Path, metricsSvc.Handler())
	// 请求限流：业务方引入 ratelimit 模块后自行实现中间件并显式注册（是否启用由本行决定）
	apiHandle.Middleware(&demo.RateLimitMiddleware{Limiter: newRateLimiter(config)})
	registerRoutes(apiHandle)

	sseHandle := sse.NewSseSvc(apiHandle, container)
	sseHandle.Middleware(trace.NewSseMiddleware(traceSvc))
	registerSseRoutes(sseHandle)

	apiHandle.Listen(sseHandle.Shutdown)
}

// newRateLimiter 根据配置选择限流后端：backend=redis 使用独立 redis 连接（集群统一计数），
// 否则用内存限流（单实例）。算法由 config.RateLimit.Algorithm 决定。
func newRateLimiter(config *app.Config) ratelimitContract.RateLimiter {
	cfg := ratelimitConfig.Config{
		Algorithm: algorithm.Algorithm(config.RateLimit.Algorithm),
		Limit:     config.RateLimit.Limit,
		Window:    config.RateLimit.Window,
	}
	if config.RateLimit.Backend == "redis" {
		client := redis.NewClient(&redis.Options{
			Addr:     config.Cache.Redis.Addr,
			Password: config.Cache.Redis.Password,
			DB:       config.Cache.Redis.DB,
			PoolSize: config.Cache.Redis.PoolSize,
		})
		return ratelimitSvc.NewRedisLimiter(client, cfg)
	}
	return ratelimitSvc.NewMemoryLimiter(cfg)
}

// ==============================
// 路由注册（集中管理）
// 各示例 API / 中间件定义在 demo/ 子包，此处只做路由挂载。
// ==============================

func registerRoutes(apiHandle iCoreApi.IApiHandle) {
	// API 路由分组，Token 鉴权中间件仅作用于该分组（SSE 不受影响）
	apiGroup := apiHandle.Group("", &demo.TokenMiddleware{})

	{
		// — 参数校验 —
		apiGroup.Post("account/add", &demo.AddAccountApi{})

		// — 数据库 CRUD —
		apiGroup.Get("account/take", &demo.TakeAccountApi{})
		apiGroup.Get("account/find", &demo.FindAccountApi{})
		apiGroup.Get("account/count", &demo.CountAccountApi{})
		apiGroup.Post("account/update", &demo.UpdateAccountApi{})
		apiGroup.Post("account/delete", &demo.DeleteAccountApi{})

		// — 多数据库切换 —
		apiGroup.Get("account/db-switch", &demo.DbSwitchApi{})

		// — 分析型数据库（ClickHouse）—
		apiGroup.Post("clickhouse/demo", &demo.ClickhouseDemoApi{})

		// — 文档型数据库（MongoDB）—
		apiGroup.Post("mongo/demo", &demo.MongoDemoApi{})

		// — 搜索型数据库（Elasticsearch）—
		apiGroup.Post("es/demo", &demo.EsDemoApi{})

		// — 缓存操作 —
		apiGroup.Get("cache/basic", &demo.CacheBasicApi{})
		apiGroup.Get("cache/hash", &demo.CacheHashApi{})
		apiGroup.Get("cache/read-through", &demo.CacheReadThroughApi{})

		// — 链路追踪 —
		apiGroup.Get("trace/demo", &demo.TraceDemoApi{})

		// — 分页 —
		apiGroup.Post("account/page", &demo.PageListApi{})
		apiGroup.Post("account/cursor", &demo.CursorListApi{})

		// — HTTP 客户端 —
		apiGroup.Get("httpclient/demo", &demo.HttpClientDemoApi{})

		// — 消息队列 —
		apiGroup.Post("mq/demo", &demo.MqDemoApi{})

		// — 分布式锁 —
		apiGroup.Get("lock/demo", &demo.LockDemoApi{})

		// — 幂等 —
		apiGroup.Get("idempotent/demo", &demo.IdempotentDemoApi{})
	}
}

func registerSseRoutes(sseHandle sseContract.ISseHandle) {
	// — SSE 推送 —
	// 通过中间件做 Accept 校验（策略性校验不内置在框架）
	sseGroup := sseHandle.Group("", &demo.SseAcceptMiddleware{}, &demo.SseCorsMiddleware{})

	// GET 方式：EventSource 标准用法
	sseGroup.Get("events/time", &demo.ExampleSse{})

	// POST 方式：适用于 fetch + ReadableStream（可携带请求体）
	sseGroup.Post("events/time", &demo.ExampleSse{})
}
