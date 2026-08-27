package demo

import (
	"context"
	"fmt"
	"time"

	logContract "github.com/hecc-blot/core/contract/log"
	sseContract "github.com/hecc-blot/sse/contract"
)

// ===== SSE 推送 =====
// 演示：ISse 接口 + Writer 写入抽象（心跳、Flusher 断言由框架处理）

// ExampleSse SSE 实时推送示例
type ExampleSse struct {
	LogSvc logContract.ILog `inject:""`
}

func (e ExampleSse) Serve(ctx context.Context, w sseContract.Writer) error {
	e.LogSvc.Info(ctx, "sse connection established")

	// 业务推送：每秒推送服务器时间
	business := time.NewTicker(1 * time.Second)
	defer business.Stop()

	for {
		select {
		case <-ctx.Done():
			// 客户端断开或心跳写入失败
			e.LogSvc.Info(ctx, "sse client disconnected")
			return nil
		case <-business.C:
			msg := fmt.Sprintf("当前服务器时间：%s", time.Now().Format(time.RFC3339))
			if err := w.Send("", "", msg); err != nil {
				return err
			}
		}
	}
}
