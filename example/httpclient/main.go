// 示例：httpclient 模块 — 统一出站 HTTP 客户端（重试 + 结构化日志 + 敏感头脱敏）
//
// 只演示 httpclient 模块的用法。示例访问 https://httpbin.org，需可访问外网。
//
// 运行：
//
//	cd example/httpclient
//	go run .
//
// 启动后可用 curl 验证：
//
//	curl http://localhost:8080/httpclient/demo
//
// 详见：github.com/hecc-blot/httpclient
package main

import (
	log "github.com/hecc-blot/core/service/log"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	httpSvc "github.com/hecc-blot/framework/service/http"
	ioc "github.com/hecc-blot/framework/service/ioc"
	"github.com/hecc-blot/guide/example/demo"
	app "github.com/hecc-blot/guide/example/internal/app"
	httpClientContract "github.com/hecc-blot/httpclient/contract"
	httpClientService "github.com/hecc-blot/httpclient/service"
)

func main() {
	config := app.InitConf("../../config.yaml")

	logSvc := app.Must(log.NewLogger(&config.Log.Local))
	httpClient := httpClientService.NewHttpClient(config.HttpClient, logSvc)

	responseSvc := httpSvc.NewResponseSvc()

	// 注册到 IOC 容器
	container := ioc.New()
	container.Set(new(httpClientContract.IHttpClient), httpClient)
	container.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)

	// — HTTP 客户端：GET / POST + 重试与结构化日志 —
	apiHandle.Get("httpclient/demo", &demo.HttpClientDemoApi{})

	apiHandle.Listen()
}
