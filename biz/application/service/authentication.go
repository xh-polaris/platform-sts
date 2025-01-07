package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/wire"
	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/xh-polaris/service-idl-gen-go/kitex_gen/platform/sts"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/xh-polaris/platform-sts/biz/infrastructure/config"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/consts"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/data/db"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/mapper"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/sdk/wechat"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/util"
	logx "github.com/xh-polaris/platform-sts/biz/infrastructure/util/log"
)

type IAuthenticationService interface {
	SignIn(ctx context.Context, req *sts.SignInReq) (*sts.SignInResp, error)
	SetPassword(ctx context.Context, req *sts.SetPasswordReq) (*sts.SetPasswordResp, error)
	SendVerifyCode(ctx context.Context, req *sts.SendVerifyCodeReq) (*sts.SendVerifyCodeResp, error)
	AddUserAuth(ctx context.Context, req *sts.AddUserAuthReq) (*sts.AddUserAuthResp, error)
}

type AuthenticationService struct {
	Config         *config.Config
	UserMapper     mapper.UserMapper
	MiniProgramMap wechat.MiniProgramMap
	Redis          *redis.Redis
}

var AuthenticationSet = wire.NewSet(
	wire.Struct(new(AuthenticationService), "*"),
	wire.Bind(new(IAuthenticationService), new(*AuthenticationService)),
)

func (s *AuthenticationService) AddUserAuth(ctx context.Context, req *sts.AddUserAuthReq) (*sts.AddUserAuthResp, error) {
	resp := &sts.AddUserAuthResp{}
	_, err := s.UserMapper.FindOne(ctx, req.UserId)
	switch err {
	case nil:
		return resp, nil
	case consts.ErrNotFound:
		oid, err := primitive.ObjectIDFromHex(req.UserId)
		if err != nil {
			return nil, consts.ErrInvalidObjectId
		}
		auth := make([]*db.Auth, 0)
		auth = append(auth, &db.Auth{
			Type:  req.Type,
			Value: req.UnionId,
		})
		err = s.UserMapper.Insert(ctx, &db.User{
			ID:       oid,
			CreateAt: time.Now(),
			UpdateAt: time.Now(),
			Auth:     auth,
		})
		if err != nil {
			return nil, err
		}
		return resp, nil
	default:
		return nil, err
	}
}

func (s *AuthenticationService) SignIn(ctx context.Context, req *sts.SignInReq) (*sts.SignInResp, error) {
	resp := &sts.SignInResp{}
	var err error
	switch req.AuthType {
	case consts.AuthTypeEmail:
		fallthrough
	case consts.AuthTypePhone:
		resp.UserId, err = s.signInByPassword(ctx, req)
	case consts.AuthTypeWechatOpenId:
		fallthrough
	case consts.AuthTypeWechat:
		resp.UserId, resp.UnionId, resp.OpenId, resp.AppId, err = s.signInByWechat(ctx, req)
	case consts.AuthTypeWechatPhone:
		resp.UserId, err = s.SignInByWechatPhone(ctx, req)
	default:
		return nil, consts.ErrInvalidArgument
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *AuthenticationService) signInByPassword(ctx context.Context, req *sts.SignInReq) (string, error) {
	UserMapper := s.UserMapper

	// 检查是否设置了验证码，若设置了检查验证码是否合法
	ok, err := s.checkVerifyCode(ctx, req.GetVerifyCode(), req.AuthId)
	if err != nil {
		return "", err
	}

	auth := &db.Auth{
		Type:  req.AuthType,
		Value: req.AuthId,
	}
	user, err := UserMapper.FindOneByAuth(ctx, auth)

	switch err {
	case nil:
	case consts.ErrNotFound:
		if !ok {
			return "", consts.ErrNoSuchUser
		}

		user = &db.User{Auth: []*db.Auth{auth}}
		err := UserMapper.Insert(ctx, user)
		if err != nil {
			return "", err
		}
		return user.ID.Hex(), nil
	default:
		return "", err
	}

	if ok {
		return user.ID.Hex(), nil
	}

	// 验证码未通过，尝试密码登录
	if user.Password == "" || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.GetPassword())) != nil {
		return "", consts.ErrWrongPassword
	}

	return user.ID.Hex(), nil
}

func (s *AuthenticationService) checkVerifyCode(ctx context.Context, except string, authValue string) (bool, error) {
	verifyCode, err := s.Redis.GetCtx(ctx, consts.VerifyCodeKeyPrefix+authValue)
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	} else if verifyCode == "" {
		logx.Info("查询到验证码为空")
		return false, nil
	} else if verifyCode != except {
		return false, nil
	} else {
		return true, nil
	}
}

func (s *AuthenticationService) SignInByWechatPhone(ctx context.Context, req *sts.SignInReq) (string, error) {
	code := req.GetVerifyCode() // 微信接口提供的换取手机号的code
	var accessToken string
	// 找到对应的小程序
	for _, conf := range s.Config.WechatApplicationConfigs {
		if req.AuthId == conf.AppID {
			res, err := util.HTTPGet(ctx, fmt.Sprintf(consts.WXAccessTokenUrl, conf.AppID, conf.AppSecret))
			logx.Info("微信AccessToken接口响应" + string(res))
			if err != nil {
				return "", err
			}
			var tokenRes map[string]any
			if err = sonic.Unmarshal(res, &tokenRes); err != nil {
				return "", err
			}
			if accessToken = tokenRes["access_token"].(string); accessToken == "" {
				return "", consts.ErrGetToken
			}
			break
		}
	}

	bodyString := fmt.Sprintf(`{"code":"%s"}`, code)
	body := strings.NewReader(bodyString)
	res, err := util.HTTPPost(ctx, fmt.Sprintf(consts.WXUserPhoneUrl, accessToken), body)
	if err != nil {
		return "", err
	}

	var phoneRes map[string]any
	if err = sonic.Unmarshal(res, &phoneRes); err != nil {
		return "", err
	} else if phoneRes["errcode"].(float64) != 0 {
		return "", errors.New(phoneRes["errmsg"].(string))
	}
	phoneInfo, ok := phoneRes["phone_info"].(map[string]any)
	if !ok {
		return "", errors.New("phone_info 类型断言失败")
	}
	// 获取到的手机号，国外的会有区号
	phone := phoneInfo["phoneNumber"].(string)

	// 这里类型用"phone", 因为本质上还是有手机登录，只不过换了一种验证方式
	UserMapper := s.UserMapper
	auth := &db.Auth{
		Type:  consts.AuthTypePhone,
		Value: phone,
	}

	user, err := UserMapper.FindOneByAuth(ctx, auth)
	switch {
	case err == nil:
		// 找到了则直接返回id即可
		return user.ID.Hex(), nil
	case errors.Is(err, consts.ErrNotFound):
		// 没找到需要创建
		auths := []*db.Auth{{
			Type:  consts.AuthTypePhone,
			Value: phone,
		}}
		user = &db.User{Auth: auths}
		err = UserMapper.Insert(ctx, user)
		if err != nil {
			return "", err
		}
		return user.ID.Hex(), nil
	default:
		return "", err
	}
}

// return userId unionId openId appid
func (s *AuthenticationService) signInByWechat(ctx context.Context, req *sts.SignInReq) (string, string, string, string, error) {
	jsCode := req.GetVerifyCode()

	var unionId string
	var openId string
	var appId string

	m := s.MiniProgramMap[req.GetAuthId()]
	if m != nil {
		// 向微信开放接口提交临时code
		res, err := m.Code2Session(ctx, jsCode)
		if err != nil {
			return "", "", "", "", err
		} else if res.ErrCode != 0 {
			return "", "", "", "", errors.New(res.ErrMsg)
		}
		unionId = res.UnionID
		openId = res.OpenID
		appId = req.GetAuthId()
	} else {
		for _, conf := range s.Config.WechatApplicationConfigs {
			if req.AuthId == conf.AppID {
				res, err := util.HTTPGet(ctx, fmt.Sprintf(consts.OAuthUrl, conf.AppID, conf.AppSecret, jsCode))
				if err != nil {
					return "", "", "", "", err
				}
				var j map[string]any
				if err = sonic.Unmarshal(res, &j); err != nil {
					return "", "", "", "", err
				}
				if id := j["unionid"]; id == "" {
					return "", "", "", "", consts.ErrWrongWechatCode
				}
				unionId = j["unionid"].(string)
				if _, ok := j["openid"]; !ok {
					return "", "", "", "", consts.ErrWrongWechatCode
				}
				openId = j["openid"].(string)
				appId = conf.AppID
			}
		}
	}

	if unionId == "" {
		return "", "", "", "", consts.ErrWrongWechatCode
	}
	UserMapper := s.UserMapper
	auth := &db.Auth{
		Type:  req.AuthType,
		Value: unionId,
	}
	if req.AuthType == consts.AuthTypeWechatOpenId {
		auth.AppId = appId
		auth.Value = openId
	}
	user, err := UserMapper.FindOneByAuth(ctx, auth)
	switch err {
	case nil:
		openAuth := &db.Auth{
			Type:  consts.AuthTypeWechatOpenId,
			Value: openId,
			AppId: appId,
		}
		_, ok := lo.Find(user.Auth, func(item *db.Auth) bool {
			return *item == *openAuth
		})
		if !ok {
			user.Auth = append(user.Auth, openAuth)
			err := UserMapper.Update(ctx, user)
			if err != nil {
				return "", "", "", "", err
			}
		}
		return user.ID.Hex(), unionId, openId, appId, nil
	case consts.ErrNotFound:
		auths := []*db.Auth{{
			Type:  consts.AuthTypeWechatOpenId,
			Value: openId,
			AppId: appId,
		}}
		if unionId != "" {
			auths = append(auths, &db.Auth{
				Type:  consts.AuthTypeWechat,
				Value: unionId,
			})
		}
		user = &db.User{Auth: auths}
		err = UserMapper.Insert(ctx, user)
		if err != nil {
			return "", "", "", "", err
		}
		return user.ID.Hex(), unionId, openId, appId, nil
	default:
		return "", "", "", "", err
	}
}

func (s *AuthenticationService) SetPassword(ctx context.Context, req *sts.SetPasswordReq) (*sts.SetPasswordResp, error) {
	user, err := s.UserMapper.FindOne(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(hashPassword)
	err = s.UserMapper.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	return &sts.SetPasswordResp{}, nil
}

func (s *AuthenticationService) SendVerifyCode(ctx context.Context, req *sts.SendVerifyCodeReq) (*sts.SendVerifyCodeResp, error) {
	var verifyCode string
	switch req.AuthType {
	case consts.AuthTypeEmail:
		c := s.Config.SMTP
		auth := smtp.PlainAuth("", c.Username, c.Password, c.Host)
		code, err := rand.Int(rand.Reader, big.NewInt(900000))
		code = code.Add(code, big.NewInt(100000))
		if err != nil {
			return nil, err
		}
		err = smtp.SendMail(c.Host+":"+strconv.Itoa(c.Port), auth, c.Username, []string{req.AuthId}, []byte(fmt.Sprintf(
			"To: %s\r\n"+
				"From: xh-polaris\r\n"+
				"Content-Type: text/plain"+"; charset=UTF-8\r\n"+
				"Subject: 验证码\r\n\r\n"+
				"您正在进行账号注册，本次注册验证码为：%s，5分钟内有效，请勿透露给其他人。\r\n", req.AuthId, code.String())))
		if err != nil {
			return nil, err
		}
		verifyCode = code.String()
	case consts.AuthTypePhone:
		c := s.Config.SMS
		code, err := rand.Int(rand.Reader, big.NewInt(900000))
		code = code.Add(code, big.NewInt(100000))
		if err != nil {
			return nil, err
		}
		phones := make([]string, 0)
		phones = append(phones, req.AuthId)
		err = callSMS(c, phones, code.String())
		if err != nil {
			return nil, err
		}
		verifyCode = code.String()

	default:
		return nil, errors.New("not implement")
	}
	err := s.Redis.SetexCtx(ctx, consts.VerifyCodeKeyPrefix+req.AuthId, verifyCode, 300)
	if err != nil {
		return nil, err
	}
	logx.Info("向%v:%v 发送验证码: %v", req.AuthType, req.AuthId, verifyCode)
	return &sts.SendVerifyCodeResp{}, nil
}

func callSMS(sms *config.SMSConfig, phones []string, code string) error {
	// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
	// 密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取
	credential := common.NewCredential(sms.SecretId, sms.SecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = sms.Host
	cpf.HttpProfile.ReqMethod = "POST"
	client := common.NewCommonClient(credential, sms.Region, cpf).WithLogger(log.Default())

	request := tchttp.NewCommonRequest("sms", sms.Version, sms.Action)
	params := make(map[string]interface{})
	params["PhoneNumberSet"] = phones
	params["SmsSdkAppId"] = sms.SmsSdkAppId
	params["TemplateId"] = sms.TemplateId
	params["SignName"] = sms.SignName
	// 模板参数
	codes := make([]string, 0)
	codes = append(codes, code)
	codes = append(codes, "5")
	params["TemplateParamSet"] = codes

	err := request.SetActionParameters(params)
	if err != nil {
		return err
	}

	response := tchttp.NewCommonResponse()
	err = client.Send(request, response)
	if err != nil {
		fmt.Println("fail to invoke api:", err.Error())
		return err
	}

	fmt.Println(string(response.GetBody()))
	return nil
}
