package demo

import (
	"fmt"
	"time"

	iCoreError "github.com/hecc-blot/framework/contract/error"
	logContract "github.com/hecc-blot/framework/contract/log"
	traceContract "github.com/hecc-blot/trace/contract"

	"github.com/gin-gonic/gin"
)

// ===== 链路追踪 =====
// 演示：FromContext / SetAttribute / RecordError / Start 子 Span / defer span.End()

// TraceDemoApi 链路追踪示例
type TraceDemoApi struct {
	TraceSvc traceContract.ITrace `inject:""`
	LogSvc   logContract.ILog     `inject:""`
}

func (a TraceDemoApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	// 1. 从 Context 获取当前请求的 Span（由 HttpTraceMiddleware 自动创建）
	currentSpan := a.TraceSvc.FromContext(ctx)

	// 2. 为当前 Span 添加自定义属性
	currentSpan.SetAttribute("business.type", "trace_demo")
	currentSpan.SetAttribute("user.id", 12345)

	// 3. 开启子 Span 追踪数据库操作
	subCtx, subSpan := a.TraceSvc.Start(ctx, "db-slow-query",
		"db.table", "account",
		"db.operation", "find",
	)
	defer subSpan.End()

	// 模拟耗时操作
	time.Sleep(10 * time.Millisecond)

	// 4. 模拟出错时记录错误到 Span
	if false { // 实际业务中将条件替换为 err != nil
		subSpan.RecordError(fmt.Errorf("模拟数据库错误"))
	}

	a.LogSvc.Info(subCtx, "trace demo completed", "span", subSpan.Name())
	return "trace demo ok", nil
}
