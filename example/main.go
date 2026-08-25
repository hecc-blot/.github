package main

import (
	"fmt"

	cacheContract "github.com/hecc-blot/cache/contract"
	cache "github.com/hecc-blot/cache/service"
	dbContract "github.com/hecc-blot/db/contract"
	db "github.com/hecc-blot/db/service"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	logContract "github.com/hecc-blot/framework/contract/log"
	httpSvc "github.com/hecc-blot/framework/service/http"
	ioc "github.com/hecc-blot/framework/service/ioc"
	log "github.com/hecc-blot/framework/service/log"
	demo "github.com/hecc-blot/guide/example/demo"
	httpClientContract "github.com/hecc-blot/httpclient/contract"
	httpClientService "github.com/hecc-blot/httpclient/service"
	idempotentContract "github.com/hecc-blot/idempotent/contract"
	idempotentService "github.com/hecc-blot/idempotent/service"
	lockContract "github.com/hecc-blot/lock/contract"
	lockService "github.com/hecc-blot/lock/service"
	logsls "github.com/hecc-blot/log-sls/service"
	metricsConfig "github.com/hecc-blot/metrics/config"
	metricsContract "github.com/hecc-blot/metrics/contract"
	metrics "github.com/hecc-blot/metrics/service"
	mqContract "github.com/hecc-blot/mq/contract"
	mqService "github.com/hecc-blot/mq/service"
	ratelimitConfig "github.com/hecc-blot/ratelimit/config"
	ratelimitContract "github.com/hecc-blot/ratelimit/contract"
	algorithm "github.com/hecc-blot/ratelimit/enum/algorithm"
	ratelimitSvc "github.com/hecc-blot/ratelimit/service"
	scheduler "github.com/hecc-blot/scheduler/service"
	sseContract "github.com/hecc-blot/sse/contract"
	sse "github.com/hecc-blot/sse/service"
	traceContract "github.com/hecc-blot/trace/contract"
	trace "github.com/hecc-blot/trace/service"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// ===== 启动入口 =====
// 演示：框架初始化全流程 — 配置 → 日志 → 追踪 → 数据库 → 缓存 → IOC → 路由 → 启动
// 各组件示例见 demo/ 子包，本文件只负责组装。

func main() {
	config := initConf("config.yaml")

	// 日志：本地日志（默认）或 SLS（加强后端，见 log-sls 模块），由组装层按配置二选一
	var logSvc logContract.ILog
	if config.Log.Sls.Enable {
		logSvc = must(logsls.NewLogger(&config.Log.Sls))
	} else {
		logSvc = must(log.NewLogger(&config.Log.Local))
	}
	traceSvc, traceClearUp := must2(trace.NewTraceSvc(&config.Trace))
	dbFactory, dbClearUp := must2(db.NewDbFactory(&config.Db, logSvc))

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
		dbClearUp()
		traceClearUp()
		if cacheFactory.Redis() != nil {
			_ = cacheFactory.Redis().Close()
		}
	}()

	// 注册到 IOC 容器（顺序无关，但必须在路由注册之前）
	container := ioc.New()

	container.Set(new(dbContract.IDbFactory), dbFactory)
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
	apiHandle.Engine().GET(metricsCfg.Path, gin.WrapH(metricsSvc.Handler()))
	// 请求限流：业务方引入 ratelimit 模块后自行实现中间件并显式注册（是否启用由本行决定）
	apiHandle.Middleware(&demo.RateLimitMiddleware{Limiter: newRateLimiter(config)})
	registerRoutes(apiHandle)

	sseHandle := sse.NewSseSvc(apiHandle.Engine(), container)
	sseHandle.Middleware(trace.NewSseMiddleware(traceSvc))
	registerSseRoutes(sseHandle)

	apiHandle.Listen(sseHandle.Shutdown)
}

// must 单返回值错误处理：出错直接 panic
func must[T any](val T, err error) T {
	if err != nil {
		panic(fmt.Errorf("初始化失败: %w", err))
	}
	return val
}

// must2 双返回值错误处理：出错直接 panic
func must2[T, U any](val T, cleanup U, err error) (T, U) {
	if err != nil {
		panic(fmt.Errorf("初始化失败: %w", err))
	}
	return val, cleanup
}

// ===== 配置加载 =====
// 演示：使用 viper 读取 config.yaml，反序列化为 Config 结构体

func initConf(configPath string) *Config {
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败: %w", err))
	}
	var conf Config
	if err := v.Unmarshal(&conf); err != nil {
		panic(fmt.Errorf("解析配置文件失败: %w", err))
	}
	return &conf
}

// newRateLimiter 根据配置选择限流后端：backend=redis 使用独立 redis 连接（集群统一计数），
// 否则用内存限流（单实例）。算法由 config.RateLimit.Algorithm 决定。
func newRateLimiter(config *Config) ratelimitContract.RateLimiter {
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
