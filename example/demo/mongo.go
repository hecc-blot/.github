package demo

import (
	"fmt"

	dbMongoContract "github.com/hecc-blot/db-mongo/contract"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	iCoreError "github.com/hecc-blot/framework/contract/error"
	"github.com/hecc-blot/framework/enum/response"
	errorSvc "github.com/hecc-blot/framework/service/error"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ===== 文档型数据库（MongoDB）=====
// 演示：集合 + 条件链式查询 + 文档写入
// 注意：依赖 MongoDB 实例（见 main.go 组装），未配置时本路由不可用。
// 详见：github.com/hecc-blot/db-mongo

// MongoDemoApi MongoDB 示例：插入文档 + 条件查询
type MongoDemoApi struct {
	Doc dbMongoContract.IDbDocument `inject:""`
}

func (a MongoDemoApi) Call(ctx iCoreApi.IContext) (interface{}, iCoreError.IError) {
	if a.Doc == nil {
		return nil, errorSvc.NewError(response.Fail, fmt.Errorf("MongoDB 未配置，请先在 config.yaml 填写 mongo 段"))
	}

	db := a.Doc.WithContext(ctx)

	// 插入文档（filter/projection/update/sort 均为 BSON 文档）
	id, err := db.Collection("users").InsertOne(bson.M{"name": "tom", "age": 20})
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// 条件查询
	var users []bson.M
	if err := db.Collection("users").
		Where(bson.M{"age": bson.M{"$gte": 18}}).
		Sort(bson.D{{Key: "age", Value: -1}}).
		Limit(10).
		Find(&users); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	return map[string]interface{}{"inserted_id": id, "users": users}, nil
}
