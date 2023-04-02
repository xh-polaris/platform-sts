package svc

import (
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/xh-polaris/sts-rpc/internal/schedule"
	"log"
	"net/http"

	"github.com/xh-polaris/sts-rpc/internal/config"
	"github.com/xh-polaris/sts-rpc/model"

	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/tencentyun/qcloud-cos-sts-sdk/go"
)

type ServiceContext struct {
	Config    config.Config
	StsClient *sts.Client
	CosClient *cos.Client
	mq        *rocketmq.PushConsumer
	UrlModel  model.UrlModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	url, err := cos.NewBucketURL(c.CosConfig.BucketName, c.CosConfig.Region, true)
	if err != nil {
		log.Fatal(err)
	}
	b := &cos.BaseURL{
		BucketURL: url,
	}
	return &ServiceContext{
		Config: c,
		StsClient: sts.NewClient(
			c.CosConfig.SecretId,
			c.CosConfig.SecretKey,
			nil),
		mq:        schedule.CreateMQConsumer(&c),
		CosClient: cos.NewClient(b, &http.Client{}),
		UrlModel:  model.NewUrlModel(c.Mongo.URL, c.Mongo.DB, c.CacheConf),
	}
}
