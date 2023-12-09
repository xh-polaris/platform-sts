package provider

import (
	"github.com/google/wire"

	"github.com/xh-polaris/platform-sts/biz/application/service"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/config"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/mapper"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/mq"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/sdk/cos"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/sdk/wechat"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/stores/redis"
)

var AllProvider = wire.NewSet(
	ApplicationSet,
	InfrastructureSet,
)

var ApplicationSet = wire.NewSet(
	service.CosSet,
	service.AuthenticationSet,
)

var InfrastructureSet = wire.NewSet(
	config.NewConfig,
	redis.NewRedis,
	MapperSet,
	SDKSet,
	MqSet,
)

var SDKSet = wire.NewSet(
	cos.CosSet,
	wechat.WechatSet,
)

var MapperSet = wire.NewSet(
	mapper.NewUrlMapper,
	mapper.NewUserMapper,
)

var MqSet = wire.NewSet(
	mq.NewMqProducer,
)
