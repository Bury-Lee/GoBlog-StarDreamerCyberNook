// Package component 把 user 服务包装为可被统一宿主装载的组件。
package component

import (
	"context"
	"net"

	userv1 "StarDreamerCyberNook/gen/user/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/host"
	"StarDreamerCyberNook/pkg/hostcfg"
	"StarDreamerCyberNook/services/user/internal"
	"StarDreamerCyberNook/utils"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type comp struct {
	addr string
	cfg  internal.Config
	srv  *grpc.Server
}

// New 依据配置构造 user 组件;未显式配置的项从 setting.yaml 派生。
func New(cfg map[string]any) (host.Component, error) {
	driver, dsn := hostcfg.DBDialect()
	driver = utils.Str(cfg, "driver", driver)
	dsn = utils.Str(cfg, "dsn", dsn)
	return &comp{
		addr: utils.Str(cfg, "addr", hostcfg.ServiceAddr("user", "127.0.0.1:9270")),
		cfg:  internal.Config{Driver: hostcfg.OrStr(driver, "sqlite"), DSN: dsn, Standalone: utils.Bool(cfg, "dbStandalone", false)},
	}, nil
}

func (c *comp) Name() string       { return "user" }
func (c *comp) Provides() []string { return []string{"user"} }
func (c *comp) Injects() []string  { return nil }

func (c *comp) Start(_ context.Context) error {
	impl, err := internal.NewServer(c.cfg)
	if err != nil {
		return err
	}
	lis, err := net.Listen("tcp", c.addr)
	if err != nil {
		return err
	}
	srv := grpcx.NewServer()
	userv1.RegisterUserServiceServer(srv, impl)
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
