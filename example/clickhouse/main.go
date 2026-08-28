// 示例：db-clickhouse 模块 — 分析型数据库追加写入 + 条件查询
//
// 只演示 db-clickhouse 模块的用法。运行前需先在 config.yaml 填写 clickhouse 段。
//
// 运行：
//
//	cd example/clickhouse
//	go run .
//
// 启动后可用 curl 验证：
//
//	curl -X POST http://localhost:8080/clickhouse/demo
//
// 详见：github.com/hecc-blot/db-clickhouse
package main

import (
	log "github.com/hecc-blot/core/service/log"
	dbClickhouseContract "github.com/hecc-blot/db-clickhouse/contract"
	dbClickhouse "github.com/hecc-blot/db-clickhouse/service"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	httpSvc "github.com/hecc-blot/framework/service/http"
	ioc "github.com/hecc-blot/framework/service/ioc"
	"github.com/hecc-blot/guide/example/demo"
	app "github.com/hecc-blot/guide/example/internal/app"
)

func main() {
	config := app.InitConf("../../config.yaml")

	logSvc := app.Must(log.NewLogger(&config.Log.Local))
	clickhouseDb, clearUp := app.Must2(dbClickhouse.NewClickhouse(&config.Clickhouse, logSvc))
	defer clearUp()

	responseSvc := httpSvc.NewResponseSvc()

	// 注册到 IOC 容器
	container := ioc.New()
	container.Set(new(dbClickhouseContract.IDbClickhouse), clickhouseDb)
	container.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)

	// — 分析型数据库（ClickHouse）：追加写入 + 条件查询 —
	apiHandle.Post("clickhouse/demo", &demo.ClickhouseDemoApi{})

	apiHandle.Listen()
}
