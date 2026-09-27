package article_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/content_service"
	"StarDreamerCyberNook/service/message_service"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	jwts "StarDreamerCyberNook/utils/jwts"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (ArticleApi) ArticleDiggView(c *gin.Context) {
	var IDRequest models.IDRequest
	if err := c.ShouldBindUri(&IDRequest); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)

	// 点赞/取消点赞的落库下沉到 content 服务;缓存计数与消息留在网关
	digged, changed, err := content_service.ToggleArticleDigg(claims.UserID, IDRequest.ID)
	if err != nil {
		if errors.Is(err, content_service.ErrNotFound) {
			response.FailWithMsg("文章不存在", c)
			return
		}
		response.FailWithMsg("操作失败", c)
		return
	}
	if changed {
		redis_count.SetCacheDigg(IDRequest.ID, digged)
	}
	if digged {
		response.OkWithMsg("点赞成功", c)
		// 发送点赞消息,失败只记录日志,不影响主流程
		if err = message_service.InsertArticleDiggMessage(models.ArticleDiggModel{
			UserID:    claims.UserID,
			ArticleID: IDRequest.ID,
		}); err != nil {
			logrus.Errorf("发送点赞消息失败: %v", err)
		}
		return
	}
	response.OkWithMsg("取消点赞成功", c)
}
