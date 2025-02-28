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
	ErrOpenIdNotFind      = status.Error(10005, "openId not find")
	ErrGetToken           = status.Error(10006, "get wx token failed")
)

const (
	AuthTypeEmail        = "email"
	AuthTypePhone        = "phone"
	AuthTypeWechat       = "wechat"
	AuthTypeWechatOpenId = "wechat-openid"
	AuthTypeWechatPhone  = "wechat-phone"
)

const (
	VerifyCodeKeyPrefix = "verify:"
	OAuthUrl            = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	WXAccessTokenUrl    = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	WXUserPhoneUrl      = "https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s"
)
