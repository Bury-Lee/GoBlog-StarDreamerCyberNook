package article_api

import (
	"fmt"
	"strconv"

	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/content_service"
	"StarDreamerCyberNook/service/message_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

// ArticleRemoveUserView 用户删除自己的文章(经 content 服务校验归属)。
func (ArticleApi) ArticleRemoveUserView(c *gin.Context) {
	var req models.RemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)

	if len(req.IDList) == 0 {
		response.FailWithMsg("ID列表不能为空", c)
		return
	}

	owner := claims.UserID
	if _, err := content_service.RemoveArticles(req.IDList, &owner); err != nil {
		response.FailWithMsg("部分文章不存在或不属于当前用户", c)
		return
	}

	response.OkWithMsg(fmt.Sprintf("成功删除 %d 篇文章", len(req.IDList)), c)
}

// ArticleRemoveView 管理员删除文章(经 content 服务)。
func (ArticleApi) ArticleRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	res, err := content_service.RemoveArticles(req.IDList, nil)
	if err != nil {
		response.FailWithMsg("删除失败", c)
		return
	}

	// 清理详情缓存
	delList := make([]string, 0, len(req.IDList))
	for _, id := range req.IDList {
		delList = append(delList, "ArticleID"+strconv.FormatUint(uint64(id), 10))
	}
	global.RedisHotPool.Del(c, delList...)

	titleList := []byte{}
	for _, title := range res.Titles {
		titleList = append(titleList, []byte(title+" 、")...)
		titleList = append(titleList, []byte("\n")...)
	}
	message_service.InsertSystemMessage(models.MessageModel{
		RevUserID:          0,
		ActionUserID:       0,
		ActionUserNickname: "系统",
		Title:              "文章删除通知",
		Content:            fmt.Sprintf("您的文章 %s等已被管理员删除", string(titleList)),
	})
	response.OkWithMsg(fmt.Sprintf("删除成功 成功删除%d条", res.Deleted), c)
}
