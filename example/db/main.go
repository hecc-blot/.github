// 示例：db 模块 — 数据库 CRUD 与多库切换
//
// 只演示 db 模块的用法，其余组件（缓存/追踪/限流等）见 example/ 下其它 main 与 demo/ 子包。
//
// 运行：
//
//	cd example/db
//	go run .
//
// 启动后可用 curl 验证（示例固定查询 id=1，请先通过 account/add 造数）：
//
//	curl -X POST http://localhost:8080/account/add -H 'Authorization: x' -d '{"account_name":"tom","password":"123","email":"tom@test.com"}'
//	curl http://localhost:8080/account/take
//	curl http://localhost:8080/account/db-switch
//
// 详见：github.com/hecc-blot/db
package main

import (
	logContract "github.com/hecc-blot/core/contract/log"
	log "github.com/hecc-blot/core/service/log"
	dbContract "github.com/hecc-blot/db/contract"
	db "github.com/hecc-blot/db/service"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	httpSvc "github.com/hecc-blot/framework/service/http"
	ioc "github.com/hecc-blot/framework/service/ioc"
	"github.com/hecc-blot/guide/example/demo"
	app "github.com/hecc-blot/guide/example/internal/app"
)

func main() {
	config := app.InitConf("../../config.yaml")

	// 日志：db 模块的构造与示例 API 都依赖 ILog，这里用本地日志（改用 SLS 见 log-sls 模块）
	logSvc := app.Must(log.NewLogger(&config.Log.Local))

	// 两种用法：
	//   - IDbFactory：多库切换（DbSwitchApi 演示，运行时 Build(ctx, dbEnum.Postgres) 指定库）
	//   - IDb：单库直连（CRUD 示例直接注入，无需工厂）
	dbFactory, dbClearUp := app.Must2(db.NewDbFactory(&config.Db, logSvc))
	mysqlDb, mysqlClearUp := app.Must2(db.NewMysql(config.Db.Mysql, logSvc))
	defer func() {
		mysqlClearUp()
		dbClearUp()
	}()

	responseSvc := httpSvc.NewResponseSvc()

	// 注册到 IOC 容器（注入依赖供 demo 子包里的 API 通过 inject tag 获取）
	container := ioc.New()
	container.Set(new(dbContract.IDbFactory), dbFactory)
	container.Set(new(dbContract.IDb), mysqlDb)
	container.Set(new(logContract.ILog), logSvc)
	container.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := httpSvc.NewApiSvc(&config.Server, responseSvc, container)

	// — 参数校验 + 数据库 CRUD + 多库切换 —
	apiHandle.Post("account/add", &demo.AddAccountApi{})
	apiHandle.Get("account/take", &demo.TakeAccountApi{})
	apiHandle.Get("account/find", &demo.FindAccountApi{})
	apiHandle.Get("account/count", &demo.CountAccountApi{})
	apiHandle.Post("account/update", &demo.UpdateAccountApi{})
	apiHandle.Post("account/delete", &demo.DeleteAccountApi{})
	apiHandle.Get("account/db-switch", &demo.DbSwitchApi{})

	apiHandle.Listen()
}
