package wechat

import (
	"github.com/google/wire"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/miniprogram"
	mpConfig "github.com/silenceper/wechat/v2/miniprogram/config"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/config"
)

type MiniProgramMap map[string]*miniprogram.MiniProgram

func NewWechatApplicationMap(config *config.Config) MiniProgramMap {
	m := make(map[string]*miniprogram.MiniProgram)
	wx := wechat.NewWechat()
	for _, conf := range config.WechatApplicationConfigs {
		if conf.Type == "miniprogram" {
			m[conf.AppID] = wx.GetMiniProgram(&mpConfig.Config{
				AppID:     conf.AppID,
				AppSecret: conf.AppSecret,
				Cache:     cache.NewMemory(),
			})
		}

	}
	return m
}

var WechatSet = wire.NewSet(
	NewWechatApplicationMap,
)
