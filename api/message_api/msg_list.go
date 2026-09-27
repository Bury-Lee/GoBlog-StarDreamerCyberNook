package message_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/message_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

type SiteMessageListViewRequest struct {
	common.PageInfo
	Type models.MessageType `form:"type" binding:"required"`
}

// SiteMessageListView 消息列表(读取后由服务端标记已读)
func (MessageApi) SiteMessageListView(c *gin.Context) {
	var req SiteMessageListViewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claim := jwts.GetClaims(c)

	items, count, capped, err := message_service.List(claim.UserID, req.Type, req.Page, req.Limit, req.Order, req.EndId)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("查询消息失败", c)
		}
		return
	}

	list := make([]models.MessageModel, 0, len(items))
	for _, it := range items {
		list = append(list, models.MessageModel{
			Model:              models.Model{ID: it.ID, CreatedAt: it.CreatedAt, UpdatedAt: it.UpdatedAt},
			Type:               it.Type,
			RevUserID:          it.RevUserID,
			ActionUserID:       it.ActionUserID,
			ActionUserNickname: it.ActionUserNickname,
			ActionUserAvatar:   it.ActionUserAvatar,
			Title:              it.Title,
			Content:            it.Content,
			ArticleID:          it.ArticleID,
			ArticleTitle:       it.ArticleTitle,
			CommentID:          it.CommentID,
			LinkTitle:          it.LinkTitle,
			LinkHref:           it.LinkHref,
			IsRead:             it.IsRead,
		})
	}
	response.OkWithListCapped(list, count, capped, c)
}
