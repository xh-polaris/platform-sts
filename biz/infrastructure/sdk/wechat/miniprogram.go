package wechat

import (
	"context"
	"fmt"
	"github.com/google/wire"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/miniprogram"
	mpConfig "github.com/silenceper/wechat/v2/miniprogram/config"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"time"
)

type RedisPlus struct {
	*redis.Redis
}

func (r RedisPlus) Get(key string) interface{} {
	data, err := r.Redis.GetCtx(context.Background(), key)
	if err != nil || data == "" {
		return nil
	}
	return data
}

func (r RedisPlus) Set(key string, val interface{}, timeout time.Duration) error {
	str := fmt.Sprintf("%v", val)
	err := r.Redis.SetexCtx(context.Background(), key, str, int(timeout))
	return err
}

func (r RedisPlus) IsExist(key string) bool {
	data, _ := r.Redis.Exists(key)
	return data
}

func (r RedisPlus) Delete(key string) error {
	_, err := r.Redis.Del(key)
	return err
}

func NewRedisPlus(conf redis.RedisConf) RedisPlus {
	return RedisPlus{
		Redis: redis.MustNewRedis(conf),
	}
}

type MiniProgramMap map[string]*miniprogram.MiniProgram

func NewWechatApplicationMap(config *config.Config) MiniProgramMap {
	m := make(map[string]*miniprogram.MiniProgram)
	wx := wechat.NewWechat()
	for _, conf := range config.WechatApplicationConfigs {
		if conf.Type == "miniprogram" {
			m[conf.AppID] = wx.GetMiniProgram(&mpConfig.Config{
				AppID:     conf.AppID,
				AppSecret: conf.AppSecret,
				Cache:     NewRedisPlus(*config.Redis),
			})
		}

	}
	return m
}

var WechatSet = wire.NewSet(
	NewWechatApplicationMap,
)
