// 示例：db-es 模块 — 搜索型数据库文档索引 + Query DSL 搜索
//
// 只演示 db-es 模块的用法。运行前需先在 config.yaml 填写 es 段。
//
// 运行：
//
//	cd example/es
//	go run .
//
// 启动后可用 curl 验证：
//
//	curl -X POST http://localhost:8080/es/demo
//
// 详见：github.com/hecc-blot/db-es
package main

import (
	log "github.com/hecc-blot/core/service/log"
	dbEsContract "github.com/hecc-blot/db-es/contract"
	dbEs "github.com/hecc-blot/db-es/service"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	httpSvc "github.com/hecc-blot/framework/service/http"
	ioc "github.com/hecc-blot/framework/service/ioc"
	"github.com/hecc-blot/guide/example/demo"
	app "github.com/hecc-blot/guide/example/internal/app"
)

func main() {
	config := app.InitConf("../../config.yaml")

	logSvc := app.Must(log.NewLogger(&config.Log.Local))
	esDb, clearUp := app.Must2(dbEs.NewEs(&config.Es, logSvc))
	defer clearUp()

	responseSvc := httpSvc.NewResponseSvc()

	// 注册到 IOC 容器
	container := ioc.New()
	container.Set(new(dbEsContract.IDbSearch), esDb)
	container.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)

	// — 搜索型数据库（Elasticsearch）：索引文档 + 搜索 —
	apiHandle.Post("es/demo", &demo.EsDemoApi{})

	apiHandle.Listen()
}
