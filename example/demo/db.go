package demo

import (
	dbContract "github.com/hecc-blot/db/contract"
	dbEnum "github.com/hecc-blot/db/enum/db"
	iCoreError "github.com/hecc-blot/framework/contract/error"
	logContract "github.com/hecc-blot/framework/contract/log"
	"github.com/hecc-blot/framework/enum/response"
	errorSvc "github.com/hecc-blot/framework/service/error"

	"github.com/gin-gonic/gin"
)

// ===== 数据库 CRUD =====
// 演示：Add / Take / Find / Select / Save / Remove / Order / Count + 事务 Begin/Commit/Rollback
// 详见：github.com/hecc-blot/db

// AddAccountApi 新增账户 + 事务演示
type AddAccountApi struct {
	DbFactory dbContract.IDbFactory `inject:""`
	LogSvc    logContract.ILog      `inject:""`
	AddAccountRequest
}

func (a AddAccountApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	account := AccountModel{
		AccountName: a.AccountName,
		Password:    a.Password,
		Email:       a.Email,
	}

	db := a.DbFactory.Build(ctx)

	// 开启事务
	tx := db.Begin()
	if err := tx.Add(&account); err != nil {
		tx.Rollback()
		return nil, errorSvc.NewError(response.Fail, err)
	}
	// 同时写入关联订单
	order := OrderModel{AccountID: account.ID, Product: "新用户礼包", Amount: 0}
	if err := tx.Add(&order); err != nil {
		tx.Rollback()
		return nil, errorSvc.NewError(response.Fail, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	a.LogSvc.Info(ctx, "account created", "id", account.ID)
	return account, nil
}

// TakeAccountApi 查询单条记录
type TakeAccountApi struct {
	DbFactory dbContract.IDbFactory `inject:""`
}

func (a TakeAccountApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	db := a.DbFactory.Build(ctx)
	var account AccountModel
	if err := db.Where("id = ?", 1).Take(&account); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}
	return account, nil
}

// FindAccountApi 查询多条记录（条件筛选 + 排序 + 字段选择）
type FindAccountApi struct {
	DbFactory dbContract.IDbFactory `inject:""`
}

func (a FindAccountApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	db := a.DbFactory.Build(ctx)
	var list []AccountModel
	if err := db.
		Select("id, account_name, email").
		Where("id >= ?", 1).
		Order("id DESC").
		Find(&list); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}
	return list, nil
}

// CountAccountApi 统计记录数
type CountAccountApi struct {
	DbFactory dbContract.IDbFactory `inject:""`
}

func (a CountAccountApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	db := a.DbFactory.Build(ctx)
	count, err := db.Query(AccountModel{}).Where("id >= ?", 1).Count()
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}
	return count, nil
}

// UpdateAccountApi 更新记录
type UpdateAccountApi struct {
	DbFactory dbContract.IDbFactory `inject:""`
}

func (a UpdateAccountApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	db := a.DbFactory.Build(ctx)
	updateData := AccountModel{AccountName: "updated_name", Email: "new@example.com"}
	if err := db.Where("id = ?", 1).Save(&updateData); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}
	return updateData, nil
}

// DeleteAccountApi 删除记录
type DeleteAccountApi struct {
	DbFactory dbContract.IDbFactory `inject:""`
}

func (a DeleteAccountApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	db := a.DbFactory.Build(ctx)
	if err := db.Where("id = ?", 1).Remove(&AccountModel{}); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}
	return nil, nil
}

// ===== 多数据库切换 =====
// 演示：SetDefault() 切换默认库、Build(ctx, dbEnum.xxx) 运行时指定数据库
// 详见：github.com/hecc-blot/db

// DbSwitchApi 多数据库切换 — 展示同时操作 MySQL 和 PostgreSQL
type DbSwitchApi struct {
	DbFactory dbContract.IDbFactory `inject:""`
}

func (a DbSwitchApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	// 方式一：使用默认数据库（通常是 MySQL）
	mysqlDB := a.DbFactory.Build(ctx)

	// 方式二：运行时指定数据库类型
	pgDB := a.DbFactory.Build(ctx, dbEnum.Postgres)

	// 分别从两个数据库查询
	var mysqlAccounts []AccountModel
	if err := mysqlDB.Where("id >= ?", 1).Find(&mysqlAccounts); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	var pgAccounts []AccountModel
	if err := pgDB.Where("id >= ?", 1).Find(&pgAccounts); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// 还可以运行时切换默认库
	// a.DbFactory.SetDefault(dbEnum.Postgres)

	return map[string]interface{}{
		"mysql": mysqlAccounts,
		"pg":    pgAccounts,
	}, nil
}
