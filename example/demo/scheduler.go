package demo

import (
	"context"

	logContract "github.com/hecc-blot/core/contract/log"
	"go.uber.org/zap"
)

// ===== 定时任务 =====
// 演示：IScheduler 的任务注册与执行
//   - 任务实现 scheduler.Job 接口（Run(ctx) error），依赖在构造时注入
//   - 覆盖超时处理、周期对账、批量清理等场景
// 详见：github.com/hecc-blot/scheduler

// CleanupJob 定时清理示例任务：周期性清理过期数据。
// 依赖在构造时传入（调度器不经过 IOC 反射注入，由业务方在组装层注入）。
type CleanupJob struct {
	logSvc logContract.ILog
}

// NewCleanupJob 构造清理任务。
func NewCleanupJob(logSvc logContract.ILog) *CleanupJob {
	return &CleanupJob{logSvc: logSvc}
}

// Run 实现 scheduler.Job 接口。
func (j *CleanupJob) Run(ctx context.Context) error {
	j.logSvc.Info(ctx, "定时清理任务执行", zap.String("job", "cleanup"))
	return nil
}
