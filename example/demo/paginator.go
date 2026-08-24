package demo

import (
	dbContract "github.com/hecc-blot/db/contract"
	iCoreError "github.com/hecc-blot/framework/contract/error"
	"github.com/hecc-blot/framework/enum/response"
	errorSvc "github.com/hecc-blot/framework/service/error"
	"github.com/hecc-blot/framework/util"

	"github.com/gin-gonic/gin"
)

// ===== 分页 =====
// 演示：Offset/limit 分页（NewPage）+ 游标分页（NewCursor）
// 详见：github.com/hecc-blot/framework

// PageRequest offset 分页请求参数
type PageRequest struct {
	Page     int `json:"page" binding:"min=1"`
	PageSize int `json:"pageSize" binding:"min=1,max=100"`
}

// PageListApi offset/limit 分页示例
type PageListApi struct {
	DbFactory dbContract.IDbFactory `inject:""`
	PageRequest
}

func (a PageListApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	opts := util.PageOpts{Page: a.Page, PageSize: a.PageSize}
	db := a.DbFactory.Build(ctx).Query(AccountModel{})

	total, err := db.Count()
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	var list []AccountModel
	offset := (opts.Page - 1) * opts.PageSize
	if err = db.Order("id DESC").Limit(opts.PageSize).Offset(offset).Find(&list); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// NewPage 自动处理 nil → []、默认 page/pageSize
	return util.NewPage(list, total, opts), nil
}

// CursorRequest 游标分页请求参数
type CursorRequest struct {
	Cursor   int `json:"cursor"`
	PageSize int `json:"pageSize" binding:"min=1,max=100"`
}

// CursorListApi 游标分页示例
type CursorListApi struct {
	DbFactory dbContract.IDbFactory `inject:""`
	CursorRequest
}

func (a CursorListApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	db := a.DbFactory.Build(ctx).Query(AccountModel{})

	// 多查一条用于判断 hasMore
	var list []AccountModel
	err := db.Where("id > ?", a.Cursor).Order("id ASC").Limit(a.PageSize + 1).Find(&list)
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// NewCursor 自动判断 hasMore 并截断多余数据
	// func(item *AccountModel) any 提取游标值（这里用 ID 作为游标）
	return util.NewCursor(list, a.PageSize, func(item *AccountModel) any {
		return item.ID
	}), nil
}
