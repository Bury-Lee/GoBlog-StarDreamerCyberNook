// Package community_service 是经 gRPC 调用独立 community 服务的客户端:
// 反馈 / 轮播图 / 友链 / 友情推广。函数签名贴合原 global.DB 调用点。
package community_service

import (
	"context"
	"sync"
	"time"

	communityv1 "StarDreamerCyberNook/gen/community/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/svc"
)

const communityTimeout = 10 * time.Second

var (
	cliOnce sync.Once
	cli     communityv1.CommunityServiceClient
	cliErr  error
)

func client() (communityv1.CommunityServiceClient, error) {
	cliOnce.Do(func() {
		var cfgs map[string][]string
		if global.Config != nil {
			cfgs = global.Config.Services
		}
		svc.Init(cfgs)
		conn, err := grpcx.Dial("community")
		if err != nil {
			cliErr = err
			return
		}
		cli = communityv1.NewCommunityServiceClient(conn)
	})
	return cli, cliErr
}

func call(fn func(context.Context, communityv1.CommunityServiceClient) error) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), communityTimeout)
	defer cancel()
	return fn(ctx, c)
}
