package demo

import (
	"context"
	"time"

	cacheContract "github.com/hecc-blot/cache/contract"
	dbContract "github.com/hecc-blot/db/contract"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	iCoreError "github.com/hecc-blot/framework/contract/error"
	"github.com/hecc-blot/framework/enum/response"
	errorSvc "github.com/hecc-blot/framework/service/error"
)

// ===== 缓存操作 =====
// 演示：本地缓存 + Redis 缓存的 Get/Set/Del/Exists、Redis Hash 操作、缓存穿透回写
// 详见：github.com/hecc-blot/cache

// CacheBasicApi 缓存基础操作
type CacheBasicApi struct {
	CacheFactory cacheContract.ICacheFactory `inject:""`
}

func (a CacheBasicApi) Call(ctx iCoreApi.IContext) (interface{}, iCoreError.IError) {
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

func (a CacheHashApi) Call(ctx iCoreApi.IContext) (interface{}, iCoreError.IError) {
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

// CacheReadThroughApi 缓存读穿透 — 编排层一行完成「查缓存 → 查库 → 回填」（内置防击穿 + 空值防穿透）
type CacheReadThroughApi struct {
	CacheFactory cacheContract.ICacheFactory `inject:""`
	Db           dbContract.IDb              `inject:""`
}

func (a CacheReadThroughApi) Call(ctx iCoreApi.IContext) (interface{}, iCoreError.IError) {
	// 只传「取数闭包」：框架统一处理缓存未命中、并发合并、回填与空值防穿透
	account, err := a.CacheFactory.Orchestrator().GetOrLoad(ctx, "account:1",
		func(ctx context.Context) (interface{}, error) {
			db := a.Db.WithContext(ctx)
			var account AccountModel
			if err := db.Where("id = ?", 1).Take(&account); err != nil {
				return nil, err
			}
			return account, nil
		})
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	return account, nil
}
