package logic

import (
	"context"
	"fmt"

	"time"

	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/xh-polaris/sts-rpc/internal/svc"
	"github.com/xh-polaris/sts-rpc/pb"
)

type GenCosStsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenCosStsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenCosStsLogic {
	return &GenCosStsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GenCosStsLogic) GenCosSts(in *pb.GenCosStsReq) (*pb.GenCosStsResp, error) {
	cosConfig := l.svcCtx.Config.CosConfig
	stsOption := &sts.CredentialOptions{
		// 临时密钥有效时长，单位是秒
		DurationSeconds: int64(10 * time.Minute.Seconds()),
		Region:          cosConfig.Region,
		Policy: &sts.CredentialPolicy{
			Statement: []sts.CredentialPolicyStatement{
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
						fmt.Sprintf("qcs::cos:%s:uid/%s:%s/users/%s/%s",
							cosConfig.Region, cosConfig.AppId, cosConfig.BucketName, in.UserId, in.Path),
					},
				},
			},
		},
	}

	res, err := l.svcCtx.StsClient.GetCredential(stsOption)
	if err != nil {
		return nil, err
	}

	return &pb.GenCosStsResp{
		SecretId:     res.Credentials.TmpSecretID,
		SecretKey:    res.Credentials.TmpSecretKey,
		SessionToken: res.Credentials.SessionToken,
		ExpiredTime:  int64(res.ExpiredTime),
		StartTime:    int64(res.StartTime),
	}, nil
}
