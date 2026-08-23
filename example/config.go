package main

import (
	cacheConfig "github.com/hecc-blot/cache/config"
	dbConfig "github.com/hecc-blot/db/config"
	logConfig "github.com/hecc-blot/log/config"
	serverConfig "github.com/hecc-blot/api/config"
	traceConfig "github.com/hecc-blot/trace/config"
)

// Config 业务方配置聚合，按模块组装各模块的配置。
type Config struct {
	Cache  cacheConfig.Config
	Db     dbConfig.Config
	Log    logConfig.Config
	Server serverConfig.Config
	Trace  traceConfig.Config
}
