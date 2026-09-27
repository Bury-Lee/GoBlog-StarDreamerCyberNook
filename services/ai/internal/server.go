package internal

import (
	"context"
	"errors"
	"io"
	"time"

	aiv1 "StarDreamerCyberNook/gen/ai/v1"

	"github.com/sashabaranov/go-openai"
)

// Config 是 AI 服务的模型参数(来自配置文件,环境变量可选覆盖)。
type Config struct {
	Host        string
	APIKey      string
	APIType     string
	Model       string
	Temperature float32
	MaxTokens   int
}

// Server 实现 AIService,持有模型客户端。
type Server struct {
	aiv1.UnimplementedAIServiceServer
	cfg Config
	cli *openai.Client
}

// NewServer 按配置构建服务。
func NewServer(cfg Config) *Server {
	oc := openai.DefaultConfig(cfg.APIKey)
	if cfg.Host != "" {
		oc.BaseURL = cfg.Host
	}
	if cfg.APIType != "" {
		oc.APIType = openai.APIType(cfg.APIType)
	}
	return &Server{cfg: cfg, cli: openai.NewClientWithConfig(oc)}
}

// buildMessages 优先使用完整消息列表;为空时用 prompt+content 构建单轮。
func (s *Server) buildMessages(req *aiv1.ChatRequest) []openai.ChatCompletionMessage {
	if len(req.GetMessages()) > 0 {
		msgs := make([]openai.ChatCompletionMessage, 0, len(req.GetMessages()))
		for _, m := range req.GetMessages() {
			msgs = append(msgs, openai.ChatCompletionMessage{Role: m.GetRole(), Content: m.GetContent()})
		}
		return msgs
	}
	return []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: req.GetPrompt()},
		{Role: openai.ChatMessageRoleUser, Content: req.GetContent()},
	}
}

// Ping 最小化探活。
func (s *Server) Ping(ctx context.Context, _ *aiv1.PingRequest) (*aiv1.PingReply, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := s.cli.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     s.cfg.Model,
		Messages:  []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "ping"}},
		MaxTokens: 1,
	})
	if err != nil {
		return &aiv1.PingReply{Ok: false, Message: err.Error()}, nil
	}
	return &aiv1.PingReply{Ok: true}, nil
}

// Chat 非流式对话。
func (s *Server) Chat(ctx context.Context, req *aiv1.ChatRequest) (*aiv1.ChatReply, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	res, err := s.cli.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       s.cfg.Model,
		Temperature: s.cfg.Temperature,
		MaxTokens:   s.cfg.MaxTokens,
		Messages:    s.buildMessages(req),
	})
	if err != nil {
		return nil, err
	}
	if len(res.Choices) == 0 {
		return nil, errors.New("模型空回复")
	}
	return &aiv1.ChatReply{Reply: res.Choices[0].Message.Content}, nil
}

// ChatStream 流式对话,逐段回传增量。
func (s *Server) ChatStream(req *aiv1.ChatRequest, stream aiv1.AIService_ChatStreamServer) error {
	ctx := stream.Context()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	st, err := s.cli.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       s.cfg.Model,
		Temperature: s.cfg.Temperature,
		MaxTokens:   s.cfg.MaxTokens,
		Messages:    s.buildMessages(req),
		Stream:      true,
	})
	if err != nil {
		return err
	}
	defer st.Close()

	for {
		res, err := st.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if len(res.Choices) == 0 {
			continue
		}
		delta := res.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		if err := stream.Send(&aiv1.ChatChunk{Delta: delta}); err != nil {
			return err
		}
	}
}
