package grpcx

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestCallMetaMDRoundTrip(t *testing.T) {
	in := CallMeta{
		TraceID:        "t1",
		SpanID:         "s1",
		ParentSpanID:   "p0",
		DeadlineUnixMs: 123,
		Depth:          3,
		IdempotencyKey: "k",
		Principal:      "u",
	}
	if out := callMetaFromMD(in.toMD()); out != in {
		t.Fatalf("round trip mismatch:\n got %+v\nwant %+v", out, in)
	}
}

func TestClientInterceptorInjectsTraceAndDepth(t *testing.T) {
	ic := UnaryClientInterceptor(WithMaxDepth(2))

	var seen metadata.MD
	err := ic(context.Background(), "/echo.v1.EchoService/Echo", nil, nil, nil,
		func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			seen, _ = metadata.FromOutgoingContext(ctx)
			return nil
		})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if seen.Get(mdTraceID) == nil {
		t.Fatal("trace id not injected")
	}
	if got := seen.Get(mdDepth); len(got) == 0 || got[0] != "1" {
		t.Fatalf("depth want 1, got %v", got)
	}
}

func TestClientInterceptorRejectsDepthOverflow(t *testing.T) {
	ic := UnaryClientInterceptor(WithMaxDepth(2))
	ctx := WithCallMeta(context.Background(), CallMeta{TraceID: "t", Depth: 2}) // 自增到 3 > 2
	err := ic(ctx, "/x/Y", nil, nil, nil,
		func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
			t.Fatal("invoker must not be called when depth exceeded")
			return nil
		})
	if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("want ResourceExhausted, got %v", err)
	}
}

func TestServerInterceptorRejectsDeep(t *testing.T) {
	si := UnaryServerInterceptor(WithMaxDepth(2))
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(mdDepth, "5"))
	_, err := si(ctx, nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
		t.Fatal("handler must not be called")
		return nil, nil
	})
	if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("want ResourceExhausted, got %v", err)
	}
}

func TestServerInterceptorRecoversPanic(t *testing.T) {
	si := UnaryServerInterceptor()
	_, err := si(context.Background(), nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
		panic("boom")
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("want Internal on panic, got %v", err)
	}
}
