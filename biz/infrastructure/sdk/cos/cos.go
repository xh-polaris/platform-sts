package cos

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/google/wire"
	"github.com/tencentyun/cos-go-sdk-v5"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
	"github.com/zeromicro/go-zero/core/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/xh-polaris/platform-sts/biz/infrastructure/config"
)

type CosSDK struct {
	stsClient *sts.Client
	cosClient *cos.Client
}

func NewCosSDK(config *config.Config) (*CosSDK, error) {
	bucketURL, err := url.Parse(config.CosConfig.CosHost())
	if err != nil {
		return nil, err
	}
	ciURL, err := url.Parse(config.CosConfig.CIHost())
	if err != nil {
		return nil, err
	}
	return &CosSDK{
		stsClient: sts.NewClient(
			config.CosConfig.SecretId,
			config.CosConfig.SecretKey,
			nil),
		cosClient: cos.NewClient(&cos.BaseURL{
			BucketURL: bucketURL,
			CIURL:     ciURL,
		}, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  config.CosConfig.SecretId,
				SecretKey: config.CosConfig.SecretKey,
			},
		}),
	}, nil
}

func (s *CosSDK) GetCredential(ctx context.Context, opt *sts.CredentialOptions) (*sts.CredentialResult, error) {
	ctx, span := trace.TracerFromContext(ctx).Start(ctx, "sts/GetCredential", oteltrace.WithTimestamp(time.Now()), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer func() {
		span.End(oteltrace.WithTimestamp(time.Now()))
	}()
	return s.stsClient.GetCredential(opt)
}

func (s *CosSDK) GetPresignedURL(ctx context.Context, httpMethod, name, ak, sk string, expired time.Duration, opt interface{}, signHost ...bool) (*url.URL, error) {
	ctx, span := trace.TracerFromContext(ctx).Start(ctx, "cos/Object/GetPresignedURL", oteltrace.WithTimestamp(time.Now()), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer func() {
		span.End(oteltrace.WithTimestamp(time.Now()))
	}()
	return s.cosClient.Object.GetPresignedURL(ctx, httpMethod, name, ak, sk, expired, opt, signHost...)
}

func (s *CosSDK) Delete(ctx context.Context, name string, opt ...*cos.ObjectDeleteOptions) (*cos.Response, error) {
	ctx, span := trace.TracerFromContext(ctx).Start(ctx, "cos/Object/Delete", oteltrace.WithTimestamp(time.Now()), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer func() {
		span.End(oteltrace.WithTimestamp(time.Now()))
	}()
	return s.cosClient.Object.Delete(ctx, name, opt...)
}

func (s *CosSDK) BatchImageAuditing(ctx context.Context, opt *cos.BatchImageAuditingOptions) (*cos.BatchImageAuditingJobResult, *cos.Response, error) {
	ctx, span := trace.TracerFromContext(ctx).Start(ctx, "cos/CI/BatchImageAuditing", oteltrace.WithTimestamp(time.Now()), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer func() {
		span.End(oteltrace.WithTimestamp(time.Now()))
	}()
	return s.cosClient.CI.BatchImageAuditing(ctx, opt)
}

var CosSet = wire.NewSet(
	NewCosSDK,
)
