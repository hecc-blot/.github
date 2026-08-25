package main

import (
	cacheConfig "github.com/hecc-blot/cache/config"
	dbConfig "github.com/hecc-blot/db/config"
	serverConfig "github.com/hecc-blot/framework/config/http"
	logConfig "github.com/hecc-blot/framework/config/log"
	httpClientConfig "github.com/hecc-blot/httpclient/config"
	mqConfig "github.com/hecc-blot/mq/config"
	schedulerConfig "github.com/hecc-blot/scheduler/config"
	traceConfig "github.com/hecc-blot/trace/config"
)

// Config 业务方配置聚合，按模块组装各模块的配置。
type Config struct {
	Cache     cacheConfig.Config
	Db        dbConfig.Config
	Log       LogConfig
	RateLimit RateLimitConfig
	Server    serverConfig.Config
	Trace     traceConfig.Config
	// HttpClient 统一出站 HTTP 客户端配置（可选，缺省使用默认超时/重试）
	HttpClient httpClientConfig.Config
	// Mq 消息队列配置（可选，Driver 为空则跳过 MQ 组装）
	Mq mqConfig.Config
	// Scheduler 定时任务调度器配置（可选，缺省使用本地时区、禁止重叠执行）
	Scheduler schedulerConfig.Config
}

// LogConfig 日志配置聚合：本地日志（framework/log）与 SLS（log-sls）按需二选一。
type LogConfig struct {
	Local logConfig.LocalConfig
	Sls   logConfig.SlsConfig
}

// RateLimitConfig 请求频率限流配置（业务方自持有，框架不内置限流）。
// Backend 决定后端（memory | redis），Algorithm/Limit/Window 透传给 ratelimit 模块。
type RateLimitConfig struct {
	Backend   string `mapstructure:"backend"`   // 后端：memory(默认) | redis
	Algorithm string `mapstructure:"algorithm"` // 算法：sliding_window(默认) | token_bucket
	Limit     int    `mapstructure:"limit"`     // 窗口内最大请求数 / 令牌桶容量
	Window    int    `mapstructure:"window"`    // 窗口时长（秒）
}
