package host

import (
	"context"

	GoTenon "GoTenon"
)

// gnPlugin 把宿主 Component 适配为 GoTenon 插件:
//   - Inject  → 组件的依赖(服务键=插件名)
//   - Apply   → 启动组件(gRPC 监听/连库)
//   - End     → 停止组件(优雅停机)
//
// 于是依赖闭包、拓扑收敛、失败回滚、卸载逆序全部由 GoTenon 负责。
type gnPlugin struct {
	c Component
}

func (p *gnPlugin) Name() string              { return p.c.Name() }
func (p *gnPlugin) Desc() map[string]string   { return nil }
func (p *gnPlugin) Inject() []string          { return p.c.Injects() }
func (p *gnPlugin) Config() map[string]string { return nil }
func (p *gnPlugin) Register() error           { return nil }

func (p *gnPlugin) Apply(_ *GoTenon.GoTenonContext, _ any) error {
	return p.c.Start(context.Background())
}

func (p *gnPlugin) Start() error { return nil }
func (p *gnPlugin) Run() error   { return nil }

func (p *gnPlugin) DealWithMessage(context.Context) error { return nil }

func (p *gnPlugin) End() error { return p.c.Stop(context.Background()) }
