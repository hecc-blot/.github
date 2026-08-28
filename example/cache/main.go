// 示例：cache 模块 — 本地/Redis 缓存基础操作、Hash、读穿透编排
//
// 只演示 cache 模块的用法，其余组件（追踪/限流等）见 example/ 下其它 main 与 demo/ 子包。
//
// 运行：
//
//	cd example/cache
//	go run .
//
// 启动后可用 curl 验证（读穿透示例需先确保 account 表有 id=1 的记录，且 Redis 已配置）：
//
//	curl http://localhost:8080/cache/basic
//	curl http://localhost:8080/cache/hash
//	curl http://localhost:8080/cache/read-through
//
// 详见：github.com/hecc-blot/cache
package main

import (
	cacheContract "github.com/hecc-blot/cache/contract"
	cache "github.com/hecc-blot/cache/service"
	logContract "github.com/hecc-blot/core/contract/log"
	log "github.com/hecc-blot/core/service/log"
	dbContract "github.com/hecc-blot/db/contract"
	db "github.com/hecc-blot/db/service"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	httpSvc "github.com/hecc-blot/framework/service/http"
	ioc "github.com/hecc-blot/framework/service/ioc"
	"github.com/hecc-blot/guide/example/demo"
	app "github.com/hecc-blot/guide/example/internal/app"
	traceContract "github.com/hecc-blot/trace/contract"
	trace "github.com/hecc-blot/trace/service"
)

func main() {
	config := app.InitConf("../../config.yaml")

	// 日志 + 追踪：cacheFactory 依赖 ILog（内部用 traceSvc 记慢查询/命中率等链路信息）
	logSvc := app.Must(log.NewLogger(&config.Log.Local))
	traceSvc, traceClearUp := app.Must2(trace.NewTraceSvc(&config.Trace))

	// db：读穿透示例（cache/read-through）需要 IDb 从 MySQL 回源
	mysqlDb, mysqlClearUp := app.Must2(db.NewMysql(config.Db.Mysql, logSvc))

	// 缓存工厂：内部按配置装配本地缓存 + Redis 缓存 + 编排层（GetOrLoad）
	cacheFactory := cache.NewCacheFactory(&config.Cache, traceSvc)

	defer func() {
		mysqlClearUp()
		traceClearUp()
		if cacheFactory.Redis() != nil {
			_ = cacheFactory.Redis().Close()
		}
	}()

	responseSvc := httpSvc.NewResponseSvc()

	// 注册到 IOC 容器
	container := ioc.New()
	container.Set(new(cacheContract.ICacheFactory), cacheFactory)
	container.Set(new(dbContract.IDb), mysqlDb)
	container.Set(new(logContract.ILog), logSvc)
	container.Set(new(traceContract.ITrace), traceSvc)
	container.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)

	// — 缓存操作：基础读写删 / Hash / 读穿透编排 —
	apiHandle.Get("cache/basic", &demo.CacheBasicApi{})
	apiHandle.Get("cache/hash", &demo.CacheHashApi{})
	apiHandle.Get("cache/read-through", &demo.CacheReadThroughApi{})

	apiHandle.Listen()
}
