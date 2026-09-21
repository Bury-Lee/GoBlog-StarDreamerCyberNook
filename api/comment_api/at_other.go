package comment_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/message_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (CommentApi) AtOther(c *gin.Context) { //@别人并且发送时前端自动调用的接口
	var req models.IDRequest
	//路由参数是 /comment/at/:id,必须用ShouldBindUri,用Query绑定拿不到ID
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claim := jwts.GetClaims(c)
	var actor models.UserModel
	err := global.DB.Take(&actor, "id = ?", claim.UserID).Error
	if err != nil {
		response.FailWithMsg("用户不存在", c)
		return
	}
	//检查被@的用户是否存在,避免产生指向不存在用户的消息
	var target models.UserModel
	if err = global.DB.Take(&target, req.ID).Error; err != nil {
		response.FailWithMsg("被@的用户不存在", c)
		return
	}
	if err = message_service.InsertAtMessage(actor, req.ID); err != nil {
		response.FailWithMsg("发送@消息失败", c)
		return
	}
	response.OkWithMsg("@消息已发送", c)
}
