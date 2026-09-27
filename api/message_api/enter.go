package message_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/message_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

type MessageApi struct{}

// SiteMessageConfView 查询站点消息列表配置
func (MessageApi) SiteMessageConfView(c *gin.Context) {
	claim := jwts.GetClaims(c)
	if claim == nil {
		response.FailWithMsg("未登录", c)
		return
	}
	conf, err := message_service.GetConf(claim.UserID)
	if err != nil {
		response.FailWithMsg("查询用户消息配置失败", c)
		return
	}
	response.OkWithData(models.UserMessageConfModel{
		UserID:             conf.UserID,
		OpenCommentMessage: conf.OpenCommentMessage,
		OpenReplyMessage:   conf.OpenReplyMessage,
		OpenDiggMessage:    conf.OpenDiggMessage,
		OpenCollectMessage: conf.OpenCollectMessage,
		OpenPrivateMessage: conf.OpenPrivateMessage,
	}, c)
}

type SiteMessageConfUpdateRequest struct {
	ID                 uint  `json:"id"`
	OpenCommentMessage *bool `json:"openCommentMessage"`
	OpenReplyMessage   *bool `json:"openReplyMessage"`
	OpenDiggMessage    *bool `json:"openDiggMessage"`
	OpenCollectMessage *bool `json:"openCollectMessage"`
	OpenPrivateMessage *bool `json:"openPrivateMessage"`
}

// SiteMessageConfUpdateView 更新用户消息配置
func (MessageApi) SiteMessageConfUpdateView(c *gin.Context) {
	claim := jwts.GetClaims(c)
	if claim == nil {
		response.FailWithMsg("未登录", c)
		return
	}
	var req SiteMessageConfUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数绑定失败", c)
		return
	}
	if req.OpenCommentMessage == nil && req.OpenReplyMessage == nil && req.OpenDiggMessage == nil &&
		req.OpenCollectMessage == nil && req.OpenPrivateMessage == nil {
		response.OkWithMsg("没有需要更新的配置", c)
		return
	}
	if err := message_service.UpdateConf(claim.UserID, req.OpenCommentMessage, req.OpenReplyMessage,
		req.OpenDiggMessage, req.OpenCollectMessage, req.OpenPrivateMessage); err != nil {
		response.FailWithMsg("更新用户消息配置失败", c)
		return
	}
	response.OkWithMsg("更新用户消息配置成功", c)
}
