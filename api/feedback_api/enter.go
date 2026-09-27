// api/feedback_api/enter.go
// 用户反馈墙:提交(登录可选/可匿名)、全站公开列表、管理员处理(增量更新)
package feedback_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/community_service"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	jwts "StarDreamerCyberNook/utils/jwts"
	"strings"

	"github.com/gin-gonic/gin"
)

// FeedbackApi 反馈接口
type FeedbackApi struct{}

type FeedbackCreateRequest struct {
	Content     string              `json:"content" binding:"required"` // 反馈内容
	Contact     string              `json:"contact"`                    // 联系方式,可选
	Type        models.FeedbackType `json:"type"`                       // 反馈类型,默认0
	IsAnonymous bool                `json:"isAnonymous"`                // 是否匿名提交
}

// FeedbackCreateView 提交反馈,登录可选(未登录记为 UserID=0);匿名提交仅隐藏展示,不改动落库内容
func (FeedbackApi) FeedbackCreateView(c *gin.Context) {
	var req FeedbackCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	content := strings.TrimSpace(xss_filter.SanitizeText(req.Content))
	if content == "" {
		response.FailWithMsg("反馈内容不能为空", c)
		return
	}
	if len([]rune(content)) > 2000 {
		response.FailWithMsg("反馈内容过长", c)
		return
	}
	if req.Type < models.FeedbackTypeOther || req.Type > models.FeedbackTypeReport {
		req.Type = models.FeedbackTypeOther
	}

	//登录可选:已登录才记录提交者ID,未登录记为0
	var userID uint
	if claims, err := jwts.ParseTokenByGin(c); err == nil && claims != nil {
		userID = claims.UserID
	}

	if err := community_service.CreateFeedback(userID, req.IsAnonymous, content, strings.TrimSpace(req.Contact), req.Type); err != nil {
		response.FailWithMsg("提交失败", c)
		return
	}
	response.OkWithMsg("感谢反馈,我们会尽快处理", c)
}

type FeedbackListRequest struct {
	common.PageInfo
	Status *models.FeedbackStatus `form:"status"` // 处理状态筛选,可选
	Type   *models.FeedbackType   `form:"type"`   // 类型筛选,可选
}

// FeedbackWallView 反馈墙列表(全站公开):服务端已做隐私过滤(联系方式/处理人/匿名者ID)
func (FeedbackApi) FeedbackWallView(c *gin.Context) {
	var req FeedbackListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	list, count, capped, err := community_service.ListFeedbacks(
		req.Status, req.Type, req.Page, req.Limit, req.Order, req.EndId)
	if err != nil {
		response.FailWithMsg("查询失败", c)
		return
	}

	//防御性再过滤一次(与服务端一致),确保不外泄隐私字段
	for i := range list {
		list[i].Contact = ""
		list[i].HandlerID = 0
		if list[i].IsAnonymous {
			list[i].UserID = 0
		}
	}
	response.OkWithListCapped(list, count, capped, c)
}

type FeedbackHandleRequest struct {
	Status models.FeedbackStatus `json:"status"` // 目标状态:0待处理 1已采纳未处理 2正在处理 3已处理
	Reply  string                `json:"reply"`  // 管理员回复
}

// FeedbackHandleView 管理员处理反馈:更新状态与回复,并返回更新后的条目供前端增量替换
func (FeedbackApi) FeedbackHandleView(c *gin.Context) {
	var uri models.IDRequest
	if err := c.ShouldBindUri(&uri); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var req FeedbackHandleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	//状态必须是合法枚举
	if req.Status < models.FeedbackStatusPending || req.Status > models.FeedbackStatusResolved {
		response.FailWithMsg("非法的状态", c)
		return
	}

	claims := jwts.GetClaims(c)
	reply := strings.TrimSpace(xss_filter.SanitizeText(req.Reply))
	feedback, err := community_service.HandleFeedback(uri.ID, req.Status, reply, claims.UserID)
	if err != nil {
		response.FailWithMsg("处理失败", c)
		return
	}

	//与列表一致的隐私过滤后返回,前端据此在列表中增量替换
	feedback.Contact = ""
	feedback.HandlerID = 0
	if feedback.IsAnonymous {
		feedback.UserID = 0
	}
	response.OkWithData(feedback, c)
}
