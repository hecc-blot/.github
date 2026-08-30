// 示例：circuitbreaker 模块 — 熔断器（连续失败熔断 + 冷却半开探测）
//
// 只演示 circuitbreaker 模块的用法。进程内本地状态，无外部依赖，无需任何服务即可运行。
//
// 运行：
//
//	cd example/circuitbreaker
//	go run .
//
// 启动后可用 curl 验证（连续 ?fail=true 触发熔断，冷却后半开探测）：
//
//	curl -H "Authorization: demo" "http://localhost:9500/circuitbreaker/demo?fail=true"
//	curl -H "Authorization: demo" "http://localhost:9500/circuitbreaker/demo"
//
// 详见：github.com/hecc-blot/circuitbreaker
package main

import (
	circuitbreakerConfig "github.com/hecc-blot/circuitbreaker/config"
	circuitbreakerContract "github.com/hecc-blot/circuitbreaker/contract"
	circuitbreakerSvc "github.com/hecc-blot/circuitbreaker/service"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	httpSvc "github.com/hecc-blot/framework/service/http"
	ioc "github.com/hecc-blot/framework/service/ioc"
	"github.com/hecc-blot/guide/example/demo"
	app "github.com/hecc-blot/guide/example/internal/app"
)

func main() {
	config := app.InitConf("../../config.yaml")

	responseSvc := httpSvc.NewResponseSvc()

	// 注册到 IOC 容器：熔断器为进程内本地状态，注册为共享实例（各请求注入同一 breaker）
	container := ioc.New()
	container.Set(new(circuitbreakerContract.Breaker), circuitbreakerSvc.New(circuitbreakerConfig.Config{FailureThreshold: 3}))
	container.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)

	// — 熔断器：Allow/Record 状态机（?fail=true 记一次失败） —
	apiHandle.Get("circuitbreaker/demo", &demo.CircuitBreakerDemoApi{})

	apiHandle.Listen()
}
