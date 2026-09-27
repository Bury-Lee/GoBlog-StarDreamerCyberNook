// Package component 把 media 服务包装为可被统一宿主装载的组件。
package component

import (
	"context"
	"net"

	mediav1 "StarDreamerCyberNook/gen/media/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/host"
	"StarDreamerCyberNook/pkg/hostcfg"
	"StarDreamerCyberNook/services/media/internal"
	"StarDreamerCyberNook/utils"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type comp struct {
	addr string
	cfg  internal.Config
	srv  *grpc.Server
}

// New 依据配置构造 media 组件;默认从 setting.yaml 的 upload/objectStorage 段派生,库同博客。
func New(cfg map[string]any) (host.Component, error) {
	backend, localDir, h, ak, sk, bucket, region := hostcfg.Media()
	driver, dsn := hostcfg.DBDialect()
	driver = utils.Str(cfg, "driver", driver)
	dsn = utils.Str(cfg, "dsn", dsn)
	return &comp{
		addr: utils.Str(cfg, "addr", hostcfg.ServiceAddr("media", "127.0.0.1:9240")),
		cfg: internal.Config{
			Backend:   utils.Str(cfg, "backend", backend),
			LocalDir:  utils.Str(cfg, "localDir", localDir),
			Host:      utils.Str(cfg, "host", h),
			AccessKey: utils.Str(cfg, "accessKey", ak),
			SecretKey: utils.Str(cfg, "secretKey", sk),
			Bucket:    utils.Str(cfg, "bucket", bucket),
			Region:    utils.Str(cfg, "region", region),
			Driver:     hostcfg.OrStr(driver, "sqlite"),
			DSN:        dsn,
			Standalone: utils.Bool(cfg, "dbStandalone", false),
		},
	}, nil
}

func (c *comp) Name() string       { return "media" }
func (c *comp) Provides() []string { return []string{"media"} }
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
	mediav1.RegisterMediaServiceServer(srv, impl)
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
