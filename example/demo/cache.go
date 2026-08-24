package demo

import (
	"time"

	cacheContract "github.com/hecc-blot/cache/contract"
	dbContract "github.com/hecc-blot/db/contract"
	iCoreError "github.com/hecc-blot/framework/contract/error"
	"github.com/hecc-blot/framework/enum/response"
	errorSvc "github.com/hecc-blot/framework/service/error"

	"github.com/gin-gonic/gin"
)

// ===== 缓存操作 =====
// 演示：本地缓存 + Redis 缓存的 Get/Set/Del/Exists、Redis Hash 操作、缓存穿透回写
// 详见：github.com/hecc-blot/cache

// CacheBasicApi 缓存基础操作
type CacheBasicApi struct {
	CacheFactory cacheContract.ICacheFactory `inject:""`
}

func (a CacheBasicApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	// 本地缓存 — Set / Get / Exists / Del
	_ = a.CacheFactory.Local().Set(ctx, "local:key", "hello", 10*time.Minute)

	if ok, _ := a.CacheFactory.Local().Exists(ctx, "local:key"); ok {
		val, _ := a.CacheFactory.Local().Get(ctx, "local:key")
		_ = a.CacheFactory.Local().Del(ctx, "local:key")
		_ = val
	}

	// Redis 缓存 — Set / Get / Del
	_ = a.CacheFactory.Redis().Set(ctx, "redis:key", "world", time.Hour)
	val, _ := a.CacheFactory.Redis().Get(ctx, "redis:key")
	_ = a.CacheFactory.Redis().Del(ctx, "redis:key")

	return val, nil
}

// CacheHashApi Redis Hash 操作
type CacheHashApi struct {
	CacheFactory cacheContract.ICacheFactory `inject:""`
}

func (a CacheHashApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	// HSet — 同时设置多个 field
	err := a.CacheFactory.Redis().HSet(ctx, "user:1", "name", "john", "email", "john@test.com")
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// HGet — 获取单个 field
	name, _ := a.CacheFactory.Redis().HGet(ctx, "user:1", "name")

	// HDel — 删除指定 field
	_ = a.CacheFactory.Redis().HDel(ctx, "user:1", "email")

	return name, nil
}

// CacheReadThroughApi 缓存读穿透 — 先查缓存，未命中则查 DB 并回写缓存
type CacheReadThroughApi struct {
	CacheFactory cacheContract.ICacheFactory `inject:""`
	DbFactory    dbContract.IDbFactory       `inject:""`
}

func (a CacheReadThroughApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	cacheKey := "account:1"

	// 1. 先从本地缓存读取
	if cached, _ := a.CacheFactory.Local().Get(ctx, cacheKey); cached != nil {
		return cached, nil
	}

	// 2. 缓存未命中，查数据库
	db := a.DbFactory.Build(ctx)
	var account AccountModel
	if err := db.Where("id = ?", 1).Take(&account); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// 3. 回写缓存（本地 + Redis 双写）
	_ = a.CacheFactory.Local().Set(ctx, cacheKey, account, 10*time.Minute)
	_ = a.CacheFactory.Redis().Set(ctx, cacheKey, account, 10*time.Minute)

	return account, nil
}
