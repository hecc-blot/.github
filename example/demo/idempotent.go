package demo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	iCoreError "github.com/hecc-blot/framework/contract/error"
	"github.com/hecc-blot/framework/enum/response"
	errorSvc "github.com/hecc-blot/framework/service/error"
	idempotentContract "github.com/hecc-blot/idempotent/contract"

	"github.com/gin-gonic/gin"
)

// ===== 幂等 =====
// 演示：同一幂等键只执行一次业务；重复调用返回首次结果，并发重复返回 ErrProcessing。
// 适用于外部回调、接口重试、消息重复投递等防重场景。
// 详见：github.com/hecc-blot/idempotent

// IdempotentDemoApi 幂等示例
type IdempotentDemoApi struct {
	Idempotent idempotentContract.IIdempotent `inject:""`
}

func (a IdempotentDemoApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	key := "idempotent:demo"
	ttl := 30 * time.Second

	result, repeated, err := a.Idempotent.Run(ctx, key, ttl, func(ctx context.Context) ([]byte, error) {
		// 副作用业务：支付 / 下单 / 消息去重等，同一 key 只执行一次
		time.Sleep(50 * time.Millisecond) // 模拟业务执行耗时
		return json.Marshal(map[string]string{"order": "1001", "status": "paid"})
	})
	if err != nil {
		if errors.Is(err, idempotentContract.ErrProcessing) {
			return nil, errorSvc.NewError(response.Fail, fmt.Errorf("请求处理中，请稍后重试"))
		}
		return nil, errorSvc.NewError(response.Fail, fmt.Errorf("幂等执行失败: %w", err))
	}

	return map[string]interface{}{
		"repeated": repeated,
		"result":   string(result),
	}, nil
}
