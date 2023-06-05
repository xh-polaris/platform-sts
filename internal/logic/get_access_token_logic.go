package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/silenceper/wechat/v2/util"

	"github.com/xh-polaris/sts-rpc/internal/svc"
	"github.com/xh-polaris/sts-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	getAccessTokenUrl = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
)

type GetAccessTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAccessTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccessTokenLogic {
	return &GetAccessTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAccessTokenLogic) GetAccessToken(in *pb.GetAccessTokenReq) (resp *pb.GetAccessTokenResp, err error) {

	var appId, appSecret string

	if in.App == "old" {
		appId = l.svcCtx.MeowchatOld.GetContext().AppID
		appSecret = l.svcCtx.MeowchatOld.GetContext().AppSecret
		println("old")
	}
	if in.App == "meowchat" {
		appId = l.svcCtx.Meowchat.GetContext().AppID
		appSecret = l.svcCtx.Meowchat.GetContext().AppSecret
		println("meowchat")
	}
	if in.App == "manager" {
		appId = l.svcCtx.Config.MeowchatManager.AppID
		appSecret = l.svcCtx.Config.MeowchatManager.AppSecret
		println("manager")
	}

	println(appId)
	data, err := util.HTTPGetContext(l.ctx, fmt.Sprintf(getAccessTokenUrl, appId, appSecret))
	if err != nil {
		return nil, err
	}
	var j map[string]any
	if err = json.Unmarshal(data, &j); err != nil {
		return nil, err
	}
	resp = &pb.GetAccessTokenResp{}
	resp.AccessToken = j["access_token"].(string)
	fmt.Printf(resp.AccessToken)
	resp.ExpiresIn = int64(j["expires_in"].(float64))
	return
}
