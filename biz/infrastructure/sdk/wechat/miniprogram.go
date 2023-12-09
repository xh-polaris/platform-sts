package wechat

import (
	"context"
	"fmt"
	"time"

	"github.com/google/wire"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/miniprogram"
	"github.com/silenceper/wechat/v2/miniprogram/auth"
	mpConfig "github.com/silenceper/wechat/v2/miniprogram/config"
	"github.com/silenceper/wechat/v2/miniprogram/security"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/xh-polaris/platform-sts/biz/infrastructure/config"
)

type RedisPlus struct {
	*redis.Redis
}

func (r *RedisPlus) Get(key string) interface{} {
	panic("implement me")
}

func (r *RedisPlus) Set(key string, val interface{}, timeout time.Duration) error {
	panic("implement me")
}

func (r *RedisPlus) IsExist(key string) bool {
	panic("implement me")
}

func (r *RedisPlus) Delete(key string) error {
	panic("implement me")
}

func (r *RedisPlus) GetContext(ctx context.Context, key string) interface{} {
	data, err := r.Redis.GetCtx(ctx, key)
	if err != nil || data == "" {
		return nil
	}
	return data
}

func (r *RedisPlus) SetContext(ctx context.Context, key string, val interface{}, timeout time.Duration) error {
	str := fmt.Sprintf("%v", val)
	err := r.Redis.SetexCtx(ctx, key, str, int(timeout.Seconds()))
	return err
}

func (r *RedisPlus) IsExistContext(ctx context.Context, key string) bool {
	data, _ := r.Redis.ExistsCtx(ctx, key)
	return data
}

func (r *RedisPlus) DeleteContext(ctx context.Context, key string) error {
	_, err := r.Redis.DelCtx(ctx, key)
	return err
}

func NewRedisPlus(conf redis.RedisConf) *RedisPlus {
	return &RedisPlus{
		Redis: redis.MustNewRedis(conf),
	}
}

type MiniProgramSDK struct {
	sdk *miniprogram.MiniProgram
}

func (s *MiniProgramSDK) MsgCheck(ctx context.Context, in *security.MsgCheckRequest) (res security.MsgCheckResponse, err error) {
	ctx, span := trace.TracerFromContext(ctx).Start(ctx, "mp/security/MsgCheck", oteltrace.WithTimestamp(time.Now()), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer func() {
		span.End(oteltrace.WithTimestamp(time.Now()))
	}()
	return s.sdk.GetSecurity().MsgCheck(in)
}

func (s *MiniProgramSDK) Code2Session(ctx context.Context, jsCode string) (result auth.ResCode2Session, err error) {
	ctx, span := trace.TracerFromContext(ctx).Start(ctx, "mp/auth/Code2Session", oteltrace.WithTimestamp(time.Now()), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer func() {
		span.End(oteltrace.WithTimestamp(time.Now()))
	}()
	return s.sdk.GetAuth().Code2SessionContext(ctx, jsCode)
}

type MiniProgramMap map[string]*MiniProgramSDK

func NewWechatApplicationMap(config *config.Config) MiniProgramMap {
	m := make(map[string]*MiniProgramSDK)
	wx := wechat.NewWechat()
	for _, conf := range config.WechatApplicationConfigs {
		if conf.Type == "miniprogram" {
			m[conf.AppID] = &MiniProgramSDK{sdk: wx.GetMiniProgram(&mpConfig.Config{
				AppID:     conf.AppID,
				AppSecret: conf.AppSecret,
				Cache:     NewRedisPlus(*config.WeChatRedis),
			})}
		}

	}
	return m
}

var WechatSet = wire.NewSet(
	NewWechatApplicationMap,
)
