// Package email_service 已从"直接 SMTP 发送"改为"经 gRPC 调用独立 notify 服务"的客户端。
// 对外函数签名保持不变,调用点无需改动;SMTP 被剥离到 services/notify。
package email_service

import (
	"context"
	"sync"
	"time"

	notifyv1 "StarDreamerCyberNook/gen/notify/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/svc"
)

const notifyTimeout = 20 * time.Second

var (
	cliOnce sync.Once
	cli     notifyv1.NotifyServiceClient
	cliErr  error
)

// client 懒加载到 notify 服务的 gRPC 连接(服务键 "notify")。
func client() (notifyv1.NotifyServiceClient, error) {
	cliOnce.Do(func() {
		var cfgs map[string][]string
		if global.Config != nil {
			cfgs = global.Config.Services
		}
		svc.Init(cfgs)
		conn, err := grpcx.Dial("notify")
		if err != nil {
			cliErr = err
			return
		}
		cli = notifyv1.NewNotifyServiceClient(conn)
	})
	return cli, cliErr
}

func ctxWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), notifyTimeout)
}

// SendRegister 发送注册验证码邮件。
func SendRegister(target string, code string) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	_, err = c.SendRegister(ctx, &notifyv1.SendCodeRequest{Target: target, Code: code})
	return err
}

// SendForgetPwd 发送重置密码验证码。
func SendForgetPwd(target string, code string) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	_, err = c.SendForgetPwd(ctx, &notifyv1.SendCodeRequest{Target: target, Code: code})
	return err
}

// SendResetEmail 发送重置邮箱验证码。
func SendResetEmail(target string, code string) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	_, err = c.SendResetEmail(ctx, &notifyv1.SendCodeRequest{Target: target, Code: code})
	return err
}

// SendEmail 发送邮件的通用函数。
func SendEmail(to, subject, text string) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	_, err = c.SendEmail(ctx, &notifyv1.SendEmailRequest{To: to, Subject: subject, Html: text})
	return err
}
