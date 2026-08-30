package demo

import (
	"fmt"

	circuitbreakerContract "github.com/hecc-blot/circuitbreaker/contract"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	iCoreError "github.com/hecc-blot/framework/contract/error"
	"github.com/hecc-blot/framework/enum/response"
	errorSvc "github.com/hecc-blot/framework/service/error"
)

// ===== 熔断器 =====
// 演示：调用外部依赖前 Allow 判断是否放行、调用后 Record 记录结果。
// 连续失败触发 open（快速失败），冷却后半开探测（half_open），探测成功闭合（closed）。
// 熔断打开时调用方自行降级（返回兜底数据）；用 ?fail=true 模拟一次失败，连续触发观察状态机流转。
// 详见：github.com/hecc-blot/circuitbreaker

// CircuitBreakerDemoApi 熔断器示例
type CircuitBreakerDemoApi struct {
	Breaker circuitbreakerContract.Breaker `inject:""`
}

func (a CircuitBreakerDemoApi) Call(ctx iCoreApi.IContext) (interface{}, iCoreError.IError) {
	// 模拟外部依赖本次调用是否失败：curl "…/circuitbreaker/demo?fail=true" 记一次失败
	fail := ctx.Query("fail") == "true"

	// 1. 调用前判断是否放行：open 且未过冷却期直接快速失败
	if !a.Breaker.Allow() {
		// 熔断打开：业务方决定降级策略（返回缓存兜底 / 默认值等），此处返回友好提示
		return map[string]any{
			"allowed": false,
			"state":   a.Breaker.State().String(),
			"message": "熔断打开，返回降级兜底结果",
		}, nil
	}

	// 2. 调用外部依赖并记录结果（nil 成功，非 nil 失败）
	var err error
	if fail {
		err = fmt.Errorf("downstream unavailable")
	}
	a.Breaker.Record(err)

	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	return map[string]any{
		"allowed": true,
		"state":   a.Breaker.State().String(),
		"message": "下游调用成功",
	}, nil
}
