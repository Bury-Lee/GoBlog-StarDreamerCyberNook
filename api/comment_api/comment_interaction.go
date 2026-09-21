package comment_api

import (
	"strconv"
	"strings"

	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

// CommentInteractionRequest 批量查询评论点赞状态的请求参数
type CommentInteractionRequest struct {
	CommentIDs string `form:"commentIDs" binding:"required"` // 逗号分隔的评论ID,如 "1,2,3"
}

// CommentInteractionResponse 当前登录用户点赞过的评论ID集合
type CommentInteractionResponse struct {
	DiggedIDs []uint `json:"diggedIDs"`
}

// CommentInteractionView 批量查询当前登录用户对一组评论的点赞状态
// 说明:评论列表是公共数据,点赞状态和访问者相关,单独开接口按需查询
func (CommentApi) CommentInteractionView(c *gin.Context) {
	var req CommentInteractionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		response.FailWithMsg("请登录", c)
		return
	}

	parts := strings.Split(req.CommentIDs, ",")
	ids := make([]uint, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseUint(part, 10, 64)
		if err != nil || id == 0 {
			continue
		}
		ids = append(ids, uint(id))
		if len(ids) >= 100 { // 单次最多查询100条,避免被刷
			break
		}
	}

	diggedIDs := make([]uint, 0)
	if len(ids) > 0 {
		if err := global.DB.Model(&models.CommentDiggModel{}).
			Where("user_id = ? and comment_id in ?", claims.UserID, ids).
			Pluck("comment_id", &diggedIDs).Error; err != nil {
			response.FailWithMsg("查询评论点赞状态失败", c)
			return
		}
	}

	response.OkWithData(CommentInteractionResponse{DiggedIDs: diggedIDs}, c)
}
