//go:build wireinject
// +build wireinject

package provider

import (
	"github.com/google/wire"
	"github.com/xh-polaris/platform-sts/biz/adaptor"
)

func NewStsServerImpl() (*adaptor.StsServerImpl, error) {
	wire.Build(
		wire.Struct(new(adaptor.StsServerImpl), "*"),
		AllProvider,
	)
	return nil, nil
}
