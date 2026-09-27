// Package grpcx 提供 gRPC 客户端拨号的统一封装:
// 把"服务键"经发现解析为实例地址,并以 round_robin 在实例间负载均衡。
package grpcx

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"StarDreamerCyberNook/pkg/discovery"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

// Scheme 是本框架自定义的解析方案,形如 sd:///content(按服务键解析实例地址)。
const Scheme = "sd"

var (
	regOnce sync.Once
	builder = &sdBuilder{}
)

// RegisterResolver 注册全局发现解析器(进程内调用一次即可)。
func RegisterResolver(r discovery.Resolver) {
	builder.set(r)
	regOnce.Do(func() { resolver.Register(builder) })
}

type sdBuilder struct {
	mu sync.RWMutex
	r  discovery.Resolver
}

func (b *sdBuilder) set(r discovery.Resolver) {
	b.mu.Lock()
	b.r = r
	b.mu.Unlock()
}

func (b *sdBuilder) get() discovery.Resolver {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.r
}

func (b *sdBuilder) Scheme() string { return Scheme }

func (b *sdBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	service := strings.TrimPrefix(target.URL.Path, "/")
	rr := &sdResolver{cc: cc, service: service, src: b}
	rr.ResolveNow(resolver.ResolveNowOptions{})
	return rr, nil
}

type sdResolver struct {
	cc      resolver.ClientConn
	service string
	src     *sdBuilder
}

func (r *sdResolver) ResolveNow(resolver.ResolveNowOptions) {
	disc := r.src.get()
	if disc == nil {
		r.cc.ReportError(fmt.Errorf("grpcx: discovery resolver not registered"))
		return
	}
	addrs, err := disc.Resolve(context.Background(), r.service)
	if err != nil {
		r.cc.ReportError(err)
		return
	}
	st := resolver.State{}
	for _, a := range addrs {
		st.Addresses = append(st.Addresses, resolver.Address{Addr: a})
	}
	_ = r.cc.UpdateState(st)
}

func (r *sdResolver) Close() {}

// Dial 按服务键拨号:sd:///<service>,默认 round_robin。
func Dial(service string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	base := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig":[{"round_robin":{}}]}`),
		grpc.WithChainUnaryInterceptor(UnaryClientInterceptor()),
	}
	base = append(base, opts...)
	return grpc.NewClient(fmt.Sprintf("%s:///%s", Scheme, service), base...)
}
