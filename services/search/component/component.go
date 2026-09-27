// Package component 把 search 服务包装为可被统一宿主装载的组件。
package component

import (
	"context"
	"net"

	searchv1 "StarDreamerCyberNook/gen/search/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/host"
	"StarDreamerCyberNook/pkg/hostcfg"
	"StarDreamerCyberNook/services/search/internal"
	"StarDreamerCyberNook/utils"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type comp struct {
	addr string
	cfg  internal.Config
	srv  *grpc.Server
}

// New 依据配置构造 search 组件;ES 地址与库默认从 setting.yaml 派生。
func New(cfg map[string]any) (host.Component, error) {
	driver, dsn := hostcfg.DBDialect()
	driver = utils.Str(cfg, "driver", driver)
	dsn = utils.Str(cfg, "dsn", dsn)

	url := utils.Str(cfg, "url", "")
	user := utils.Str(cfg, "username", "")
	pass := utils.Str(cfg, "password", "")
	if url == "" {
		if en, u, us, pw := hostcfg.ES(); en {
			url, user, pass = u, us, pw
		}
	}
	return &comp{
		addr: utils.Str(cfg, "addr", hostcfg.ServiceAddr("search", "127.0.0.1:9220")),
		cfg: internal.Config{
			URL: url, UserName: user, Password: pass,
			Driver: hostcfg.OrStr(driver, "sqlite"), DSN: dsn,
			Standalone: utils.Bool(cfg, "dbStandalone", false),
		},
	}, nil
}

func (c *comp) Name() string       { return "search" }
func (c *comp) Provides() []string { return []string{"search"} }
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
	searchv1.RegisterSearchServiceServer(srv, impl)
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
