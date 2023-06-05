package svc

import (
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"log"
	"net/http"

	"github.com/silenceper/wechat/v2/miniprogram"
	mpConfig "github.com/silenceper/wechat/v2/miniprogram/config"
	"github.com/xh-polaris/sts-rpc/internal/config"
	"github.com/xh-polaris/sts-rpc/model"

	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/tencentyun/qcloud-cos-sts-sdk/go"
)

type A interface {
	miniprogram.MiniProgram
}

type ServiceContext struct {
	Config      config.Config
	StsClient   *sts.Client
	CosClient   *cos.Client
	UrlModel    model.UrlModel
	Meowchat    *miniprogram.MiniProgram
	MeowchatOld *miniprogram.MiniProgram
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
		CosClient: cos.NewClient(b, &http.Client{}),
		UrlModel:  model.NewUrlModel(c.Mongo.URL, c.Mongo.DB, c.CacheConf),
		Meowchat: wechat.NewWechat().GetMiniProgram(&mpConfig.Config{
			AppID:     c.Meowchat.AppID,
			AppSecret: c.Meowchat.AppSecret,
			Cache:     cache.NewMemory(),
		}),
		MeowchatOld: wechat.NewWechat().GetMiniProgram(&mpConfig.Config{
			AppID:     c.MeowchatOld.AppID,
			AppSecret: c.MeowchatOld.AppSecret,
			Cache:     cache.NewMemory(),
		}),
	}
}
