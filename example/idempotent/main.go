// 示例：idempotent 模块 — 幂等（同一幂等键只执行一次业务）
//
// 只演示 idempotent 模块的用法。依赖 Redis（复用 cache 模块的 redis 连接），运行前需先配置 config.yaml 的 redis 段。
//
// 运行：
//
//	cd example/idempotent
//	go run .
//
// 启动后可用 curl 验证（重复请求返回首次结果，repeated=true）：
//
//	curl http://localhost:8080/idempotent/demo
//
// 详见：github.com/hecc-blot/idempotent
package main

import (
	cache "github.com/hecc-blot/cache/service"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	httpSvc "github.com/hecc-blot/framework/service/http"
	ioc "github.com/hecc-blot/framework/service/ioc"
	"github.com/hecc-blot/guide/example/demo"
	app "github.com/hecc-blot/guide/example/internal/app"
	idempotentContract "github.com/hecc-blot/idempotent/contract"
	idempotentService "github.com/hecc-blot/idempotent/service"
	trace "github.com/hecc-blot/trace/service"
)

func main() {
	config := app.InitConf("../../config.yaml")

	traceSvc, traceClearUp := app.Must2(trace.NewTraceSvc(&config.Trace))
	cacheFactory := cache.NewCacheFactory(&config.Cache, traceSvc)
	defer func() {
		traceClearUp()
		if cacheFactory.Redis() != nil {
			_ = cacheFactory.Redis().Close()
		}
	}()

	responseSvc := httpSvc.NewResponseSvc()

	// 注册到 IOC 容器：幂等复用 cache 的 redis 连接（SetNX/Get/Eval 原子原语）
	container := ioc.New()
	container.Set(new(idempotentContract.IIdempotent), idempotentService.NewRedisIdempotent(cacheFactory.Redis()))
	container.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)

	// — 幂等：同一 key 只执行一次副作用业务 —
	apiHandle.Get("idempotent/demo", &demo.IdempotentDemoApi{})

	apiHandle.Listen()
}
