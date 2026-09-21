package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/review_service"

	"github.com/gin-gonic/gin"
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

	if err = review_service.ApplyArticleReview(&article, req.Status, req.Msg); err != nil {
		response.FailWithMsg("审核失败:"+err.Error(), c)
		return
	}
	response.OkWithMsg("审核成功", c)
}

// ArticleAIReviewRequest AI审核请求
type ArticleAIReviewRequest struct {
	ArticleID uint   `json:"articleID"` //单个文章ID,可选
	IDList    []uint `json:"IDList"`    //批量文章ID,可选;与ArticleID都为空时审核全部待审核文章
	Limit     int    `json:"limit"`     //批量审核上限,默认10,最大20
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

	list, err := review_service.ReviewPendingArticles(req.IDList, req.Limit)
	if err != nil {
		response.FailWithMsg("查询待审核文章失败", c)
		return
	}
	if len(list) == 0 {
		response.OkWithMsg("没有需要AI审核的文章", c)
		return
	}

	successCount := 0
	for _, item := range list {
		if item.Error == "" {
			successCount++
		}
	}

	response.OkWithData(gin.H{"list": list, "count": successCount, "total": len(list)}, c)
}
