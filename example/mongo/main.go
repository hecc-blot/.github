// 示例：db-mongo 模块 — 文档型数据库集合写入 + 条件链式查询
//
// 只演示 db-mongo 模块的用法。运行前需先在 config.yaml 填写 mongo 段。
//
// 运行：
//
//	cd example/mongo
//	go run .
//
// 启动后可用 curl 验证：
//
//	curl -X POST http://localhost:8080/mongo/demo
//
// 详见：github.com/hecc-blot/db-mongo
package main

import (
	log "github.com/hecc-blot/core/service/log"
	dbMongoContract "github.com/hecc-blot/db-mongo/contract"
	dbMongo "github.com/hecc-blot/db-mongo/service"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	httpSvc "github.com/hecc-blot/framework/service/http"
	ioc "github.com/hecc-blot/framework/service/ioc"
	"github.com/hecc-blot/guide/example/demo"
	app "github.com/hecc-blot/guide/example/internal/app"
)

func main() {
	config := app.InitConf("../../config.yaml")

	logSvc := app.Must(log.NewLogger(&config.Log.Local))
	mongoDb, clearUp := app.Must2(dbMongo.NewMongo(&config.Mongo, logSvc))
	defer clearUp()

	responseSvc := httpSvc.NewResponseSvc()

	// 注册到 IOC 容器
	container := ioc.New()
	container.Set(new(dbMongoContract.IDbDocument), mongoDb)
	container.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)

	// — 文档型数据库（MongoDB）：插入文档 + 条件查询 —
	apiHandle.Post("mongo/demo", &demo.MongoDemoApi{})

	apiHandle.Listen()
}
