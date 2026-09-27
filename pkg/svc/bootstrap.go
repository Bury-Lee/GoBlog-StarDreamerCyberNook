// Package svc 提供进程级服务发现引导:把"服务键 → 实例地址"表交给 grpcx 的
// 自定义解析器(sd:///<service>)。
//
// 配置以**配置文件**为主(conf.Config.Services),环境变量 SVC_<KEY> 可选覆盖。
// 合并顺序:内置默认 < 配置文件 < 环境变量。
package svc

import (
	"os"
	"strings"
	"sync"

	"StarDreamerCyberNook/pkg/discovery"
	"StarDreamerCyberNook/pkg/grpcx"
)

// 内置开发默认(配置文件未提供时使用)。
var defaults = map[string][]string{
	"ai":        {"127.0.0.1:9210"},
	"search":    {"127.0.0.1:9220"},
	"notify":    {"127.0.0.1:9230"},
	"media":     {"127.0.0.1:9240"},
	"auth":      {"127.0.0.1:9250"},
	"content":   {"127.0.0.1:9260"},
	"user":      {"127.0.0.1:9270"},
	"message":   {"127.0.0.1:9280"},
	"community": {"127.0.0.1:9290"},
	"log":       {"127.0.0.1:9295"},
	"chat":      {"127.0.0.1:9296"},
}

var (
	once sync.Once
	res  *discovery.Static
)

// Init 用配置文件的"服务键 → 实例地址"表初始化发现(进程内一次)。
// services 可为 nil。
func Init(services map[string][]string) {
	once.Do(func() {
		m := map[string][]string{}
		for k, v := range defaults {
			m[k] = append([]string(nil), v...)
		}
		for k, v := range services {
			if len(v) > 0 {
				m[strings.ToLower(k)] = append([]string(nil), v...)
			}
		}
		// 环境变量可选覆盖(部署/临时调试用)
		for _, e := range os.Environ() {
			k, v, ok := strings.Cut(e, "=")
			if !ok || !strings.HasPrefix(k, "SVC_") {
				continue
			}
			key := strings.ToLower(strings.TrimPrefix(k, "SVC_"))
			if addrs := splitAddrs(v); len(addrs) > 0 {
				m[key] = addrs
			}
		}
		res = discovery.NewStatic(m)
		grpcx.RegisterResolver(res)
	})
}

// Set 运行期覆盖某服务键地址(测试 / 动态用)。
func Set(service string, addrs ...string) {
	Init(nil)
	res.Set(service, addrs...)
}

func splitAddrs(s string) []string {
	var out []string
	for _, a := range strings.Split(s, ",") {
		if a = strings.TrimSpace(a); a != "" {
			out = append(out, a)
		}
	}
	return out
}
