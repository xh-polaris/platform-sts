package mq

import (
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/config"
)

func NewMqProducer(config *config.Config) (rocketmq.Producer, error) {
	produce, err := rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver(config.RocketMq.URL)),
		producer.WithRetry(config.RocketMq.Retry),
		producer.WithGroupName(config.RocketMq.GroupName),
	)
	if err != nil {
		return nil, err
	}
	err = produce.Start()
	if err != nil {
		return nil, err
	}
	return produce, nil
}
