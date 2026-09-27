package message_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/service/message_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OneKeyReadRequest struct {
	CommentMessage        bool `json:"commentMessage"`
	DiggAndCollectMessage bool `json:"diggAndCollectMessage"`
	PrivateMessage        bool `json:"privateMessage"`
	SystemMessage         bool `json:"systemMessage"`
}

// OneKeyClearView 一键已读
func (MessageApi) OneKeyClearView(c *gin.Context) {
	var req OneKeyReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if !req.CommentMessage && !req.DiggAndCollectMessage && !req.PrivateMessage && !req.SystemMessage {
		response.OkWithMsg("成功", c)
		return
	}
	claim := jwts.GetClaims(c)
	if err := message_service.OneKeyRead(claim.UserID, req.CommentMessage, req.DiggAndCollectMessage, req.PrivateMessage, req.SystemMessage); err != nil {
		logrus.Error("一键已读失败", err)
		response.FailWithMsg("服务器错误", c)
		return
	}
	response.OkWithMsg("成功", c)
}
