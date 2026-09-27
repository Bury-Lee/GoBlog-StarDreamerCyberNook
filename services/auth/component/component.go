// Package component 把 auth 服务包装为可被统一宿主装载的组件。
package component

import (
	"context"
	"net"

	authv1 "StarDreamerCyberNook/gen/auth/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/host"
	"StarDreamerCyberNook/pkg/hostcfg"
	"StarDreamerCyberNook/services/auth/internal"
	"StarDreamerCyberNook/utils"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type comp struct {
	addr string
	cfg  internal.Config
	srv  *grpc.Server
}

// New 依据配置构造 auth 组件;密钥/过期默认从 setting.yaml 的 jwt 段派生。
func New(cfg map[string]any) (host.Component, error) {
	acc, ref, accMin, refH, iss := hostcfg.JWT()
	return &comp{
		addr: utils.Str(cfg, "addr", hostcfg.ServiceAddr("auth", "127.0.0.1:9250")),
		cfg: internal.Config{
			AccessSecret:        utils.Str(cfg, "accessSecret", acc),
			RefreshSecret:       utils.Str(cfg, "refreshSecret", ref),
			AccessExpireMinutes: utils.Int(cfg, "accessExpireMinutes", hostcfg.OrInt(accMin, 60)),
			RefreshExpireHours:  utils.Int(cfg, "refreshExpireHours", hostcfg.OrInt(refH, 168)),
			Issuer:              utils.Str(cfg, "issuer", hostcfg.OrStr(iss, "stardreamer")),
		},
	}, nil
}

func (c *comp) Name() string       { return "auth" }
func (c *comp) Provides() []string { return []string{"auth"} }
func (c *comp) Injects() []string  { return nil }

func (c *comp) Start(_ context.Context) error {
	lis, err := net.Listen("tcp", c.addr)
	if err != nil {
		return err
	}
	srv := grpcx.NewServer()
	authv1.RegisterAuthServiceServer(srv, internal.NewServer(c.cfg))
	c.srv = srv
	go func() {
		if err := srv.Serve(lis); err != nil {
			logrus.Errorf("%s gRPC 服务退出: %v", c.Name(), err)
		}
	}()
	return nil
}

func (c *comp) Stop(_ context.Context) error {
	if c.srv != nil {
		c.srv.GracefulStop()
	}
	return nil
}
