package ai_service

import (
	"StarDreamerCyberNook/global"
	"context"
	"errors"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

// aiPingTimeout AI服务可用性探测的超时时间,比正式审核的60秒短,避免不可用时长时间阻塞定时任务
const aiPingTimeout = 10 * time.Second

// Available 探测AI服务是否可用
// 说明:用一次最小化的对话请求做探活,任何错误(未启用/未初始化/请求失败)都视为不可用,
// 返回错误信息便于调用方记录日志
func Available() (bool, error) {
	if !global.Config.AI.Enable {
		return false, errors.New("AI功能未启用")
	}
	if global.AIClient == nil {
		return false, errors.New("AI客户端未初始化")
	}

	ctx, cancel := context.WithTimeout(context.Background(), aiPingTimeout)
	defer cancel()

	_, err := global.AIClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: global.Config.AI.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "ping",
				},
			},
			MaxTokens: 1,
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

// 输入内容和提示词,返回模型回复和错误信息
func CreateSingleReply(content string, prompt string) (string, error) {
	//TODO:改为配置中设置
	//设置超时,避免AI服务不可用时请求长期挂起
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 构建完整的消息列表
	var messages []openai.ChatCompletionMessage

	// 添加系统提示词作为第一条消息

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: prompt,
	})

	// 添加对话历史消息
	messages = append(messages,
		openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: content,
		})

	// 创建非流式请求
	res, err := global.AIClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       global.Config.AI.Model,
			Temperature: global.Config.AI.Temperature,
			MaxTokens:   global.Config.AI.MaxTokens,
			Messages:    messages,
		},
	)
	if err != nil {
		if global.Config.System.RunMode == "debug" {
			logrus.Errorf("创建单条回复失败:%v", err)
		}
		return "", err
	}
	if global.Config.System.RunMode == "debug" {
		logrus.Debugf("输入内容:%s,提示词:%s", content, prompt)
	}
	if len(res.Choices) == 0 {
		// logrus.Error("模型空回复")
		return "", errors.New("模型空回复")
	}
	if global.Config.System.RunMode == "debug" {
		logrus.Debugf("模型回复:%s", res.Choices[0].Message.Content)
	}
	return res.Choices[0].Message.Content, nil
}
