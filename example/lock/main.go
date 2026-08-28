// 示例：lock 模块 — 分布式锁（TryLock / Lock / Renew / Unlock）
//
// 只演示 lock 模块的用法。依赖 Redis（复用 cache 模块的 redis 连接），运行前需先配置 config.yaml 的 redis 段。
//
// 运行：
//
//	cd example/lock
//	go run .
//
// 启动后可用 curl 验证：
//
//	curl http://localhost:8080/lock/demo
//
// 详见：github.com/hecc-blot/lock
package main

import (
	cache "github.com/hecc-blot/cache/service"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	httpSvc "github.com/hecc-blot/framework/service/http"
	ioc "github.com/hecc-blot/framework/service/ioc"
	"github.com/hecc-blot/guide/example/demo"
	app "github.com/hecc-blot/guide/example/internal/app"
	lockContract "github.com/hecc-blot/lock/contract"
	lockService "github.com/hecc-blot/lock/service"
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

	// 注册到 IOC 容器：分布式锁复用 cache 的 redis 连接（SetNX/Eval 原子原语）
	container := ioc.New()
	container.Set(new(lockContract.ILocker), lockService.NewRedisLocker(cacheFactory.Redis()))
	container.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)

	// — 分布式锁：非阻塞/阻塞获取、续期、释放 —
	apiHandle.Get("lock/demo", &demo.LockDemoApi{})

	apiHandle.Listen()
}
