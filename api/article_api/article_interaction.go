package article_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/content_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

// ArticleInteractionResponse 当前登录用户对某篇文章的互动状态
type ArticleInteractionResponse struct {
	Digged    bool `json:"digged"`    // 是否已点赞
	Collected bool `json:"collected"` // 是否已收藏
}

// ArticleInteractionView 查询当前登录用户对指定文章的点赞/收藏状态
// 说明:这类"和访问者相关"的状态单独开接口,不写进公共的文章详情缓存
func (ArticleApi) ArticleInteractionView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		response.FailWithMsg("请登录", c)
		return
	}

	digged, collected, err := content_service.GetArticleInteraction(claims.UserID, req.ID)
	if err != nil {
		response.FailWithMsg("查询互动状态失败", c)
		return
	}

	response.OkWithData(ArticleInteractionResponse{
		Digged:    digged,
		Collected: collected,
	}, c)
}
