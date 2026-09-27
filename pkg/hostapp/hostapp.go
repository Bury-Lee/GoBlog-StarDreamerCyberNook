// Package hostapp 组装统一宿主的组件注册表,并提供"按 setting.yaml 启动"的入口。
//
// 设计(单一配置源):组件的启用开关内置于 setting.yaml 的 components 段,
// 地址/库/密钥同样从 setting.yaml 派生(pkg/hostcfg)。宿主本身也登记为一个组件("host")。
package hostapp

import (
	"context"

	"StarDreamerCyberNook/conf"
	"StarDreamerCyberNook/core"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/pkg/host"
	"StarDreamerCyberNook/pkg/hostcfg"
	aiC "StarDreamerCyberNook/services/ai/component"
	authC "StarDreamerCyberNook/services/auth/component"
	blogC "StarDreamerCyberNook/services/blog/component"
	chatC "StarDreamerCyberNook/services/chat/component"
	communityC "StarDreamerCyberNook/services/community/component"
	contentC "StarDreamerCyberNook/services/content/component"
	logC "StarDreamerCyberNook/services/log/component"
	mediaC "StarDreamerCyberNook/services/media/component"
	messageC "StarDreamerCyberNook/services/message/component"
	notifyC "StarDreamerCyberNook/services/notify/component"
	searchC "StarDreamerCyberNook/services/search/component"
	userC "StarDreamerCyberNook/services/user/component"
)

// known 全部可选组件名(含宿主自身)。
var known = []string{
	"host", "blog", "ai", "search", "notify", "media", "auth",
	"content", "user", "message", "community", "log", "chat",
}

// hostComponent 是"宿主"组件的声明式占位:宿主编排由 main/cmd-host 承担。
type hostComponent struct{}

func newHostComponent(map[string]any) (host.Component, error) { return hostComponent{}, nil }

func (hostComponent) Name() string                { return "host" }
func (hostComponent) Provides() []string          { return []string{"host"} }
func (hostComponent) Injects() []string           { return nil }
func (hostComponent) Start(context.Context) error { return nil }
func (hostComponent) Stop(context.Context) error  { return nil }

// Registry 返回登记了全部组件类型(含 host 自身)的注册表。
func Registry() *host.Registry {
	reg := host.NewRegistry()
	reg.Register("host", []string{"host"}, nil, newHostComponent)
	reg.Register("blog", []string{"blog"}, nil, blogC.New)
	reg.Register("ai", []string{"ai"}, nil, aiC.New)
	reg.Register("search", []string{"search"}, nil, searchC.New)
	reg.Register("notify", []string{"notify"}, nil, notifyC.New)
	reg.Register("media", []string{"media"}, nil, mediaC.New)
	reg.Register("auth", []string{"auth"}, nil, authC.New)
	reg.Register("content", []string{"content"}, nil, contentC.New)
	reg.Register("user", []string{"user"}, nil, userC.New)
	reg.Register("message", []string{"message"}, nil, messageC.New)
	reg.Register("community", []string{"community"}, nil, communityC.New)
	reg.Register("log", []string{"log"}, nil, logC.New)
	reg.Register("chat", []string{"chat"}, nil, chatC.New)
	return reg
}

// 蓝图设置依据 setting.yaml 的 components组件配置段生成装配清单:
// 未配置 components 时默认全部启用;配置了则"列出的按其布尔值,未列出的视为关闭"。
func BlueprintFromConfig() host.Blueprint {
	bp := host.Blueprint{Components: map[string]host.CompConfig{}}
	var explicit map[string]conf.ComponentConfig
	if global.Config != nil {
		explicit = global.Config.Components
	}
	hasExplicit := len(explicit) > 0
	for _, name := range known {
		enabled := true
		var cfg map[string]any
		if hasExplicit {
			cc, ok := explicit[name]
			enabled = ok && cc.Enabled
			cfg = cc.Config
		}
		bp.Components[name] = host.CompConfig{Enabled: enabled, Config: cfg}
	}
	return bp
}

// Run 按 setting.yaml 启动宿主(阻塞至退出信号或 ctx 结束)。
func Run(ctx context.Context, path string) error {
	hostcfg.SetSettingPath(path)
	if global.Config == nil {
		global.Config = core.ReadConf(path)
	}
	// 宿主统一初始化共享数据库(幂等),供各组件复用;即使关闭 blog 组件也能保证 global.DB 就绪
	if global.DB == nil {
		global.DB = core.InitDB()
	}
	return Registry().Run(ctx, BlueprintFromConfig())
}
