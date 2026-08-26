package demo

import (
	"fmt"

	dbEsContract "github.com/hecc-blot/db-es/contract"
	iCoreError "github.com/hecc-blot/framework/contract/error"
	"github.com/hecc-blot/framework/enum/response"
	errorSvc "github.com/hecc-blot/framework/service/error"

	"github.com/gin-gonic/gin"
)

// ===== 搜索型数据库（Elasticsearch）=====
// 演示：文档索引 + Query DSL 搜索
// 注意：依赖 Elasticsearch 实例（见 main.go 组装），未配置时本路由不可用。
// 详见：github.com/hecc-blot/db-es

// EsDemoApi Elasticsearch 示例：索引文档 + 搜索
type EsDemoApi struct {
	Search dbEsContract.IDbSearch `inject:""`
}

func (a EsDemoApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	if a.Search == nil {
		return nil, errorSvc.NewError(response.Fail, fmt.Errorf("Elasticsearch 未配置，请先在 config.yaml 填写 es 段"))
	}

	db := a.Search.WithContext(ctx)

	// 索引文档
	if err := db.Index("articles").IndexDoc("1", &ArticleModel{Title: "golang"}); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// Query DSL 搜索（命中文档 _source 解码到 []ArticleModel）
	query := map[string]interface{}{
		"match": map[string]interface{}{"title": "golang"},
	}
	var articles []ArticleModel
	if err := db.Index("articles").Search(query, &articles); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}
	return articles, nil
}
