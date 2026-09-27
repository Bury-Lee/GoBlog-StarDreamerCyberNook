package jwts

import (
	"context"
	"errors"
	"sync"
	"time"

	authv1 "StarDreamerCyberNook/gen/auth/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/svc"
)

const authTimeout = 5 * time.Second

var (
	cliOnce sync.Once
	cli     authv1.AuthServiceClient
	cliErr  error
)

// client 懒加载到 auth 服务的 gRPC 连接(服务键 "auth")。
func client() (authv1.AuthServiceClient, error) {
	cliOnce.Do(func() {
		if global.Config == nil {
			cliErr = errors.New("jwts: config not ready")
			return
		}
		svc.Init(global.Config.Services)
		conn, err := grpcx.Dial("auth")
		if err != nil {
			cliErr = err
			return
		}
		cli = authv1.NewAuthServiceClient(conn)
	})
	return cli, cliErr
}

func ctxTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), authTimeout)
}

func issueReq(userID uint, username string, role int) *authv1.IssueRequest {
	return &authv1.IssueRequest{UserId: uint64(userID), Username: username, Role: int32(role)}
}

func issueRefreshReq(userID uint) *authv1.RefreshRequest {
	return &authv1.RefreshRequest{UserId: uint64(userID)}
}
