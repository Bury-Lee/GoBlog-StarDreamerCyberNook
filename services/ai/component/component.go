// Package component 把 ai 服务包装为可被统一宿主装载的组件。
package component

import (
	"context"
	"net"

	aiv1 "StarDreamerCyberNook/gen/ai/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/host"
	"StarDreamerCyberNook/pkg/hostcfg"
	"StarDreamerCyberNook/services/ai/internal"
	"StarDreamerCyberNook/utils"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type comp struct {
	addr string
	cfg  internal.Config
	srv  *grpc.Server
}

// New 依据配置构造 ai 组件;默认从 setting.yaml 的 ai 段派生。
func New(cfg map[string]any) (host.Component, error) {
	h, key, apiType, model, temp, maxTok := hostcfg.AI()
	return &comp{
		addr: utils.Str(cfg, "addr", hostcfg.ServiceAddr("ai", "127.0.0.1:9210")),
		cfg: internal.Config{
			Host:        utils.Str(cfg, "host", h),
			APIKey:      utils.Str(cfg, "apiKey", key),
			APIType:     utils.Str(cfg, "apiType", apiType),
			Model:       utils.Str(cfg, "model", hostcfg.OrStr(model, "gpt-3.5-turbo")),
			Temperature: float32(utils.Float(cfg, "temperature", hostcfg.OrFloat(float64(temp), 0.7))),
			MaxTokens:   utils.Int(cfg, "maxTokens", hostcfg.OrInt(maxTok, 1024)),
		},
	}, nil
}

func (c *comp) Name() string       { return "ai" }
func (c *comp) Provides() []string { return []string{"ai"} }
func (c *comp) Injects() []string  { return nil }

func (c *comp) Start(_ context.Context) error {
	lis, err := net.Listen("tcp", c.addr)
	if err != nil {
		return err
	}
	srv := grpcx.NewServer()
	aiv1.RegisterAIServiceServer(srv, internal.NewServer(c.cfg))
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
