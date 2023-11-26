package config

import (
	"fmt"
	"os"

	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/cache"
)

type CosConfig struct {
	AppId      string
	BucketName string
	Region     string
	SecretId   string
	SecretKey  string
}

func (c *CosConfig) CosHost() string {
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com", c.BucketName, c.Region)
}

func (c *CosConfig) CIHost() string {
	return fmt.Sprintf("https://%s.ci.%s.myqcloud.com", c.BucketName, c.Region)
}

type Config struct {
	service.ServiceConf
	ListenOn                 string
	CosConfig                *CosConfig
	CacheConf                cache.CacheConf
	WechatApplicationConfigs []*WechatApplicationConfig
	Redis                    *redis.RedisConf
	WeChatRedis              *redis.RedisConf
	DefaultWechatUser        *DefaultWechatUser
	SMTP                     *struct {
		Username string
		Password string
		Host     string
		Port     int
	}
	Mongo *struct {
		DB  string
		URL string
	}
	RocketMq *struct {
		URL       []string
		Retry     int
		GroupName string
	}
}

type DefaultWechatUser struct {
	AppId  string
	OpenId string
}

type WechatApplicationConfig struct {
	AppID     string
	AppSecret string
	Type      string
}

func NewConfig() (*Config, error) {
	c := new(Config)
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "etc/config.yaml"
	}
	err := conf.Load(path, c)
	if err != nil {
		return nil, err
	}
	err = c.SetUp()
	if err != nil {
		return nil, err
	}
	return c, nil
}
