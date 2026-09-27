package ai_api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/service/ai_service"

	"github.com/gin-gonic/gin"
)

type UserAIRequest struct {
	Messages  []ai_service.Message `json:"messages"`   // 对话历史消息
	UserInput string               `json:"user_input"` // 用户输入
	ImageID   uint                 `json:"image_ID"`   // 图片ID(可选)
	Model     string               `json:"model"`      // 模型名称(可选)
}

// AIResponse 表示AI响应的结果
type AIResponse struct {
	Success bool   `json:"success"`
	Content string `json:"content"`
	Error   string `json:"error"`
}

func (AIApi) Chat(c *gin.Context) {
	var req UserAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("请求参数错误", c)
		return
	}

	// 历史消息只允许 user / assistant 角色,防止注入 system
	for _, v := range req.Messages {
		if v.Role != "user" && v.Role != "assistant" {
			response.FailWithMsg("对话历史消息中包含非用户或助手角色", c)
			return
		}
	}
	if strings.Contains(req.UserInput, global.SystemPromptMainSite.String()) {
		response.FailWithMsg("用户输入中包含system_prompt", c)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 组装消息:system(看板娘) + 历史 + 本轮用户输入,交由 ai 微服务完成。
	msgs := make([]ai_service.Message, 0, len(req.Messages)+2)
	msgs = append(msgs, ai_service.Message{Role: "system", Content: global.SystemPromptMainSite.String()})
	msgs = append(msgs, req.Messages...)
	msgs = append(msgs, ai_service.Message{Role: "user", Content: req.UserInput})

	content, err := ai_service.ChatMessages(ctx, msgs)
	if err != nil {
		response.FailWithMsg(fmt.Sprintf("AI服务调用失败: %v", err), c)
		return
	}

	response.OkWithData(AIResponse{Success: true, Content: content}, c)
}
