package demo

import (
	"context"
	"fmt"
	"time"

	iCoreApi "github.com/hecc-blot/framework/contract/api"
	iCoreError "github.com/hecc-blot/framework/contract/error"
	"github.com/hecc-blot/framework/enum/response"
	errorSvc "github.com/hecc-blot/framework/service/error"
	lockContract "github.com/hecc-blot/lock/contract"
)

// ===== 分布式锁 =====
// 演示：基于 Redis 的分布式锁，用于共享资源并发互斥（扣减计数、去重、防重复触发）。
// TryLock 非阻塞，拿不到立即返回 ErrLocked；Lock 阻塞等待；Unlock 释放；Renew 手动续期。
// 详见：github.com/hecc-blot/lock

// LockDemoApi 分布式锁示例
type LockDemoApi struct {
	Locker lockContract.ILocker `inject:""`
}

func (a LockDemoApi) Call(ctx iCoreApi.IContext) (interface{}, iCoreError.IError) {
	key := "lock:demo"
	ttl := 30 * time.Second

	// 1. 非阻塞获取锁：拿不到立即返回 ErrLocked（errors.Is 判定被占用）
	lk, err := a.Locker.TryLock(ctx, key, ttl)
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, fmt.Errorf("获取锁失败: %w", err))
	}
	// 解锁用独立 context，避免请求 ctx 取消导致释放失败（锁 TTL 兜底）
	defer lk.Unlock(context.Background())

	// 2. 临界区业务逻辑（此处模拟长任务）
	time.Sleep(50 * time.Millisecond)

	// 3. 手动续期：长任务在临界区内调用，避免锁过期（或用 WithAutoRenew 自动续期）
	if err := lk.Renew(ctx); err != nil {
		return nil, errorSvc.NewError(response.Fail, fmt.Errorf("续期失败: %w", err))
	}

	// 4. 阻塞获取：Lock 会等待直到拿到锁或超时
	lk2, err := a.Locker.Lock(ctx, "lock:demo-blocking", ttl, time.Second)
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, fmt.Errorf("阻塞获取失败: %w", err))
	}
	defer lk2.Unlock(context.Background())

	return "分布式锁获取/释放成功", nil
}
