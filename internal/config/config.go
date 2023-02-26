package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	CosConfig struct {
		AppId      string
		BucketName string
		Region     string
		SecretId   string
		SecretKey  string
	}
	Mongo struct {
		DB  string
		URL string
	}
	CacheConf cache.CacheConf
}
