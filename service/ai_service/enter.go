package ai_service

import (
	"context"
	"errors"
	"sync"
	"time"

	aiv1 "StarDreamerCyberNook/gen/ai/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/svc"
)

// 本包是"经 gRPC 调用独立 ai 服务"的客户端。
// 模型调用被剥离到 services/ai;对外函数签名保持稳定。

const (
	aiPingTimeout = 10 * time.Second
	aiChatTimeout = 60 * time.Second
)

// Message 是一条对话消息(role: system / user / assistant)。
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var (
	cliOnce sync.Once
	cli     aiv1.AIServiceClient
	cliErr  error
)

// client 懒加载到 ai 服务的 gRPC 连接(服务键 "ai")。
func client() (aiv1.AIServiceClient, error) {
	cliOnce.Do(func() {
		var cfgs map[string][]string
		if global.Config != nil {
			cfgs = global.Config.Services
		}
		svc.Init(cfgs)
		conn, err := grpcx.Dial("ai")
		if err != nil {
			cliErr = err
			return
		}
		cli = aiv1.NewAIServiceClient(conn)
	})
	return cli, cliErr
}

func toPB(msgs []Message) []*aiv1.Message {
	out := make([]*aiv1.Message, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, &aiv1.Message{Role: m.Role, Content: m.Content})
	}
	return out
}

// Available 探测 AI 服务是否可用。
func Available() (bool, error) {
	if !global.Config.AI.Enable {
		return false, errors.New("AI功能未启用")
	}
	c, err := client()
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), aiPingTimeout)
	defer cancel()
	rep, err := c.Ping(ctx, &aiv1.PingRequest{})
	if err != nil {
		return false, err
	}
	return rep.GetOk(), nil
}

// CreateSingleReply 输入内容与提示词,返回模型回复(单轮)。
func CreateSingleReply(content string, prompt string) (string, error) {
	c, err := client()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), aiChatTimeout)
	defer cancel()
	rep, err := c.Chat(ctx, &aiv1.ChatRequest{Content: content, Prompt: prompt})
	if err != nil {
		return "", err
	}
	if rep.GetReply() == "" {
		return "", errors.New("模型空回复")
	}
	return rep.GetReply(), nil
}

// ChatMessages 多轮对话(消息列表含 system),返回完整回复。
func ChatMessages(ctx context.Context, msgs []Message) (string, error) {
	c, err := client()
	if err != nil {
		return "", err
	}
	rep, err := c.Chat(ctx, &aiv1.ChatRequest{Messages: toPB(msgs)})
	if err != nil {
		return "", err
	}
	if rep.GetReply() == "" {
		return "", errors.New("模型空回复")
	}
	return rep.GetReply(), nil
}

// ChatStream 多轮流式对话,返回增量流。调用方负责 Recv 直到 io.EOF。
func ChatStream(ctx context.Context, msgs []Message) (aiv1.AIService_ChatStreamClient, error) {
	c, err := client()
	if err != nil {
		return nil, err
	}
	return c.ChatStream(ctx, &aiv1.ChatRequest{Messages: toPB(msgs)})
}
