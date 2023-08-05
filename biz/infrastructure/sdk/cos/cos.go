package cos

import (
	"net/http"
	"net/url"

	"github.com/google/wire"
	"github.com/tencentyun/cos-go-sdk-v5"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"

	"github.com/xh-polaris/platform-sts/biz/infrastructure/config"
)

func NewStsClient(config *config.Config) *sts.Client {
	return sts.NewClient(
		config.CosConfig.SecretId,
		config.CosConfig.SecretKey,
		nil)
}

func NewCosClient(config *config.Config) (*cos.Client, error) {
	//bucketURL, err := cos.NewBucketURL(config.CosConfig.BucketName, config.CosConfig.Region, true)
	bucketURL, err := url.Parse(config.CosConfig.CosHost())
	if err != nil {
		return nil, err
	}
	ciURL, err := url.Parse(config.CosConfig.CIHost())
	if err != nil {
		return nil, err
	}
	return cos.NewClient(&cos.BaseURL{
		BucketURL: bucketURL,
		CIURL:     ciURL,
	}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.CosConfig.SecretId,
			SecretKey: config.CosConfig.SecretKey,
		},
	}), nil
}

var CosSet = wire.NewSet(
	NewStsClient,
	NewCosClient,
)
