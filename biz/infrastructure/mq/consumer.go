package mq

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	consumer2 "github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/config"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/consts"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/mapper"
	cos2 "github.com/xh-polaris/platform-sts/biz/infrastructure/sdk/cos"
	"github.com/zeromicro/go-zero/core/jsonx"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"net/url"
	"sync"
)

type DelayUrlMessage url.URL

var (
	urlModel mapper.UrlMapper
	umu      sync.Mutex

	cosClient *cos.Client
	cmu       sync.Mutex
)

func NewMqConsumer(c *config.Config) rocketmq.PushConsumer {
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
	return consumer
}

func DelayMessageHandler(c *config.Config, b []byte) error {
	checkSingletonModel(c)
	checkSingletonCos(c)
	msg := DelayUrlMessage{}
	err := jsonx.Unmarshal(b, &msg)
	if err != nil {
		return err
	}
	// 通过路径查找
	m, err := urlModel.FindOneByPath(context.Background(), msg.Path)
	if err != nil {
		if err == consts.ErrNotFound {
			return nil
		}
		return err
	} else {
		// 最大重复三次，如果不行就留存数据库，后期可以通过脚本删除
		res, err := cosClient.Object.Delete(context.Background(), msg.Path)
		suc := false
		if err != nil || res.StatusCode != 200 {
			for i := 0; i < 2; i++ {
				res, err = cosClient.Object.Delete(context.Background(), msg.Path)
				if err == nil && res.StatusCode == 200 {
					suc = true
					break
				}
			}
		}
		if suc {
			_ = urlModel.Delete(context.Background(), m.ID)
		}
	}
	return nil
}

func UsedUrlMessageHandler(c *config.Config, b []byte) error {
	checkSingletonModel(c)
	checkSingletonCos(c)
	var msg []url.URL
	err := jsonx.Unmarshal(b, &msg)
	if err != nil {
		return err
	}
	for _, u := range msg {
		m, err := urlModel.FindOneByPath(context.Background(), u.Path)
		if err != nil {
			if err == consts.ErrNotFound {
				return nil
			}
			return err
		}
		_ = urlModel.Delete(context.Background(), m.ID)
	}
	return nil
}

func checkSingletonModel(c *config.Config) {
	if urlModel == nil {
		umu.Lock()
		if urlModel == nil {
			Model := mapper.NewUrlMapper(c)
			urlModel = Model
		}
		umu.Unlock()
	}
}

func checkSingletonCos(c *config.Config) {
	if cosClient == nil {
		cmu.Lock()
		if cosClient == nil {
			Client, err := cos2.NewCosClient(c)
			if err != nil {
				log.Fatal(err)
			}
			cosClient = Client
		}
		cmu.Unlock()
	}
}
