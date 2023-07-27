package consts

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"google.golang.org/grpc/status"
)

var (
	ErrNotFound        = mon.ErrNotFound
	ErrInvalidObjectId = errors.New("invalid objectId")
)

var (
	ErrCannotDeleteObject = status.Error(10000, "can not delete object")
	ErrNoSuchUser         = status.Error(10001, "no such user")
	ErrWrongWechatCode    = status.Error(10002, "wrong wechat code")
	ErrInvalidArgument    = status.Error(10003, "invalid argument")
	ErrWrongPassword      = status.Error(10004, "wrong password")
)

const (
	AuthTypeEmail  = "email"
	AuthTypePhone  = "phone"
	AuthTypeWechat = "wechat"
)

const (
	VerifyCodeKeyPrefix = "verify:"
	OAuthUrl            = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
)
