package demo

import (
	"fmt"
	"net/http"
	"strings"

	iCoreApi "github.com/hecc-blot/framework/contract/api"
	"github.com/hecc-blot/framework/enum/response"
	errorSvc "github.com/hecc-blot/framework/service/error"
	ratelimitContract "github.com/hecc-blot/ratelimit/contract"
)

// ===== 中间件 =====
// 演示：定义 Token 校验中间件，中间件中通过 inject tag 注入依赖

// TokenMiddleware Token 鉴权中间件
type TokenMiddleware struct {
	ResponseSvc iCoreApi.IResponse `inject:""`
}

func (t TokenMiddleware) Middleware() iCoreApi.MiddlewareFunc {
	return func(ctx iCoreApi.IContext) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			t.ResponseSvc.Regular(ctx, nil, errorSvc.NewError(response.TokenInvalid, fmt.Errorf("缺少 Authorization 头")))
			ctx.Abort()
			return
		}
		// 实际项目可在此解析 JWT、查询用户信息等
		ctx.Set("token", token)
		ctx.Next()
	}
}

// RateLimitMiddleware 请求限流中间件 — 业务方实现。
// 演示：限流不在框架内置，业务方引入 ratelimit 模块后自行实现中间件，
// 按客户端 IP 限流，超限返回 429 + 统一响应格式（code=40006）。
type RateLimitMiddleware struct {
	Limiter ratelimitContract.RateLimiter
}

func (r *RateLimitMiddleware) Middleware() iCoreApi.MiddlewareFunc {
	return func(ctx iCoreApi.IContext) {
		if !r.Limiter.Allow(ctx, ctx.ClientIP()) {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, map[string]any{
				"code":    response.RateLimit,
				"message": response.CodeMap[response.RateLimit],
				"data":    nil,
			})
			return
		}
		ctx.Next()
	}
}

// SseAcceptMiddleware SSE Accept 头校验中间件
// 演示：策略性校验（如 Accept 头）通过中间件实现，而非框架内置
type SseAcceptMiddleware struct{}

func (m SseAcceptMiddleware) Middleware() iCoreApi.MiddlewareFunc {
	return func(ctx iCoreApi.IContext) {
		if !strings.Contains(ctx.GetHeader("Accept"), "text/event-stream") {
			ctx.String(http.StatusNotAcceptable, "Accept: text/event-stream required")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// SseCorsMiddleware SSE CORS 中间件
// 演示：浏览器 EventSource 跨域需要 CORS 响应头，策略性配置通过中间件实现
type SseCorsMiddleware struct{}

func (m SseCorsMiddleware) Middleware() iCoreApi.MiddlewareFunc {
	return func(ctx iCoreApi.IContext) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Accept, Content-Type, Last-Event-Id")
		if ctx.Method() == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	}
}
