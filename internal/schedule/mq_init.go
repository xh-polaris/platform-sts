package schedule

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	consumer2 "github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/xh-polaris/sts-rpc/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
)

func CreateMQConsumer(c *config.Config) *rocketmq.PushConsumer {
	consumer, err := rocketmq.NewPushConsumer(
		consumer2.WithNsResolver(primitive.NewPassthroughResolver(c.RocketMq.URL)),
		consumer2.WithGroupName(c.RocketMq.GroupName),
	)
	if err != nil {
		log.Fatal(err)
	}
	err = consumer.Subscribe(
		"sts-self",
		consumer2.MessageSelector{},
		func(ctx context.Context, ext ...*primitive.MessageExt) (consumer2.ConsumeResult, error) {
			for i := range ext {
				err := DelayMessageHandler(c, ext[i].Body)
				if err != nil {
					logx.Alert(fmt.Sprintf("%v", err))
				}
			}
			return consumer2.ConsumeSuccess, nil
		},
	)
	err = consumer.Subscribe("sts_used_url",
		consumer2.MessageSelector{},
		func(ctx context.Context, ext ...*primitive.MessageExt) (consumer2.ConsumeResult, error) {
			for i := range ext {
				err := UsedUrlMessageHandler(c, ext[i].Body)
				if err != nil {
					logx.Alert(fmt.Sprintf("%v", err))
				}
			}
			return consumer2.ConsumeSuccess, nil
		})
	if err != nil {
		log.Fatal(err)
	}
	err = consumer.Start()
	if err != nil {
		log.Fatal(err)
	}
	return &consumer
}
