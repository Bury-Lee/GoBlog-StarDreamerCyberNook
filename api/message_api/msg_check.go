package message_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/service/message_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

// SiteMessageCheckView 各类型未读消息数量
func (MessageApi) SiteMessageCheckView(c *gin.Context) {
	claim := jwts.GetClaims(c)
	counts, err := message_service.UnreadCounts(claim.UserID)
	if err != nil {
		response.FailWithMsg("未读消息查询失败", c)
		return
	}
	response.OkWithData(counts, c)
}
