package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/ai_service"
	"StarDreamerCyberNook/service/message_service"
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ArticleReviewListViewRequest struct {
	common.PageInfo
	UserID uint `form:"userID"` //可以指定选择谁的文章
}

func (ArticleApi) ArticleReviewListView(c *gin.Context) {
	var req ArticleReviewListViewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	option := common.Options{
		PageInfo:      req.PageInfo,
		Likes:         []string{"title"},
		Preloads:      []string{"UserModel"},
		Where:         global.DB.Where("status = ?", models.StatusPending), //待审核=用户提交后等待审核的文章
		AllowedOrders: []string{"id", "created_at", "status"},
	}
	if req.UserID != 0 {
		option.Where = option.Where.Where("user_id = ?", req.UserID)
	}
	list, count, err := common.ListQuery(models.ArticleModel{}, option)
	if err != nil {
		response.FailWithMsg("查询失败", c)
		return
	}
	response.OkWithList(list, count, c)
}

type ArticleReviewRequest struct {
	ArticleID uint          `json:"articleID" binding:"required"`
	Status    models.Status `json:"status" binding:"required"` //审核状态,2为通过,1,3为不通过
	Msg       string        `json:"msg"`                       // 为4的时候，传递进来
}

func (ArticleApi) ArticleReviewView(c *gin.Context) {
	var req ArticleReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	var article models.ArticleModel
	err := global.DB.Take(&article, req.ArticleID).Error
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}

	if err = applyArticleReview(&article, req.Status, req.Msg); err != nil {
		response.FailWithMsg("审核失败:"+err.Error(), c)
		return
	}
	response.OkWithMsg("审核成功", c)
}

// applyArticleReview 更新文章审核状态、清理详情缓存并给作者发送系统通知
// 说明:人工审核与AI审核共用,保证两条链路的缓存清理与通知行为一致
func applyArticleReview(article *models.ArticleModel, status models.Status, msg string) error {
	if err := global.DB.Model(article).Update("status", status).Error; err != nil {
		return err
	}
	article.Status = status

	//状态变更后清理详情缓存,避免已下线文章继续可见或新通过的文章一直不可见
	global.RedisHotPool.Del(context.Background(), "ArticleID"+strconv.FormatUint(uint64(article.ID), 10))

	var content string
	switch status {
	case models.StatusPublished:
		content = fmt.Sprintf("文章通过审核:%s\n备注:%s", article.Title, msg)
	case models.StatusDraft:
		content = fmt.Sprintf("文章不通过审核:%s\n备注:%s", article.Title, msg)
	case models.StatusOffline:
		content = fmt.Sprintf("文章已下线:%s\n备注:%s", article.Title, msg)
	case models.StatusPending:
		content = fmt.Sprintf("文章进入待审核状态:%s\n备注:%s", article.Title, msg)
	}

	message := models.MessageModel{
		RevUserID:          article.UserID,
		ActionUserID:       0,
		ActionUserNickname: "系统",
		ActionUserAvatar:   "", //这里可以改用站内默认头像
		Title:              "文章审核通知",
		ArticleID:          article.ID,
		ArticleTitle:       article.Title,
		Content:            content,
	}
	if err := message_service.InsertSystemMessage(message); err != nil {
		logrus.Error("发送系统消息失败", err)
	}
	return nil
}

// ArticleAIReviewRequest AI审核请求
type ArticleAIReviewRequest struct {
	ArticleID uint   `json:"articleID"` //单个文章ID,可选
	IDList    []uint `json:"IDList"`    //批量文章ID,可选;与ArticleID都为空时审核全部待审核文章
	Limit     int    `json:"limit"`     //批量审核上限,默认10,最大20
}

// ArticleAIReviewItem AI审核结果
type ArticleAIReviewItem struct {
	ArticleID uint          `json:"articleID"`
	Title     string        `json:"title"`
	AIResult  string        `json:"aiResult"` //AI的原始判定:通过/拒绝/其它
	Status    models.Status `json:"status"`   //审核后的文章状态
	Error     string        `json:"error,omitempty"`
}

// ArticleAIReviewView 对处于"审核中"状态的文章执行AI审核
// 审核规则:通过->已发布, 拒绝->草稿(退回作者), 无法判定->保持审核中等待人工处理
func (ArticleApi) ArticleAIReviewView(c *gin.Context) {
	if !global.Config.AI.Enable {
		response.FailWithMsg("AI功能未启用", c)
		return
	}
	var req ArticleAIReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if req.Limit <= 0 || req.Limit > 20 {
		req.Limit = 10
	}
	if req.ArticleID != 0 {
		req.IDList = append(req.IDList, req.ArticleID)
	}

	//只处理审核中的文章
	query := global.DB.Where("status = ?", models.StatusPending)
	if len(req.IDList) > 0 {
		query = query.Where("id in ?", req.IDList)
	}
	var articles []models.ArticleModel
	if err := query.Order("id asc").Limit(req.Limit).Find(&articles).Error; err != nil {
		response.FailWithMsg("查询待审核文章失败", c)
		return
	}
	if len(articles) == 0 {
		response.OkWithMsg("没有需要AI审核的文章", c)
		return
	}

	list := make([]ArticleAIReviewItem, 0, len(articles))
	successCount := 0
	for i := range articles {
		article := &articles[i]
		item := ArticleAIReviewItem{ArticleID: article.ID, Title: article.Title, Status: article.Status}

		reply, err := ai_service.ReviewArticle(article.Title, article.Abstract, article.Content)
		if err != nil {
			logrus.Errorf("AI审核文章 %d 失败: %v", article.ID, err)
			item.Error = err.Error()
			list = append(list, item)
			continue
		}
		item.AIResult = reply
		switch reply {
		case ai_service.AIReviewPass:
			item.Status = models.StatusPublished
		case ai_service.AIReviewReject:
			item.Status = models.StatusDraft
		default:
			//AI无法判定,保持待审核状态,交给人工处理
			item.Status = models.StatusPending
		}

		if err = applyArticleReview(article, item.Status, "AI审核结果:"+reply); err != nil {
			logrus.Errorf("AI审核文章 %d 状态更新失败: %v", article.ID, err)
			item.Error = err.Error()
			list = append(list, item)
			continue
		}
		successCount++
		list = append(list, item)
	}

	response.OkWithData(gin.H{"list": list, "count": successCount, "total": len(articles)}, c)
}
