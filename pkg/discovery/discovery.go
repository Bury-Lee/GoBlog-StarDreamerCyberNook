// Package discovery 定义"服务键 → 实例地址列表"的发现抽象。
//
// 注意:实例发现属于微服务层,不属于插件内核;这里只提供最小实现,
// 后续可替换为 etcd / Nacos / Consul 适配器,对调用方无感。

// TODO:后续会提供更好的抽象和文档,以此适配更加复杂的场景。
package discovery

import (
	"context"
	"fmt"
	"sync"
)

// Resolver 把服务键解析为若干实例地址(host:port)。
type Resolver interface {
	Resolve(ctx context.Context, service string) ([]string, error)
}

// Static 是内存静态实现,用于本地开发与骨架验证。后续可以提供 etcd / Nacos / Consul 适配器
type Static struct {
	mu   sync.RWMutex
	data map[string][]string
}

// NewStatic 用初始映射构造。
func NewStatic(init map[string][]string) *Static {
	m := make(map[string][]string, len(init))
	for k, v := range init {
		m[k] = append([]string(nil), v...)
	}
	return &Static{data: m}
}

// Resolve 返回服务键对应的实例地址。
func (s *Static) Resolve(_ context.Context, service string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	addrs := s.data[service]
	if len(addrs) == 0 {
		return nil, fmt.Errorf("discovery: no instance for %q", service)
	}
	return append([]string(nil), addrs...), nil
}

// Set 覆盖某服务键的实例列表(模拟实例注册/摘除)。
func (s *Static) Set(service string, addrs ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[service] = append([]string(nil), addrs...)
}
