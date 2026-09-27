package grpcx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"

	"google.golang.org/grpc/metadata"
)

// 调用元数据在 gRPC metadata 中的键。
const (
	mdTraceID    = "x-trace-id"
	mdSpanID     = "x-span-id"
	mdParentSpan = "x-parent-span-id"
	mdDeadline   = "x-deadline-unix-ms"
	mdDepth      = "x-depth"
	mdIdemKey    = "x-idempotency-key"
	mdPrincipal  = "x-principal"
)

// CallMeta 是跨服务调用元数据的进程内表示(与 gen/common/v1.CallMeta 对应)。
type CallMeta struct {
	TraceID        string
	SpanID         string
	ParentSpanID   string
	DeadlineUnixMs int64
	Depth          int32
	IdempotencyKey string
	Principal      string
}

// NewID 生成 16 字节随机十六进制 ID(用于 trace/span)。
func NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b[:])
}

type metaCtxKey struct{}

// WithCallMeta 把调用元数据放入上下文(调用方/服务端共享)。
func WithCallMeta(ctx context.Context, m CallMeta) context.Context {
	return context.WithValue(ctx, metaCtxKey{}, m)
}

// CallMetaFromContext 读取调用元数据。
func CallMetaFromContext(ctx context.Context) (CallMeta, bool) {
	m, ok := ctx.Value(metaCtxKey{}).(CallMeta)
	return m, ok
}

func (m CallMeta) toMD() metadata.MD {
	md := metadata.MD{}
	set := func(k, v string) {
		if v != "" {
			md.Set(k, v)
		}
	}
	set(mdTraceID, m.TraceID)
	set(mdSpanID, m.SpanID)
	set(mdParentSpan, m.ParentSpanID)
	set(mdIdemKey, m.IdempotencyKey)
	set(mdPrincipal, m.Principal)
	if m.DeadlineUnixMs > 0 {
		md.Set(mdDeadline, strconv.FormatInt(m.DeadlineUnixMs, 10))
	}
	if m.Depth != 0 {
		md.Set(mdDepth, strconv.FormatInt(int64(m.Depth), 10))
	}
	return md
}

func callMetaFromMD(md metadata.MD) CallMeta {
	first := func(k string) string {
		if v := md.Get(k); len(v) > 0 {
			return v[0]
		}
		return ""
	}
	m := CallMeta{
		TraceID:        first(mdTraceID),
		SpanID:         first(mdSpanID),
		ParentSpanID:   first(mdParentSpan),
		IdempotencyKey: first(mdIdemKey),
		Principal:      first(mdPrincipal),
	}
	if s := first(mdDeadline); s != "" {
		m.DeadlineUnixMs, _ = strconv.ParseInt(s, 10, 64)
	}
	if s := first(mdDepth); s != "" {
		d, _ := strconv.ParseInt(s, 10, 32)
		m.Depth = int32(d)
	}
	return m
}
