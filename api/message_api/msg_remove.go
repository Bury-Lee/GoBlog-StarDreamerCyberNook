package message_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/service/message_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

type MessageRemoveRequest struct {
	MessageID []uint `json:"messageID"`
}

// MessageRemoveView 删除消息
func (MessageApi) MessageRemoveView(c *gin.Context) {
	var req MessageRemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if len(req.MessageID) == 0 || len(req.MessageID) > 100 {
		response.FailWithMsg("删除的消息数量必须在1-100之间", c)
		return
	}
	claim := jwts.GetClaims(c)
	if err := message_service.Remove(claim.UserID, req.MessageID); err != nil {
		response.FailWithMsg("删除消息失败", c)
		return
	}
	response.OkWithMsg("删除消息成功", c)
}
