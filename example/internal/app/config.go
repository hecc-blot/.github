package app

import (
	"fmt"

	cacheConfig "github.com/hecc-blot/cache/config"
	logConfig "github.com/hecc-blot/core/config/log"
	clickhouseConfig "github.com/hecc-blot/db-clickhouse/config"
	esConfig "github.com/hecc-blot/db-es/config"
	mongoConfig "github.com/hecc-blot/db-mongo/config"
	dbConfig "github.com/hecc-blot/db/config"
	serverConfig "github.com/hecc-blot/framework/config/http"
	httpClientConfig "github.com/hecc-blot/httpclient/config"
	slsConfig "github.com/hecc-blot/log-sls/config"
	metricsConfig "github.com/hecc-blot/metrics/config"
	mqConfig "github.com/hecc-blot/mq/config"
	schedulerConfig "github.com/hecc-blot/scheduler/config"
	traceConfig "github.com/hecc-blot/trace/config"

	"github.com/spf13/viper"
)

// Config 业务方配置聚合，按模块组装各模块的配置。
// 各示例 main 共用同一份 config.yaml，按需取用其中的字段。
type Config struct {
	Cache cacheConfig.Config
	Db    dbConfig.Config
	// Clickhouse 分析型数据库配置（可选，Ip 为空则跳过装配）
	Clickhouse clickhouseConfig.ClickhouseConfig
	// Mongo 文档型数据库配置（可选，Uri 与 Ip 均为空则跳过装配）
	Mongo mongoConfig.MongoConfig
	// Es 搜索型数据库配置（可选，Addresses 为空则跳过装配）
	Es        esConfig.EsConfig
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
	// Metrics 监控指标配置（可选，缺省端点 /metrics、无命名空间前缀）
	Metrics metricsConfig.Config
}

// LogConfig 日志配置聚合：本地日志（core/service/log）与 SLS（log-sls）按需二选一。
type LogConfig struct {
	Local logConfig.LocalConfig
	Sls   slsConfig.SlsConfig
}

// RateLimitConfig 请求频率限流配置（业务方自持有，框架不内置限流）。
// Backend 决定后端（memory | redis），Algorithm/Limit/Window 透传给 ratelimit 模块。
type RateLimitConfig struct {
	Backend   string `mapstructure:"backend"`   // 后端：memory(默认) | redis
	Algorithm string `mapstructure:"algorithm"` // 算法：sliding_window(默认) | token_bucket
	Limit     int    `mapstructure:"limit"`     // 窗口内最大请求数 / 令牌桶容量
	Window    int    `mapstructure:"window"`    // 窗口时长（秒）
}

// InitConf 使用 viper 读取 config.yaml，反序列化为 Config。
// configPath 相对各 main 所在目录传入（全量组装为 "config.yaml"，子模块为 "../../config.yaml"）。
func InitConf(configPath string) *Config {
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

// Must 单返回值错误处理：出错直接 panic。
func Must[T any](val T, err error) T {
	if err != nil {
		panic(fmt.Errorf("初始化失败: %w", err))
	}
	return val
}

// Must2 双返回值错误处理：出错直接 panic。
func Must2[T, U any](val T, cleanup U, err error) (T, U) {
	if err != nil {
		panic(fmt.Errorf("初始化失败: %w", err))
	}
	return val, cleanup
}
