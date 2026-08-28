// 示例：scheduler 模块 — 定时任务调度器（cron 注册 + 启动 + 优雅退出）
//
// 只演示 scheduler 模块的用法。调度器无 HTTP 端点，启动后周期执行任务、打印日志，Ctrl+C 优雅退出。
//
// 运行：
//
//	cd example/scheduler
//	go run .
//
// 启动后每 5 秒执行一次 cleanup 任务，观察日志输出（正式用法可将 cron 换成 "0 */5 * * * *" 等）。
//
// 详见：github.com/hecc-blot/scheduler
package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/hecc-blot/core/service/log"
	"github.com/hecc-blot/guide/example/demo"
	app "github.com/hecc-blot/guide/example/internal/app"
	scheduler "github.com/hecc-blot/scheduler/service"
	trace "github.com/hecc-blot/trace/service"
)

func main() {
	config := app.InitConf("../../config.yaml")

	logSvc := app.Must(log.NewLogger(&config.Log.Local))
	traceSvc, traceClearUp := app.Must2(trace.NewTraceSvc(&config.Trace))
	defer traceClearUp()

	schedulerSvc := app.Must(scheduler.NewScheduler(&config.Scheduler, logSvc, traceSvc, nil))

	// 注册定时任务：任务实现 scheduler.Job 接口（见 demo/scheduler.go 的 CleanupJob）
	_ = schedulerSvc.Add("cleanup", "*/5 * * * * *", demo.NewCleanupJob(logSvc))

	schedulerSvc.Start()
	defer schedulerSvc.Stop()

	// 阻塞直到收到 Ctrl+C / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
}
