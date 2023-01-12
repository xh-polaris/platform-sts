package logic

import (
	"context"
	"time"

	"github.com/xh-polaris/sts-rpc/internal/svc"
	"github.com/xh-polaris/sts-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenSignedUrlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenSignedUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenSignedUrlLogic {
	return &GenSignedUrlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GenSignedUrlLogic) GenSignedUrl(in *pb.GenSignedUrlReq) (*pb.GenSignedUrlResp, error) {
	url, err := l.svcCtx.CosClient.Object.GetPresignedURL(l.ctx, in.Method, in.Path, in.SecretId, in.SecretKey, time.Minute, nil)
	if err != nil {
		return nil, err
	}
	return &pb.GenSignedUrlResp{SignedUrl: url.String()}, nil
}
