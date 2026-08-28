// 示例：mq 模块 — 消息队列生产/消费（Kafka/NSQ）
//
// 只演示 mq 模块的用法。运行前需先在 config.yaml 配置 mq 段（driver + broker）。
//
// 运行：
//
//	cd example/mq
//	go run .
//
// 启动后可用 curl 验证（向 example-topic 生产一条消息）：
//
//	curl -X POST http://localhost:8080/mq/demo
//
// 详见：github.com/hecc-blot/mq
package main

import (
	log "github.com/hecc-blot/core/service/log"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	httpSvc "github.com/hecc-blot/framework/service/http"
	ioc "github.com/hecc-blot/framework/service/ioc"
	"github.com/hecc-blot/guide/example/demo"
	app "github.com/hecc-blot/guide/example/internal/app"
	mqContract "github.com/hecc-blot/mq/contract"
	mqService "github.com/hecc-blot/mq/service"
	trace "github.com/hecc-blot/trace/service"
)

func main() {
	config := app.InitConf("../../config.yaml")

	logSvc := app.Must(log.NewLogger(&config.Log.Local))
	traceSvc, traceClearUp := app.Must2(trace.NewTraceSvc(&config.Trace))

	// 消息队列工厂：依赖 broker（Kafka/NSQ），未配置时构造会返回错误
	mqFactory, mqCleanup := app.Must2(mqService.NewMqFactory(&config.Mq, logSvc, traceSvc))
	defer func() {
		mqCleanup()
		traceClearUp()
	}()

	responseSvc := httpSvc.NewResponseSvc()

	// 注册到 IOC 容器
	container := ioc.New()
	container.Set(new(mqContract.IMqFactory), mqFactory)
	container.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)

	// — 消息队列：生产消息 + 按需延迟消息 —
	apiHandle.Post("mq/demo", &demo.MqDemoApi{})

	apiHandle.Listen()
}
