// Package component 把 notify 服务包装为可被统一宿主装载的组件。
package component

import (
	"context"
	"net"

	notifyv1 "StarDreamerCyberNook/gen/notify/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/host"
	"StarDreamerCyberNook/pkg/hostcfg"
	"StarDreamerCyberNook/services/notify/internal"
	"StarDreamerCyberNook/utils"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type comp struct {
	addr string
	cfg  internal.Config
	srv  *grpc.Server
}

// New 依据配置构造 notify 组件;默认从 setting.yaml 的 email 段派生。
func New(cfg map[string]any) (host.Component, error) {
	domain, port, sendEmail, authCode, nick := hostcfg.Email()
	return &comp{
		addr: utils.Str(cfg, "addr", hostcfg.ServiceAddr("notify", "127.0.0.1:9230")),
		cfg: internal.Config{
			Domain:       utils.Str(cfg, "domain", domain),
			Port:         utils.Int(cfg, "port", hostcfg.OrInt(port, 465)),
			SendEmail:    utils.Str(cfg, "sendEmail", sendEmail),
			AuthCode:     utils.Str(cfg, "authCode", authCode),
			SendNickname: utils.Str(cfg, "sendNickname", nick),
			SiteTitle:    utils.Str(cfg, "siteTitle", ""),
		},
	}, nil
}

func (c *comp) Name() string       { return "notify" }
func (c *comp) Provides() []string { return []string{"notify"} }
func (c *comp) Injects() []string  { return nil }

func (c *comp) Start(_ context.Context) error {
	lis, err := net.Listen("tcp", c.addr)
	if err != nil {
		return err
	}
	srv := grpcx.NewServer()
	notifyv1.RegisterNotifyServiceServer(srv, internal.NewServer(c.cfg))
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
