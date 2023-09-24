package adaptor

import (
	"context"

	"github.com/xh-polaris/platform-sts/biz/application/service"

	"github.com/xh-polaris/platform-sts/biz/infrastructure/config"

	"github.com/xh-polaris/service-idl-gen-go/kitex_gen/platform/sts"
)

type StsServerImpl struct {
	*config.Config
	CosService            service.ICosService
	AuthenticationService service.IAuthenticationService
}

func (s *StsServerImpl) GenCosSts(ctx context.Context, req *sts.GenCosStsReq) (res *sts.GenCosStsResp, err error) {
	return s.CosService.GenCosSts(ctx, req)
}

func (s *StsServerImpl) GenSignedUrl(ctx context.Context, req *sts.GenSignedUrlReq) (res *sts.GenSignedUrlResp, err error) {
	return s.CosService.GenSignedUrl(ctx, req)
}

func (s *StsServerImpl) DeleteObject(ctx context.Context, req *sts.DeleteObjectReq) (res *sts.DeleteObjectResp, err error) {
	return s.CosService.DeleteObject(ctx, req)
}

func (s *StsServerImpl) TextCheck(ctx context.Context, req *sts.TextCheckReq) (res *sts.TextCheckResp, err error) {
	return s.CosService.TextCheck(ctx, req)
}

func (s *StsServerImpl) PhotoCheck(ctx context.Context, req *sts.PhotoCheckReq) (res *sts.PhotoCheckResp, err error) {
	return s.CosService.PhotoCheck(ctx, req)
}

func (s *StsServerImpl) SignIn(ctx context.Context, req *sts.SignInReq) (res *sts.SignInResp, err error) {
	return s.AuthenticationService.SignIn(ctx, req)
}

func (s *StsServerImpl) SetPassword(ctx context.Context, req *sts.SetPasswordReq) (res *sts.SetPasswordResp, err error) {
	return s.AuthenticationService.SetPassword(ctx, req)
}

func (s *StsServerImpl) SendVerifyCode(ctx context.Context, req *sts.SendVerifyCodeReq) (res *sts.SendVerifyCodeResp, err error) {
	return s.AuthenticationService.SendVerifyCode(ctx, req)
}
