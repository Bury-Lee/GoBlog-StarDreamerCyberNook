package article_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/message_service"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (ArticleApi) ArticleDiggView(c *gin.Context) {
	var IDRequest models.IDRequest
	if err := c.ShouldBindUri(&IDRequest); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	var article models.ArticleModel
	err := global.DB.Take(&article, "status = ? and id = ?", models.StatusPublished, IDRequest.ID).Error
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}

	claims := jwts.GetClaims(c)

	// 查一下之前有没有点过
	var userDiggArticle models.ArticleDiggModel
	err = global.DB.Take(&userDiggArticle, "user_id = ? and article_id = ?", claims.UserID, article.ID).Error
	if err != nil {
		// 点赞
		digg := models.ArticleDiggModel{
			UserID:    claims.UserID,
			ArticleID: IDRequest.ID,
		}
		if err = global.DB.Create(&digg).Error; err != nil {
			response.FailWithMsg("点赞失败", c)
			return
		}
		redis_count.SetCacheDigg(IDRequest.ID, true)
		response.OkWithMsg("点赞成功", c)
		// 发送点赞消息,失败只记录日志,不影响主流程
		if err = message_service.InsertArticleDiggMessage(digg); err != nil {
			logrus.Errorf("发送点赞消息失败: %v", err)
		}
		return
	}
	// 取消点赞
	redis_count.SetCacheDigg(IDRequest.ID, false)
	global.DB.Delete(&models.ArticleDiggModel{}, "user_id = ? and article_id = ?", claims.UserID, article.ID) //没必要进行错误处理
	response.OkWithMsg("取消点赞成功", c)
}
