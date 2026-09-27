// Package host 是"统一的宿主",一个进程按蓝图配置装载/启停多个微服务组件。
//
// 内核使用 GoTenon(进程内插件内核):依赖闭包、拓扑收敛、失败回滚、卸载逆序
// 全部交给 GoTenon;宿主只负责"把组件包装成插件 + 按蓝图启用"。
//
// 蓝图示例:
//
//	components:
//	  content: { enabled: true,  config: { addr: "127.0.0.1:9260", driver: "sqlite", dsn: "temp/test.db" } }
//	  search:  { enabled: false }
package host

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"

	GoTenon "GoTenon"
)

// Component 是可被宿主装载的微服务组件。
type Component interface {
	Name() string
	Provides() []string
	Injects() []string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Factory 工厂函数,依据配置构造组件。
type Factory func(config map[string]any) (Component, error)

// 组件结构体
type spec struct {
	provides []string //提供功能
	injects  []string //依赖功能
	factory  Factory  //工厂构造函数
}

// Registry 登记组件类型。
type Registry struct {
	mu    sync.Mutex
	specs map[string]spec
}

// NewRegistry 创建注册表。
func NewRegistry() *Registry { return &Registry{specs: map[string]spec{}} }

// Register 登记一个组件类型及其 provide/inject 声明。
func (r *Registry) Register(name string, provides, injects []string, f Factory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.specs[name] = spec{provides: provides, injects: injects, factory: f}
}

// CompConfig 蓝图中的单个组件配置。
type CompConfig struct {
	Enabled bool           `yaml:"enabled"`
	Config  map[string]any `yaml:"config"`
}

// Blueprint 装配蓝图。
type Blueprint struct {
	Components map[string]CompConfig `yaml:"components"`
}

// Run 按蓝图启停组件,阻塞至 ctx 结束或收到退出信号。
// 启用集与依赖闭包、拓扑顺序、失败回滚均由 GoTenon 内核负责。
func (r *Registry) Run(ctx context.Context, bp Blueprint) error {
	r.mu.Lock()
	specs := make(map[string]spec, len(r.specs))
	for k, v := range r.specs {
		specs[k] = v
	}
	r.mu.Unlock()

	m := GoTenon.NewManager(GoTenon.New("host"))

	// 构造并注册全部组件(仅构造);GoTenon 按插件名解析 Inject组件对象
	names := make([]string, 0, len(specs))
	for n := range specs {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		c, err := specs[n].factory(bp.Components[n].Config)
		if err != nil {
			return fmt.Errorf("host: build %q: %w", n, err)
		}
		if _, err := m.Register(&gnPlugin{c: c}, nil); err != nil {
			return fmt.Errorf("host: register %q: %w", n, err)
		}
	}

	// 启用集:GoTenon 自动拉起依赖闭包、波次收敛、失败回滚
	enabled := make([]string, 0, len(bp.Components))
	for _, n := range names {
		if bp.Components[n].Enabled {
			enabled = append(enabled, n)
		}
	}
	if len(enabled) == 0 {
		return fmt.Errorf("host: no component enabled")
	}
	for _, n := range enabled {
		if err := m.Enable(n); err != nil {
			stopEnabled(m, enabled)
			return fmt.Errorf("host: enable %q: %w", n, err)
		}
	}

	// 等待退出,逆序 Disable(触发插件 End → 组件 Stop)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	select {
	case <-ctx.Done():
	case <-sig:
	}
	stopEnabled(m, enabled)
	return nil
}

// 停止运行的函数
func stopEnabled(m *GoTenon.Manager, enabled []string) {
	for i := len(enabled) - 1; i >= 0; i-- {
		_ = m.Disable(enabled[i])
	}
}
