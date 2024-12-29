package util

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/zeromicro/go-zero/core/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func HTTPGet(ctx context.Context, rawURL string) ([]byte, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	// 获取 query 之前的部分
	baseURL := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
	ctx, span := trace.TracerFromContext(ctx).Start(ctx, fmt.Sprintf("http/%s", baseURL), oteltrace.WithTimestamp(time.Now()), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer func() {
		span.End(oteltrace.WithTimestamp(time.Now()))
	}()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http get error : rawURL=%v , statusCode=%v", rawURL, response.StatusCode)
	}
	return io.ReadAll(response.Body)
}

func HTTPPost(ctx context.Context, rawURL string, body io.Reader) ([]byte, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	baseURL := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
	ctx, span := trace.TracerFromContext(ctx).Start(ctx, fmt.Sprintf("http/%s", baseURL), oteltrace.WithTimestamp(time.Now()), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer func() {
		span.End(oteltrace.WithTimestamp(time.Now()))
	}()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, body)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http post error : rawURL=%v , statusCode=%v", rawURL, response.StatusCode)
	}
	return io.ReadAll(response.Body)
}
