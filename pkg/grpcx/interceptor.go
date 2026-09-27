package grpcx

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const defaultMaxDepth = 8

type config struct {
	maxDepth int32
}

// Option 定制拦截器行为。
type Option func(*config)

// WithMaxDepth 设置调用链最大深度(超过则拒绝,用于防环)。
func WithMaxDepth(d int32) Option {
	return func(c *config) {
		if d > 0 {
			c.maxDepth = d
		}
	}
}

func newConfig(opts []Option) *config {
	c := &config{maxDepth: defaultMaxDepth}
	for _, o := range opts {
		o(c)
	}
	return c
}

// UnaryClientInterceptor 注入/透传调用元数据,并累加调用深度。
//
// - trace:无 trace 则生成,span 每次调用新建,parent 为上游 span;
// - depth:自增,超过上限直接以 ResourceExhausted 拒绝(防环);
// - deadline:取 ctx 的 deadline 写入,供服务端派生超时。
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	cfg := newConfig(opts)
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {

		m, _ := CallMetaFromContext(ctx)
		parent := m.SpanID
		if m.TraceID == "" {
			m.TraceID = NewID()
		}
		m.SpanID = NewID()
		m.ParentSpanID = parent
		m.Depth++
		if m.Depth > cfg.maxDepth {
			return status.Errorf(codes.ResourceExhausted,
				"grpcx: call depth %d exceeds max %d (possible cycle)", m.Depth, cfg.maxDepth)
		}
		if dl, ok := ctx.Deadline(); ok {
			m.DeadlineUnixMs = dl.UnixMilli()
		}

		if existing, ok := metadata.FromOutgoingContext(ctx); ok {
			ctx = metadata.NewOutgoingContext(ctx, metadata.Join(existing, m.toMD()))
		} else {
			ctx = metadata.NewOutgoingContext(ctx, m.toMD())
		}
		// 写回上下文:下游子调用/日志可读到本次 span
		ctx = WithCallMeta(ctx, m)
		return invoker(ctx, method, req, reply, cc, callOpts...)
	}
}

// UnaryServerInterceptor 读取调用元数据,校验深度、派生 deadline,并做 panic 兜底。
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	cfg := newConfig(opts)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {

		md, _ := metadata.FromIncomingContext(ctx)
		m := callMetaFromMD(md)
		if m.Depth > cfg.maxDepth {
			return nil, status.Errorf(codes.ResourceExhausted,
				"grpcx: call depth %d exceeds max %d", m.Depth, cfg.maxDepth)
		}
		if m.DeadlineUnixMs > 0 {
			dl := time.UnixMilli(m.DeadlineUnixMs)
			if !dl.After(time.Now()) {
				return nil, status.Error(codes.DeadlineExceeded, "grpcx: deadline already exceeded")
			}
			var cancel context.CancelFunc
			ctx, cancel = context.WithDeadline(ctx, dl)
			defer cancel()
		}
		ctx = WithCallMeta(ctx, m)

		defer func() {
			if r := recover(); r != nil {
				log.Printf("grpcx: panic in %s: %v\n%s", info.FullMethod, r, debug.Stack())
				err = status.Errorf(codes.Internal, "grpcx: recovered panic: %v", r)
			}
		}()
		return handler(ctx, req)
	}
}

// NewServer 创建带默认拦截器链的 gRPC 服务端。
func NewServer(opts ...Option) *grpc.Server {
	return grpc.NewServer(grpc.ChainUnaryInterceptor(UnaryServerInterceptor(opts...)))
}
