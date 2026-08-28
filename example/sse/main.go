// 示例：sse 模块 — Server-Sent Events 推送（与 API 共享端口）
//
// 只演示 sse 模块的用法。
//
// 运行：
//
//	cd example/sse
//	go run .
//
// 启动后可用 curl 验证（每秒推送一次服务器时间）：
//
//	curl -N http://localhost:8080/events/time
//
// 详见：github.com/hecc-blot/sse
package main

import (
	logContract "github.com/hecc-blot/core/contract/log"
	log "github.com/hecc-blot/core/service/log"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	httpSvc "github.com/hecc-blot/framework/service/http"
	ioc "github.com/hecc-blot/framework/service/ioc"
	"github.com/hecc-blot/guide/example/demo"
	app "github.com/hecc-blot/guide/example/internal/app"
	sse "github.com/hecc-blot/sse/service"
)

func main() {
	config := app.InitConf("../../config.yaml")

	logSvc := app.Must(log.NewLogger(&config.Log.Local))
	responseSvc := httpSvc.NewResponseSvc()

	// 注册到 IOC 容器
	container := ioc.New()
	container.Set(new(logContract.ILog), logSvc)
	container.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)

	// SSE 复用 API 的路由注册（不感知具体 HTTP 内核），Shutdown 作为 apiHandle.Listen 的退出回调
	sseHandle := sse.NewSseSvc(apiHandle, container)
	sseGroup := sseHandle.Group("", &demo.SseAcceptMiddleware{}, &demo.SseCorsMiddleware{})

	// — SSE 推送：GET（EventSource）+ POST（fetch + ReadableStream）—
	sseGroup.Get("events/time", &demo.ExampleSse{})
	sseGroup.Post("events/time", &demo.ExampleSse{})

	apiHandle.Listen(sseHandle.Shutdown)
}
