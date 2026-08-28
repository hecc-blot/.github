package demo

import (
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	iCoreError "github.com/hecc-blot/framework/contract/error"
	"github.com/hecc-blot/framework/enum/response"
	errorSvc "github.com/hecc-blot/framework/service/error"
	httpClientContract "github.com/hecc-blot/httpclient/contract"
)

// ===== HTTP 客户端 =====
// 演示：统一 HTTP 客户端（内置重试 + 结构化日志 + 敏感头脱敏 + traceId 透传）
//   - GET 默认重试（幂等），POST 默认不重试（非幂等）
//   - WithHeader / WithTimeout / WithRetry 等 Option 覆盖单次请求行为
// 注意：示例访问 https://httpbin.org，需可访问外网；替换成你自己的目标 URL 即可。
// 详见：github.com/hecc-blot/httpclient

// HttpClientDemoApi HTTP 客户端示例：GET + POST
type HttpClientDemoApi struct {
	HttpClient httpClientContract.IHttpClient `inject:""`
}

func (a HttpClientDemoApi) Call(ctx iCoreApi.IContext) (interface{}, iCoreError.IError) {
	// 1. GET：默认重试（幂等），可通过 WithHeader 追加请求头
	getResp, err := a.HttpClient.Get(ctx, "https://httpbin.org/get",
		httpClientContract.WithHeader("X-Demo", "hecc-blot"),
	)
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// 2. POST：body 传任意 Go 值，内部自动 JSON 序列化并补 Content-Type
	postResp, err := a.HttpClient.Post(ctx, "https://httpbin.org/post", map[string]any{
		"name": "hecc-blot",
		"note": "POST 自动 JSON 序列化",
	})
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	return map[string]any{
		"get": map[string]any{
			"status": getResp.StatusCode,
			"body":   string(getResp.Body),
		},
		"post": map[string]any{
			"status":  postResp.StatusCode,
			"body":    string(postResp.Body),
			"retries": postResp.Retries,
		},
	}, nil
}
