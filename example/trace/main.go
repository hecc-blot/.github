// 示例：trace 模块 — 链路追踪 Span 的使用（FromContext / SetAttribute / 子 Span / RecordError）
//
// 只演示 trace 模块的用法。
//
// 运行：
//
//	cd example/trace
//	go run .
//
// 启动后可用 curl 验证：
//
//	curl http://localhost:8080/trace/demo
//
// 详见：github.com/hecc-blot/trace
package main

import (
	logContract "github.com/hecc-blot/core/contract/log"
	log "github.com/hecc-blot/core/service/log"
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

	logSvc := app.Must(log.NewLogger(&config.Log.Local))
	traceSvc, clearUp := app.Must2(trace.NewTraceSvc(&config.Trace))
	defer clearUp()

	responseSvc := httpSvc.NewResponseSvc()

	// 注册到 IOC 容器
	container := ioc.New()
	container.Set(new(traceContract.ITrace), traceSvc)
	container.Set(new(logContract.ILog), logSvc)
	container.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)

	// 注册 trace HTTP 中间件：TraceDemoApi 通过 FromContext 读取当前请求 Span
	apiHandle.Middleware(trace.NewHttpMiddleware(traceSvc))

	// — 链路追踪：Span / SetAttribute / 子 Span / RecordError —
	apiHandle.Get("trace/demo", &demo.TraceDemoApi{})

	apiHandle.Listen()
}
