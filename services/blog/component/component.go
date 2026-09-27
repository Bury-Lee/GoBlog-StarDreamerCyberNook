// Package component 把博客主程序包裹为可被统一宿主装载的组件:
// 这样"博客 + 各微服务组件"可在同一进程中按蓝图一起拉起。
package component

import (
	"context"

	"StarDreamerCyberNook/pkg/blog"
	"StarDreamerCyberNook/pkg/host"
	"StarDreamerCyberNook/pkg/hostcfg"
	"StarDreamerCyberNook/utils"
)

type comp struct {
	setting string
	app     *blog.App
}

// New 依据配置构造 blog 组件(config.setting 缺省用当前宿主使用的配置文件)。
func New(cfg map[string]any) (host.Component, error) {
	return &comp{setting: utils.Str(cfg, "setting", hostcfg.OrStr(hostcfg.SettingPath(), "setting.yaml"))}, nil
}

func (c *comp) Name() string       { return "blog" }
func (c *comp) Provides() []string { return []string{"blog"} }
func (c *comp) Injects() []string  { return nil }

func (c *comp) Start(ctx context.Context) error {
	app, err := blog.Run(ctx, c.setting)
	if err != nil {
		return err
	}
	c.app = app
	return nil
}

func (c *comp) Stop(ctx context.Context) error {
	if c.app == nil {
		return nil
	}
	return c.app.Stop(ctx)
}
