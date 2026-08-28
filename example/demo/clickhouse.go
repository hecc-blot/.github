package demo

import (
	"fmt"

	dbClickhouseContract "github.com/hecc-blot/db-clickhouse/contract"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	iCoreError "github.com/hecc-blot/framework/contract/error"
	"github.com/hecc-blot/framework/enum/response"
	errorSvc "github.com/hecc-blot/framework/service/error"
)

// ===== 分析型数据库（ClickHouse）=====
// 演示：追加写入 + 条件查询（无事务、无更新删除）
// 注意：依赖 ClickHouse 实例（见 main.go 组装），未配置时本路由不可用。
// 详见：github.com/hecc-blot/db-clickhouse

// ClickhouseDemoApi ClickHouse 示例：写入日志 + 条件查询
type ClickhouseDemoApi struct {
	Ch dbClickhouseContract.IDbClickhouse `inject:""`
}

func (a ClickhouseDemoApi) Call(ctx iCoreApi.IContext) (interface{}, iCoreError.IError) {
	if a.Ch == nil {
		return nil, errorSvc.NewError(response.Fail, fmt.Errorf("ClickHouse 未配置，请先在 config.yaml 填写 clickhouse 段"))
	}

	db := a.Ch.WithContext(ctx)

	// 追加写入
	if err := db.BatchInsert(&[]LogModel{
		{Date: "2026-08-26", Msg: "hello"},
		{Date: "2026-08-27", Msg: "world"},
	}); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// 条件查询
	var rows []LogModel
	if err := db.Where("date >= ?", "2026-08-01").Limit(100).Find(&rows); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}
	return rows, nil
}
