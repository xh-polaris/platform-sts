package logic

import (
	"context"
	"github.com/xh-polaris/sts-rpc/errorx"

	"github.com/xh-polaris/sts-rpc/internal/svc"
	"github.com/xh-polaris/sts-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteObjectLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteObjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteObjectLogic {
	return &DeleteObjectLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteObjectLogic) DeleteObject(in *pb.DeleteObjectReq) (*pb.DeleteObjectResp, error) {
	res, err := l.svcCtx.CosClient.Object.Delete(l.ctx, in.Path)
	if res.StatusCode != 200 || err != nil {
		return nil, errorx.ErrCannotDeleteObject
	}
	return &pb.DeleteObjectResp{}, nil
}
