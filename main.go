package main

import (
	"net"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	"github.com/xh-polaris/gopkg/kitex/middleware"
	logx "github.com/xh-polaris/gopkg/util/log"
	"github.com/xh-polaris/service-idl-gen-go/kitex_gen/platform/sts/stsservice"

	"github.com/xh-polaris/platform-sts/biz/infrastructure/util/log"
	"github.com/xh-polaris/platform-sts/provider"
)

func main() {
	klog.SetLogger(logx.NewKlogLogger())
	s, err := provider.NewStsServerImpl()
	if err != nil {
		panic(err)
	}
	addr, err := net.ResolveTCPAddr("tcp", s.ListenOn)
	if err != nil {
		panic(err)
	}
	svr := stsservice.NewServer(
		s,
		server.WithMiddleware(middleware.LogMiddleware(s.Name)),
		server.WithServiceAddr(addr),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: s.Name}),
	)

	err = svr.Run()

	if err != nil {
		log.Error(err.Error())
	}
}
