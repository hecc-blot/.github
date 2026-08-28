package demo

import (
	"context"
	"fmt"
	"time"

	logContract "github.com/hecc-blot/core/contract/log"
	iCoreApi "github.com/hecc-blot/framework/contract/api"
	iCoreError "github.com/hecc-blot/framework/contract/error"
	"github.com/hecc-blot/framework/enum/response"
	errorSvc "github.com/hecc-blot/framework/service/error"
	mqContract "github.com/hecc-blot/mq/contract"
)

// ===== 消息队列 =====
// 演示：IMqFactory / IProducer / IConsumer 的注入与使用
//   - Producer.Send 发送消息；Consumer.Subscribe 订阅消费（长期运行，见 startMqConsumer）
//   - 高级能力（延迟消息 / 顺序消费）按后端是否实现，通过类型断言启用
// 注意：实际收发依赖 Kafka/NSQ broker（见 main.go 的 MQ 组装），未配置 broker 时本路由不可用。
// 详见：github.com/hecc-blot/mq

// MqDemoApi 消息队列示例：生产一条消息 + 按需发送延迟消息
type MqDemoApi struct {
	MqFactory mqContract.IMqFactory `inject:""`
}

func (a MqDemoApi) Call(ctx iCoreApi.IContext) (interface{}, iCoreError.IError) {
	if a.MqFactory == nil {
		return nil, errorSvc.NewError(response.Fail, fmt.Errorf("MQ 未配置 broker，请先在 config.yaml 配置 mq 段"))
	}

	const topic = "example-topic"

	// 1. 生产：构造统一消息体（序列化由业务自行处理）
	producer := a.MqFactory.Producer()
	msg := &mqContract.Message{
		Topic: topic,
		Key:   "order-1",
		Body:  []byte(`{"order_id": 1, "amount": 99}`),
	}
	if err := producer.Send(ctx, msg); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// 2. 延迟消息（仅部分后端支持，如 NSQ / RocketMQ，Kafka 不实现）——类型断言启用
	delayed := "not supported by current driver"
	if p, ok := producer.(mqContract.IDelayedProducer); ok {
		if err := p.SendDelay(ctx, &mqContract.Message{Topic: topic, Body: []byte("later")}, time.Now().Add(time.Minute)); err != nil {
			return nil, errorSvc.NewError(response.Fail, err)
		}
		delayed = "sent"
	}

	return map[string]any{
		"topic":   topic,
		"sent":    true,
		"delayed": delayed,
	}, nil
}

// startMqConsumer 消费端订阅示例（长期运行，通常在 main 组装后以 goroutine 启动）。
// 本示例未接入 broker，故不自动启动；此处仅作用法参考。
func startMqConsumer(factory mqContract.IMqFactory, logger logContract.ILog) {
	consumer := factory.Consumer()
	_ = consumer.Subscribe(context.Background(), "example-topic", "example-group",
		func(c context.Context, msg *mqContract.Message) error {
			logger.Info(c, "consume message", "topic", msg.Topic, "body", string(msg.Body))
			// 显式确认；省略时按 handler 返回值自动结算（nil=确认，非 nil=拒绝重投）
			return msg.Ack(c)
		},
		mqContract.WithConcurrency(2), // 消费并发数
	)
}
