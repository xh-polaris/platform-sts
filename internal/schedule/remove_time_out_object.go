package schedule

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/xh-polaris/sts-rpc/internal/config"
	"github.com/xh-polaris/sts-rpc/model"
	"github.com/zeromicro/go-zero/core/jsonx"
	"github.com/zeromicro/go-zero/core/stores/monc"
)

var (
	prod *rocketmq.Producer
	pmu  sync.Mutex

	urlModel *model.UrlModel
	umu      sync.Mutex

	cosClient *cos.Client
	cmu       sync.Mutex
)

type DelayUrlMessage url.URL

func checkSingleProducer(c *config.Config) {
	if prod == nil {
		pmu.Lock()
		if prod == nil {
			produce, err := rocketmq.NewProducer(
				producer.WithNsResolver(primitive.NewPassthroughResolver(c.RocketMq.URL)),
				producer.WithRetry(c.RocketMq.Retry),
				producer.WithGroupName(c.RocketMq.GroupName),
			)
			if err != nil {
				log.Fatal()
			}
			err = produce.Start()
			if err != nil {
				return
			}
			prod = &produce
		}
		pmu.Unlock()
	}
}

// SendDelayMessage 发送延迟消息，最多重试两次，如果两次都无法发送，认为MQ宕机，不再发送
func SendDelayMessage(c *config.Config, message interface{}) {
	checkSingleProducer(c)
	json, _ := jsonx.Marshal(message)
	msg := &primitive.Message{
		Topic: "sts-self",
		Body:  json,
	}
	// level 8 means delay 5min
	msg.WithDelayTimeLevel(8)
	res, err := (*prod).SendSync(context.Background(), msg)
	if err != nil || res.Status != primitive.SendOK {
		for i := 0; i < 2; i++ {
			res, err := (*prod).SendSync(context.Background(), msg)
			if err == nil && res.Status == primitive.SendOK {
				break
			}
		}
	}
}

func checkSingletonModel(c *config.Config) {
	if urlModel == nil {
		umu.Lock()
		if urlModel == nil {
			Model := model.NewUrlModel(c.Mongo.URL, c.Mongo.DB, c.CacheConf)
			urlModel = &Model
		}
		umu.Unlock()
	}
}

func checkSingletonCos(c *config.Config) {
	if cosClient == nil {
		cmu.Lock()
		if cosClient == nil {
			URL, err := cos.NewBucketURL(c.CosConfig.BucketName, c.CosConfig.Region, true)
			if err != nil {
				log.Fatal(err)
			}
			b := &cos.BaseURL{
				BucketURL: URL,
			}
			cosClient = cos.NewClient(b, &http.Client{})
		}
		cmu.Unlock()
	}
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
	m, err := (*urlModel).FindOneByPath(context.Background(), msg.Path)
	if err != nil {
		if err == monc.ErrNotFound {
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
			_ = (*urlModel).Delete(context.Background(), m.ID.String())
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
		m, err := (*urlModel).FindOneByPath(context.Background(), u.Path)
		if err != nil {
			if err == monc.ErrNotFound {
				return nil
			}
			return err
		}
		_ = (*urlModel).Delete(context.Background(), m.ID.String())
	}
	return nil
}
