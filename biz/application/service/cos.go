package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/google/wire"
	"github.com/silenceper/wechat/v2/miniprogram/security"
	"github.com/silenceper/wechat/v2/miniprogram/subscribe"
	cossdk "github.com/tencentyun/cos-go-sdk-v5"
	cossts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
	"github.com/xh-polaris/service-idl-gen-go/kitex_gen/platform/sts"

	"github.com/xh-polaris/platform-sts/biz/infrastructure/config"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/consts"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/mapper"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/sdk/cos"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/sdk/wechat"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/util"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/util/log"
)

type ICosService interface {
	GenCosSts(ctx context.Context, req *sts.GenCosStsReq) (*sts.GenCosStsResp, error)
	GenSignedUrl(ctx context.Context, req *sts.GenSignedUrlReq) (*sts.GenSignedUrlResp, error)
	DeleteObject(ctx context.Context, req *sts.DeleteObjectReq) (*sts.DeleteObjectResp, error)
	TextCheck(ctx context.Context, req *sts.TextCheckReq) (*sts.TextCheckResp, error)
	PhotoCheck(ctx context.Context, req *sts.PhotoCheckReq) (*sts.PhotoCheckResp, error)
	SendMessage(ctx context.Context, req *sts.SendMessageReq) (*sts.SendMessageResp, error)
}

type CosService struct {
	Config         *config.Config
	CosSDK         *cos.CosSDK
	UrlMapper      mapper.UrlMapper
	UserMapper     mapper.UserMapper
	MiniProgramMap wechat.MiniProgramMap
}

var CosSet = wire.NewSet(
	wire.Struct(new(CosService), "*"),
	wire.Bind(new(ICosService), new(*CosService)),
)

func (s *CosService) SendMessage(ctx context.Context, req *sts.SendMessageReq) (*sts.SendMessageResp, error) {
	user := req.User.WechatUserMeta
	if user.AppId == "" {
		user.AppId = s.Config.DefaultWechatUser.AppId
	}
	mp := s.MiniProgramMap[user.AppId]
	if mp == nil {
		log.CtxError(ctx, "[SendMessage] appId not found")
		return &sts.SendMessageResp{}, nil
	}
	toUser := ""
	targetUser, err := s.UserMapper.FindOne(ctx, req.TargetUserId)
	if err != nil {
		return nil, err
	}
	for _, v := range targetUser.Auth {
		if v.Type == "wechat-openid" {
			toUser = v.Value
		}
	}
	if toUser == "" {
		return nil, consts.ErrOpenIdNotFind
	}
	data := make(map[string]*subscribe.DataItem)
	if req.MessageType == sts.MessageType_TypeFollowed {
		data["name1"] = &subscribe.DataItem{Value: req.SourceUserName}
		data["time2"] = &subscribe.DataItem{Value: time.Unix(req.CreateAt, 0)}
		data["thing3"] = &subscribe.DataItem{Value: "有新的朋友关注您, 快来看看"}
	} else if req.MessageType == sts.MessageType_TypeLiked {
		data["thing1"] = &subscribe.DataItem{Value: req.SourceContent}
		data["time2"] = &subscribe.DataItem{Value: time.Unix(req.CreateAt, 0)}
		data["thing3"] = &subscribe.DataItem{Value: req.SourceUserName}
	} else if req.MessageType == sts.MessageType_TypeComment {
		data["thing15"] = &subscribe.DataItem{Value: req.SourceContent}
		data["time3"] = &subscribe.DataItem{Value: time.Unix(req.CreateAt, 0)}
		data["thing5"] = &subscribe.DataItem{Value: req.SourceUserName}
		data["thing2"] = &subscribe.DataItem{Value: req.CommentText}
	}
	err = mp.Send(ctx, &subscribe.Message{
		ToUser:     toUser,
		TemplateID: s.Config.TemplateId[req.MessageType-1],
		Data:       data,
	})
	if err != nil {
		return nil, err
	}
	return &sts.SendMessageResp{}, nil
}

func (s *CosService) GenCosSts(ctx context.Context, req *sts.GenCosStsReq) (*sts.GenCosStsResp, error) {
	cosConfig := s.Config.CosConfig
	stsOption := &cossts.CredentialOptions{
		// 临时密钥有效时长，单位是秒
		DurationSeconds: int64(10 * time.Minute.Seconds()),
		Region:          cosConfig.Region,
		Policy: &cossts.CredentialPolicy{
			Statement: []cossts.CredentialPolicyStatement{
				{
					// 密钥的权限列表。简单上传和分片需要以下的权限，其他权限列表请看 https://cloud.tencent.com/document/product/436/31923
					Action: []string{
						// 简单上传
						"name/cos:PostObject",
						"name/cos:PutObject",
						// 分片上传
						"name/cos:InitiateMultipartUpload",
						"name/cos:ListMultipartUploads",
						"name/cos:ListParts",
						"name/cos:UploadPart",
						"name/cos:CompleteMultipartUpload",
					},
					Effect: "allow",
					// 密钥可控制的资源列表。此处开放名字为用户ID的文件夹及其子文件夹
					Resource: []string{
						fmt.Sprintf("qcs::cos:%s:uid/%s:%s/%s",
							cosConfig.Region, cosConfig.AppId, cosConfig.BucketName, req.Path),
					},
				},
			},
		},
	}

	res, err := s.CosSDK.GetCredential(ctx, stsOption)
	if err != nil {
		return nil, err
	}

	return &sts.GenCosStsResp{
		SecretId:     res.Credentials.TmpSecretID,
		SecretKey:    res.Credentials.TmpSecretKey,
		SessionToken: res.Credentials.SessionToken,
		ExpiredTime:  int64(res.ExpiredTime),
		StartTime:    int64(res.StartTime),
	}, nil
}

func (s *CosService) GenSignedUrl(ctx context.Context, req *sts.GenSignedUrlReq) (*sts.GenSignedUrlResp, error) {
	signedUrl, err := s.CosSDK.GetPresignedURL(ctx, req.Method, req.Path, req.SecretId, req.SecretKey, time.Minute, nil)
	if err != nil {
		return nil, err
	}
	//s.SendDelayMessage(s.Config, signedUrl)
	return &sts.GenSignedUrlResp{SignedUrl: signedUrl.String()}, nil
}

func (s *CosService) DeleteObject(ctx context.Context, req *sts.DeleteObjectReq) (*sts.DeleteObjectResp, error) {
	res, err := s.CosSDK.Delete(ctx, req.Path)
	if err != nil || res.StatusCode != 200 {
		return nil, consts.ErrCannotDeleteObject
	}
	return &sts.DeleteObjectResp{}, nil
}

func (s *CosService) TextCheck(ctx context.Context, req *sts.TextCheckReq) (*sts.TextCheckResp, error) {
	user := req.User.WechatUserMeta
	if user.AppId == "" {
		user.AppId = s.Config.DefaultWechatUser.AppId
		user.OpenId = s.Config.DefaultWechatUser.OpenId
	}
	mp := s.MiniProgramMap[user.AppId]
	if mp == nil {
		log.CtxError(ctx, "[TextCheck] appId not found")
		return &sts.TextCheckResp{Pass: false}, nil
	}
	checkRes, err := mp.MsgCheck(ctx, &security.MsgCheckRequest{
		OpenID:  user.OpenId,
		Scene:   security.MsgScene(req.GetScene()),
		Content: req.GetText(),
		Title:   req.GetTitle(),
	})
	if err != nil {
		return nil, err
	}
	if checkRes.ErrCode != 0 {
		return nil, errors.New(checkRes.Error())
	}
	if checkRes.Result.Suggest != security.CheckSuggestPass {
		log.CtxInfo(ctx, "[TextCheck] don't pass, label=%s", checkRes.Result.Label.String())
		return &sts.TextCheckResp{Pass: false}, nil
	}
	return &sts.TextCheckResp{Pass: true}, nil
}

func (s *CosService) PhotoCheck(ctx context.Context, req *sts.PhotoCheckReq) (*sts.PhotoCheckResp, error) {
	var input []cossdk.ImageAuditingInputOptions
	for key, rawUrl := range req.GetUrl() {
		input = append(input, cossdk.ImageAuditingInputOptions{
			DataId: strconv.Itoa(key),
			Url:    rawUrl,
		})
	}
	opt := &cossdk.BatchImageAuditingOptions{
		Input: input,
		Conf:  &cossdk.ImageAuditingJobConf{},
	}
	res, _, err := s.CosSDK.BatchImageAuditing(ctx, opt)
	log.CtxInfo(ctx, "[PhotoCheck] res=%s, err=%v", util.JSONF(res), err)
	if err != nil {
		return nil, err
	}

	for key := range req.GetUrl() {
		if res.JobsDetail[key].Result != 0 {
			for _, rawUrl := range req.GetUrl() {
				u, err := url.Parse(rawUrl)
				if err != nil {
					return nil, err
				}
				_, err = s.CosSDK.Delete(ctx, u.Path)
				if err != nil {
					return nil, err
				}
			}
			return &sts.PhotoCheckResp{Pass: false}, nil
		}
	}
	return &sts.PhotoCheckResp{Pass: true}, nil
}
